package node

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"

	"github.com/oktasecuritylabs/sgt/handlers/auth"
	"github.com/oktasecuritylabs/sgt/handlers/response"
	"github.com/oktasecuritylabs/sgt/logger"
	"github.com/oktasecuritylabs/sgt/osquery_types"
	log "github.com/sirupsen/logrus"
	"strconv"
)

type NodeDB interface {
	SearchByHostIdentifier(hid string) ([]osquery_types.OsqueryClient, error)
	UpsertClient(oc osquery_types.OsqueryClient) error
	ValidNode(nodeKey string) error
	SearchByNodeKey(nk string) (osquery_types.OsqueryClient, error)
	GetNamedConfig(configName string) (osquery_types.OsqueryNamedConfig, error)
	//BuildOsqueryPackAsJSON(nc osquery_types.OsqueryNamedConfig) (json.RawMessage)
	BuildNamedConfig(configName string) (osquery_types.OsqueryNamedConfig, error)
}

const (
	nodeInvalid = true
	nodeValid   = false
)

type EnrollRequestResponse struct {
	NodeKey     string `json:"node_key"`
	NodeInvalid bool   `json:"node_invalid"`
}

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
func NodeEnrollRequest(dyn NodeDB, config *osquery_types.ServerConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleRequest := func() error {

			//test if enrol secret is correct

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

			nodeEnrollRequestLogger := logger.WithFields(log.Fields{
				"node_ip_address": r.RemoteAddr,
				"user_agent":      r.UserAgent(),
			})

			body, err := ioutil.ReadAll(r.Body)
			defer r.Body.Close()
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

			if data.NodeKey != "" {
				return fmt.Errorf("node key '%s' already exists in data", data.NodeKey)
			}

			//Need more error handling here.  what if node key is valid but hostname is duplicate?
			ans, err := dyn.SearchByHostIdentifier(data.HostIdentifier)
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
				autoApprove, err := strconv.ParseBool(config.AutoApproveNodes)
				if err != nil {
					return err
				}
				osc := osquery_types.OsqueryClient{
					ConfigName:     "default",
					HostDetails:    data.HostDetails,
					HostIdentifier: data.HostIdentifier,
					HostName:       data.HostDetails["system_info"]["computer_name"],
					NodeKey:        nodeKey,
					//if autoenroll enabled, set pending to false
					PendingRegistrationApproval: !autoApprove,
				}

				osc.SetTimestamp()
				// might be good to check for dupe hostnames here before ACTUALLY issuing new key
				err = dyn.UpsertClient(osc)
				if err != nil {
					nodeEnrollRequestLogger.WithFields(log.Fields{
						"hostname": data.HostIdentifier,
					}).Info("failed to upsert node")
					return fmt.Errorf("node upsert failed: %s", err)
				}

				//return invalid node response to client
				response.WriteCustomJSON(w, EnrollRequestResponse{NodeKey: nodeKey, NodeInvalid: nodeInvalid})
			default:
				nodeEnrollRequestLogger.WithFields(log.Fields{
					"hostname": data.HostIdentifier,
				}).Info("host already exists, setting host to existing node_key")
				osc := ans[0]
				osc.SetTimestamp()
				err := dyn.UpsertClient(osc)
				if err != nil {
					nodeEnrollRequestLogger.Error(err)
					return fmt.Errorf("node upsert failed: %s", err)
				}
				// might be good to check for dupe hostnames here before ACTUALLY issuing new key
				//return a valid node response to client
				response.WriteCustomJSON(w, EnrollRequestResponse{NodeKey: ans[0].NodeKey, NodeInvalid: nodeValid})
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
			logger.Error(errString)
			//response.WriteError(respWriter, errString)
		}

	})
}

// NodeConfigureRequest configures a node.  Returns a json body of either a full osquery config, or a node_invalide = True to
// indicate need for re-enrollment
func NodeConfigureRequest(dyn NodeDB, config *osquery_types.ServerConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerLogger := logger.WithFields(log.Fields{
			"node_ip_address": r.RemoteAddr,
			"user_agent":      r.UserAgent(),
		})

		handleRequest := func() (interface{}, error) {
			//to recieve a valid config, node must have both a valid sekret and
			//a node_key that is valid
			body, err := ioutil.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				return nil, fmt.Errorf("failed to read request body: %s", err)
			}

			var data NodeConfigurePost
			// unmarshal post data into data
			// if invalid json, return invalid json error
			err = json.Unmarshal(body, &data)
			if err != nil {
				logger.Error(err)
				err = errors.New("Invalid Json body")
				return nil, err
			}

			//dynSvc := dyndb.DbInstance()
			// if node invalid, return invalid_node -> true
			err = dyn.ValidNode(data.NodeKey)
			if err != nil {

				return nil, fmt.Errorf("node validation failed for node with key '%s': %s", data.NodeKey, err)
			}

			//query dyndb for node state with key
			logger.WithFields(log.Fields{
				"hostname": data.HostIdentifier,
				"node_key": data.NodeKey,
			}).Debug("valid node")

			//get type of config for endpoint, return config
			osqNode, err := dyn.SearchByNodeKey(data.NodeKey)
			if err != nil {
				return nil, fmt.Errorf("could not find node with key '%s': %s", data.NodeKey, err)
			}

			osqNode.SetTimestamp()
			err = dyn.UpsertClient(osqNode)
			if err != nil {
				return nil, fmt.Errorf("node upsert failed: %s", err)
			}

			namedConfig := osquery_types.OsqueryNamedConfig{}
			if osqNode.ConfigName != "" {
				namedConfig, err = dyn.BuildNamedConfig(osqNode.ConfigName)
				if err != nil {
					return nil, fmt.Errorf("could not get config with name '%s': \n %s", osqNode.ConfigName, err)
				}
			} else {
				handlerLogger.Info("No named config found, setting default config")
				namedConfig, err = dyn.BuildNamedConfig("default")
				if err != nil {
					return nil, fmt.Errorf("could not get default config: %s", err)
				}
			}

			//config, err := osquery_types.GetServerConfig("config.json")
			//if err != nil {
			//return nil, fmt.Errorf("could not get server config: %s", err)
			//}
			//oc, err := dyn.BuildNamedConfig(osqNode.ConfigName)
			//logger.Infof("OC: %+v", oc)
			//if err != nil {
			//logger.Error(err)
			//}

			namedConfig.OsqueryConfig.Options.AwsAccessKeyID = config.FirehoseAWSAccessKeyID
			namedConfig.OsqueryConfig.Options.AwsSecretAccessKey = config.FirehoseAWSSecretAccessKey
			if namedConfig.OsqueryConfig.Options.AwsFirehoseStream == "" {
				namedConfig.OsqueryConfig.Options.AwsFirehoseStream = config.FirehoseStreamName
			}

			//namedConfig.OsqueryConfig = oc
			return namedConfig.OsqueryConfig, nil
		}

		result, err := handleRequest()
		if err != nil {
			handlerLogger.Error(err)
			errString := fmt.Sprintf("[NodeConfigureRequest] node configuration failed: %s", err)
			logger.Error(errString)
			result := EnrollRequestResponse{NodeInvalid: nodeInvalid}
			response.WriteCustomJSON(w, result)
		} else {
			response.WriteCustomJSON(w, result)
		}

	})
}
