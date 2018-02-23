package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/mux"
	"github.com/oktasecuritylabs/sgt/dyndb"
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
func GetNamedConfigs(w http.ResponseWriter, r *http.Request) {
	dynDBInstance := dyndb.DbInstance()
	ans, err := dyndb.GetNamedConfigs(dynDBInstance)
	if err != nil {
		logger.Error(err)
		w.Write([]byte(`{"error": true}`))
		return
	}
	js, err := json.Marshal(ans)
	if err != nil {
		logger.Error(err)
		w.Write([]byte(`{"error": true}`))
		return
	}
	w.Write(js)
}

//ConfigurationRequest accepts a json post body of a NamedConfig
func ConfigurationRequest(respwritter http.ResponseWriter, request *http.Request) {
	mu := sync.Mutex{}
	dynDBInstance := dyndb.DbInstance()
	vars := mux.Vars(request)
	if request.Method == "GET" {
		if vars["config_name"] == "" {
			ans, err := dyndb.GetNamedConfig(dynDBInstance, vars["config_name"])
			if err != nil {
				logger.Error(err)
			}
			js, err := json.Marshal(ans)
			if err != nil {
				logger.Error(err)
			}
			respwritter.Write(js)
			//return list of configs
		} else {
			//return named config
			ans, err := dyndb.GetNamedConfig(dynDBInstance, vars["config_name"])
			if err != nil {
				logger.Error(err)
			}
			js, err := json.Marshal(ans)
			if err != nil {
				logger.Error(err)
			}
			respwritter.Write(js)
		}
	}
	if request.Method == "POST" {
		if vars["config_name"] == "" {
			respwritter.Write([]byte(`{"result":"failure"}`))
			return
		}
		//Create base osquery named config with default options
		namedConfig := osquery_types.OsqueryNamedConfig{}
		namedConfig.OsqueryConfig.Options = osquery_types.NewOsqueryOptions()
		existingNamedConfig, err := dyndb.GetNamedConfig(dynDBInstance, vars["config_name"])
		//now merge what's already in teh database with our defaults
		js, err := json.Marshal(existingNamedConfig)
		if err != nil {
			logger.Error(err)
			return
		}
		err = json.Unmarshal(js, &namedConfig)
		if err != nil {
			logger.Error(err)
			return
		}
		// finally...
		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			panic(err)
		}
		//overlay default + existing with options provided by user
		err = json.Unmarshal(body, &namedConfig)
		if err != nil {
			panic(fmt.Sprintln(err, os.Stdout))
		}
		if vars["config_name"] != namedConfig.ConfigName {
			respwritter.Write([]byte(`{"result":"failure", "reason": "named config endpoint does not match posted data config_name"}`))
			return
		}

		err = dyndb.UpsertNamedConfig(dynDBInstance, &namedConfig, &mu)
		if err != nil {
			logger.Error(err)
			return
		}
	}
	return
}

//GetNodes returns json reponse of a list of nodes
func GetNodes(respwritter http.ResponseWriter, request *http.Request) {
	dynDBInstance := dyndb.DbInstance()
	//nodes := osquery_types.OsqueryClient{}
	if request.Method == "GET" {
		results, err := dyndb.SearchByHostIdentifier("", dynDBInstance)
		if err != nil {
			logger.Error(err)
			return
		}
		js, err := json.Marshal(results)
		if err != nil {
			logger.Error(err)
			return
		}
		respwritter.Write(js)
		return
	}
}

//ConfigureNode accepts json node configuration
func ConfigureNode(respwritter http.ResponseWriter, request *http.Request) {
	dynDBInstance := dyndb.DbInstance()
	mu := sync.Mutex{}
	vars := mux.Vars(request)
	nodeKey := vars["node_key"]
	client := osquery_types.OsqueryClient{}
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		panic(err)
	}
	//set posted osquery client = client
	err = json.Unmarshal(body, &client)
	logger.Infof("%+v", client)
	if request.Method == "GET" {
		if vars["node_key"] != "" {
			result, err := dyndb.SearchByNodeKey(client.NodeKey, dynDBInstance)
			if err != nil {
				logger.Error(err)
			}
			if result.HostIdentifier != "" {
				js, err := json.Marshal(result)
				if err != nil {
					logger.Error(err)
				}
				respwritter.Write(js)
				return
			}
		}
	}
	if request.Method == "POST" {
		if vars["node_key"] != "" {
			//verify that node exists
			//get existing client config
			//rewrite this crappy shit
			existingClient, err := dyndb.SearchByNodeKey(nodeKey, dynDBInstance)
			if err != nil {
				logger.Error(err)
			}
			logger.Infof("%+v", existingClient)
			if err != nil {
				logger.Error(err)
				respwritter.Write([]byte(`{"error": "node invalid"}`))
				return
			}
			if len(existingClient.NodeKey) > 0 {
				//map missing keys to client
				client.NodeKey = nodeKey
				if len(client.ConfigName) <= 0 {
					client.ConfigName = existingClient.ConfigName
				}
				client.HostIdentifier = existingClient.HostIdentifier
				if client.NodeInvalid != true {
					if client.NodeInvalid != false {
						client.NodeInvalid = existingClient.NodeInvalid
					}
				}
				client.HostDetails = existingClient.HostDetails
				if client.PendingRegistrationApproval != false {
					if client.PendingRegistrationApproval != true {
						client.PendingRegistrationApproval = existingClient.PendingRegistrationApproval
					}
				}
				if len(client.Tags) < 1 {
					client.Tags = existingClient.Tags
				}
				logger.Warn("%v", client)
				err := dyndb.UpsertClient(client, dynDBInstance, &mu)
				if err != nil {
					logger.Error(err)
					respwritter.Write([]byte(`{"error": "update failed"}`))
					return
				}
				js, err := json.Marshal(client)
				if err != nil {
					logger.Error(err)
					return
				}
				respwritter.Write(js)
				return
			}
		}
		//Results := dyndb.SearchByHostIdentifier("", dynDBInstance)
	}
	if request.Method == "GET" {
		//nodeKey := vars["nodeKey"]
		client, err := dyndb.SearchByNodeKey(nodeKey, dynDBInstance)
		if err != nil {
			logger.Error(err)
		}
		js, err := json.Marshal(client)
		if err != nil {
			logger.Error(err)
			return
		}
		respwritter.Write(js)
		return
	}
}

