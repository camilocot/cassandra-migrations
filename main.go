package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
)

const (
	ServerPort = "7777"
)

func CheckErr(err error) {
	if err != nil {
		panic(err)
	}
}

type CassandraMigration struct {
	called  bool
	command string
	args    []string
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

	output, execErr := exec.Command(binary, c.args...).CombinedOutput()
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

func main() {
	c := CassandraMigration{
		called:  false,
		command: "echo",
		args:    []string{"-n", "test"},
	}

	c.HandleRequests()

	fmt.Printf("Starting server for testing HTTP POST...\n")
	if err := http.ListenAndServe(":"+ServerPort, nil); err != nil {
		log.Fatal(err)
	}
}
