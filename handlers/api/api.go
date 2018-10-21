package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/oktasecuritylabs/sgt/handlers/response"
	"github.com/oktasecuritylabs/sgt/logger"
	"github.com/oktasecuritylabs/sgt/osquery_types"
	log "github.com/sirupsen/logrus"
)

type ApiDB interface {
	GetNamedConfigs() ([]osquery_types.OsqueryNamedConfig, error)
	GetNamedConfig(configName string) (osquery_types.OsqueryNamedConfig, error)
	UpsertNamedConfig(onc *osquery_types.OsqueryNamedConfig) error
	UpsertClient(oc osquery_types.OsqueryClient) error
	SearchByHostIdentifier(hid string) ([]osquery_types.OsqueryClient, error)
	ApprovePendingNode(nodeKey string) error
	ValidNode(nodeKey string) error
	SearchByNodeKey(nk string) (osquery_types.OsqueryClient, error)
	APIGetPackQueries() ([]osquery_types.PackQuery, error)
	APISearchPackQueries(searchString string) ([]osquery_types.PackQuery, error)
	GetPackQuery(queryName string) (osquery_types.PackQuery, error)
	UpsertPackQuery(pq osquery_types.PackQuery) error
	GetPackByName(packName string) (osquery_types.Pack, error)
	SearchQueryPacks(searchString string) ([]osquery_types.QueryPack, error)
	NewQueryPack(qp osquery_types.QueryPack) error
	DeleteQueryPack(queryPackName string) error
	UpsertPack(qp osquery_types.QueryPack) error
	SearchDistributedNodeKey(nk string) (osquery_types.DistributedQuery, error)
	NewDistributedQuery(dq osquery_types.DistributedQuery) error
	DeleteDistributedQuery(dq osquery_types.DistributedQuery) error
	AppendDistributedQuery(dq osquery_types.DistributedQuery) error
	UpsertDistributedQuery(dq osquery_types.DistributedQuery) error
	NewUser(u osquery_types.User) error
	GetUser(username string) (osquery_types.User, error)
}

func init() {
	//logger.SetFormatter(&logger.JSONFormatter{/I//})
	logger.WithFields(log.Fields{
		"Module": "API",
	})
}

//GetNamedConfigs returns all named configs in a json list
func GetNamedConfigsHandler(db ApiDB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleRequest := func() (interface{}, error) {

			//dynDBInstance := dyndb.DbInstance()
			ans, err := db.GetNamedConfigs()
			if err != nil {
				return nil, fmt.Errorf("could not get named configs: %s", err)
			}

			return ans, nil
		}

		result, err := handleRequest()
		if err != nil {
			logger.Error(err)
			errString := fmt.Sprintf("[GetNamedConfigs] failed to get named configs: %s", err)
			response.WriteError(w, errString)
		} else {
			response.WriteCustomJSON(w, result)
		}

	})
}

// ConfigurationRequestHandler accepts a json post body of a NamedConfig
func ConfigurationRequestHandler(db ApiDB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleRequest := func() (interface{}, error) {

			vars := mux.Vars(r)
			configName, ok := vars["config_name"]
			if !ok || configName == "" {
				return nil, errors.New("no config name specified")
			}

			// get the named config
			//dynDBInstance := dyndb.DbInstance()
			existingNamedConfig, err := db.GetNamedConfig(configName)
			if err != nil {
				return nil, fmt.Errorf("failed to get config with name [%s]: %s", configName, err)
			}

			switch r.Method {
			case http.MethodGet:

				return existingNamedConfig, nil

			case http.MethodPost:

				//now merge what's already in teh database with our defaults
				existingNamedConfig.OsqueryConfig.Options = osquery_types.NewOsqueryOptions()

				// finally...
				body, err := ioutil.ReadAll(r.Body)
				defer r.Body.Close()
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

				//err = dyndb.UpsertNamedConfig(dynDBInstance, &existingNamedConfig)
				err = db.UpsertNamedConfig(&existingNamedConfig)
				if err != nil {
					return nil, fmt.Errorf("dynamo named config upsert failed: %s", err)
				}

				return existingNamedConfig, nil
			}

			return nil, fmt.Errorf("method not supported: %s", r.Method)
		}

		result, err := handleRequest()
		if err != nil {
			logger.Error(err)
			errString := fmt.Sprintf("[ConfigurationRequest] failed to handle named config in %s request: %s", r.Method, err)
			response.WriteError(w, errString)
		} else {
			response.WriteCustomJSON(w, result)
		}

	})
}