//ApproveNode helper function as a shortcut to approving node.  Takes no body input
func ApproveNode(respwritter http.ResponseWriter, request *http.Request) {
	dynDBInstance := dyndb.DbInstance()
	mu := sync.Mutex{}
	vars := mux.Vars(request)
	if request.Method == "POST" {
		logger.Warn("posting approval")
		err := dyndb.ApprovePendingNode(vars["node_key"], dynDBInstance, &mu)
		logger.Infof("[ApproveNode] vars: %+v", vars)
		if err != nil {
			logger.Error(err)
			respwritter.Write([]byte(`{"result": "error"}`))
			return
		}
		respwritter.Write([]byte(`{"result": "success"}`))
		return
	}
}

//GetPackQueries returns json response of a list of packqueries
func GetPackQueries(respwritter http.ResponseWriter, request *http.Request) {
	type packQueryList struct {
		packqueries []osquery_types.PackQuery `json:"packqueries"`
	}
	//pql := packQueryList{}
	dynDBInstance := dyndb.DbInstance()
	if request.Method == "GET" {
		results, err := dyndb.APIGetPackQueries(dynDBInstance)
		if err != nil {
			logger.Error(err)
		}
		js, err := json.Marshal(results)
		if err != nil {
			logger.Error(err)
		}
		respwritter.Write(js)
	}
}

//SearchPackQueries searches all packqueries by substring
func SearchPackQueries(respwritter http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	type packQueryList struct {
		packqueries []osquery_types.PackQuery `json:"packqueries"`
	}
	//pql := packQueryList{}
	dynDBInstance := dyndb.DbInstance()
	if request.Method == "GET" {
		results, err := dyndb.APISearchPackQueries(vars["search_string"], dynDBInstance)
		if err != nil {
			logger.Error(err)
		}
		js, err := json.Marshal(results)
		if err != nil {
			logger.Error(err)
		}
		respwritter.Write(js)
	}
}

//GetQueryPacks returns all querypacks
func GetQueryPacks(respwritter http.ResponseWriter, request *http.Request) {
	type packQueryList struct {
		QueryPacks []osquery_types.QueryPack `json:"query_packs"`
	}
	dynDBInstance := dyndb.DbInstance()
	if request.Method == "GET" {
		results, err := dyndb.SearchQueryPacks("", dynDBInstance)
		if err != nil {
			logger.Error(err)
		}
		js, err := json.Marshal(results)
		if err != nil {
			logger.Error(err)
		}
		respwritter.Write(js)
	}
}

//SearchQueryPacks search for substring in query pack name
func SearchQueryPacks(respwritter http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	type packQueryList struct {
		QueryPacks []osquery_types.QueryPack `json:"query_packs"`
	}
	dynDBInstance := dyndb.DbInstance()
	if request.Method == "GET" {
		results, err := dyndb.SearchQueryPacks(vars["search_string"], dynDBInstance)
		if err != nil {
			logger.Error(err)
		}
		js, err := json.Marshal(results)
		if err != nil {
			logger.Error(err)
		}
		respwritter.Write(js)
	}
}

//ConfigurePack configures named pack
func ConfigurePack(respwritter http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	if len(vars["pack_name"]) < 1 {
		respwritter.Write([]byte(`No pack specified`))
		logger.Error("No pack specified")
		return
	}
	//pname := vars["pack_name"]
	dynDBInstance := dyndb.DbInstance()
	if request.Method == "POST" {
		body, err := ioutil.ReadAll(request.Body)
		defer request.Body.Close()
		if err != nil {
			logger.Error(err)
		}
		querypack := osquery_types.QueryPack{}
		err = json.Unmarshal(body, &querypack)
		if err != nil {
			logger.Error(err)
			return
		}
		err = dyndb.UpsertPack(querypack, dynDBInstance)
		if err != nil {
			logger.Error(err)
			return
		}
	}

}

//ConfigurePackQuery accepts post body with packquery config
func ConfigurePackQuery(respwritter http.ResponseWriter, request *http.Request) {
	mu := sync.Mutex{}
	vars := mux.Vars(request)
	if len(vars["query_name"]) < 1 {
		respwritter.Write([]byte(`No pack specified`))
		logger.Error("No pack specified")
		return
	}
	qname := vars["query_name"]
	dynDBInstance := dyndb.DbInstance()
	if request.Method == "GET" {
		packQuery, err := dyndb.GetPackQuery(qname, dynDBInstance)
		if err != nil {
			logger.Error(err)
			return
		}
		js, err := json.Marshal(packQuery)
		if err != nil {
			logger.Error(err)
			return
		}
		respwritter.Write(js)

	}
	if request.Method == "POST" {
		body, err := ioutil.ReadAll(request.Body)
		defer request.Body.Close()
		if err != nil {
			logger.Error(err)
		}
		var postData osquery_types.PackQuery
		err = json.Unmarshal(body, &postData)
		if err != nil {
			logger.Error(err)
			return
		}
		err = dyndb.UpsertPackQuery(postData, dynDBInstance, &mu)
		if err != nil {
			logger.Error(err)
			return
		}
	}
}
