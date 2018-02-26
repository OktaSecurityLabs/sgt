package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/oktasecuritylabs/sgt/handlers/api"
	"github.com/oktasecuritylabs/sgt/handlers/auth"
	"github.com/oktasecuritylabs/sgt/handlers/distributed"
	"github.com/oktasecuritylabs/sgt/handlers/node"
	"github.com/urfave/negroni"
)

// Serve will create the server listen
func Serve() error {
	router := mux.NewRouter()
	//node endpoint
	nodeAPI := router.PathPrefix("/node").Subrouter()
	nodeAPI.Path("/configure").HandlerFunc(node.NodeConfigureRequest)
	nodeAPI.Path("/enroll").HandlerFunc(node.NodeEnrollRequest)
	//protect with uiAuth
	//Configuration (management) endpoint
	apiRouter := mux.NewRouter().PathPrefix("/api/v1/configuration").Subrouter()

	apiRouter.HandleFunc("/configs", api.GetNamedConfigs).Methods(http.MethodGet, http.MethodPost)
	apiRouter.HandleFunc("/configs/{config_name}", api.ConfigurationRequest).Methods(http.MethodPost)
	//Nodes
	apiRouter.HandleFunc("/nodes", api.GetNodes).Methods(http.MethodGet)
	apiRouter.HandleFunc("/nodes/{node_key}", api.ConfigureNode).Methods(http.MethodPost, http.MethodGet)
	apiRouter.HandleFunc("/nodes/{node_key}/approve", api.ApproveNode).Methods(http.MethodPost)
	//apiRouter.HandleFunc("/nodes/approve/_bulk", api.Placeholder).Methods("POST)
	//Packs
	apiRouter.HandleFunc("/packs", api.GetQueryPacks).Methods(http.MethodGet)
	apiRouter.HandleFunc("/packs/search/{search_string}", api.SearchQueryPacks).Methods(http.MethodGet)
	apiRouter.HandleFunc("/packs/{pack_name}", api.ConfigurePack).Methods(http.MethodPost)
	//PackQueries
	apiRouter.HandleFunc("/packqueries", api.GetPackQueries).Methods(http.MethodGet)
	apiRouter.HandleFunc("/packqueries/{query_name}", api.ConfigurePackQuery)
	apiRouter.HandleFunc("/packqueries/search/{search_string}", api.SearchPackQueries)
	apiRouter.HandleFunc("/distributed/add", distributed.DistributedQueryAdd)
	//Enforce uiAuth for all our api configuration endpoints
	router.PathPrefix("/api/v1/configuration").Handler(negroni.New(
		negroni.NewRecovery(),
		negroni.HandlerFunc(auth.AnotherValidation),
		negroni.Wrap(apiRouter),
	))
	//token
	router.HandleFunc("/api/v1/get-token", auth.GetTokenHandler)
	//Distributed endpoint
	distributedRouter := mux.NewRouter().PathPrefix("/distributed").Subrouter()
	distributedRouter.HandleFunc("/read", distributed.DistributedQueryRead)
	distributedRouter.HandleFunc("/write", distributed.DistributedQueryWrite)
	//auth for distributed read/write
	router.PathPrefix("/distributed").Handler(negroni.New(
		negroni.NewRecovery(),
		negroni.HandlerFunc(auth.ValidNodeKey),
		negroni.Wrap(distributedRouter),
	))
	//Enforce auth for all our api configuration endpoints
	router.PathPrefix("/api/v1/configuration").Handler(negroni.New(
		negroni.NewRecovery(),
		negroni.HandlerFunc(auth.AnotherValidation),
		negroni.Wrap(apiRouter),
	))
	err := http.ListenAndServeTLS(":443",
		"fullchain.pem", "privkey.pem",  router)
		//"fullchain.pem", "privkey.pem", handlers.LoggingHandler(os.Stdout, router))
	return err
}
