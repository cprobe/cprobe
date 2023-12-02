package exporter

import (
	"fmt"
	"regexp"
	"strconv"
	"sync"

	"github.com/Shopify/sarama"
	"github.com/cprobe/cprobe/lib/logger"
	"github.com/krallistic/kazoo-go"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rcrowley/go-metrics"
)

const (
	clientID = "kafka_exporter"
)

func init() {
	// https://l1905.github.io/golang/2020/04/30/golang-kafka-sarama/
	metrics.UseNilMetrics = true
}

// Exporter collects Kafka stats from the given server and exports them using
// the prometheus metrics package.
type Exporter struct {
	client                sarama.Client
	topicFilter           *regexp.Regexp
	topicExclude          *regexp.Regexp
	topicExcludeString    string
	groupFilter           *regexp.Regexp
	groupExclude          *regexp.Regexp
	groupExcludeString    string
	mu                    sync.Mutex
	useZooKeeperLag       bool
	zookeeperClient       *kazoo.Kazoo
	offsetShowAll         bool
	topicWorkers          int
	consumerGroupFetchAll bool
}

type KafkaOpts struct {
	Namespace                string
	Uri                      []string
	UseSASL                  bool
	UseSASLHandshake         bool
	SaslUsername             string
	SaslPassword             string
	SaslMechanism            string
	SaslDisablePAFXFast      bool
	UseTLS                   bool
	TlsServerName            string
	TlsCAFile                string
	TlsCertFile              string
	TlsKeyFile               string
	ServerUseTLS             bool
	ServerMutualAuthEnabled  bool
	ServerTlsCAFile          string
	ServerTlsCertFile        string
	ServerTlsKeyFile         string
	TlsInsecureSkipTLSVerify bool
	KafkaVersion             string
	UseZooKeeperLag          bool
	UriZookeeper             []string
	Labels                   string
	ServiceName              string
	KerberosConfigPath       string
	Realm                    string
	KeyTabPath               string
	KerberosAuthType         string
	OffsetShowAll            bool
	TopicWorkers             int
}

// NewExporter returns an initialized Exporter.
func NewExporter(opts KafkaOpts, topicFilter string, topicExclude string, groupFilter string, groupExclude string) (*Exporter, error) {
	initDesc(opts.Namespace)

	config := sarama.NewConfig()
	config.ClientID = clientID
	kafkaVersion, err := sarama.ParseKafkaVersion(opts.KafkaVersion)
	if err != nil {
		return nil, err
	}
	config.Version = kafkaVersion

	if err := fillSaslFields(config, opts); err != nil {
		return nil, err
	}

	if err := fillTlsFields(config, opts); err != nil {
		return nil, err
	}

	config.Metadata.AllowAutoTopicCreation = false
	client, err := sarama.NewClient(opts.Uri, config)
	if err != nil {
		return nil, errors.Wrap(err, "cannot init kafka client")
	}

	var zookeeperClient *kazoo.Kazoo
	if opts.UseZooKeeperLag {
		zookeeperClient, err = kazoo.NewKazoo(opts.UriZookeeper, nil)
		if err != nil {
			return nil, errors.Wrap(err, "cannot connect to zookeeper")
		}
	}

	exp := &Exporter{
		client:                client,
		topicExcludeString:    topicExclude,
		groupExcludeString:    groupExclude,
		topicFilter:           regexp.MustCompile(topicFilter),
		groupFilter:           regexp.MustCompile(groupFilter),
		useZooKeeperLag:       opts.UseZooKeeperLag,
		zookeeperClient:       zookeeperClient,
		offsetShowAll:         opts.OffsetShowAll,
		topicWorkers:          opts.TopicWorkers,
		consumerGroupFetchAll: config.Version.IsAtLeast(sarama.V2_0_0_0),
	}

	if topicExclude != "" {
		exp.topicExclude = regexp.MustCompile(topicExclude)
	}

	if groupExclude != "" {
		exp.groupExclude = regexp.MustCompile(groupExclude)
	}

	return exp, nil
}

func (e *Exporter) CloseClient() {
	if e.client != nil {
		e.client.Close()
	}
}