// GetNodes returns json reponse of a list of nodes
func GetNodesHandler(db ApiDB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			return
		}

		handleRequest := func() (interface{}, error) {

			results, err := db.SearchByHostIdentifier("")
			if err != nil {
				return nil, fmt.Errorf("failed to get all nodes: %s", err)
			}

			return results, nil
		}

		result, err := handleRequest()
		if err != nil {
			logger.Error(err)
			errString := fmt.Sprintf("[GetNodes] %s", err)
			response.WriteError(w, errString)
		} else {
			response.WriteCustomJSON(w, result)
		}

	})
}

/*
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
*/

func ConfigureNodeHandler(db ApiDB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleRequest := func() (interface{}, error) {

			vars := mux.Vars(r)
			nodeKey, ok := vars["node_key"]
			if !ok || nodeKey == "" {
				return nil, errors.New("request does not contain node_key")
			}

			existingClient, err := db.SearchByNodeKey(nodeKey)
			if err != nil {
				return nil, fmt.Errorf("failed to find node by key [%s]: %s", nodeKey, err)
			}

			logger.Infof("existing client: %+v", existingClient)

			switch r.Method {
			case http.MethodGet:

				return existingClient, nil

			case http.MethodPost:

				if existingClient.NodeKey == "" {
					return nil, errors.New("existing client node_key is empty")
				}

				body, err := ioutil.ReadAll(r.Body)
				defer r.Body.Close()
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

				err = db.UpsertClient(client)
				if err != nil {
					return nil, fmt.Errorf("client update in dynamo failed: %s", err)
				}

				return client, nil
			}

			return nil, fmt.Errorf("method not supported: %s", r.Method)
		}

		result, err := handleRequest()
		if err != nil {
			logger.Error(err)
			errString := fmt.Sprintf("[ConfigureNode] failed to configure node in %s request: %s", r.Method, err)
			response.WriteError(w, errString)
		} else {
			response.WriteCustomJSON(w, result)
		}
	})
}

/*
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

*/

func ApproveNode(db ApiDB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			return
		}

		handleRequest := func() error {

			vars := mux.Vars(r)
			nodeKey, ok := vars["node_key"]
			if !ok || nodeKey == "" {
				return errors.New("request does not contain node_key")
			}

			logger.Warn("posting approval")

			err := db.ApprovePendingNode(nodeKey)
			if err != nil {
				return fmt.Errorf("approval of pending node failed: %s", err)
			}

			return nil
		}

		err := handleRequest()
		if err != nil {
			logger.Error(err)
			errString := fmt.Sprintf("[ApproveNode] failed to approve node: %s", err)
			response.WriteError(w, errString)
		} else {
			response.WriteSuccess(w, "")
		}

	})
}

/*
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
*/

func GetPackQueries(db ApiDB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			return
		}

		handleRequest := func() (interface{}, error) {

			results, err := db.APIGetPackQueries()
			if err != nil {
				return nil, fmt.Errorf("failed to get query packs: %s", err)
			}

			return results, nil
		}

		result, err := handleRequest()
		if err != nil {
			logger.Error(err)
			errString := fmt.Sprintf("[GetPackQueries] %s", err)
			response.WriteError(w, errString)
		} else {
			response.WriteCustomJSON(w, result)
		}

	})
}

/*
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
*/

