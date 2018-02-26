package response

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/oktasecuritylabs/sgt/logger"
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

func writeResponseJSON(respWriter http.ResponseWriter, response interface{}) {

	// Set the header type
	respWriter.Header().Set("Content-Type", "application/json")

	respJSON, err := json.Marshal(response)
	if err != nil {
		// If there is an error marshaling the interface, write a basic error
		errString := fmt.Sprintf("response failed to marshal to json: %s", err)
		logger.Error(errString)
		http.Error(respWriter, errString, http.StatusInternalServerError)
		return
	}

	// Write the response to the http.ResponseWriter using io.WriteString
	_, err = io.WriteString(respWriter, string(respJSON))

	if err != nil {
		// If there is an error writing the repsonse, write an error
		errString := fmt.Sprintf("failed to write response: %s", err)
		logger.Error(errString)
		http.Error(respWriter, errString, http.StatusInternalServerError)
	}
}

// WriteError will write the passed error to the http response writer
func WriteError(respWriter http.ResponseWriter, errorString string) {
	writeResponseJSON(respWriter, sgtBaseResponse{Message: errorString, Status: statusError})
}

// WriteSuccess will write the a success status and optional message to the http response writer
func WriteSuccess(respWriter http.ResponseWriter, optionalMessage string) {
	writeResponseJSON(respWriter, sgtBaseResponse{Message: optionalMessage, Status: statusSuccess})
}

// WriteCustomJSON will write the custom response to the http response writer as json
func WriteCustomJSON(respWriter http.ResponseWriter, resp interface{}) {
	writeResponseJSON(respWriter, resp)
}
