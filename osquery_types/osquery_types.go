package osquery_types

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type OsqueryClient struct {
	HostIdentifier              string                       `json:"host_identifier"`
	NodeKey                     string                       `json:"node_key"`
	NodeInvalid                 bool                         `json:"node_invalid"`
	HostDetails                 map[string]map[string]string `json:"host_details"`
	PendingRegistrationApproval bool                         `json:"pending_registration_approval"`
	Tags                        []string                     `json:"tags,omitempty"`
	ConfigurationGroup          string                       `json:"configuration_group,omitempty"`
	ConfigName                  string                       `json:"config_name"`
	LastUpdated                 string                       `json:"last_updated"`
}

// SetTimestamp sets the current timestamp with the proper format
func (os *OsqueryClient) SetTimestamp() {
	os.LastUpdated = time.Now().UTC().Format("Mon, 01/02/06, 03:04:05PM")
}

type OsqueryOptions struct {
	//Audit
	AuditAllowConfig  bool `json:"audit_allow_config"`
	AuditAllowSockets bool `json:"audit_allow_sockets"`
	AuditPersist      bool `json:"audit_persist"`
	//aws options
	AwsAccessKeyID               string `json:"aws_access_key_id,omitempty"`
	AwsFirehosePeriod            int    `json:"aws_firehose_period,omitempty"`
	AwsFirehoseStream            string `json:"aws_firehose_stream"`
	AwsKinesisPeriod             int    `json:"aws_kinesis_period,omitempty"`
	AwsKinesisRandomPartitionKey bool   `json:"aws_kinesis_random_partition_key,omitempty"`
	AwsKinesisStream             string `json:"aws_kinesis_stream,omitempty"`
	AwsProfileName               string `json:"aws_profile_name,omitempty"`
	AwsRegion                    string `json:"aws_region,omitempty"`
	AwsSecretAccessKey           string `json:"aws_secret_access_key,omitempty"`
	AwsSTSARNRole                string `json:"aws_sts_arn_role,omitempty"`
	AwsSTSRegion                 string `json:"aws_sts_region,omitempty"`
	AwsSTSSessionName            string `json:"aws_sts_session_name,omitempty"`
	AwsSTSTimeout                string `json:"aws_sts_timeout,omitempty"`
	//Carver settings
	CarverBlockSize        int    `json:"carver_block_size,omitempty"`
	CarverContinueEndpoint string `json:"carver_continue_endpoint,omitempty"`
	CarverStartEndpoint    string `json:"carver_start_endpoint,omitempty"`
	CarverDisableFunction  bool   `json:"carver_disable_function"`
	//config_settings
	ConfigRefresh int  `json:"config_refresh"`
	CSV           bool `json:"csv,omitempty"`

	//Disables
	DisableAudit        bool `json:"disable_audit"`
	DisableCaching      bool `json:"disable_caching"`
	DisableCarver       bool `json:"disable_carver"`
	DisableDatabase     bool `json:"disable_database"`
	DisableDecorators   bool `json:"disable_decorators"`
	DisableDistributed  bool `json:"disable_distributed"`
	DisableEnrollment   bool `json:"disable_enrollment"`
	DisableEvents       bool `json:"disable_events"`
	DisableExtensions   bool `json:"disable_extensions"`
	DisableForensic     bool `json:"disable_forensic"`
	DisableKernel       bool `json:"disable_kernel"`
	DisableLogging      bool `json:"disable_logging"`
	DisableMemory       bool `json:"disable_memory"`
	DisableReenrollment bool `json:"disable_reenrollment"`
	DisableTables       bool `json:"disable_tables"`
	DisableWatchdog     bool `json:"disable_watchdog"`

	//Distributed
	DistributedInterval         int    `json:"distributed_interval,omitempty"`
	DistributedPlugin           string `json:"distributed_plugin,omitempty"`
	DistributedTLSMaxAttempts   int    `json:"distributed_tls_max_attempts,omitempty"`
	DistributedTLSReadEndpoint  string `json:"distributed_tls_read_endpoint,omitempty"`
	DistributedTLSWriteEndpoint string `json:"distributed_tls_write_endpoint,omitempty"`
	//Enables

	EnableForeign bool `json:"enable_foreign"`
	EnableMonitor bool `json:"enable_monitor"`
	EnableSyslog  bool `json:"enable_syslog"`

	//Enroll (these are handled by flags on host, not set in tls config

	//Events
	EventsExpiry   int  `json:"events_expiry"`
	EventsMax      int  `json:"events_max"`
	EventsOptimize bool `json:"events_optimize"`

	//Extensions
	ExtensionsAutoload bool   `json:"extenstions_autoload,omitempty"`
	ExtensionsInterval int    `json:"extensions_interval,omitempty"`
	ExtensionsRequire  string `json:"extensions_require,omitempty"`
	ExtensionsTimeout  int    `json:"extensions_timeout,omitempty"`

	Force                 bool   `json:"force,omitempty"`
	HardwareDisabledTypes string `json:"hardware_disabled_types,omitempty"`
	Header                bool   `json:"header,omitempty"`
	HostIdentifier        string `json:"host_identifier"`
	//output
	JSON bool `json:"json,omitempty"`
	Line bool `json:"line,omitempty"`
	List bool `json:"list,omitempty"`

	//Logger
	LoggerEventType bool   `json:"logger_event_type,omitempty"`
	LoggerMinStatus int    `json:"logger_min_status,omitempty"`
	LoggerMode      int    `json:"logger_mode,omitempty"`
	LoggerPath      string `json:"logger_path,omitempty"`
	LoggerPlugin    string `json:"logger_plugin"`

	LoggerSecondaryStatusOnly bool `json:"logger_secondary_status_only,omitempty"`
	LoggerSnapshotEventType   bool `json:"logger_snapshot_event_type,omitempty"`
	LoggerStatusSync          bool `json:"logger_status_sync,omitempty"`

	LoggerSyslogFacility   int  `json:"logger_syslog_facility,omitempty"`
	LoggerSyslogPrependCee bool `json:"logger_syslog_prepend_cee,omitempty"`
	LoggerTLSCompress      bool `json:"logger_tls_compress,omitempty"`
	//Endpoints provided by flags
	LoggerTLSMax    int  `json:"logger_tls_max,omitempty"`
	LoggerTLSPeriod int  `json:"logger_tls_period,omitempty"`
	Logtostderr     bool `json:"logtostderr,omitempty"`
	//Schedule
	ScheduleDefaultInterval int `json:"schedule_default_interval,omitempty"`
	ScheduleSplayPercent    int `json:"schedule_splay_percent,omitempty"`
	//Syslog
	SyslogEventsExpiry int    `json:"syslog_events_expiry,omitempty"`
	SyslogEventsMax    int    `json:"syslog_events_max,omitempty"`
	SyslogPipePath     string `json:"syslog_pipe_path,omitempty"`
	SyslogRateLimit    int    `json:"syslog_rate_limit,omitempty"`
	//TLS settings should be specified in flags file, since there is no guarantee of tls communcation without it
	UTC     bool `json:"utc,omitempty"`
	Verbose bool `json:"verbose"`
	//Watchdog
	WatchdogLevel            int `json:"watchdog_level,omitempty"`
	WatchdogMemoryLimit      int `json:"watchdog_memory_limit,omitempty"`
	WatchdogUtilizationLimit int `json:"watchdog_utilization_limit,omitempty"`
}

