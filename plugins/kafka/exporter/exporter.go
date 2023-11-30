package exporter

import (
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

// Exporter collects Kafka stats from the given server and exports them using
// the prometheus metrics package.
type Exporter struct {
	client                sarama.Client
	topicFilter           *regexp.Regexp
	topicExclude          *regexp.Regexp
	groupFilter           *regexp.Regexp
	groupExclude          *regexp.Regexp
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
	var zookeeperClient *kazoo.Kazoo
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

	if opts.UseZooKeeperLag {
		logger.Infof("Using zookeeper lag, so connecting to zookeeper")
		zookeeperClient, err = kazoo.NewKazoo(opts.UriZookeeper, nil)
		if err != nil {
			return nil, errors.Wrap(err, "error connecting to zookeeper")
		}
	}

	config.Metadata.AllowAutoTopicCreation = false

	client, err := sarama.NewClient(opts.Uri, config)

	if err != nil {
		return nil, errors.Wrap(err, "Error Init Kafka Client")
	}

	logger.Infof("Done Init Clients")
	// Init our exporter.
	return &Exporter{
		client:                client,
		topicFilter:           regexp.MustCompile(topicFilter),
		topicExclude:          regexp.MustCompile(topicExclude),
		groupFilter:           regexp.MustCompile(groupFilter),
		groupExclude:          regexp.MustCompile(groupExclude),
		useZooKeeperLag:       opts.UseZooKeeperLag,
		zookeeperClient:       zookeeperClient,
		offsetShowAll:         opts.OffsetShowAll,
		topicWorkers:          opts.TopicWorkers,
		consumerGroupFetchAll: config.Version.IsAtLeast(sarama.V2_0_0_0),
	}, nil
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
	var wg = sync.WaitGroup{}
	ch <- prometheus.MustNewConstMetric(
		clusterBrokers, prometheus.GaugeValue, float64(len(e.client.Brokers())),
	)
	for _, b := range e.client.Brokers() {
		ch <- prometheus.MustNewConstMetric(
			clusterBrokerInfo, prometheus.GaugeValue, 1, strconv.Itoa(int(b.ID())), b.Addr(),
		)
	}

	offset := make(map[string]map[int32]int64)

	topics, err := e.client.Topics()
	//logger.Infof("kafka topics: ", topics)
	if err != nil {
		logger.Errorf("Cannot get topics: %v", err)
		return
	}

	topicChannel := make(chan string)

	getTopicMetrics := func(topic string) {
		defer wg.Done()
		//logger.Infof("开始获取topic指标", topic)

		// 如果该topic不在filter里面并且在exclude里面,不采集指标
		if !e.topicFilter.MatchString(topic) || e.topicExclude.MatchString(topic) {
			return
		}

		partitions, err := e.client.Partitions(topic)
		//logger.Infof("the partition of topics", partitions, topics)
		if err != nil {
			logger.Errorf("Cannot get partitions of topic %s: %v", topic, err)
			return
		}
		ch <- prometheus.MustNewConstMetric(
			topicPartitions, prometheus.GaugeValue, float64(len(partitions)), topic,
		)
		e.mu.Lock()
		offset[topic] = make(map[int32]int64, len(partitions))
		e.mu.Unlock()
		for _, partition := range partitions {
			broker, err := e.client.Leader(topic, partition)
			if err != nil {
				logger.Errorf("Cannot get leader of topic %s partition %d: %v", topic, partition, err)
			} else {
				ch <- prometheus.MustNewConstMetric(
					topicPartitionLeader, prometheus.GaugeValue, float64(broker.ID()), topic, strconv.FormatInt(int64(partition), 10),
				)
			}

			currentOffset, err := e.client.GetOffset(topic, partition, sarama.OffsetNewest)
			if err != nil {
				logger.Errorf("Cannot get current offset of topic %s partition %d: %v", topic, partition, err)
			} else {
				e.mu.Lock()
				offset[topic][partition] = currentOffset
				e.mu.Unlock()
				ch <- prometheus.MustNewConstMetric(
					topicCurrentOffset, prometheus.GaugeValue, float64(currentOffset), topic, strconv.FormatInt(int64(partition), 10),
				)
			}

			oldestOffset, err := e.client.GetOffset(topic, partition, sarama.OffsetOldest)
			if err != nil {
				logger.Errorf("Cannot get oldest offset of topic %s partition %d: %v", topic, partition, err)
			} else {
				ch <- prometheus.MustNewConstMetric(
					topicOldestOffset, prometheus.GaugeValue, float64(oldestOffset), topic, strconv.FormatInt(int64(partition), 10),
				)
			}

			replicas, err := e.client.Replicas(topic, partition)
			if err != nil {
				logger.Errorf("Cannot get replicas of topic %s partition %d: %v", topic, partition, err)
			} else {
				ch <- prometheus.MustNewConstMetric(
					topicPartitionReplicas, prometheus.GaugeValue, float64(len(replicas)), topic, strconv.FormatInt(int64(partition), 10),
				)
			}

			inSyncReplicas, err := e.client.InSyncReplicas(topic, partition)
			if err != nil {
				logger.Errorf("Cannot get in-sync replicas of topic %s partition %d: %v", topic, partition, err)
			} else {
				ch <- prometheus.MustNewConstMetric(
					topicPartitionInSyncReplicas, prometheus.GaugeValue, float64(len(inSyncReplicas)), topic, strconv.FormatInt(int64(partition), 10),
				)
			}

			if broker != nil && replicas != nil && len(replicas) > 0 && broker.ID() == replicas[0] {
				ch <- prometheus.MustNewConstMetric(
					topicPartitionUsesPreferredReplica, prometheus.GaugeValue, float64(1), topic, strconv.FormatInt(int64(partition), 10),
				)
			} else {
				ch <- prometheus.MustNewConstMetric(
					topicPartitionUsesPreferredReplica, prometheus.GaugeValue, float64(0), topic, strconv.FormatInt(int64(partition), 10),
				)
			}

			if replicas != nil && inSyncReplicas != nil && len(inSyncReplicas) < len(replicas) {
				ch <- prometheus.MustNewConstMetric(
					topicUnderReplicatedPartition, prometheus.GaugeValue, float64(1), topic, strconv.FormatInt(int64(partition), 10),
				)
			} else {
				ch <- prometheus.MustNewConstMetric(
					topicUnderReplicatedPartition, prometheus.GaugeValue, float64(0), topic, strconv.FormatInt(int64(partition), 10),
				)
			}

			if e.useZooKeeperLag {
				ConsumerGroups, err := e.zookeeperClient.Consumergroups()

				if err != nil {
					logger.Errorf("Cannot get consumer group %v", err)
				}

				for _, group := range ConsumerGroups {
					offset, _ := group.FetchOffset(topic, partition)
					if offset > 0 {

						consumerGroupLag := currentOffset - offset
						ch <- prometheus.MustNewConstMetric(
							consumergroupLagZookeeper, prometheus.GaugeValue, float64(consumerGroupLag), group.Name, topic, strconv.FormatInt(int64(partition), 10),
						)
					}
				}
			}
		}
	}

	loopTopics := func() {
		ok := true
		for ok {
			topic, open := <-topicChannel
			//logger.Warnf("open", open)
			ok = open
			if open {
				getTopicMetrics(topic)
			}
		}
	}

	minx := func(x int, y int) int {
		if x < y {
			return x
		} else {
			return y
		}
	}

	N := len(topics)
	if N > 1 {
		N = minx(N/2, e.topicWorkers)
	}

	for w := 1; w <= N; w++ {
		//logger.Infof("准备获取topic指标")
		go loopTopics()
	}

	for _, topic := range topics {
		// 如果topic在filter里并且未被排除,则采集指标
		//logger.Infof("topic filter", e.topicFilter, topic)

		if e.topicFilter.MatchString(topic) && !e.topicExclude.MatchString(topic) {
			wg.Add(1)
			topicChannel <- topic
		}
	}
	close(topicChannel)

	wg.Wait()

	getConsumerGroupMetrics := func(broker *sarama.Broker) {
		defer wg.Done()
		if err := broker.Open(e.client.Config()); err != nil && err != sarama.ErrAlreadyConnected {
			logger.Errorf("Cannot connect to broker %d: %v", broker.ID(), err)
			return
		}
		defer broker.Close()

		groups, err := broker.ListGroups(&sarama.ListGroupsRequest{})
		if err != nil {
			logger.Errorf("Cannot get consumer group: %v", err)
			return
		}
		groupIds := make([]string, 0)
		for groupId := range groups.Groups {
			if e.groupFilter.MatchString(groupId) && !e.groupExclude.MatchString(groupId) {
				groupIds = append(groupIds, groupId)
			}
		}

		describeGroups, err := broker.DescribeGroups(&sarama.DescribeGroupsRequest{Groups: groupIds})
		if err != nil {
			logger.Errorf("Cannot get describe groups: %v", err)
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
						logger.Errorf("Cannot get GetMemberAssignment of group member %v : %v", member, err)
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
				logger.Errorf("Cannot get offset of group %s: %v", group.GroupId, err)
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
						logger.Errorf("Error for  partition %d :%v", partition, err.Error())
						continue
					}
					currentOffset := offsetFetchResponseBlock.Offset
					currentOffsetSum += currentOffset
					ch <- prometheus.MustNewConstMetric(
						consumergroupCurrentOffset, prometheus.GaugeValue, float64(currentOffset), group.GroupId, topic, strconv.FormatInt(int64(partition), 10),
					)
					e.mu.Lock()
					if offset, ok := offset[topic][partition]; ok {
						// If the topic is consumed by that consumer group, but no offset associated with the partition
						// forcing lag to -1 to be able to alert on that
						var lag int64
						if offsetFetchResponseBlock.Offset == -1 {
							lag = -1
						} else {
							lag = offset - offsetFetchResponseBlock.Offset
							lagSum += lag
						}
						ch <- prometheus.MustNewConstMetric(
							consumergroupLag, prometheus.GaugeValue, float64(lag), group.GroupId, topic, strconv.FormatInt(int64(partition), 10),
						)
					} else {
						logger.Errorf("No offset of topic %s partition %d, cannot get consumer group lag", topic, partition)
					}
					e.mu.Unlock()
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

	logger.Infof("Fetching consumer group metrics")
	if len(e.client.Brokers()) > 0 {
		for _, broker := range e.client.Brokers() {
			wg.Add(1)
			go getConsumerGroupMetrics(broker)
		}
		wg.Wait()
	} else {
		logger.Errorf("No valid broker, cannot get consumer group metrics")
	}
}

func init() {
	metrics.UseNilMetrics = true
}

func Setup(topicFilter string, topicExclude string, groupFilter string, groupExclude string, opts KafkaOpts) (*Exporter, error) {
	initDesc(opts.Namespace)
	return NewExporter(opts, topicFilter, topicExclude, groupFilter, groupExclude)
}
