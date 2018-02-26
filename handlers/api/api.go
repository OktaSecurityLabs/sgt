package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/oktasecuritylabs/sgt/dyndb"
	"github.com/oktasecuritylabs/sgt/handlers/response"
	"github.com/oktasecuritylabs/sgt/logger"
	"github.com/oktasecuritylabs/sgt/osquery_types"
	log "github.com/sirupsen/logrus"
)

func init() {
	//logger.SetFormatter(&logger.JSONFormatter{/I//})
	logger.WithFields(log.Fields{
		"Module": "API",
	})
}

//GetNamedConfigs returns all named configs in a json list
func GetNamedConfigs(respWriter http.ResponseWriter, r *http.Request) {

	handleRequest := func() (interface{}, error) {

		dynDBInstance := dyndb.DbInstance()
		ans, err := dyndb.GetNamedConfigs(dynDBInstance)
		if err != nil {
			return nil, fmt.Errorf("could not get named configs: %s", err)
		}

		return ans, nil
	}

	result, err := handleRequest()
	if err != nil {
		logger.Error(err)
		errString := fmt.Sprintf("[GetNamedConfigs] failed to get named configs: %s", err)
		response.WriteError(respWriter, errString)
	} else {
		response.WriteCustomJSON(respWriter, result)
	}
}

// ConfigurationRequest accepts a json post body of a NamedConfig
func ConfigurationRequest(respWriter http.ResponseWriter, request *http.Request) {

	handleRequest := func() (interface{}, error) {

		vars := mux.Vars(request)
		configName, ok := vars["config_name"]
		if !ok || configName == "" {
			return nil, errors.New("no config name specified")
		}

		// get the named config
		dynDBInstance := dyndb.DbInstance()
		existingNamedConfig, err := dyndb.GetNamedConfig(dynDBInstance, configName)
		if err != nil {
			return nil, fmt.Errorf("failed to get config with name [%s]: %s", configName, err)
		}

		switch request.Method {
		case http.MethodGet:

			return existingNamedConfig, nil

		case http.MethodPost:

			//now merge what's already in teh database with our defaults
			existingNamedConfig.OsqueryConfig.Options = osquery_types.NewOsqueryOptions()

			// finally...
			body, err := ioutil.ReadAll(request.Body)
			defer request.Body.Close()
			if err != nil {
				return nil, fmt.Errorf("failed to read request body: %s", err)
			}

			//overlay default + existing with options provided by user
			err = json.Unmarshal(body, &existingNamedConfig)
			if err != nil {
				return nil, fmt.Errorf("merging of named configs failed: %s", err)
			}

			if configName != existingNamedConfig.ConfigName {
				return nil, errors.New("named config endpoint does not match posted data config_name")
			}

			err = dyndb.UpsertNamedConfig(dynDBInstance, &existingNamedConfig)
			if err != nil {
				return nil, fmt.Errorf("dynamo named config upsert failed: %s", err)
			}

			return existingNamedConfig, nil
		}

		return nil, fmt.Errorf("method not supported: %s", request.Method)
	}

	result, err := handleRequest()
	if err != nil {
		logger.Error(err)
		errString := fmt.Sprintf("[ConfigurationRequest] failed to handle named config in %s request: %s", request.Method, err)
		response.WriteError(respWriter, errString)
	} else {
		response.WriteCustomJSON(respWriter, result)
	}
}

// GetNodes returns json reponse of a list of nodes
func GetNodes(respWriter http.ResponseWriter, request *http.Request) {
	// Only handle GET requests
	if request.Method != http.MethodGet {
		return
	}

	handleRequest := func() (interface{}, error) {

		results, err := dyndb.SearchByHostIdentifier("", dyndb.DbInstance())
		if err != nil {
			return nil, fmt.Errorf("failed to get all nodes: %s", err)
		}

		return results, nil
	}

	result, err := handleRequest()
	if err != nil {
		logger.Error(err)
		errString := fmt.Sprintf("[GetNodes] %s", err)
		response.WriteError(respWriter, errString)
	} else {
		response.WriteCustomJSON(respWriter, result)
	}
}

