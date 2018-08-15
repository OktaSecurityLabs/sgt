package helpers

import (
	"net/url"
	"net/http/httptest"
	"testing"
	"net/http"
	"strings"
)

type HandleTester func(
	method string,
	params url.Values,
) *httptest.ResponseRecorder

// Given the current test runner and an http.Handler, generate a
// HandleTester which will test its given input against the
// handler.

func GenerateHandleTester(
	t *testing.T,
	handleFunc http.Handler,
) HandleTester {

	// Given a method type ("GET", "POST", etc) and
	// parameters, serve the response against the handler and
	// return the ResponseRecorder.

	return func(
		method string,
		params url.Values,
	) *httptest.ResponseRecorder {

		req, err := http.NewRequest(
			method,
			"",
			strings.NewReader(params.Encode()),
		)
		if err != nil {
			t.Errorf("%v", err)
		}
		req.Header.Set(
			"Content-Type",
			"application/x-www-form-urlencoded; param=value",
		)
		w := httptest.NewRecorder()
		handleFunc.ServeHTTP(w, req)
		return w
	}
}