package osquery_types

import (
	"fmt"
	"encoding/json"
	"os"
	"github.com/oktasecuritylabs/sgt/logger"
	"strings"
	"strconv"
	"golang.org/x/crypto/bcrypt"
	"time"
)


type OsqueryClient struct {
	Host_identifier string `json:"host_identifier"`
	Node_key string `json:"node_key"`
	Node_invalid bool `json:"node_invalid"`
	HostDetails map[string]map[string]string `json:"host_details"`
	Pending_registration_approval bool `json:"pending_registration_approval"`
	Tags []string `json:"tags,omitempty"`
	Configuration_group string `json:"configuration_group,omitempty"`
	Config_name string `json:"config_name"`
	LastUpdated string `json:"last_updated"`
}

func (os *OsqueryClient) Timestamp() {
	os.LastUpdated = time.Now().UTC().Format("Mon, 01/02/06, 03:04:05PM")
}

type OsqueryOptions struct {
	//Audit
	Audit_allow_config bool `json:"audit_allow_config"`
	AuditAllowSockets bool `json:"audit_allow_sockets"`
	AuditPersist bool `json:"audit_persist"`
	//aws options
	Aws_access_key_id string `json:"aws_access_key_id,omitempty"`
	Aws_firehose_period int `json:"aws_firehose_period,omitempty"`
	Aws_firehose_stream string `json:"aws_firehose_stream"`
	Aws_kinesis_period int `json:"aws_kinesis_period,omitempty"`
	Aws_kinesis_random_partition_key bool `json:"aws_kinesis_random_partition_key,omitempty"`
	Aws_kinesis_stream string `json:"aws_kinesis_stream,omitempty"`
	Aws_profile_name string `json:"aws_profile_name,omitempty"`
	Aws_region string `json:"aws_region,omitempty"`
	Aws_secret_access_key string `json:"aws_secret_access_key,omitempty"`
	Aws_sts_arn_role string `json:"aws_sts_arn_role,omitempty"`
	Aws_sts_region string `json:"aws_sts_region,omitempty"`
	Aws_sts_session_name string `json:"aws_sts_session_name,omitempty"`
	Aws_sts_timeout string `json:"aws_sts_timeout,omitempty"`
	//Carver settings
	Carver_block_size int `json:"carver_block_size,omitempty"`
	Carver_continue_endpoint string `json:"carver_continue_endpoint,omitempty"`
	Carver_start_endpoint string `json:"carver_start_endpoint,omitempty"`
	//config_settings
	Config_refresh int `json:"config_refresh"`
	CSV bool `json:"csv,omitempty"`
	//disables

	Disable_audit bool `json:"disable_audit"`
	Disable_caching bool `json:"disable_caching"`
	Disable_carver bool `json:"disable_carver"`
	Disable_database bool `json:"disable_database"`
	Disable_decorators bool `json:"disable_decorators"`
	Disable_distributed bool `json:"disable_distributed"`
	Disable_enrollment bool `json:"disable_enrollment"`
	Disable_events bool `json:"disable_events"`
	Disable_extensions bool `json:"disable_extensions"`
	Disable_forensic bool `json:"disable_forensic"`
	Disable_kernel bool `json:"disable_kernel"`
	Disable_logging bool `json:"disable_logging"`
	Disable_memory bool `json:"disable_memory"`
	Disable_reenrollment bool `json:"disable_reenrollment"`
	Disable_tables bool `json:"disable_tables"`
	Disable_watchdog bool `json:"disable_watchdog"`

	//Distributed
	Distributed_interval int `json:"distributed_interval,omitempty"`
	Distributed_plugin string `json:"distributed_plugin,omitempty"`
	Distributed_tls_max_attempts int `json:"distributed_tls_max_attempts,omitempty"`
	Distributed_tls_read_endpoint string `json:"distributed_tls_read_endpoint,omitempty"`
	Distributed_tls_write_endpoint string `json:"distributed_tls_write_endpoint,omitempty"`
	//Enables

	Enable_foreign bool `json:"enable_foreign"`
	Enable_monitor bool `json:"enable_monitor"`
	Enable_syslog bool `json:"enable_syslog"`

	//Enroll (these are handled by flags on host, not set in tls config

	//Events
	Events_expiry int `json:"events_expiry"`
	Events_max int `json:"events_max"`
	Events_optimize bool `json:"events_optimize"`
	
	//Extensions
	Extensions_autoload bool `json:"extenstions_autoload,omitempty"`
	Extensions_interval int `json:"extensions_interval,omitempty"`
	Extensions_require string `json:"extensions_require,omitempty"`
	Extensions_timeout int `json:"extensions_timeout,omitempty"`
	
	Force bool `json:"force,omitempty"`	
	Hardware_disabled_types string `json:"hardware_disabled_types,omitempty"`
	Header bool `json:"header,omitempty"`
	Host_identifier string `json:"host_identifier"`
	//output
	Json bool `json:"json,omitempty"`
	Line bool `json:"line,omitempty"`
	List bool `json:"list,omitempty"`
	
	//Logger
	Logger_event_type bool `json:"logger_event_type,omitempty"`
	Logger_min_status int `json:"logger_min_status,omitempty"`
	Logger_mode int `json:"logger_mode,omitempty"`
	Logger_path string `json:"logger_path,omitempty"`
	Logger_plugin string `json:"logger_plugin"`
	
	Logger_secondary_status_only bool `json:"logger_secondary_status_only,omitempty"`
	Logger_status_sync bool `json:"logger_status_sync,omitempty"`
	
	Logger_syslog_facility int `json:"logger_syslog_facility,omitempty"`
	Logger_syslog_prepend_cee bool `json:"logger_syslog_prepend_cee,omitempty"`
	Logger_tls_compress bool `json:"logger_tls_compress,omitempty"`
	//Endpoints provided by flags
	Logger_tls_max int `json:"logger_tls_max,omitempty"`
	Logger_tls_period int `json:"logger_tls_period,omitempty"`
	Logtostderr bool `json:"logtostderr,omitempty"`
	//Schedule
	Schedule_default_interval int `json:"schedule_default_interval,omitempty"`
	Schedule_splay_percent int `json:"schedule_splay_percent,omitempty"`
	//Syslog
	Syslog_events_expiry int `json:"syslog_events_expiry,omitempty"`
	Syslog_events_max int `json:"syslog_events_max,omitempty"`
	Syslog_pipe_path string `json:"syslog_pipe_path,omitempty"`
	Syslog_rate_limit int `json:"syslog_rate_limit,omitempty"`
	//TLS settings should be specified in flags file, since there is no guarantee of tls communcation without it
	Utc bool `json:"utc,omitempty"`
	Verbose bool `json:"verbose"`
	//Watchdog
	Watchdog_level int `json:"watchdog_level,omitempty"`
	Watchdog_memory_limit int `json:"watchdog_memory_limit,omitempty"`
	Watchdog_utilization_limit int `json:"watchdog_utilization_limit,omitempty"`
}

