package node

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/oktasecuritylabs/sgt/dyndb"
	"github.com/oktasecuritylabs/sgt/handlers/auth"
	"github.com/oktasecuritylabs/sgt/handlers/response"
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

// NodeEnrollRequest enrolls a node given the host identifier
func NodeEnrollRequest(respWriter http.ResponseWriter, request *http.Request) {

	handleRequest := func() error {

		//test if enrol secret is correct
		dump, err := httputil.DumpRequest(request, true)
		logger.Info(string(dump))
		if err != nil {
			return fmt.Errorf("could not dump request: %s", err)
		}

		sekret, err := auth.GetNodeSecret()
		if err != nil {
			return fmt.Errorf("could not get node secret: %s", err)
		}

		if len(sekret) <= 3 {
			return fmt.Errorf("node secret too short: %s", sekret)
		}

		//check if enroll secret is accurate
		//if enroll secret is correct check if hostname registered
		//if hostname registered, send config
		//if not, send back to pending registration
		type EnrollRequest struct {
			NodeConfigurePost
			PlatformType string                       `json:"platform_type"`
			HostDetails  map[string]map[string]string `json:"host_details"`
		}
		type EnrollRequestResponse struct {
			NodeKey     string `json:"node_key"`
			NodeInvalid bool   `json:"node_invalid"`
		}

		nodeEnrollRequestLogger := logger.WithFields(log.Fields{
			"node_ip_address": request.RemoteAddr,
			"user_agent":      request.UserAgent(),
		})

		body, err := ioutil.ReadAll(request.Body)
		logger.Info(string(body))
		defer request.Body.Close()
		if err != nil {
			nodeEnrollRequestLogger.Error(err)
			return fmt.Errorf("failed to read request body: %s", err)
		}

		data := EnrollRequest{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			nodeEnrollRequestLogger.Error(err)
			return fmt.Errorf("unmarshal failed: %s", err)
		}

		if string(data.EnrollSecret) != sekret {
			return errors.New("node secret does not match enroll secret")
		}

		nodeEnrollRequestLogger.WithFields(log.Fields{
			"hostname": data.HostIdentifier,
		}).Info("Correct sekret received")

		dynSvc := dyndb.DbInstance()

		if data.NodeKey != "" {
			return fmt.Errorf("node key '%s' already exists in data", data.NodeKey)
		}

		//Need more error handling here.  what if node key is valid but hostname is duplicate?
		ans, err := dyndb.SearchByHostIdentifier(data.HostIdentifier, dynSvc)
		if err != nil {
			return fmt.Errorf("failed to get node for host identifier '%s': %s", data.HostIdentifier, err)
		}

		switch len(ans) {
		case 0:
			// this will trigger a new enrollment request.  Upon creation of this request, it might be
			//a good idea to generate a post event to an endpoint to notify of a pending enrollment
			//approval
			// this could also be taken care of by a separate worker that sweeps for pending enrollments.
			// work may be better, as this task can be assigned to a lambda and gives more flexibility to the
			//post endpoint
			nodeKey := RandomString(20)
			nodeEnrollRequestLogger.WithFields(log.Fields{
				"hostname": data.HostIdentifier,
			}).Info("generating new node_key")

			// Handle enrollment defaults here.  Default configs for widerps, osux, Linux
			osc := osquery_types.OsqueryClient{
				ConfigName:                  "default",
				HostDetails:                 data.HostDetails,
				HostIdentifier:              data.HostIdentifier,
				NodeKey:                     nodeKey,
				PendingRegistrationApproval: true,
			}

			osc.SetTimestamp()
			// might be good to check for dupe hostnames here before ACTUALLY issuing new key
			err := dyndb.UpsertClient(osc, dynSvc)
			if err != nil {
				nodeEnrollRequestLogger.WithFields(log.Fields{
					"hostname": data.HostIdentifier,
				}).Info("failed to upsert node")
				return fmt.Errorf("node upsert failed: %s", err)
			}

			response.WriteCustomJSON(respWriter, EnrollRequestResponse{NodeKey: nodeKey, NodeInvalid: false})
		default:
			nodeEnrollRequestLogger.WithFields(log.Fields{
				"hostname": data.HostIdentifier,
			}).Info("host already exists, setting host to existing node_key")
			osc := ans[0]
			osc.SetTimestamp()
			err := dyndb.UpsertClient(osc, dynSvc)
			if err != nil {
				nodeEnrollRequestLogger.Error(err)
				return fmt.Errorf("node upsert failed: %s", err)
			}
			// might be good to check for dupe hostnames here before ACTUALLY issuing new key
			response.WriteCustomJSON(respWriter, EnrollRequestResponse{NodeKey: ans[0].NodeKey, NodeInvalid: false})
		}

		// TODO:
		// if sekret is correct, check if node_key already configured
		// if NOT node key, generate node key and create upsert to add node_key to pending
		// if node key INVALID, return node_invalid
		// if node key configured, check

		return nil
	}

	err := handleRequest()
	if err != nil {
		logger.Error(err)
		errString := fmt.Sprintf("[NodeEnrollRequest] node enrolling failed: %s", err)
		response.WriteError(respWriter, errString)
	}
}

