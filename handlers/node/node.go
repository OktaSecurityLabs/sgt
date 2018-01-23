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

type NodeConfigurePost struct {
	Enroll_secret   string `json:"enroll_secret"`
	Node_key        string `json:"node_key"`
	Host_identifier string `json:"host_identifier"`
}

func RandomString(strlen int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	var result []string
	for i := 0; i < strlen; i++ {
		result = append(result, string(chars[rand.Intn(len(chars))]))
	}
	return strings.Join(result, "")
}

func NodeEnrollRequest(respwritter http.ResponseWriter, request *http.Request) {
	dump, _ := httputil.DumpRequest(request, true)
	logger.Debug(string(dump))
	dyn_svc := dyndb.DbInstance()
	var (
		mut sync.Mutex
	)
	respwritter.Header().Set("Content-Type", "application/json")
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
		respwritter.Write([]byte(fmt.Sprintf("actual secret: %s", sekret)))
		return
	}
	//check if enroll secret is accurate
	//if enroll secret is correct check if hostname registered
	//if hostname registered, send config
	//if not, send back to pending registration
	type EnrollRequest struct {
		Enroll_secret   string                       `json:"enroll_secret"`
		Node_key        string                       `json:"node_key"`
		Host_identifier string                       `json:"host_identifier"`
		PlatformType    string                       `json:"platform_type"`
		HostDetails     map[string]map[string]string `json:"host_details"`
	}
	type EnrollRequestResponse struct {
		Node_key     string `json:"node_key"`
		Node_invalid bool   `json:"node_invalid"`
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
	if string(data.Enroll_secret) == sekret {
		nodeEnrollRequestLogger.WithFields(log.Fields{
			"hostname": data.Host_identifier,
		}).Info("Correct sekret received")
		//Need more error handling here.  what if node key is valid but hostname is duplicate?
		if data.Node_key == "" {
			ans, err := dyndb.SearchByHostIdentifier(data.Host_identifier, dyn_svc)
			if err != nil {
				logger.Error(err)
				return
			}
			if len(ans) > 0 {
				nodeEnrollRequestLogger.WithFields(log.Fields{
					"hostname": data.Host_identifier,
				}).Info("host already exists, setting host to existing node_key")
				enroll_request_response := EnrollRequestResponse{ans[0].Node_key, false}
				osc := ans[0]
				osc.Timestamp()
				err := dyndb.UpsertClient(osc, dyn_svc, mut)
				if err != nil {
					nodeEnrollRequestLogger.Error(err)
					return
				}
				// might be good to check for dupe hostnames here before ACTUALLY issuing new key
				js, _ := json.Marshal(enroll_request_response)
				respwritter.Write(js)
			} else {
				// this will trigger a new enrollment request.  Upon creation of this request, it might be
				//a good idea to generate a post event to an endpoint to notify of a pending enrollment
				//approval
				// this could also be taken care of by a separate worker that sweeps for pending enrollments.
				// work may be better, as this task can be assigned to a lambda and gives more flexibility to the
				//post endpoint
				node_key := RandomString(20)
				nodeEnrollRequestLogger.WithFields(log.Fields{
					"hostname": data.Host_identifier,
				}).Info("generating new  node_key")
				enroll_request_response := EnrollRequestResponse{node_key, false}
				// Handle enrollment defaults here.  Default configs for widerps, osux, Linux
				osc := osquery_types.OsqueryClient{
					Host_identifier:               data.Host_identifier,
					Node_key:                      node_key,
					Node_invalid:                  false,
					HostDetails:                   data.HostDetails,
					Pending_registration_approval: true,
					Tags:                []string{},
					Config_name:         "default",
					Configuration_group: "",
				}
				osc.Timestamp()
				// might be good to check for dupe hostnames here before ACTUALLY issuing new key
				err := dyndb.UpsertClient(osc, dyn_svc, mut)
				if err != nil {
					nodeEnrollRequestLogger.WithFields(log.Fields{
						"hostname": data.Host_identifier,
					}).Info("failed to upsert  node")
				}
				//respwritter.Write([]byte(`{"node_invalid": true, "node_key": nodekey}`))
				js, _ := json.Marshal(enroll_request_response)
				respwritter.Write(js)
				return

			}
		}
		// if sekret is correct, check if node_key already configured
		// if NOT node key, generate node key and create upsert to add node_key to pending
		// if node key INVALID, return node_invalid
		// if node key configured, check
	} else {
		respwritter.Header().Set("Content-Type", "application/json")
		respwritter.Write([]byte(`{"node_invalid": true}`))
		return
	}
	//fmt.Println(string(dump))
}

func NodeConfigureRequest(respwritter http.ResponseWriter, request *http.Request) {
	dump, _ := httputil.DumpRequest(request, true)
	logger.Debug(string(dump))
	dyn_svc := dyndb.DbInstance()
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
	valid_node, err := dyndb.ValidNode(data.Node_key, dyn_svc)
	if err != nil {
		logger.Error(err)
		return
	}
	//query dyndb for node state with key
	if valid_node {
		logger.WithFields(log.Fields{
			"hostname": data.Host_identifier,
			"node_key": data.Node_key,
		}).Debug("valid node")
		//get type of config for endpoint, return config
		osq_node, err := dyndb.SearchByNodeKey(data.Node_key, dyn_svc)
		osq_node.Timestamp()
		if err != nil {
			logger.Panic(err)
		}
		mu := sync.Mutex{}
		err = dyndb.UpsertClient(osq_node, dyn_svc, mu)
		if err != nil {
			logger.Error(err)
			return
		}
		named_config := osquery_types.OsqueryNamedConfig{}
		if len(osq_node.Config_name) > 0 {
			named_config, err = dyndb.GetNamedConfig(dyn_svc, osq_node.Config_name)
			if err != nil {
				logger.Panicf("[node.go]: Error returned when getting Named config: \n %s", err)
			}
		} else {
			logger.Info("No named config found, setting default config")
			named_config, err = dyndb.GetNamedConfig(dyn_svc, "default")
			if err != nil {
				logger.Panic(err)
			}
		}
		config, err := osquery_types.GetServerConfig("config.json")
		if err != nil {
			logger.Error(err)
			return
		}
		//named_config.Osquery_config.Options.Aws_access_key_id = os.Getenv("FIREHOSE_AWS_ACCESS_KEY_ID")
		//named_config.Osquery_config.Options.Aws_secret_access_key = os.Getenv("FIREHOSE_AWS_SECRET_ACCESS_KEY")
		//named_config.Osquery_config.Options.Aws_firehose_stream = os.Getenv("AWS_FIREHOSE_STREAM")
		logger.Debug(config)
		named_config.Osquery_config.Options.Aws_access_key_id = config.FirehoseAWSAccessKeyID
		named_config.Osquery_config.Options.Aws_secret_access_key = config.FirehoseAWSSecretAccessKey
		if named_config.Osquery_config.Options.Aws_firehose_stream == "" {
			named_config.Osquery_config.Options.Aws_firehose_stream = config.FirehoseStreamName
		}
		raw_pack_json := dyndb.BuildOsqueryPacksAsJSON(named_config)
		named_config.Osquery_config.Packs = &raw_pack_json
		js, err := json.Marshal(named_config.Osquery_config)
		if err != nil {
			logger.Error(err)
		}
		respwritter.Header().Set("Content-Type", "application/json")
		respwritter.Write(js)
		return
	} else {
		respwritter.Header().Set("Content-Type", "application/json")
		respwritter.Write([]byte(`{"node_invalid": true}`))
		return
	}
}
