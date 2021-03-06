package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"

	billy "gopkg.in/src-d/go-billy.v4"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/storage"
)

func TestExecuteHandler(t *testing.T) {

	req, err := http.NewRequest("GET", "/health-check", nil)
	if err != nil {
		t.Fatal(err)
	}

	c := CassandraMigration{
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
	http.DefaultServeMux = new(http.ServeMux)

	c := CassandraMigration{
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

type timeTrack struct {
	start   time.Time
	elapsed time.Duration
}

func (tt *timeTrack) SecondsElapsed() int {
	return int(tt.elapsed.Seconds())
}

func TestHandleConcurrentRequest(t *testing.T) {
	http.DefaultServeMux = new(http.ServeMux)
	sleepDuration := 1

	c := CassandraMigration{
		command: "sleep",
		args:    []string{strconv.Itoa(sleepDuration)},
	}

	var wg sync.WaitGroup

	rr := httptest.NewRecorder()

	t1 := timeTrack{
		start: time.Now(),
	}

	t2 := timeTrack{
		start: t1.start,
	}

	http.HandleFunc("/run", c.ExecuteHandler)

	req1, err := http.NewRequest("GET", "/run", nil)
	if err != nil {
		t.Fatal(err)
	}

	req2, err := http.NewRequest("GET", "/run", nil)
	if err != nil {
		t.Fatal(err)
	}

	wg.Add(1)
	go func(tt *timeTrack) {
		defer wg.Done()
		http.DefaultServeMux.ServeHTTP(rr, req1)
		tt.elapsed = time.Since(tt.start)
	}(&t1)

	wg.Add(1)
	go func(tt *timeTrack) {
		defer wg.Done()
		http.DefaultServeMux.ServeHTTP(rr, req2)
		tt.elapsed = time.Since(tt.start)
	}(&t2)

	wg.Wait()

	if (sleepDuration != t1.SecondsElapsed() && sleepDuration != t2.SecondsElapsed()) || (sleepDuration*2 != t1.SecondsElapsed() && sleepDuration*2 != t2.SecondsElapsed()) {
		t.Errorf("Concurrent requests were not run sequentially")
	}
}

func CloneMock(s storage.Storer, f billy.Filesystem, op *git.CloneOptions) (*git.Repository, error) {

	if op.URL != MigrationRepositoryUrl {
		return nil, errors.New("Wrong repository URL")
	}

	return &git.Repository{
		Storer: s,
	}, nil

}

func TestGitClone(t *testing.T) {
	c := CassandraMigration{}
	c.RepositoryClone(CloneMock)
	if &c.repository == nil {
		t.Errorf("Repository wasn't initialized")
	}
}

func stringInSlice(a string, list []interface{}) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func TestUnmarshalBody(t *testing.T) {
	body := UnmarshalBody([]byte(`{"Name":"Wednesday","Parents":["Morticia"]}`))

	f := interface{}(body).(map[string]interface{})

	if !stringInSlice("Morticia", f["Parents"].([]interface{})) {
		t.Errorf("Morticia not found in parents")
	}

	if f["Name"] != "Wednesday" {
		t.Errorf("Wednesday is not found")
	}
}

func TestGetConfigJson(t *testing.T) {
	expected := `{"userName":"fake username","password":"fake password"}`

	cc := CassandraMigrationConfig{}
	cs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, expected)
	}))
	defer cs.Close()

	err := GetConfigJson(cs.URL, &cc)
	CheckIfError(err)

	if cc.UserName != "fake username" {
		t.Errorf("Username unexpected: " + cc.UserName)
	}

	if cc.Password != "fake password" {
		t.Errorf("Password unexpected: " + cc.Password)
	}
}
