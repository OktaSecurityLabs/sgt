package node

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"strings"
	"sync"

	"github.com/oktasecuritylabs/sgt/dyndb"
	"github.com/oktasecuritylabs/sgt/handlers/auth"
	"github.com/oktasecuritylabs/sgt/logger"
	"github.com/oktasecuritylabs/sgt/osquery_types"
	log "github.com/sirupsen/logrus"
)

// NodeConfigurePost type for handling post requests made by node
type NodeConfigurePost struct {
	EnrollSecret   string `json:"enroll_secret"`
	NodeKey        string `json:"node_key"`
	HostIdentifier string `json:"host_identifier"`
}

func RandomString(strlen int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	var result []string
	for i := 0; i < strlen; i++ {
		result = append(result, string(chars[rand.Intn(len(chars))]))
	}
	return strings.Join(result, "")
}

func NodeEnrollRequest(respWriter http.ResponseWriter, request *http.Request) {
	dump, _ := httputil.DumpRequest(request, true)
	logger.Debug(string(dump))
	dynSvc := dyndb.DbInstance()

	mu := sync.Mutex{}

	respWriter.Header().Set("Content-Type", "application/json")
	//test if enrol secret is correct
	dump, err := httputil.DumpRequest(request, true)
	logger.Info(string(dump))
	if err != nil {
		logger.Error(string(dump))
		logger.Error(err)
		return
	}
	sekret, err := auth.GetNodeSecret()
	if err != nil {
		logger.Error(err)
	}
	if len(sekret) <= 3 {
		logger.Warn("Node secret too short, exiting.")
		respWriter.Write([]byte(fmt.Sprintf("actual secret: %s", sekret)))
		return
	}
	//check if enroll secret is accurate
	//if enroll secret is correct check if hostname registered
	//if hostname registered, send config
	//if not, send back to pending registration
	type EnrollRequest struct {
		EnrollSecret   string                       `json:"enroll_secret"`
		NodeKey        string                       `json:"node_key"`
		HostIdentifier string                       `json:"host_identifier"`
		PlatformType   string                       `json:"platform_type"`
		HostDetails    map[string]map[string]string `json:"host_details"`
	}
	type EnrollRequestResponse struct {
		NodeKey     string `json:"node_key"`
		NodeInvalid bool   `json:"node_invalid"`
	}
	//fmt.Println(request.Body)
	nodeEnrollRequestLogger := logger.WithFields(log.Fields{
		"node_ip_address": request.RemoteAddr,
		"user_agent":      request.UserAgent(),
	})
	body, err := ioutil.ReadAll(request.Body)
	logger.Info(string(body))
	defer request.Body.Close()
	if err != nil {
		nodeEnrollRequestLogger.Error(err)
	}
	data := EnrollRequest{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		nodeEnrollRequestLogger.Error(err)
	}
	if string(data.EnrollSecret) == sekret {
		nodeEnrollRequestLogger.WithFields(log.Fields{
			"hostname": data.HostIdentifier,
		}).Info("Correct sekret received")
		//Need more error handling here.  what if node key is valid but hostname is duplicate?
		if data.NodeKey == "" {
			ans, err := dyndb.SearchByHostIdentifier(data.HostIdentifier, dynSvc)
			if err != nil {
				logger.Error(err)
				return
			}
			if len(ans) > 0 {
				nodeEnrollRequestLogger.WithFields(log.Fields{
					"hostname": data.HostIdentifier,
				}).Info("host already exists, setting host to existing node_key")
				enrollRequestResponse := EnrollRequestResponse{ans[0].NodeKey, false}
				osc := ans[0]
				osc.Timestamp()
				err := dyndb.UpsertClient(osc, dynSvc, &mu)
				if err != nil {
					nodeEnrollRequestLogger.Error(err)
					return
				}
				// might be good to check for dupe hostnames here before ACTUALLY issuing new key
				js, _ := json.Marshal(enrollRequestResponse)
				respWriter.Write(js)
			} else {
				// this will trigger a new enrollment request.  Upon creation of this request, it might be
				//a good idea to generate a post event to an endpoint to notify of a pending enrollment
				//approval
				// this could also be taken care of by a separate worker that sweeps for pending enrollments.
				// work may be better, as this task can be assigned to a lambda and gives more flexibility to the
				//post endpoint
				nodeKey := RandomString(20)
				nodeEnrollRequestLogger.WithFields(log.Fields{
					"hostname": data.HostIdentifier,
				}).Info("generating new  node_key")
				enrollRequestResponse := EnrollRequestResponse{nodeKey, false}
				// Handle enrollment defaults here.  Default configs for widerps, osux, Linux
				osc := osquery_types.OsqueryClient{
					HostIdentifier:              data.HostIdentifier,
					NodeKey:                     nodeKey,
					NodeInvalid:                 false,
					HostDetails:                 data.HostDetails,
					PendingRegistrationApproval: true,
					Tags:               []string{},
					ConfigName:         "default",
					ConfigurationGroup: "",
				}
				osc.Timestamp()
				// might be good to check for dupe hostnames here before ACTUALLY issuing new key
				err := dyndb.UpsertClient(osc, dynSvc, &mu)
				if err != nil {
					nodeEnrollRequestLogger.WithFields(log.Fields{
						"hostname": data.HostIdentifier,
					}).Info("failed to upsert  node")
				}

				js, _ := json.Marshal(enrollRequestResponse)
				respWriter.Write(js)
				return

			}
		}
		// if sekret is correct, check if node_key already configured
		// if NOT node key, generate node key and create upsert to add node_key to pending
		// if node key INVALID, return node_invalid
		// if node key configured, check
	} else {
		respWriter.Header().Set("Content-Type", "application/json")
		respWriter.Write([]byte(`{"node_invalid": true}`))
		return
	}
	//fmt.Println(string(dump))
}

