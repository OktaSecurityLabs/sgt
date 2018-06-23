package helpers

import (
	"github.com/oktasecuritylabs/sgt/osquery_types"
	"encoding/json"
)


type MockDB struct {

}

func NewMockDB() *MockDB {
	db := MockDB{}
	return &db
}

var testPackQuery1 = osquery_types.PackQuery{
	QueryName: "test1",
	Query: "select * from users;",
	Interval: "60",
	Version: "1.1.1.1",
	Description: "test1 description",
	Value: "some value",
	}
var testPackQuery2 = osquery_types.PackQuery{
	"test2",
	"select * from installed_packages",
	"60",
	"1.1.1",
	"test2 description",
	"some value",
	"true",
	}
	var testQueryPack1 = osquery_types.QueryPack{
		"test-pack",
		[]string{"select * from users"},
	}
var testUser1 = osquery_types.User{
	"testuser1",
	[]byte("password"),
	"user",
}
var testClient1 = osquery_types.OsqueryClient{
	"host1",
	"3lkjsdf0jdfoiasdjf",
	false,
	map[string]map[string]string{},
	false,
	[]string{"a", "b"},
	"default",
	"default",
	"erlkjer",
}
var testDistributedQuery = osquery_types.DistributedQuery{
	"dlfkjadflikjerkj",
	[]string{"select * from users;"},
	false,

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

func (m MockDB) APIGetPackQueries() ([]osquery_types.PackQuery, error) {
	results := []osquery_types.PackQuery{
		testPackQuery1,
		testPackQuery2,
		}
	return results, nil
}

func (m MockDB) APISearchPackQueries(searchString string) ([]osquery_types.PackQuery, error) {
	results := []osquery_types.PackQuery{
		testPackQuery1,
		testPackQuery2,
	}
	return results, nil
}

func (m MockDB) AppendDistributedQuery(dq osquery_types.DistributedQuery) (error) {
	return nil
}

func (m MockDB) ApprovePendingNode(nodeKey string) (error) {
	return nil
}

func (m MockDB) DeleteDistributedQuery(dq osquery_types.DistributedQuery) (error) {
	return nil
}

func (m MockDB) DeleteQueryPack(queryPackName string) (error) {
	return nil
}

func (m MockDB) GetPackByName(packName string) (osquery_types.Pack, error) {
	p := osquery_types.Pack{
		"pack1",
		[]osquery_types.PackQuery{testPackQuery1},
	}
	return p, nil
}

func (m MockDB) GetPackQuery(queryName string) (osquery_types.PackQuery, error) {
	return testPackQuery1, nil
}

func (m MockDB) GetUser(username string) (osquery_types.User, error) {
	return testUser1, nil
}

func (m MockDB) NewDistributedQuery(dq osquery_types.DistributedQuery) (error) {
	return nil
}

func (m MockDB) NewQueryPack(qp osquery_types.QueryPack) (error) {
	return nil
}

func (m MockDB) NewUser(u osquery_types.User) (error) {
	return nil
}

func (m MockDB) SearchByHostIdentifier(hid string) ([]osquery_types.OsqueryClient, error) {
	return []osquery_types.OsqueryClient{testClient1}, nil
}

func (m MockDB) SearchByNodeKey(nk string) (osquery_types.OsqueryClient, error) {
	return testClient1, nil
}

func (m MockDB) SearchDistributedNodeKey(nk string) (osquery_types.DistributedQuery, error) {
	return testDistributedQuery, nil
}

func (m MockDB) SearchQueryPacks(searchString string) ([]osquery_types.QueryPack, error) {
	return []osquery_types.QueryPack{testQueryPack1}, nil
}

func (m MockDB) UpsertClient(oc osquery_types.OsqueryClient) (error) {
	return nil
}

func (m MockDB) UpsertDistributedQuery(dq osquery_types.DistributedQuery) (error) {
	return nil
}

func (m MockDB) UpsertPackQuery(pq osquery_types.PackQuery) (error)  {
	return nil
}

func (m MockDB) UpsertPack(qp osquery_types.QueryPack) (error) {
	return nil
}

func (m MockDB) ValidNode(nodeKey string) (error) {
	return nil
}

func (m MockDB) BuildOsqueryPackAsJSON(nc osquery_types.OsqueryNamedConfig) (json.RawMessage) {
	return json.RawMessage{}
}

func (m MockDB) BuildNamedConfig(configName string) (osquery_types.OsqueryNamedConfig, error) {
	return osquery_types.OsqueryNamedConfig{}, nil
}