// Describe describes all the metrics ever exported by the Kafka exporter. It
// implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- clusterBrokers
	ch <- topicCurrentOffset
	ch <- topicOldestOffset
	ch <- topicPartitions
	ch <- topicPartitionLeader
	ch <- topicPartitionReplicas
	ch <- topicPartitionInSyncReplicas
	ch <- topicPartitionUsesPreferredReplica
	ch <- topicUnderReplicatedPartition
	ch <- consumergroupCurrentOffset
	ch <- consumergroupCurrentOffsetSum
	ch <- consumergroupLag
	ch <- consumergroupLagZookeeper
	ch <- consumergroupLagSum
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(
		clusterBrokers, prometheus.GaugeValue, float64(len(e.client.Brokers())),
	)

	for _, b := range e.client.Brokers() {
		ch <- prometheus.MustNewConstMetric(
			clusterBrokerInfo, prometheus.GaugeValue, 1, strconv.Itoa(int(b.ID())), b.Addr(),
		)
	}

	topics, err := e.client.Topics()
	if err != nil {
		logger.Errorf("cannot get topics: %v", err)
		return
	}

	// topic -> partition -> offset
	offset := make(map[string]map[int32]int64)

	// 创建一个闭包函数，用于获取某个 topic 的监控指标
	getTopicMetrics := func(topic string) {
		partitions, err := e.client.Partitions(topic)
		if err != nil {
			logger.Errorf("cannot get partitions of topic %s: %v", topic, err)
			return
		}

		ch <- prometheus.MustNewConstMetric(
			topicPartitions, prometheus.GaugeValue, float64(len(partitions)), topic,
		)

		partitionOffsets := make(map[int32]int64, len(partitions))

		for _, partition := range partitions {
			broker, err := e.client.Leader(topic, partition)
			if err != nil {
				logger.Errorf("cannot get leader of topic %s partition %d: %v", topic, partition, err)
			} else {
				ch <- prometheus.MustNewConstMetric(
					topicPartitionLeader, prometheus.GaugeValue, float64(broker.ID()), topic, fmt.Sprint(partition),
				)
			}

			currentOffset, err := e.client.GetOffset(topic, partition, sarama.OffsetNewest)
			if err != nil {
				logger.Errorf("cannot get current offset of topic %s partition %d: %v", topic, partition, err)
			} else {
				partitionOffsets[partition] = currentOffset

				ch <- prometheus.MustNewConstMetric(
					topicCurrentOffset, prometheus.GaugeValue, float64(currentOffset), topic, fmt.Sprint(partition),
				)

				if e.useZooKeeperLag {
					ConsumerGroups, err := e.zookeeperClient.Consumergroups()
					if err != nil {
						logger.Errorf("cannot get consumer group %v", err)
					} else {
						for _, group := range ConsumerGroups {
							groupOffset, _ := group.FetchOffset(topic, partition)
							if groupOffset > 0 {
								consumerGroupLag := currentOffset - groupOffset
								ch <- prometheus.MustNewConstMetric(
									consumergroupLagZookeeper, prometheus.GaugeValue, float64(consumerGroupLag), group.Name, topic, fmt.Sprint(partition),
								)
							}
						}
					}
				}
			}

			oldestOffset, err := e.client.GetOffset(topic, partition, sarama.OffsetOldest)
			if err != nil {
				logger.Errorf("cannot get oldest offset of topic %s partition %d: %v", topic, partition, err)
			} else {
				ch <- prometheus.MustNewConstMetric(
					topicOldestOffset, prometheus.GaugeValue, float64(oldestOffset), topic, fmt.Sprint(partition),
				)
			}

			replicas, err := e.client.Replicas(topic, partition)
			if err != nil {
				logger.Errorf("cannot get replicas of topic %s partition %d: %v", topic, partition, err)
			} else {
				ch <- prometheus.MustNewConstMetric(
					topicPartitionReplicas, prometheus.GaugeValue, float64(len(replicas)), topic, fmt.Sprint(partition),
				)
			}

			inSyncReplicas, err := e.client.InSyncReplicas(topic, partition)
			if err != nil {
				logger.Errorf("cannot get in-sync replicas of topic %s partition %d: %v", topic, partition, err)
			} else {
				ch <- prometheus.MustNewConstMetric(
					topicPartitionInSyncReplicas, prometheus.GaugeValue, float64(len(inSyncReplicas)), topic, fmt.Sprint(partition),
				)
			}

			if broker != nil && replicas != nil && len(replicas) > 0 && broker.ID() == replicas[0] {
				ch <- prometheus.MustNewConstMetric(
					topicPartitionUsesPreferredReplica, prometheus.GaugeValue, float64(1), topic, fmt.Sprint(partition),
				)
			} else {
				ch <- prometheus.MustNewConstMetric(
					topicPartitionUsesPreferredReplica, prometheus.GaugeValue, float64(0), topic, fmt.Sprint(partition),
				)
			}

			if replicas != nil && inSyncReplicas != nil && len(inSyncReplicas) < len(replicas) {
				ch <- prometheus.MustNewConstMetric(
					topicUnderReplicatedPartition, prometheus.GaugeValue, float64(1), topic, fmt.Sprint(partition),
				)
			} else {
				ch <- prometheus.MustNewConstMetric(
					topicUnderReplicatedPartition, prometheus.GaugeValue, float64(0), topic, fmt.Sprint(partition),
				)
			}

		}

		e.mu.Lock()
		offset[topic] = partitionOffsets
		e.mu.Unlock()
	}

	// 控制一下并发
	semaphone := make(chan struct{}, e.topicWorkers)
	topicWG := sync.WaitGroup{}

	// 对于匹配的 topic 挨个采集指标
	for _, topic := range topics {
		if e.topicFilter.MatchString(topic) && (e.topicExcludeString == "" || !e.topicExclude.MatchString(topic)) {
			semaphone <- struct{}{}
			topicWG.Add(1)
			go func(topic string) {
				defer func() {
					<-semaphone
					topicWG.Done()
				}()
				getTopicMetrics(topic)
			}(topic)
		}
	}

	topicWG.Wait()
	close(semaphone)

	getConsumerGroupMetrics := func(broker *sarama.Broker) {
		if err := broker.Open(e.client.Config()); err != nil && err != sarama.ErrAlreadyConnected {
			logger.Errorf("broker(%v:%v) cannot connect: %v", broker.ID(), broker.Addr(), err)
			return
		}

		defer broker.Close()

		groups, err := broker.ListGroups(&sarama.ListGroupsRequest{})
		if err != nil {
			logger.Errorf("broker(%v:%v) cannot get consumer group: %v", broker.ID(), broker.Addr(), err)
			return
		}

		var groupIds []string
		for groupId := range groups.Groups {
			if e.groupFilter.MatchString(groupId) && (e.groupExcludeString == "" || !e.groupExclude.MatchString(groupId)) {
				groupIds = append(groupIds, groupId)
			}
		}

		describeGroups, err := broker.DescribeGroups(&sarama.DescribeGroupsRequest{Groups: groupIds})
		if err != nil {
			logger.Errorf("broker(%v:%v) cannot get describe groups: %v", broker.ID(), broker.Addr(), err)
			return
		}

		for _, group := range describeGroups.Groups {
			offsetFetchRequest := sarama.OffsetFetchRequest{ConsumerGroup: group.GroupId, Version: 1}
			if e.offsetShowAll {
				for topic, partitions := range offset {
					for partition := range partitions {
						offsetFetchRequest.AddPartition(topic, partition)
					}
				}
			} else {
				for _, member := range group.Members {
					assignment, err := member.GetMemberAssignment()
					if err != nil {
						logger.Errorf("broker(%v:%v) cannot get GetMemberAssignment of group member %v : %v", broker.ID(), broker.Addr(), member, err)
						return
					}

					for topic, partions := range assignment.Topics {
						for _, partition := range partions {
							offsetFetchRequest.AddPartition(topic, partition)
						}
					}
				}
			}

			ch <- prometheus.MustNewConstMetric(
				consumergroupMembers, prometheus.GaugeValue, float64(len(group.Members)), group.GroupId,
			)

			offsetFetchResponse, err := broker.FetchOffset(&offsetFetchRequest)
			if err != nil {
				logger.Errorf("broker(%v:%v) cannot get offset of group %s: %v", broker.ID(), broker.Addr(), group.GroupId, err)
				continue
			}

			for topic, partitions := range offsetFetchResponse.Blocks {
				// If the topic is not consumed by that consumer group, skip it
				topicConsumed := false
				for _, offsetFetchResponseBlock := range partitions {
					// Kafka will return -1 if there is no offset associated with a topic-partition under that consumer group
					if offsetFetchResponseBlock.Offset != -1 {
						topicConsumed = true
						break
					}
				}
				if !topicConsumed {
					continue
				}

				var currentOffsetSum int64
				var lagSum int64
				for partition, offsetFetchResponseBlock := range partitions {
					err := offsetFetchResponseBlock.Err
					if err != sarama.ErrNoError {
						logger.Errorf("broker(%v:%v) error for partition %d : %v", broker.ID(), broker.Addr(), partition, err.Error())
						continue
					}

					currentOffset := offsetFetchResponseBlock.Offset
					currentOffsetSum += currentOffset
					ch <- prometheus.MustNewConstMetric(
						consumergroupCurrentOffset, prometheus.GaugeValue, float64(currentOffset), group.GroupId, topic, strconv.FormatInt(int64(partition), 10),
					)

					// e.mu.Lock()
					if currentOffset, ok := offset[topic][partition]; ok {
						// If the topic is consumed by that consumer group, but no offset associated with the partition
						// forcing lag to -1 to be able to alert on that
						var lag int64
						if offsetFetchResponseBlock.Offset == -1 {
							lag = -1
						} else {
							lag = currentOffset - offsetFetchResponseBlock.Offset
							lagSum += lag
						}
						ch <- prometheus.MustNewConstMetric(
							consumergroupLag, prometheus.GaugeValue, float64(lag), group.GroupId, topic, strconv.FormatInt(int64(partition), 10),
						)
					} else {
						logger.Errorf("broker(%v:%v) no offset of topic %s partition %d, cannot get consumer group lag", broker.ID(), broker.Addr(), topic, partition)
					}
					// e.mu.Unlock()
				}

				ch <- prometheus.MustNewConstMetric(
					consumergroupCurrentOffsetSum, prometheus.GaugeValue, float64(currentOffsetSum), group.GroupId, topic,
				)

				ch <- prometheus.MustNewConstMetric(
					consumergroupLagSum, prometheus.GaugeValue, float64(lagSum), group.GroupId, topic,
				)
			}
		}
	}

	brokers := e.client.Brokers()
	if len(brokers) == 0 {
		return
	}

	if len(brokers) == 1 {
		getConsumerGroupMetrics(brokers[0])
		return
	}

	brokerWG := sync.WaitGroup{}
	for i := range brokers {
		brokerWG.Add(1)
		go func(broker *sarama.Broker) {
			defer brokerWG.Done()
			getConsumerGroupMetrics(broker)
		}(brokers[i])
	}
	brokerWG.Wait()
}