// ConfigureNode accepts json node configuration
func ConfigureNode(respWriter http.ResponseWriter, request *http.Request) {

	handleRequest := func() (interface{}, error) {

		vars := mux.Vars(request)
		nodeKey, ok := vars["node_key"]
		if !ok || nodeKey == "" {
			return nil, errors.New("request does not contain node_key")
		}

		dynDBInstance := dyndb.DbInstance()
		existingClient, err := dyndb.SearchByNodeKey(nodeKey, dynDBInstance)
		if err != nil {
			return nil, fmt.Errorf("failed to find node by key [%s]: %s", nodeKey, err)
		}

		logger.Infof("existing client: %+v", existingClient)

		switch request.Method {
		case http.MethodGet:

			return existingClient, nil

		case http.MethodPost:

			if existingClient.NodeKey == "" {
				return nil, errors.New("existing client node_key is empty")
			}

			body, err := ioutil.ReadAll(request.Body)
			defer request.Body.Close()
			if err != nil {
				return nil, fmt.Errorf("failed to read request body: %s", err)
			}

			//set posted osquery client = client
			client := osquery_types.OsqueryClient{}
			err = json.Unmarshal(body, &client)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal request body [%s]: %s", string(body), err)
			}

			logger.Infof("new client: %+v", client)

			//map missing keys to client
			client.NodeKey = nodeKey
			if client.ConfigName == "" {
				client.ConfigName = existingClient.ConfigName
			}
			client.HostIdentifier = existingClient.HostIdentifier
			client.NodeInvalid = client.NodeInvalid || existingClient.NodeInvalid
			client.PendingRegistrationApproval = client.PendingRegistrationApproval && existingClient.PendingRegistrationApproval
			client.HostDetails = existingClient.HostDetails

			if len(client.Tags) == 0 {
				client.Tags = existingClient.Tags
			}

			err = dyndb.UpsertClient(client, dynDBInstance)
			if err != nil {
				return nil, fmt.Errorf("client update in dynamo failed: %s", err)
			}

			return client, nil
		}

		return nil, fmt.Errorf("method not supported: %s", request.Method)
	}

	result, err := handleRequest()
	if err != nil {
		logger.Error(err)
		errString := fmt.Sprintf("[ConfigureNode] failed to configure node in %s request: %s", request.Method, err)
		response.WriteError(respWriter, errString)
	} else {
		response.WriteCustomJSON(respWriter, result)
	}
}

// ApproveNode helper function as a shortcut to approving node.  Takes no body input
func ApproveNode(respWriter http.ResponseWriter, request *http.Request) {

	// Only handle POST requests
	if request.Method != http.MethodPost {
		return
	}

	handleRequest := func() error {

		vars := mux.Vars(request)
		nodeKey, ok := vars["node_key"]
		if !ok || nodeKey == "" {
			return errors.New("request does not contain node_key")
		}

		logger.Warn("posting approval")

		dynDBInstance := dyndb.DbInstance()
		err := dyndb.ApprovePendingNode(nodeKey, dynDBInstance)
		if err != nil {
			return fmt.Errorf("approval of pending node failed: %s", err)
		}

		return nil
	}

	err := handleRequest()
	if err != nil {
		logger.Error(err)
		errString := fmt.Sprintf("[ApproveNode] failed to approve node: %s", err)
		response.WriteError(respWriter, errString)
	} else {
		response.WriteSuccess(respWriter, "")
	}
}

// GetPackQueries returns json response of a list of packqueries
func GetPackQueries(respWriter http.ResponseWriter, request *http.Request) {

	// Only handle GET requests
	if request.Method != http.MethodGet {
		return
	}

	handleRequest := func() (interface{}, error) {

		dynDBInstance := dyndb.DbInstance()
		results, err := dyndb.APIGetPackQueries(dynDBInstance)
		if err != nil {
			return nil, fmt.Errorf("failed to get query packs: %s", err)
		}

		return results, nil
	}

	result, err := handleRequest()
	if err != nil {
		logger.Error(err)
		errString := fmt.Sprintf("[GetPackQueries] %s", err)
		response.WriteError(respWriter, errString)
	} else {
		response.WriteCustomJSON(respWriter, result)
	}
}

// SearchPackQueries searches all packqueries by substring
func SearchPackQueries(respWriter http.ResponseWriter, request *http.Request) {

	// Only handle GET requests
	if request.Method != http.MethodGet {
		return
	}

	handleRequest := func() (interface{}, error) {

		vars := mux.Vars(request)
		dynDBInstance := dyndb.DbInstance()
		searchString := vars["search_string"]
		results, err := dyndb.APISearchPackQueries(searchString, dynDBInstance)
		if err != nil {
			return nil, fmt.Errorf("failed to search pack queries for string '%s': %s", searchString, err)
		}

		return results, nil
	}

	result, err := handleRequest()
	if err != nil {
		logger.Error(err)
		errString := fmt.Sprintf("[SearchPackQueries] %s", err)
		response.WriteError(respWriter, errString)
	} else {
		response.WriteCustomJSON(respWriter, result)
	}
}