func NewOsqueryOptions() (OsqueryOptions){
	options := OsqueryOptions{}
	options.Audit_allow_config = false
	options.AuditAllowSockets = false
	options.AuditPersist = true
	options.Config_refresh = 300
	options.Disable_audit = true
	options.Disable_caching = false
	options.Disable_carver = true
	options.Disable_database = false
	options.Disable_decorators = false
	options.Disable_distributed = false
	options.Disable_enrollment = false
	options.Disable_events = false
	options.Disable_extensions = false
	options.Disable_forensic = false
	options.Disable_kernel = false
	options.Disable_logging = false
	options.Disable_memory = false
	options.Disable_reenrollment = false
	options.Disable_tables = false
	options.Disable_watchdog = false
	options.Distributed_interval = 60
	options.Distributed_tls_max_attempts = 5
	options.Enable_foreign = false
	options.Enable_monitor = false
	options.Enable_syslog = false
	options.Events_expiry = 14400
	options.Events_max = 100000
	options.Events_optimize = true
	options.Extensions_autoload = false
	options.Host_identifier = "hostname"
	options.Logger_plugin = "firehose"
	options.Verbose = false
	return options
}

type OsqueryPacks struct {
	Fedramp string `json:"fedramp"`

}

//type OsqueryDecorators struct {
	//Load []OsqueryQuery `json:"load"`
	//Always []OsqueryQuery `json:"always"`
//}

type OsqueryDecorators struct {
	Load []string `json:"load,omitempty"`
	Always []string `json:"always,omitempty"`
}
type OsqueryQuery struct {
	Query string `json:"query"`
}

type Time struct {
	Query string `json:"query"`
	Interval int `json:"interval"`
	Removed string `json:"removed"`
}

type OsquerySchedule struct {
	Time Time `json:"time"`
}

type OsqueryConfig struct {
	//Node_invalid string
	Node_invalid bool
	Options OsqueryOptions `json:"options"`
	Decorators OsqueryDecorators `json:"decorators,omitemtpy"`
	Schedule OsquerySchedule `json:"schedule,omitempty"`
	//Packs OsqueryPacks `json:"packs"`
	Packs *json.RawMessage `json:"packs,omitempty"`
}

