package node

import (
	"github.com/oktasecuritylabs/sgt/handlers/helpers"
	"github.com/oktasecuritylabs/sgt/osquery_types"
	"net/http"
	"net/url"
	"testing"
)

func init() {
}

func TestNodeEnrollRequest(t *testing.T) {
	config := &osquery_types.ServerConfig{}
	mockdb := helpers.NewMockDB()

	handler := NodeEnrollRequest(mockdb, config)

	test := helpers.GenerateHandleTester(t, handler)

	w := test("POST", "", url.Values{}, nil)

	if w.Code != http.StatusOK {
		t.Errorf("NodeEnrollRequest returned: %+v, expected: %+v", w.Code, http.StatusOK)
	}
}
