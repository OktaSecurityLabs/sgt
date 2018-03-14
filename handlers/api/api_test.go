package api

import (
	"testing"
	"github.com/oktasecuritylabs/sgt/osquery_types"
	"net/http"
	"net/http/httptest"
)

type MockDB struct {

}

func (m MockDB) GetNamedConfigs() ([]osquery_types.OsqueryNamedConfig, error) {
	results := []osquery_types.OsqueryNamedConfig{}
	nc := osquery_types.OsqueryNamedConfig{
		ConfigName: "test-config",

	}
	results = append(results, nc)
	return results, nil
}

func (m MockDB) GetNamedConfig(cn string) (osquery_types.OsqueryNamedConfig, error) {
	nc := osquery_types.OsqueryNamedConfig{
		ConfigName: "test-config",

	}
	return nc, nil
}

func (m MockDB) UpsertNamedConfig(nc *osquery_types.OsqueryNamedConfig) (error) {
	return nil
}


func TestGetNamedConfigsHandler(t *testing.T) {
	db := MockDB{}
	handler := GetNamedConfigsHandler(db)
	req, _ := http.NewRequest("GET", "/api/v1/configuration/configs", nil)

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Add records did not return %+v", http.StatusOK)
	}

}


func TestConfigurationRequestHandler(t *testing.T) {
	db := MockDB{}
	handler := ConfigurationRequestHandler(db)
	req, _ := http.NewRequest("POST", "/api/v1/configuration/configs/test-config", nil)

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Add records did not return %+v", http.StatusOK)
	}
}

func TestGetNodesHandler(t *testing.T) {
	db := MockDB{}
	handler := GetNodesHandler(db)
	req, _ := http.NewRequest("GET", "/api/v1/configuration/nodes", nil)

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Add records did not return %+v", http.StatusOK)
	}

}