// NewOsqueryOptions returns some default options for osquery
func NewOsqueryOptions() OsqueryOptions {
	options := OsqueryOptions{
		AuditPersist:              true,
		ConfigRefresh:             300,
		DisableAudit:              true,
		DisableCarver:             true,
		DistributedInterval:       60,
		DistributedTLSMaxAttempts: 5,
		EventsExpiry:              14400,
		EventsMax:                 100000,
		EventsOptimize:            true,
		HostIdentifier:            "hostname",
		LoggerPlugin:              "firehose",
		//LoggerSnapshotEventType:	true,
	}
	return options
}

type OsqueryDecorators struct {
	Load   []string `json:"load,omitempty"`
	Always []string `json:"always,omitempty"`
}
type OsqueryQuery struct {
	Query string `json:"query"`
}

type Time struct {
	Query    string `json:"query"`
	Interval int    `json:"interval"`
	Removed  string `json:"removed"`
}

type OsquerySchedule struct {
	Time Time `json:"time"`
}

type OsqueryConfig struct {
	//Node_invalid string
	NodeInvalid bool
	Options     OsqueryOptions    `json:"options"`
	Decorators  OsqueryDecorators `json:"decorators,omitemtpy"`
	Schedule    OsquerySchedule   `json:"schedule,omitempty"`
	//Packs OsqueryPacks `json:"packs"`
	Packs map[string]map[string]map[string]map[string]string `json:"packs"`
}

type OsqueryUploadConfig struct {
	//Node_invalid string
	NodeInvalid bool
	Options     OsqueryOptions    `json:"options"`
	Decorators  OsqueryDecorators `json:"decorators,omitemtpy"`
	Schedule    OsquerySchedule   `json:"schedule,omitempty"`
	Packs       []string          `json:"packs"`
	//Packs OsqueryPacks `json:"packs"`
}

type OsqueryNamedConfig struct {
	ConfigName    string        `json:"config_name"`
	OsqueryConfig OsqueryConfig `json:"osquery_config"`
	OsType        string        `json:"os_type"`
	PackList      []string      `json:"pack_list"`
}

