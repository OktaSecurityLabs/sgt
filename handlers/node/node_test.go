package node

import (
	"github.com/oktasecuritylabs/sgt/handlers/helpers"
	"net/http"
	"net/url"
	"testing"
)

func TestNodeEnrollRequest(t *testing.T) {
	mockdb := helpers.NewMockDB()

	handler := NodeEnrollRequest(mockdb)

	test := helpers.GenerateHandleTester(t, handler)

	w := test("POST", url.Values{})

	if w.Code != http.StatusOK {
		t.Errorf("NodeEnrollRequest returned: %+v, expected: %+v", w.Code, http.StatusOK)
	}
}
