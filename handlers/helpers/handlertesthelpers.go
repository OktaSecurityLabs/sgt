package helpers

import (
	"net/url"
	"net/http/httptest"
	"testing"
	"net/http"
	"io"
	"fmt"
)

type HandleTester func(method, url string, params url.Values, body io.Reader) *httptest.ResponseRecorder

// Given the current test runner and an http.Handler, generate a
// HandleTester which will test its given input against the
// handler.

func GenerateHandleTester(t *testing.T, handleFunc http.Handler) HandleTester {

	// Given a method type ("GET", "POST", etc) and
	// parameters, serve the response against the handler and
	// return the ResponseRecorder.

	return func(method, url string, params url.Values, body io.Reader) *httptest.ResponseRecorder {

		u := fmt.Sprintf("%s?%s", params.Encode())

		req, err := http.NewRequest(
			method,
			u,
			body,
		)
		if err != nil {
			t.Errorf("%v", err)
		}
		req.Header.Set(
			"Content-Type",
			"application/json",
		)
		w := httptest.NewRecorder()
		handleFunc.ServeHTTP(w, req)
		return w
	}
}