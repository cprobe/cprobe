package exporter

import "github.com/prometheus/client_golang/prometheus"

func initMetricDesc() map[string]zookeeperMetric {

	return map[string]zookeeperMetric{
		"zk_avg_latency": {
			desc:    prometheus.NewDesc("zk_avg_latency", "Average latency of requests", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_max_latency": {
			desc:    prometheus.NewDesc("zk_max_latency", "Maximum seen latency of requests", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_min_latency": {
			desc:    prometheus.NewDesc("zk_min_latency", "Minimum seen latency of requests", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_packets_received": {
			desc:    prometheus.NewDesc("zk_packets_received", "Number of packets received", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.CounterValue,
		},
		"zk_packets_sent": {
			desc:    prometheus.NewDesc("zk_packets_sent", "Number of packets sent", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.CounterValue,
		},
		"zk_num_alive_connections": {
			desc:    prometheus.NewDesc("zk_num_alive_connections", "Number of active connections", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_outstanding_requests": {
			desc:    prometheus.NewDesc("zk_outstanding_requests", "Number of outstanding requests", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_server_state": {
			desc:    prometheus.NewDesc("zk_server_state", "Server state (leader/follower)", []string{"state"}, nil),
			extract: func(s string) float64 { return 1 },
			extractLabels: func(s string) []string {
				return []string{s}
			},
			valType: prometheus.UntypedValue,
		},
		"zk_znode_count": {
			desc:    prometheus.NewDesc("zk_znode_count", "Number of znodes", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_watch_count": {
			desc:    prometheus.NewDesc("zk_watch_count", "Number of watches", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_approximate_data_size": {
			desc:    prometheus.NewDesc("zk_approximate_data_size", "Approximate size of data set", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_open_file_descriptor_count": {
			desc:    prometheus.NewDesc("zk_open_file_descriptor_count", "Number of open file descriptors", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_max_file_descriptor_count": {
			desc:    prometheus.NewDesc("zk_max_file_descriptor_count", "Maximum number of open file descriptors", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.CounterValue,
		},
		"zk_followers": {
			desc:    prometheus.NewDesc("zk_followers", "Number of followers", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_synced_followers": {
			desc:    prometheus.NewDesc("zk_synced_followers", "Number of followers in sync", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_pending_syncs": {
			desc:    prometheus.NewDesc("zk_pending_syncs", "Number of followers with syncronizations pending", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_global_sessions": {
			desc:    prometheus.NewDesc("zk_global_sessions", "Number of followers with syncronizations pending", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_bytes_received_count": {
			desc:    prometheus.NewDesc("zk_bytes_received_count", "The number of bytes received", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_close_session_prep_time": {
			desc:    prometheus.NewDesc("zk_sum_close_session_prep_time", "Sum of closesessionprep_time", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_close_session_prep_time": {
			desc:    prometheus.NewDesc("zk_cnt_close_session_prep_time", "Total count of closesessionprep_time", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_commit_commit_proc_req_queued": {
			desc:    prometheus.NewDesc("zk_sum_commit_commit_proc_req_queued", "Sum of commitcommitprocreqqueued", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_commit_commit_proc_req_queued": {
			desc:    prometheus.NewDesc("zk_cnt_commit_commit_proc_req_queued", "Total count of commitcommitprocreqqueued", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_commit_count": {
			desc:    prometheus.NewDesc("zk_commit_count", "The number of commits performed on leader", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_commit_process_time": {
			desc:    prometheus.NewDesc("zk_sum_commit_process_time", "Sum of commitprocesstime", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_commit_process_time": {
			desc:    prometheus.NewDesc("zk_cnt_commit_process_time", "Total count of commitprocesstime", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_commit_propagation_latency": {
			desc:    prometheus.NewDesc("zk_sum_commit_propagation_latency", "Sum of commitpropagationlatency", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_commit_propagation_latency": {
			desc:    prometheus.NewDesc("zk_cnt_commit_propagation_latency", "Total count of commitpropagationlatency", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_concurrent_request_processing_in_commit_processor": {
			desc:    prometheus.NewDesc("zk_sum_concurrent_request_processing_in_commit_processor", "Sum of concurrentrequestprocessingincommit_processor", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_concurrent_request_processing_in_commit_processor": {
			desc:    prometheus.NewDesc("zk_cnt_concurrent_request_processing_in_commit_processor", "Total count of concurrentrequestprocessingincommit_processor", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_connection_drop_count": {
			desc:    prometheus.NewDesc("zk_connection_drop_count", "Count of connection drops", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_connection_drop_probability": {
			desc:    prometheus.NewDesc("zk_connection_drop_probability", "Connection drop probability", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_connection_rejected": {
			desc:    prometheus.NewDesc("zk_connection_rejected", "Connection rejected counts", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_connection_request_count": {
			desc:    prometheus.NewDesc("zk_connection_request_count", "Number of incoming client connection requests", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_connection_revalidate_count": {
			desc:    prometheus.NewDesc("zk_connection_revalidate_count", "Count of connection revalidations", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_connection_token_deficit": {
			desc:    prometheus.NewDesc("zk_sum_connection_token_deficit", "Sum of connectiontokendeficit", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_connection_token_deficit": {
			desc:    prometheus.NewDesc("zk_cnt_connection_token_deficit", "Total count of connectiontokendeficit", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_dbinittime": {
			desc:    prometheus.NewDesc("zk_sum_dbinittime", "Sum of dbinittime Time to reload database", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_dbinittime": {
			desc:    prometheus.NewDesc("zk_cnt_dbinittime", "Total count of dbinittime Time to reload database", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_dead_watchers_cleaner_latency": {
			desc:    prometheus.NewDesc("zk_sum_dead_watchers_cleaner_latency", "Sum of dbinittime deadwatcherscleaner_latency", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_dead_watchers_cleaner_latency": {
			desc:    prometheus.NewDesc("zk_cnt_dead_watchers_cleaner_latency", "Total count of deadwatcherscleaner_latency", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_dead_watchers_cleared": {
			desc:    prometheus.NewDesc("zk_dead_watchers_cleared", "", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_dead_watchers_queued": {
			desc:    prometheus.NewDesc("zk_dead_watchers_queued", "", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_diff_count": {
			desc:    prometheus.NewDesc("zk_diff_count", "Number of diff syncs performed", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_digest_mismatches_count": {
			desc:    prometheus.NewDesc("zk_digest_mismatches_count", "", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_election_time": {
			desc:    prometheus.NewDesc("zk_sum_election_time", "Sum of Time between entering and leaving election", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_election_time": {
			desc:    prometheus.NewDesc("zk_cnt_election_time", "Total count of Time between entering and leaving election", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_ensemble_auth_fail": {
			desc:    prometheus.NewDesc("zk_ensemble_auth_fail", "", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_ensemble_auth_skip": {
			desc:    prometheus.NewDesc("zk_ensemble_auth_skip", "", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_ensemble_auth_success": {
			desc:    prometheus.NewDesc("zk_ensemble_auth_success", "", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_ephemerals_count": {
			desc:    prometheus.NewDesc("zk_ephemerals_count", "Number of ephemeral nodes", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_follower_sync_time": {
			desc:    prometheus.NewDesc("zk_sum_follower_sync_time", "Sum of Time for follower to sync with leader", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_follower_sync_time": {
			desc:    prometheus.NewDesc("zk_cnt_follower_sync_time", "Total count of Time for follower to sync with leader", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_fsynctime": {
			desc:    prometheus.NewDesc("zk_sum_fsynctime", "Sum of Time to fsync transaction log", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_fsynctime": {
			desc:    prometheus.NewDesc("zk_cnt_fsynctime", "Total count of Time to fsync transaction log", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_large_requests_rejected": {
			desc:    prometheus.NewDesc("zk_large_requests_rejected", "", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_last_client_response_size": {
			desc:    prometheus.NewDesc("zk_last_client_response_size", "", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_learner_commit_received_count": {
			desc:    prometheus.NewDesc("zk_learner_commit_received_count", "", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_learner_proposal_received_count": {
			desc:    prometheus.NewDesc("zk_learner_proposal_received_count", "", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_local_sessions": {
			desc:    prometheus.NewDesc("zk_local_sessions", "Count of local sessions", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_local_write_committed_time_ms": {
			desc:    prometheus.NewDesc("zk_sum_local_write_committed_time_ms", "Sum of localwritecommittedtimems", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_local_write_committed_time_ms": {
			desc:    prometheus.NewDesc("zk_cnt_local_write_committed_time_ms", "Total count of localwritecommittedtimems", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_looking_count": {
			desc:    prometheus.NewDesc("zk_looking_count", "Number of transitions into looking state", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_max_client_response_size": {
			desc:    prometheus.NewDesc("zk_max_client_response_size", "", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_min_client_response_size": {
			desc:    prometheus.NewDesc("zk_min_client_response_size", "", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_netty_queued_buffer_capacity": {
			desc:    prometheus.NewDesc("zk_sum_netty_queued_buffer_capacity", "Sum of nettyqueuedbuffer_capacity", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_netty_queued_buffer_capacity": {
			desc:    prometheus.NewDesc("zk_cnt_netty_queued_buffer_capacity", "Total count of nettyqueuedbuffer_capacity", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_node_changed_watch_count": {
			desc:    prometheus.NewDesc("zk_sum_node_changed_watch_count", "Sum of nodechangedwatch_count", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_node_changed_watch_count": {
			desc:    prometheus.NewDesc("zk_cnt_node_changed_watch_count", "Total count of nodechangedwatch_count", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_node_children_watch_count": {
			desc:    prometheus.NewDesc("zk_sum_node_children_watch_count", "Sum of nodechildrenwatch_count", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_node_children_watch_count": {
			desc:    prometheus.NewDesc("zk_cnt_node_children_watch_count", "Total count of nodechildrenwatch_count", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_node_created_watch_count": {
			desc:    prometheus.NewDesc("zk_sum_node_created_watch_count", "Sum of nodecreatedwatch_count", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_node_created_watch_count": {
			desc:    prometheus.NewDesc("zk_cnt_node_created_watch_count", "Total count of nodecreatedwatch_count", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_node_deleted_watch_count": {
			desc:    prometheus.NewDesc("zk_sum_node_deleted_watch_count", "Sum of nodedeletedwatch_count", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_node_deleted_watch_count": {
			desc:    prometheus.NewDesc("zk_cnt_node_deleted_watch_count", "Total count of nodedeletedwatch_count", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_om_commit_process_time_ms": {
			desc:    prometheus.NewDesc("zk_sum_om_commit_process_time_ms", "Sum of omcommitprocesstimems", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_om_commit_process_time_ms": {
			desc:    prometheus.NewDesc("zk_cnt_om_commit_process_time_ms", "Total count of omcommitprocesstimems", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_om_proposal_process_time_ms": {
			desc:    prometheus.NewDesc("zk_sum_om_proposal_process_time_ms", "Sum of omproposalprocesstimems", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_outstanding_changes_queued": {
			desc:    prometheus.NewDesc("zk_outstanding_changes_queued", "", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_outstanding_changes_removed": {
			desc:    prometheus.NewDesc("zk_outstanding_changes_removed", "", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_outstanding_tls_handshake": {
			desc:    prometheus.NewDesc("zk_outstanding_tls_handshake", "", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_pending_session_queue_size": {
			desc:    prometheus.NewDesc("zk_sum_pending_session_queue_size", "Sum of pendingsessionqueue_size", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_pending_session_queue_size": {
			desc:    prometheus.NewDesc("zk_cnt_pending_session_queue_size", "Total count of pendingsessionqueue_size", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_prep_process_time": {
			desc:    prometheus.NewDesc("zk_sum_prep_process_time", "Sum of prepprocesstime", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_prep_process_time": {
			desc:    prometheus.NewDesc("zk_cnt_prep_process_time", "Total count of prepprocesstime", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_prep_processor_queue_size": {
			desc:    prometheus.NewDesc("zk_sum_prep_processor_queue_size", "Sum of prepprocessorqueue_size", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_prep_processor_queue_size": {
			desc:    prometheus.NewDesc("zk_cnt_prep_processor_queue_size", "Total count of prepprocessorqueue_size", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_prep_processor_queue_time_ms": {
			desc:    prometheus.NewDesc("zk_sum_prep_processor_queue_time_ms", "Sum of prepprocessorqueuetimems", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_prep_processor_queue_time_ms": {
			desc:    prometheus.NewDesc("zk_cnt_prep_processor_queue_time_ms", "Total count of prepprocessorqueuetimems", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_prep_processor_request_queued": {
			desc:    prometheus.NewDesc("zk_prep_processor_request_queued", "", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_propagation_latency": {
			desc:    prometheus.NewDesc("zk_sum_propagation_latency", "End-to-end latency for updates, from proposal on leader to committed-to-datatree on a given host", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_propagation_latency": {
			desc:    prometheus.NewDesc("zk_cnt_propagation_latency", "End-to-end latency for updates, from proposal on leader to committed-to-datatree on a given host", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_proposal_ack_creation_latency": {
			desc:    prometheus.NewDesc("zk_sum_proposal_ack_creation_latency", "Sum of proposalackcreation_latency", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_proposal_ack_creation_latency": {
			desc:    prometheus.NewDesc("zk_cnt_proposal_ack_creation_latency", "Total count of proposalackcreation_latency", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_proposal_count": {
			desc:    prometheus.NewDesc("zk_proposal_count", "", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_proposal_latency": {
			desc:    prometheus.NewDesc("zk_sum_proposal_latency", "Sum of proposal_latency", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_proposal_latency": {
			desc:    prometheus.NewDesc("zk_cnt_proposal_latency", "Total count of proposal_latency", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_quit_leading_due_to_disloyal_voter": {
			desc:    prometheus.NewDesc("zk_quit_leading_due_to_disloyal_voter", "", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_quorum_ack_latency": {
			desc:    prometheus.NewDesc("zk_sum_quorum_ack_latency", "Sum of quorumacklatency", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_quorum_ack_latency": {
			desc:    prometheus.NewDesc("zk_cnt_quorum_ack_latency", "Total count of quorumacklatency", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_read_commit_proc_issued": {
			desc:    prometheus.NewDesc("zk_sum_read_commit_proc_issued", "Sum of readcommitproc_issued", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_read_commit_proc_issued": {
			desc:    prometheus.NewDesc("zk_cnt_read_commit_proc_issued", "Total count of readcommitproc_issued", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_read_commit_proc_req_queued": {
			desc:    prometheus.NewDesc("zk_sum_read_commit_proc_req_queued", "Sum of readcommitprocreqqueued", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_read_commit_proc_req_queued": {
			desc:    prometheus.NewDesc("zk_cnt_read_commit_proc_req_queued", "Total count of readcommitprocreqqueued", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_read_commitproc_time_ms": {
			desc:    prometheus.NewDesc("zk_sum_read_commitproc_time_ms", "Sum of readfinalproctimems", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_read_commitproc_time_ms": {
			desc:    prometheus.NewDesc("zk_cnt_read_commitproc_time_ms", "Total count of readfinalproctimems", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_read_final_proc_time_ms": {
			desc:    prometheus.NewDesc("zk_sum_read_final_proc_time_ms", "Sum of readfinalproctimems", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_read_final_proc_time_ms": {
			desc:    prometheus.NewDesc("zk_cnt_read_final_proc_time_ms", "Total count of readfinalproctimems", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_readlatency": {
			desc:    prometheus.NewDesc("zk_sum_readlatency", "Sum of readlatency Read request latency", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_readlatency": {
			desc:    prometheus.NewDesc("zk_cnt_readlatency", "Total count of readlatency Read request latency", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_reads_after_write_in_session_queue": {
			desc:    prometheus.NewDesc("zk_sum_reads_after_write_in_session_queue", "Sum of readsafterwriteinsession_queue", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_reads_after_write_in_session_queue": {
			desc:    prometheus.NewDesc("zk_cnt_reads_after_write_in_session_queue", "Total count of readsafterwriteinsession_queue", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_reads_issued_from_session_queue": {
			desc:    prometheus.NewDesc("zk_sum_reads_issued_from_session_queue", "Sum of readsafterwriteinsession_queue", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_reads_issued_from_session_queue": {
			desc:    prometheus.NewDesc("zk_cnt_reads_issued_from_session_queue", "Total count of readsafterwriteinsession_queue", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_request_commit_queued": {
			desc:    prometheus.NewDesc("zk_request_commit_queued", "Count of request commits queued", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_request_throttle_wait_count": {
			desc:    prometheus.NewDesc("zk_request_throttle_wait_count", "", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_requests_in_session_queue": {
			desc:    prometheus.NewDesc("zk_sum_requests_in_session_queue", "Sum of requestsinsession_queue", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_requests_in_session_queue": {
			desc:    prometheus.NewDesc("zk_cnt_requests_in_session_queue", "Total count of requestsinsession_queue", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_response_packet_cache_hits": {
			desc:    prometheus.NewDesc("zk_response_packet_cache_hits", "", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_response_packet_cache_misses": {
			desc:    prometheus.NewDesc("zk_response_packet_cache_misses", "", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_response_packet_get_children_cache_hits": {
			desc:    prometheus.NewDesc("zk_response_packet_get_children_cache_hits", "", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_response_packet_get_children_cache_misses": {
			desc:    prometheus.NewDesc("zk_response_packet_get_children_cache_misses", "", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_revalidate_count": {
			desc:    prometheus.NewDesc("zk_revalidate_count", "", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_server_write_committed_time_ms": {
			desc:    prometheus.NewDesc("zk_sum_server_write_committed_time_ms", "Sum of serverwritecommittedtimems", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_server_write_committed_time_ms": {
			desc:    prometheus.NewDesc("zk_cnt_server_write_committed_time_ms", "Total count of serverwritecommittedtimems", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_session_queues_drained": {
			desc:    prometheus.NewDesc("zk_sum_session_queues_drained", "Sum of sessionqueuesdrained", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_session_queues_drained": {
			desc:    prometheus.NewDesc("zk_cnt_session_queues_drained", "Total count of sessionqueuesdrained", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sessionless_connections_expired": {
			desc:    prometheus.NewDesc("zk_sessionless_connections_expired", "", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_snap_count": {
			desc:    prometheus.NewDesc("zk_snap_count", "Number of snap syncs performed", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_snapshottime": {
			desc:    prometheus.NewDesc("zk_sum_snapshottime", "Sum of snapshottime Time to write a snapshot", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_snapshottime": {
			desc:    prometheus.NewDesc("zk_cnt_snapshottime", "Total count of snapshottime Time to write a snapshot", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_stale_replies": {
			desc:    prometheus.NewDesc("zk_stale_replies", "", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_stale_requests": {
			desc:    prometheus.NewDesc("zk_stale_requests", "", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_stale_requests_dropped": {
			desc:    prometheus.NewDesc("zk_stale_requests_dropped", "", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_stale_sessions_expired": {
			desc:    prometheus.NewDesc("zk_stale_sessions_expired", "", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_startup_snap_load_time": {
			desc:    prometheus.NewDesc("zk_sum_startup_snap_load_time", "Sum of startupsnapload_time", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_startup_snap_load_time": {
			desc:    prometheus.NewDesc("zk_cnt_startup_snap_load_time", "Total count of startupsnapload_time", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_startup_txns_load_time": {
			desc:    prometheus.NewDesc("zk_sum_startup_txns_load_time", "Sum of startuptxnsload_time", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_startup_txns_load_time": {
			desc:    prometheus.NewDesc("zk_cnt_startup_txns_load_time", "Total count of startuptxnsload_time", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_startup_txns_loaded": {
			desc:    prometheus.NewDesc("zk_sum_startup_txns_loaded", "Sum of startuptxnsloaded", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_startup_txns_loaded": {
			desc:    prometheus.NewDesc("zk_cnt_startup_txns_loaded", "Total count of startuptxnsloaded", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_sync_process_time": {
			desc:    prometheus.NewDesc("zk_sum_sync_process_time", "Sum of syncprocesstime", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_sync_process_time": {
			desc:    prometheus.NewDesc("zk_cnt_sync_process_time", "Total count of syncprocesstime", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_sync_processor_batch_size": {
			desc:    prometheus.NewDesc("zk_sum_sync_processor_batch_size", "Sum of syncprocessorbatch_size", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_sync_processor_batch_size": {
			desc:    prometheus.NewDesc("zk_cnt_sync_processor_batch_size", "Total count of syncprocessorbatch_size", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_sync_processor_queue_and_flush_time_ms": {
			desc:    prometheus.NewDesc("zk_sum_sync_processor_queue_and_flush_time_ms", "Sum of syncprocessorqueueandflushtimems", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_sync_processor_queue_and_flush_time_ms": {
			desc:    prometheus.NewDesc("zk_cnt_sync_processor_queue_and_flush_time_ms", "Total count of syncprocessorqueueandflushtimems", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_sync_processor_queue_flush_time_ms": {
			desc:    prometheus.NewDesc("zk_sum_sync_processor_queue_flush_time_ms", "Sum of syncprocessorqueueflushtime_ms", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_sync_processor_queue_flush_time_ms": {
			desc:    prometheus.NewDesc("zk_cnt_sync_processor_queue_flush_time_ms", "Total count of syncprocessorqueueflushtime_ms", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_sync_processor_queue_size": {
			desc:    prometheus.NewDesc("zk_sum_sync_processor_queue_size", "Sum of syncprocessorqueue_size", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_sync_processor_queue_size": {
			desc:    prometheus.NewDesc("zk_cnt_sync_processor_queue_size", "Total count of syncprocessorqueue_size", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_sync_processor_queue_time_ms": {
			desc:    prometheus.NewDesc("zk_sum_sync_processor_queue_time_ms", "Sum of syncprocessorqueuetimems", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_sync_processor_queue_time_ms": {
			desc:    prometheus.NewDesc("zk_cnt_sync_processor_queue_time_ms", "Total count of syncprocessorqueuetimems", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sync_processor_request_queued": {
			desc:    prometheus.NewDesc("zk_sync_processor_request_queued", "", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_time_waiting_empty_pool_in_commit_processor_read_ms": {
			desc:    prometheus.NewDesc("zk_sum_time_waiting_empty_pool_in_commit_processor_read_ms", "Sum of timewaitingemptypoolincommitprocessorreadms", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_time_waiting_empty_pool_in_commit_processor_read_ms": {
			desc:    prometheus.NewDesc("zk_cnt_time_waiting_empty_pool_in_commit_processor_read_ms", "Total count of timewaitingemptypoolincommitprocessorreadms", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_tls_handshake_exceeded": {
			desc:    prometheus.NewDesc("zk_tls_handshake_exceeded", "", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_unrecoverable_error_count": {
			desc:    prometheus.NewDesc("zk_unrecoverable_error_count", "", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_updatelatency": {
			desc:    prometheus.NewDesc("zk_sum_updatelatency", "Sum of updatelatency Update request latency", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_updatelatency": {
			desc:    prometheus.NewDesc("zk_cnt_updatelatency", "Total count of updatelatency Update request latency", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_uptime": {
			desc:    prometheus.NewDesc("zk_uptime", "Uptime that a peer has been in a table leading/following/observing state", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_write_batch_time_in_commit_processor": {
			desc:    prometheus.NewDesc("zk_sum_write_batch_time_in_commit_processor", "Sum of writebatchtimeincommit_processor", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_write_batch_time_in_commit_processor": {
			desc:    prometheus.NewDesc("zk_cnt_write_batch_time_in_commit_processor", "Total count of writebatchtimeincommit_processor", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_write_commit_proc_issued": {
			desc:    prometheus.NewDesc("zk_sum_write_commit_proc_issued", "Sum of writecommitproc_issued", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_write_commit_proc_issued": {
			desc:    prometheus.NewDesc("zk_cnt_write_commit_proc_issued", "Total count of writecommitproc_issued", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_write_commit_proc_req_queued": {
			desc:    prometheus.NewDesc("zk_sum_write_commit_proc_req_queued", "Sum of writecommitprocreqqueued", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_write_commit_proc_req_queued": {
			desc:    prometheus.NewDesc("zk_cnt_write_commit_proc_req_queued", "Total count of writecommitprocreqqueued", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_write_commitproc_time_ms": {
			desc:    prometheus.NewDesc("zk_sum_write_commitproc_time_ms", "Sum of writecommitproctime_ms", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_write_commitproc_time_ms": {
			desc:    prometheus.NewDesc("zk_cnt_write_commitproc_time_ms", "Total count of writecommitproctime_ms", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_sum_write_final_proc_time_ms": {
			desc:    prometheus.NewDesc("zk_sum_write_final_proc_time_ms", "Sum of writefinalproctimems", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
		"zk_cnt_write_final_proc_time_ms": {
			desc:    prometheus.NewDesc("zk_cnt_write_final_proc_time_ms", "Total count of writefinalproctimems", nil, nil),
			extract: func(s string) float64 { return parseFloatOrZero(s) },
			valType: prometheus.GaugeValue,
		},
	}
}