// NodeConfigureRequest configures a node
func NodeConfigureRequest(respWriter http.ResponseWriter, request *http.Request) {

	handlerLogger := logger.WithFields(log.Fields{
		"node_ip_address": request.RemoteAddr,
		"user_agent":      request.UserAgent(),
	})

	handleRequest := func() (interface{}, error) {

		dump, _ := httputil.DumpRequest(request, true)
		logger.Debug(string(dump))

		//to recieve a valid config, node must have both a valid sekret and
		//a node_key that is valid
		body, err := ioutil.ReadAll(request.Body)
		defer request.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %s", err)
		}

		var data NodeConfigurePost
		// unmarshal post data into data
		err = json.Unmarshal(body, &data)
		if err != nil {
			return nil, fmt.Errorf("unmarshal failed: %s", err)
		}

		dynSvc := dyndb.DbInstance()
		err = dyndb.ValidNode(data.NodeKey, dynSvc)
		if err != nil {
			return nil, fmt.Errorf("node validation failed for node with key '%s': %s", data.NodeKey, err)
		}

		//query dyndb for node state with key
		logger.WithFields(log.Fields{
			"hostname": data.HostIdentifier,
			"node_key": data.NodeKey,
		}).Debug("valid node")

		//get type of config for endpoint, return config
		osqNode, err := dyndb.SearchByNodeKey(data.NodeKey, dynSvc)
		if err != nil {
			return nil, fmt.Errorf("could not find node with key '%s': %s", data.NodeKey, err)
		}

		osqNode.SetTimestamp()
		err = dyndb.UpsertClient(osqNode, dynSvc)
		if err != nil {
			return nil, fmt.Errorf("node upsert failed: %s", err)
		}

		namedConfig := osquery_types.OsqueryNamedConfig{}
		if osqNode.ConfigName != "" {
			namedConfig, err = dyndb.GetNamedConfig(dynSvc, osqNode.ConfigName)
			if err != nil {
				return nil, fmt.Errorf("could not get config with name '%s': \n %s", osqNode.ConfigName, err)
			}
		} else {
			handlerLogger.Info("No named config found, setting default config")
			namedConfig, err = dyndb.GetNamedConfig(dynSvc, "default")
			if err != nil {
				return nil, fmt.Errorf("could not get default config: %s", err)
			}
		}

		config, err := osquery_types.GetServerConfig("config.json")
		if err != nil {
			return nil, fmt.Errorf("could not get server config: %s", err)
		}
		//namedConfig.OsqueryConfig.Options.AwsAccessKeyID = os.Getenv("FIREHOSE_AWS_ACCESS_KEY_ID")
		//namedConfig.OsqueryConfig.Options.AwsSecretAccessKey = os.Getenv("FIREHOSE_AWS_SECRET_ACCESS_KEY")
		//namedConfig.OsqueryConfig.Options.AwsFirehoseStream = os.Getenv("AWS_FIREHOSE_STREAM")
		handlerLogger.Debug(config)
		namedConfig.OsqueryConfig.Options.AwsAccessKeyID = config.FirehoseAWSAccessKeyID
		namedConfig.OsqueryConfig.Options.AwsSecretAccessKey = config.FirehoseAWSSecretAccessKey
		if namedConfig.OsqueryConfig.Options.AwsFirehoseStream == "" {
			namedConfig.OsqueryConfig.Options.AwsFirehoseStream = config.FirehoseStreamName
		}
		rawPackJSON := dyndb.BuildOsqueryPacksAsJSON(namedConfig)
		namedConfig.OsqueryConfig.Packs = &rawPackJSON

		return namedConfig.OsqueryConfig, nil
	}

	result, err := handleRequest()
	if err != nil {
		handlerLogger.Error(err)
		errString := fmt.Sprintf("[NodeConfigureRequest] node configuration failed: %s", err)
		response.WriteError(respWriter, errString)
	} else {
		response.WriteCustomJSON(respWriter, result)
	}
}