// GetQueryPacks returns all querypacks
func GetQueryPacks(respWriter http.ResponseWriter, request *http.Request) {

	// Only handle GET requests
	if request.Method != http.MethodGet {
		return
	}

	handleRequest := func() (interface{}, error) {

		dynDBInstance := dyndb.DbInstance()
		results, err := dyndb.SearchQueryPacks("", dynDBInstance)
		if err != nil {
			return nil, fmt.Errorf("failed to get all query packs: %s", err)
		}

		return results, nil
	}

	result, err := handleRequest()
	if err != nil {
		logger.Error(err)
		errString := fmt.Sprintf("[GetQueryPacks] %s", err)
		response.WriteError(respWriter, errString)
	} else {
		response.WriteCustomJSON(respWriter, result)
	}
}

// SearchQueryPacks search for substring in query pack name
func SearchQueryPacks(respWriter http.ResponseWriter, request *http.Request) {

	// Only handle GET requests
	if request.Method != http.MethodGet {
		return
	}

	handleRequest := func() (interface{}, error) {

		vars := mux.Vars(request)
		dynDBInstance := dyndb.DbInstance()
		searchString := vars["search_string"]
		results, err := dyndb.SearchQueryPacks(searchString, dynDBInstance)
		if err != nil {
			return nil, fmt.Errorf("failed to search query packs for string '%s': %s", searchString, err)
		}

		return results, nil
	}

	result, err := handleRequest()
	if err != nil {
		logger.Error(err)
		errString := fmt.Sprintf("[SearchQueryPacks] %s", err)
		response.WriteError(respWriter, errString)
	} else {
		response.WriteCustomJSON(respWriter, result)
	}
}

// ConfigurePack configures named pack
func ConfigurePack(respWriter http.ResponseWriter, request *http.Request) {

	// Only handle POST requests
	if request.Method != http.MethodPost {
		return
	}

	handleRequest := func() error {

		vars := mux.Vars(request)
		packName, ok := vars["pack_name"]
		if !ok || packName == "" {
			return errors.New("no pack specified")
		}

		body, err := ioutil.ReadAll(request.Body)
		defer request.Body.Close()
		if err != nil {
			return fmt.Errorf("failed to read request body: %s", err)
		}

		querypack := osquery_types.QueryPack{}
		err = json.Unmarshal(body, &querypack)
		if err != nil {
			return fmt.Errorf("failed to unmarshal request body [%s]: %s", string(body), err)
		}

		dynDBInstance := dyndb.DbInstance()
		err = dyndb.UpsertPack(querypack, dynDBInstance)
		if err != nil {
			return fmt.Errorf("dynamo pack upsert failed: %s", err)
		}

		return nil
	}

	err := handleRequest()
	if err != nil {
		logger.Error(err)
		errString := fmt.Sprintf("[ConfigurePack] %s", err)
		response.WriteError(respWriter, errString)
	}
}

// ConfigurePackQuery accepts post body with packquery config
func ConfigurePackQuery(respWriter http.ResponseWriter, request *http.Request) {

	handleRequest := func() error {

		dynDBInstance := dyndb.DbInstance()

		switch request.Method {
		case http.MethodGet:

			vars := mux.Vars(request)
			qName, ok := vars["query_name"]
			if !ok || qName == "" {
				return errors.New("no pack specified")
			}

			packQuery, err := dyndb.GetPackQuery(qName, dynDBInstance)
			if err != nil {
				return fmt.Errorf("failed to get pack query: %s", err)
			}

			response.WriteCustomJSON(respWriter, packQuery)

		case http.MethodPost:
			body, err := ioutil.ReadAll(request.Body)
			defer request.Body.Close()
			if err != nil {
				return fmt.Errorf("failed to read request body: %s", err)
			}

			var postData osquery_types.PackQuery
			err = json.Unmarshal(body, &postData)
			if err != nil {
				return fmt.Errorf("failed to unmarshal request body [%s]: %s", string(body), err)
			}

			err = dyndb.UpsertPackQuery(postData, dynDBInstance)
			if err != nil {
				return fmt.Errorf("dynamo pack query upsert failed: %s", err)
			}
		}
		return nil
	}

	err := handleRequest()
	if err != nil {
		logger.Error(err)
		errString := fmt.Sprintf("[ConfigurePackQuery] %s request failed: %s", request.Method, err)
		response.WriteError(respWriter, errString)
	}
}
