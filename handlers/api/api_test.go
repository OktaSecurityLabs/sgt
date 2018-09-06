package api

import (
	"github.com/oktasecuritylabs/sgt/handlers/helpers"
	"net/http"
	"net/url"
	"testing"
)

func TestGetNamedConfigsHandler(t *testing.T) {
	mockdb := helpers.NewMockDB()

	handler := GetNamedConfigsHandler(mockdb)

	test := helpers.GenerateHandleTester(t, handler)

	w := test("GET", url.Values{})

	if w.Code != http.StatusOK {
		t.Errorf("Add records did not return %+v", http.StatusOK)
	}

}

func TestConfigurationRequestHandler(t *testing.T) {

	mockdb := helpers.NewMockDB()
	handler := ConfigurationRequestHandler(mockdb)
	test := helpers.GenerateHandleTester(t, handler)

	v := url.Values{}
	v.Add("config_name", "default")
	w := test("POST", v)

	if w.Code != http.StatusOK {
		t.Errorf("Add records did not return %+v", http.StatusOK)
	}
}

func TestGetNodesHandler(t *testing.T) {
	mockdb := helpers.NewMockDB()
	handler := GetNodesHandler(mockdb)
	test := helpers.GenerateHandleTester(t, handler)

	w := test("GET", url.Values{})

	if w.Code != http.StatusOK {
		t.Errorf("Add records did not return %+v", http.StatusOK)
	}

}
