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
		namedConfig.Osquery_config.Options = osquery_types.NewOsqueryOptions()
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
		if vars["config_name"] != namedConfig.Config_name {
			respwritter.Write([]byte(`{"result":"failure", "reason": "named config endpoint does not match posted data config_name"}`))
			return
		}
		ans := dyndb.UpsertNamedConfig(dynDBInstance, &namedConfig, mu)
		if ans {
			js, err := json.Marshal(namedConfig)
			if err != nil {
				respwritter.Write([]byte(`{"result":"failure"}`))
				return
			}
			respwritter.Write(js)
		}
		fmt.Println(ans)
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
	nodeKey := vars["nodeKey"]
	client := osquery_types.OsqueryClient{}
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		panic(err)
	}
	//set posted osquery client = client
	err = json.Unmarshal(body, &client)
	logger.Warn("%v", client)
	if request.Method == "GET" {
		if vars["nodeKey"] != "" {
			result, err := dyndb.SearchByNodeKey(client.Node_key, dynDBInstance)
			if err != nil {
				logger.Error(err)
			}
			if result.Host_identifier != "" {
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
		if vars["nodeKey"] != "" {
			//verify that node exists
			//get existing client config
			existingClient, err := dyndb.SearchByNodeKey(nodeKey, dynDBInstance)
			if err != nil {
				logger.Error(err)
			}
			logger.Warn("%v", existingClient)
			if err != nil {
				logger.Error(err)
				respwritter.Write([]byte(`{"error": "node invalid"}`))
				return
			}
			if len(existingClient.Node_key) > 0 {
				//map missing keys to client
				client.Node_key = nodeKey
				if len(client.Config_name) <= 0 {
					client.Config_name = existingClient.Config_name
				}
				client.Host_identifier = existingClient.Host_identifier
				if client.Node_invalid != true {
					if client.Node_invalid != false {
						client.Node_invalid = existingClient.Node_invalid
					}
				}
				client.HostDetails = existingClient.HostDetails
				if client.Pending_registration_approval != false {
					if client.Pending_registration_approval != true {
						client.Pending_registration_approval = existingClient.Pending_registration_approval
					}
				}
				if len(client.Tags) < 1 {
					client.Tags = existingClient.Tags
				}
				logger.Warn("%v", client)
				err := dyndb.UpsertClient(client, dynDBInstance, mu)
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
		err := dyndb.ApprovePendingNode(vars["nodeKey"], dynDBInstance, mu)
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
	mu := sync.Mutex{}
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
		err = dyndb.UpsertPack(querypack, dynDBInstance, mu)
		if err != nil {
			logger.Error(err)
			return
		}
	}

}

//ConfigurePackQuery accepts post body with packquery config
func ConfigurePackQuery(respwritter http.ResponseWriter, request *http.Request) {
	mut := sync.Mutex{}
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
		ok, err := dyndb.UpsertPackQuery(postData, dynDBInstance, mut)
		if ok {
			return
		}
		if err != nil {
			logger.Error(err)
			return
		}
	}
}
