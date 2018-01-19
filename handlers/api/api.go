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

func APIConfigurationRequest(respwritter http.ResponseWriter, request *http.Request) {
	mu := sync.Mutex{}
	dyn_svc := dyndb.DbInstance()
	vars := mux.Vars(request)
	if request.Method == "GET" {
		if vars["config_name"] == "" {
			ans, err := dyndb.GetNamedConfig(dyn_svc, vars["config_name"])
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
			ans, err := dyndb.GetNamedConfig(dyn_svc, vars["config_name"])
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
		named_config := osquery_types.OsqueryNamedConfig{}
		named_config.Osquery_config.Options = osquery_types.NewOsqueryOptions()
		existing_named_config, err := dyndb.GetNamedConfig(dyn_svc, vars["config_name"])
		//now merge what's already in teh database with our defaults
		js, err := json.Marshal(existing_named_config)
		if err != nil {
			logger.Error(err)
			return
		}
		err = json.Unmarshal(js, &named_config)
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
		err = json.Unmarshal(body, &named_config)
		if err != nil {
			panic(fmt.Sprintln(err, os.Stdout))
		}
		if vars["config_name"] != named_config.Config_name {
			respwritter.Write([]byte(`{"result":"failure", "reason": "named config endpoint does not match posted data config_name"}`))
			return
		}
		ans := dyndb.UpsertNamedConfig(dyn_svc, &named_config, mu)
		if ans {
			js, err := json.Marshal(named_config)
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

//(/api/v1/configure/node/{node_key}
//node

func APIGetNodes(respwritter http.ResponseWriter, request *http.Request) {
	dyn_svc := dyndb.DbInstance()
	//nodes := osquery_types.OsqueryClient{}
	if request.Method == "GET" {
		results, err := dyndb.SearchByHostIdentifier("", dyn_svc)
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

func APIConfigureNode(respwritter http.ResponseWriter, request *http.Request) {
	dyn_svc := dyndb.DbInstance()
	mu := sync.Mutex{}
	vars := mux.Vars(request)
	node_key := vars["node_key"]
	client := osquery_types.OsqueryClient{}
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		panic(err)
	}
	//set posted osquery client = client
	err = json.Unmarshal(body, &client)
	logger.Warn("%v", client)
	if request.Method == "GET" {
		if vars["node_key"] != "" {
			result, err := dyndb.SearchByNodeKey(client.Node_key, dyn_svc)
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
		if vars["node_key"] != "" {
			//verify that node exists
			//get existing client config
			existing_client, err := dyndb.SearchByNodeKey(node_key, dyn_svc)
			if err != nil {
				logger.Error(err)
			}
			logger.Warn("%v", existing_client)
			if err != nil {
				logger.Error(err)
				respwritter.Write([]byte(`{"error": "node invalid"}`))
				return
			}
			if len(existing_client.Node_key) > 0 {
				//map missing keys to client
				client.Node_key = node_key
				if len(client.Config_name) <= 0 {
					client.Config_name = existing_client.Config_name
				}
				client.Host_identifier = existing_client.Host_identifier
				if client.Node_invalid != true {
					if client.Node_invalid != false {
						client.Node_invalid = existing_client.Node_invalid
					}
				}
				client.HostDetails = existing_client.HostDetails
				if client.Pending_registration_approval != false {
					if client.Pending_registration_approval != true {
						client.Pending_registration_approval = existing_client.Pending_registration_approval
					}
				}
				if len(client.Tags) < 1 {
					client.Tags = existing_client.Tags
				}
				logger.Warn("%v", client)
				err := dyndb.UpsertClient(client, dyn_svc, mu)
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
		//Results := dyndb.SearchByHostIdentifier("", dyn_svc)
	}
	if request.Method == "GET" {
		//node_key := vars["node_key"]
		client, err := dyndb.SearchByNodeKey(node_key, dyn_svc)
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

func APIApproveNode(respwritter http.ResponseWriter, request *http.Request) {
	dync_svc := dyndb.DbInstance()
	mu := sync.Mutex{}
	vars := mux.Vars(request)
	if request.Method == "POST" {
		logger.Warn("posting approval")
		err := dyndb.ApprovePendingNode(vars["node_key"], dync_svc, mu)
		if err != nil {
			logger.Error(err)
			respwritter.Write([]byte(`{"result": "error"}`))
			return
		}
		respwritter.Write([]byte(`{"result": "success"}`))
		return
	}
}

//func APIGet

func APIGetPacks(respwritter http.ResponseWriter, request *http.Request) {
	//dyn_svc := dyndb.DbInstance()
	//if request.Method == "GET" {
	//res, err := dyndb.Get
	//}
}

func APIGetPackQueries(respwritter http.ResponseWriter, request *http.Request) {
	type pack_query_list struct {
		packqueries []osquery_types.PackQuery `json:"packqueries"`
	}
	//pql := pack_query_list{}
	dyn_svc := dyndb.DbInstance()
	if request.Method == "GET" {
		results, err := dyndb.APIGetPackQueries(dyn_svc)
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

func APISearchPackQueries(respwritter http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	type pack_query_list struct {
		packqueries []osquery_types.PackQuery `json:"packqueries"`
	}
	//pql := pack_query_list{}
	dyn_svc := dyndb.DbInstance()
	if request.Method == "GET" {
		results, err := dyndb.APISearchPackQueries(vars["search_string"], dyn_svc)
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

//(/api/v1/configure/packs/{pack_name}

func APIGetQueryPacks(respwritter http.ResponseWriter, request *http.Request) {
	type pack_query_list struct {
		QueryPacks []osquery_types.QueryPack `json:"query_packs"`
	}
	dyn_svc := dyndb.DbInstance()
	if request.Method == "GET" {
		results, err := dyndb.SearchQueryPacks("", dyn_svc)
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

func APISearchQueryPacks(respwritter http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	type pack_query_list struct {
		QueryPacks []osquery_types.QueryPack `json:"query_packs"`
	}
	dyn_svc := dyndb.DbInstance()
	if request.Method == "GET" {
		results, err := dyndb.SearchQueryPacks(vars["search_string"], dyn_svc)
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

func APIConfigurePack(respwritter http.ResponseWriter, request *http.Request) {
	mu := sync.Mutex{}
	vars := mux.Vars(request)
	if len(vars["pack_name"]) < 1 {
		respwritter.Write([]byte(`No pack specified`))
		logger.Error("No pack specified")
		return
	}
	//pname := vars["pack_name"]
	dyn_svc := dyndb.DbInstance()
	//if request.Method == "GET" {
	//logger.Println("GOT GET")
	//s, err := dyndb.GetPackByName(pname, dyn_svc)
	//logger.Println(s, err)
	//if err != nil {
	//logger.Println(err)
	//return
	//}
	//logger.Println(s)
	//respwritter.Write([]byte(s))
	//}
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
		err = dyndb.UpsertPack(querypack, dyn_svc, mu)
		if err != nil {
			logger.Error(err)
			return
		}
	}

}

func APIConfigurePackQuery(respwritter http.ResponseWriter, request *http.Request) {
	mut := sync.Mutex{}
	vars := mux.Vars(request)
	if len(vars["query_name"]) < 1 {
		respwritter.Write([]byte(`No pack specified`))
		logger.Error("No pack specified")
		return
	}
	qname := vars["query_name"]
	dyn_svc := dyndb.DbInstance()
	if request.Method == "GET" {
		pack_query, err := dyndb.GetPackQuery(qname, dyn_svc)
		if err != nil {
			logger.Error(err)
			return
		}
		js, err := json.Marshal(pack_query)
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
		var post_data osquery_types.PackQuery
		err = json.Unmarshal(body, &post_data)
		if err != nil {
			logger.Error(err)
			return
		}
		ok, err := dyndb.UpsertPackQuery(post_data, dyn_svc, mut)
		if ok {
			return
		}
		if err != nil {
			logger.Error(err)
			return
		}
	}
}
