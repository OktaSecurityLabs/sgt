package osquery_types

import (
	"testing"
	"reflect"
)

var (
	q1 = PackQuery{
		QueryName: "testquery",
		Query: "select * from test;",
		Interval: "60",
		Version: "1.4.0",
		Description: "test description",
		Value: "a test",
		Snapshot: "true",
	}
	p1 = Pack{
		PackName: "test pack",
		Queries: []PackQuery{q1,},
	}

)
//removing function
/*
func TestPackQuery_AsString(t *testing.T) {
	expected := fmt.Sprintf(`%q: {"query": %q, "interval": %q, "version": %q, "description": %q, "value": %q, "snapshot": %q}`,
		"testquery", "select * from test;", "60", "1.4.0", "test description", "a test", "true")
	if q1.AsString() != expected {
		t.Errorf("Got: \n\t%s, expected: \n\t%s", q1.AsString(), expected)
	}
}*/

func TestPack_AsMap(t *testing.T) {
	packMap := p1.AsMap()
	expectedMap := map[string]map[string]map[string]string{
		"queries": {
			"testquery": {
				"query": "select * from test;",
				"interval": "60",
				"version": "1.4.0",
				"description": "test description",
				"value": "a test",
				"snapshot": "true",
			},
		},
	}
	eq := reflect.DeepEqual(packMap, expectedMap)
	if !eq {
		t.Errorf("maps not equal")
	}
}



