package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"sync"

	billy "gopkg.in/src-d/go-billy.v4"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/storage"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

const (
	ServerPort             = "7777"
	MigrationRepositoryUrl = "https://github.com/camilocot/cassandra-migrations"
)

type CassandraMigrationConfig struct {
	UserName string `json:"userName"`
	Password string `json:"password"`
}

// CheckIfError should be used to naively panics if an error is not nil.
func CheckIfError(err error) {
	if err == nil {
		return
	}

	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
	os.Exit(1)
}

// Info should be used to describe the commands that are about to run.
func Info(format string, args ...interface{}) {
	fmt.Printf("\x1b[34;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
}

// Warning should be used to display a warning
func Warning(format string, args ...interface{}) {
	fmt.Printf("\x1b[36;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
}

type CassandraMigration struct {
	// Parallel executions of Casssandra Migrations could cause inconsistencies in th DB
	// @TODO: implement a lock per Keyspace instead of blocking the entirely command execution
	sync.Mutex
	command    string
	args       []string
	repository git.Repository
}

func (c *CassandraMigration) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, `{"alive": true}`)
}

func (c *CassandraMigration) ExecuteHandler(w http.ResponseWriter, r *http.Request) {
	binary, lookErr := exec.LookPath(c.command)
	if lookErr != nil {
		panic(lookErr)
	}

	c.Lock()
	output, execErr := exec.Command(binary, c.args...).CombinedOutput()
	c.Unlock()

	if execErr != nil {
		panic(execErr)
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	io.WriteString(w, `{"output": "`+string(output)+`"}`)
}

func (c *CassandraMigration) HandleRequests() {
	http.HandleFunc("/health", c.HealthCheckHandler)
	http.HandleFunc("/run", c.ExecuteHandler)
}

type Clone func(storage.Storer, billy.Filesystem, *git.CloneOptions) (*git.Repository, error)

func (c *CassandraMigration) RepositoryClone(clone Clone) {

	Info("git clone " + MigrationRepositoryUrl)

	repository, err := clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL: MigrationRepositoryUrl,
	})

	CheckIfError(err)

	c.repository = *repository
}

func GetConfigJson(url string, target interface{}) error {
	res, err := http.Get(url)
	CheckIfError(err)
	defer res.Body.Close()

	return json.NewDecoder(res.Body).Decode(target)
}

func UnmarshalBody(body []byte) (m map[string]interface{}) {
	err := json.Unmarshal(body, &m)
	CheckIfError(err)
	return
}

func main() {

	c := CassandraMigration{
		command: "echo",
		args:    []string{"-n", "test"},
	}

	c.HandleRequests()
	c.RepositoryClone(git.Clone)

	fmt.Printf("Starting server for testing HTTP POST...\n")
	err := http.ListenAndServe(":"+ServerPort, nil)
	CheckIfError(err)
}
