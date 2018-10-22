package api

import (
	"net/http"
	"fmt"
	"github.com/oktasecuritylabs/sgt/handlers/response"
	"github.com/oktasecuritylabs/sgt/logger"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

func DeleteNodeHandler(db ApiDB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleRequest := func() (error) {

			vars := mux.Vars(r)
			nodeKey, ok := vars["node_key"]
			if !ok || nodeKey == "" {
				return errors.New("request did not contain node_key")
			}
			err := db.DeleteNodeByNodekey(nodeKey)
			//dynDBInstance := dyndb.DbInstance()
			if err != nil {
				return errors.Errorf("could not get named configs: %s", err)
			}

			return nil
		}

		err := handleRequest()
		if err != nil {
			logger.Error(err)
			errString := fmt.Sprintf("DeleteNodeKey: ", err)
			response.WriteError(w, errString)
		} else {
			response.WriteSuccess(w, "")
		}

	})
}
