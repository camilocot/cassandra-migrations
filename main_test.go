package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExecuteHandler(t *testing.T) {

	req, err := http.NewRequest("GET", "/health-check", nil)
	if err != nil {
		t.Fatal(err)
	}

	c := CassandraMigration{
		called:  false,
		command: "echo",
		args:    []string{"-n", "test"},
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(c.ExecuteHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := `{"output": "test"}`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestHealthCheckHandler(t *testing.T) {
	c := CassandraMigration{
		called:  false,
		command: "echo",
		args:    []string{"-n", "test"},
	}

	req, err := http.NewRequest("GET", "/health-check", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(c.HealthCheckHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := `{"alive": true}`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func ExecuteHandlerMock(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, `called`)
}

func HealthCheckHandlerMock(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, `called`)
}

func TestHandleRequests(t *testing.T) {
	c := CassandraMigration{
		called:  false,
		command: "echo",
		args:    []string{"-n", "test"},
	}

	c.HandleRequests()

	req, err := http.NewRequest("GET", "/non-existent", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusNotFound)
	}

	req, err = http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr = httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	req, err = http.NewRequest("GET", "/run", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr = httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

}