type Pack struct {
	PackName string `json:"pack_name"`
	//QueryList []string `json:"query_list"`
	Queries []PackQuery `json:"queries"`
}

// deprecated
//func (p Pack) AsString() string {
//s := fmt.Sprintf("%q: ", p.PackName)
//s += `{"queries": `
//s += BuildPackQueries(p.Queries)
//s += "}}"
//return s
//}

// deprecated
//func (p Pack) AsRawJson() json.RawMessage {
//return json.RawMessage(p.AsString())
//}

func (p Pack) AsMap() map[string]map[string]map[string]string {
	m := map[string]map[string]map[string]string{}
	m["queries"] = map[string]map[string]string{}
	for _, packQuery := range p.Queries {
		pq := map[string]string{}
		pq["query"] = packQuery.Query
		pq["interval"] = packQuery.Interval
		//if len(packQuery.Version) > 0 {
		pq["version"] = packQuery.Version
		//}
		//if len(packQuery.Description) > 0 {
		pq["description"] = packQuery.Description
		//}
		//if len(packQuery.Value) > 0 {
		pq["value"] = packQuery.Value
		//}
		//if len(packQuery.Snapshot) > 0 {
		pq["snapshot"] = packQuery.Snapshot
		//}
		m["queries"][packQuery.QueryName] = pq
	}
	return m
}

type QueryPack struct {
	PackName string   `json:"pack_name"`
	Queries  []string `json:"queries"`
}

type PackQuery struct {
	QueryName   string `json:"query_name"`
	Query       string `json:"query"`
	Interval    string `json:"interval"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Value       string `json:"value"`
	Snapshot    string `json:"snapshot"`
}

func (pq PackQuery) AsString() string {
	s := fmt.Sprintf(`"%s": {"query": %q, "interval": %q, "version": %q, "description": %q, "value": %q}`, pq.QueryName, pq.Query, pq.Interval, pq.Version, pq.Description, pq.Value)
	return s
}

func PackQueryToString(p *PackQuery) string {
	s := fmt.Sprintf(`"%s": {"query": %q, "interval": %q, "version": %q, "description": %q, "value": %q}`, p.QueryName, p.Query, p.Interval, p.Version, p.Description, p.Value)
	return s
}

func BuildPackQueries(pqs []PackQuery) string {
	queriesString := "{"
	for c, i := range pqs {
		switch c {
		case 0:
			queriesString += i.AsString()
		case len(pqs):
			queriesString += i.AsString()
			queriesString += "}"
			return queriesString
		default:
			queriesString += ", "
			queriesString += i.AsString()
		}
	}
	return queriesString
}

type DistributedQuery struct {
	NodeKey     string   `json:"node_key"`
	Queries     []string `json:"queries"`
	NodeInvalid bool     `json:"node_invalid"`
}

// ToJSON returns a formatted version of the DistributedQuery
func (dq DistributedQuery) ToJSON() string {
	result := make(map[string]interface{})
	querylist := make(map[string]string)
	for i, j := range dq.Queries {
		querylist[fmt.Sprintf("id%d", i+1)] = j
	}
	result["queries"] = querylist
	result["node_invalid"] = strconv.FormatBool(dq.NodeInvalid)

	js, _ := json.Marshal(result)

	return string(js)
}

type ServerConfig struct {
	FirehoseAWSAccessKeyID                   string   `json:"firehose_aws_access_key_id"`
	FirehoseAWSSecretAccessKey               string   `json:"firehose_aws_secret_access_key"`
	FirehoseStreamName                       string   `json:"firehose_stream_name"`
	DistributedQueryLogger                   []string `json:"distributed_query_logger"`
	DistributedQueryLoggerS3BucketName       string   `json:"distributed_query_logger_s3_bucket_name"`
	DistributedQueryLoggerFirehoseStreamName string   `json:"distributed_query_logger_firehose_stream_name"`
	DistributedQueryLoggerFilesytemPath      string   `json:"distributed_query_logger_filesytem_path"`
	AutoApproveNodes						 string	  `json:"auto_approve_nodes"`
}

func GetServerConfig(fn string) (*ServerConfig, error) {

	config := ServerConfig{}
	file, err := os.Open(fn)
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

type User struct {
	Username string `json:"username"`
	Password []byte `json:"password"`
	Role     string `json:"role"`
}

func (u User) Validate(plaintext_pw string) error {
	return bcrypt.CompareHashAndPassword(u.Password, []byte(plaintext_pw))
}

type DistributedQueryResult struct {
	Name           string            `json:"name"`
	CalendarTime   string            `json:"calendarTime"`
	Action         string            `json:"action"`
	LogType        string            `json:"log_type"`
	Columns        map[string]string `json:"columns"`
	HostIdentifier string            `json:"host_identifier"`
}