// SearchPackQueries searches all packqueries by substring
func SearchPackQueries(db ApiDB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			return
		}

		handleRequest := func() (interface{}, error) {

			vars := mux.Vars(r)
			searchString := vars["search_string"]
			results, err := db.APISearchPackQueries(searchString)
			if err != nil {
				return nil, fmt.Errorf("failed to search pack queries for string '%s': %s", searchString, err)
			}

			return results, nil
		}

		result, err := handleRequest()
		if err != nil {
			logger.Error(err)
			errString := fmt.Sprintf("[SearchPackQueries] %s", err)
			response.WriteError(w, errString)
		} else {
			response.WriteCustomJSON(w, result)
		}

	})
}

/*
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
*/

// GetQueryPacks returns all querypacks
func GetQueryPacks(db ApiDB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			return
		}

		handleRequest := func() (interface{}, error) {

			results, err := db.SearchQueryPacks("")
			if err != nil {
				return nil, fmt.Errorf("failed to get all query packs: %s", err)
			}

			return results, nil
		}

		result, err := handleRequest()
		if err != nil {
			logger.Error(err)
			errString := fmt.Sprintf("[GetQueryPacks] %s", err)
			response.WriteError(w, errString)
		} else {
			response.WriteCustomJSON(w, result)
		}

	})
}

/*
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
*/

// SearchQueryPacks search for substring in query pack name
func SearchQueryPacks(db ApiDB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			return
		}

		handleRequest := func() (interface{}, error) {

			vars := mux.Vars(r)
			searchString := vars["search_string"]
			results, err := db.SearchQueryPacks(searchString)
			if err != nil {
				return nil, fmt.Errorf("failed to search query packs for string '%s': %s", searchString, err)
			}

			return results, nil
		}

		result, err := handleRequest()
		if err != nil {
			logger.Error(err)
			errString := fmt.Sprintf("[SearchQueryPacks] %s", err)
			response.WriteError(w, errString)
		} else {
			response.WriteCustomJSON(w, result)
		}

	})
}

/*
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
*/

// ConfigurePack configures named pack
func ConfigurePack(db ApiDB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			return
		}

		handleRequest := func() error {

			vars := mux.Vars(r)
			packName, ok := vars["pack_name"]
			if !ok || packName == "" {
				return errors.New("no pack specified")
			}

			body, err := ioutil.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				return fmt.Errorf("failed to read request body: %s", err)
			}

			querypack := osquery_types.QueryPack{}
			err = json.Unmarshal(body, &querypack)
			if err != nil {
				return fmt.Errorf("failed to unmarshal request body [%s]: %s", string(body), err)
			}

			err = db.UpsertPack(querypack)
			if err != nil {
				return fmt.Errorf("dynamo pack upsert failed: %s", err)
			}

			return nil
		}

		err := handleRequest()
		if err != nil {
			logger.Error(err)
			errString := fmt.Sprintf("[ConfigurePack] %s", err)
			response.WriteError(w, errString)
		}

	})
}

/*
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
*/

// ConfigurePackQuery accepts post body with packquery config
func ConfigurePackQuery(db ApiDB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleRequest := func() error {

			switch r.Method {
			case http.MethodGet:

				vars := mux.Vars(r)
				qName, ok := vars["query_name"]
				if !ok || qName == "" {
					return errors.New("no pack specified")
				}

				packQuery, err := db.GetPackQuery(qName)
				if err != nil {
					return fmt.Errorf("failed to get pack query: %s", err)
				}

				response.WriteCustomJSON(w, packQuery)

			case http.MethodPost:
				body, err := ioutil.ReadAll(r.Body)
				defer r.Body.Close()
				if err != nil {
					return fmt.Errorf("failed to read request body: %s", err)
				}

				var postData osquery_types.PackQuery
				err = json.Unmarshal(body, &postData)
				if err != nil {
					return fmt.Errorf("failed to unmarshal request body [%s]: %s", string(body), err)
				}

				err = db.UpsertPackQuery(postData)
				if err != nil {
					return fmt.Errorf("dynamo pack query upsert failed: %s", err)
				}
			}
			return nil
		}

		err := handleRequest()
		if err != nil {
			logger.Error(err)
			errString := fmt.Sprintf("[ConfigurePackQuery] %s request failed: %s", r.Method, err)
			response.WriteError(w, errString)
		}

	})
}

/*
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
*/
