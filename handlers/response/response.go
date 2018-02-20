package response

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	statusError   = "error"
	statusFailed  = "failed"
	statusSuccess = "success"
)

// sgtBaseResponse stores some basic endpoint response values
type sgtBaseResponse struct {
	Message string `json:"message,omitempty"`
	Status  string `json:"status"`
}

// SGTCustomResponse is an alias for a map that allows for
// storing custom endpoint response values
type SGTCustomResponse map[string]interface{}

func getResponseJSON(response interface{}) []byte {

	respJSON, err := json.Marshal(response)
	if err != nil {
		// If there is an error marshaling the interface, return a basic error
		errString := fmt.Sprintf("response failed to marshal to json: %s", err)
		newResp := sgtBaseResponse{Message: errString, Status: errString}
		respJSON, _ = json.Marshal(newResp)
	}
	return respJSON
}

// WriteError will write the passed error to the http response writer
func WriteError(respWriter http.ResponseWriter, errorString string) {
	respWriter.Header().Set("Content-Type", "application/json")
	respWriter.Write(getResponseJSON(sgtBaseResponse{Message: errorString, Status: statusError}))
}

// WriteSuccess will write the a success status and optional message to the http response writer
func WriteSuccess(respWriter http.ResponseWriter, optionalMessage string) {
	respWriter.Header().Set("Content-Type", "application/json")
	respWriter.Write(getResponseJSON(sgtBaseResponse{Message: optionalMessage, Status: statusSuccess}))
}

// WriteCustom will write the custom response to the http response writer
func WriteCustom(respWriter http.ResponseWriter, resp interface{}) {
	respWriter.Header().Set("Content-Type", "application/json")
	respWriter.Write(getResponseJSON(resp))
}
