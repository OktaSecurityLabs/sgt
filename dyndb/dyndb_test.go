package dyndb

import "github.com/oktasecuritylabs/sgt/osquery_types"

type MockDB struct {
}

func NewMockDB() *MockDB {
	db := MockDB{}
	return &db
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

func (m MockDB) UpsertNamedConfig(nc *osquery_types.OsqueryNamedConfig) error {
	return nil
}
