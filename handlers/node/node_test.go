package node

import (
	"testing"
	"net/http/httptest"
	"encoding/json"
	"bytes"
	"net/http"
)

func TestNodeEnrollRequestOK(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer ts.Close()
	postData := NodeConfigurePost{
		EnrollSecret: "test-secret",
		NodeKey: "",
		HostIdentifier: "dev.local",
	}
	req, _ := http.Request{
		URL:
	}
	js, err := json.Marshal(postData)
	if err != nil {
		t.Error(err)
	}

	if r.Method != "POST" {
		t.Errorf("Expected 'POST', got %s", r.Method)
		}

	}
}