type OsqueryNamedConfig struct {
	Config_name string `json:"config_name"`
	Osquery_config OsqueryConfig `json:"osquery_config"`
	Os_type string `json:"os_type"`
	PackList []string `json:"pack_list"`
}



type Pack struct {
	PackName string `json:"pack_name"`
	//QueryList []string `json:"query_list"`
	Queries []PackQuery `json:"queries"`
}

func (p Pack) AsString() string {
	s := fmt.Sprintf("%q: ", p.PackName)
	s += `{"queries": `
	s += BuildPackQueries(p.Queries)
	s += "}}"
	return s
}

func (p Pack) AsRawJson() json.RawMessage {
	return json.RawMessage(p.AsString())
}

type QueryPack struct {
		PackName string `json:"pack_name"`
		Queries []string `json:"queries"`
	}

type PackQuery struct {
	QueryName string `json:"query_name"`
	Query string `json:"query"`
	Interval string `json:"interval"`
	Version string `json:"version"`
	Description string `json:"description"`
	Value string `json:"value"`
}

func (pq PackQuery) AsString() string {
	s := fmt.Sprintf(`"%s": {"query": %q, "interval": %q, "version": %q, "description": %q, "value": %q}`, pq.QueryName, pq.Query, pq.Interval, pq.Version, pq.Description, pq.Value)
	return s
}

func PackQueryToString(p *PackQuery) (string) { s := fmt.Sprintf(`"%s": {"query": %q, "interval": %q, "version": %q, "description": %q, "value": %q}`, p.QueryName, p.Query, p.Interval, p.Version, p.Description, p.Value)
	return s
}

func BuildPackQueries(pqs []PackQuery)(string) {
	queries_string := "{"
	for c, i := range pqs{
		switch c {
		case 0:
			queries_string += i.AsString()
		case len(pqs):
			queries_string += i.AsString()
			queries_string += "}"
			return queries_string
		default:
			queries_string += ", "
			queries_string += i.AsString()
		}
	}
	return queries_string
}


type DistributedQuery struct {
	NodeKey string `json:"node_key"`
	Queries []string `json:"queries"`
	NodeInvalid bool `json:"node_invalid"`
}

func (dq DistributedQuery) ToJson() (string){
	js := `{"queries": {`
	var querylist []string
	for i, j := range dq.Queries {
		querylist = append(querylist, fmt.Sprintf(`"id%d": "%s"`, i+1, j))
	}
	qstrings := strings.Join(querylist, ",")
	js += qstrings
	js += fmt.Sprintf(`}, "node_invalid": %s}`, strconv.FormatBool(dq.NodeInvalid))
	return js
}


type ServerConfig struct {
	FirehoseAWSAccessKeyID                   string `json:"firehose_aws_access_key_id"`
	FirehoseAWSSecretAccessKey               string `json:"firehose_aws_secret_access_key"`
	FirehoseStreamName                       string `json:"firehose_stream_name"`
	DistributedQueryLogger                   []string `json:"distributed_query_logger"`
	DistributedQueryLoggerS3BucketName       string `json:"distributed_query_logger_s3_bucket_name"`
	DistributedQueryLoggerFirehoseStreamName string `json:"distributed_query_logger_firehose_stream_name"`
	DistributedQueryLoggerFilesytemPath      string `json:"distributed_query_logger_filesytem_path"`
}

func GetServerConfig(fn string) (ServerConfig, error) {
	file, err := os.Open(fn)
	if err != nil {
		logger.Error(err)
	}
	decoder := json.NewDecoder(file)
	config := ServerConfig{}
	err = decoder.Decode(&config)
	if err != nil {
		logger.Error(err)
		return config, err
	}
	return config, nil
}

type User struct {
    Username string `json:"username"`
    Password []byte `json:"password"`
    Role string `json:"role"`
}

func (u User) Validate(plaintext_pw string) (error){
	err := bcrypt.CompareHashAndPassword(u.Password, []byte(plaintext_pw))
	if err != nil {
		logger.Error(err)
		return err
	}
	return nil
}

type DistributedQueryResult struct {
	Name string `json:"name"`
	CalendarTime string `json:"calendarTime"`
	Action string `json:"action"`
	LogType string `json:"log_type"`
	Columns map[string]string `json:"columns"`
	HostIdentifier string `json:"host_identifier"`
}