func NodeConfigureRequest(respWriter http.ResponseWriter, request *http.Request) {
	dump, _ := httputil.DumpRequest(request, true)
	logger.Debug(string(dump))
	dynSvc := dyndb.DbInstance()
	//to recieve a valid config, node must have both a valid sekret and
	//a node_key that is valid
	body, err := ioutil.ReadAll(request.Body)
	defer request.Body.Close()
	logger := logger.WithFields(log.Fields{
		"node_ip_address": request.RemoteAddr,
		"user_agent":      request.UserAgent(),
	})
	if err != nil {
		logger.Error(err)
	}
	var data NodeConfigurePost
	// unmarshal post data into data
	err = json.Unmarshal(body, &data)
	if err != nil {
		logger.Warn("unmarshal error")
	}

	err = dyndb.ValidNode(data.NodeKey, dynSvc)
	if err != nil {
		logger.Error(err)
		respWriter.Header().Set("Content-Type", "application/json")
		respWriter.Write([]byte(`{"node_invalid": true}`))
		return
	}

	//query dyndb for node state with key
	logger.WithFields(log.Fields{
		"hostname": data.HostIdentifier,
		"node_key": data.NodeKey,
	}).Debug("valid node")
	//get type of config for endpoint, return config
	osqNode, err := dyndb.SearchByNodeKey(data.NodeKey, dynSvc)
	osqNode.Timestamp()
	if err != nil {
		logger.Panic(err)
	}
	mu := sync.Mutex{}
	err = dyndb.UpsertClient(osqNode, dynSvc, &mu)
	if err != nil {
		logger.Error(err)
		return
	}
	namedConfig := osquery_types.OsqueryNamedConfig{}
	if len(osqNode.ConfigName) > 0 {
		namedConfig, err = dyndb.GetNamedConfig(dynSvc, osqNode.ConfigName)
		if err != nil {
			logger.Panicf("[node.go]: Error returned when getting Named config: \n %s", err)
		}
	} else {
		logger.Info("No named config found, setting default config")
		namedConfig, err = dyndb.GetNamedConfig(dynSvc, "default")
		if err != nil {
			logger.Panic(err)
		}
	}
	config, err := osquery_types.GetServerConfig("config.json")
	if err != nil {
		logger.Error(err)
		return
	}
	//namedConfig.OsqueryConfig.Options.AwsAccessKeyID = os.Getenv("FIREHOSE_AWS_ACCESS_KEY_ID")
	//namedConfig.OsqueryConfig.Options.AwsSecretAccessKey = os.Getenv("FIREHOSE_AWS_SECRET_ACCESS_KEY")
	//namedConfig.OsqueryConfig.Options.AwsFirehoseStream = os.Getenv("AWS_FIREHOSE_STREAM")
	logger.Debug(config)
	namedConfig.OsqueryConfig.Options.AwsAccessKeyID = config.FirehoseAWSAccessKeyID
	namedConfig.OsqueryConfig.Options.AwsSecretAccessKey = config.FirehoseAWSSecretAccessKey
	if namedConfig.OsqueryConfig.Options.AwsFirehoseStream == "" {
		namedConfig.OsqueryConfig.Options.AwsFirehoseStream = config.FirehoseStreamName
	}
	rawPackJSON := dyndb.BuildOsqueryPacksAsJSON(namedConfig)
	namedConfig.OsqueryConfig.Packs = &rawPackJSON
	js, err := json.Marshal(namedConfig.OsqueryConfig)
	if err != nil {
		logger.Error(err)
	}

	respWriter.Header().Set("Content-Type", "application/json")
	respWriter.Write(js)
}
