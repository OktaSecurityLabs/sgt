package response

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteResponseJSONValid(t *testing.T) {

	writer := httptest.NewRecorder()

	response := map[string]string{"foo": "bar"}
	writeResponseJSON(writer, response)

	result := writer.Result()

	if result.StatusCode != 200 {
		t.Error("Incorrect status code")
	}

	if result.Header.Get("Content-Type") != "application/json" {
		t.Error("Incorrect Content-Type in header")
	}

	body, err := ioutil.ReadAll(result.Body)
	defer result.Body.Close()
	if err != nil {
		t.Error(err)
	}

	if string(body) != `{"foo":"bar"}` {
		t.Error("Incorrect body content")
	}
}

func TestWriteResponseJSONInvalid(t *testing.T) {

	writer := httptest.NewRecorder()

	writeResponseJSON(writer, make(chan int))

	result := writer.Result()

	if result.StatusCode != 500 {
		t.Error("Incorrect status code")
	}

	if result.Header.Get("Content-Type") != "text/plain; charset=utf-8" {
		t.Error("Incorrect Content-Type in header")
	}

	body, err := ioutil.ReadAll(result.Body)
	defer result.Body.Close()
	if err != nil {
		t.Error(err)
	}

	var errString = "response failed to marshal to json: json: unsupported type: chan int\n"
	if string(body) != errString {
		t.Error("Incorrect body content")
	}
}

// FakeWriter implements http.ResponseWriter but
// will produce an error on Write()
type FakeWriter struct {
	headers http.Header
	status  int
}

func (f *FakeWriter) Header() http.Header {
	return f.headers
}

func (FakeWriter) Write(b []byte) (int, error) {
	return 0, errors.New("bad thing happened")
}

func (f *FakeWriter) WriteHeader(statusCode int) {
	f.status = statusCode
}

func TestWriteResponseJSONIOError(t *testing.T) {

	writer := &FakeWriter{}
	writer.headers = make(map[string][]string)

	writeResponseJSON(writer, "data")

	if writer.status != http.StatusInternalServerError {
		t.Error("Incorrect status code")
	}

	if writer.Header().Get("Content-Type") != "text/plain; charset=utf-8" {
		t.Error("Incorrect Content-Type in header")
	}

	if writer.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("Incorrect X-Content-Type-Options in header")
	}
}

func TestWriteError(t *testing.T) {

	writer := httptest.NewRecorder()

	WriteError(writer, "err message")

	result := writer.Result()

	if result.StatusCode != 200 {
		t.Error("Incorrect status code")
	}

	// t.Logf("%+v", result)

	body, _ := ioutil.ReadAll(result.Body)
	defer result.Body.Close()

	if string(body) != `{"message":"err message","status":"error"}` {
		t.Error("Incorrect body content")
	}
}

func TestWriteSuccess(t *testing.T) {

	writer := httptest.NewRecorder()

	WriteSuccess(writer, "")

	result := writer.Result()

	if result.StatusCode != 200 {
		t.Error("Incorrect status code")
	}

	body, _ := ioutil.ReadAll(result.Body)
	defer result.Body.Close()

	if string(body) != `{"status":"success"}` {
		t.Error("Incorrect body content")
	}
}

func TestWriteCustomJSON(t *testing.T) {

	writer := httptest.NewRecorder()

	output := map[string]int{"total": 100}
	WriteCustomJSON(writer, output)

	result := writer.Result()

	if result.StatusCode != 200 {
		t.Error("Incorrect status code")
	}

	body, _ := ioutil.ReadAll(result.Body)
	defer result.Body.Close()

	if string(body) != `{"total":100}` {
		t.Error("Incorrect body content")
	}
}
