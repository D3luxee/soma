package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/nahanni/go-ucl"
)

// Config struct
type Config struct {
	API        string       `json:"api"`
	Cert       string       `json:"cert"`
	LogDir     string       `json:"logdir"`
	Timeout    uint         `json:"timeout,string"`
	Activation string       `json:"activation.mode"`
	AsyncWait  bool         `json:"async.wait,string"`
	JobSave    bool         `json:"save.jobs,string"`
	ProcJSON   string       `json:"json.output.processor"`
	Auth       AuthConfig   `json:"auth"`
	AdminAuth  AuthConfig   `json:"admin.auth"`
	BoltDB     ConfigBoltDB `json:"boltdb"`
	Run        RunTimeConfig
}

// AuthConfig struct
type AuthConfig struct {
	User  string `json:"user"`
	Pass  string `json:"pass"`
	Token string `json:"token"`
}

// ConfigBoltDB struct
type ConfigBoltDB struct {
	Path    string `json:"path"`
	File    string `json:"file"`
	Mode    string `json:"mode"`
	Timeout uint   `json:"open.timeout,string"`
}

// RunTimeConfig struct
type RunTimeConfig struct {
	SomaAPI       *url.URL      `json:"-"`
	PathLogs      string        `json:"-"`
	PathBoltDB    string        `json:"-"`
	ModeBoltDB    uint64        `json:"-"`
	CertPath      string        `json:"-"`
	TimeoutBoltDB time.Duration `json:"-"`
	TimeoutResty  time.Duration `json:"-"`
	Logger        *log.Logger   `json:"-"`
}

func (c *Config) populateFromFile(fname string) error {
	file, err := ioutil.ReadFile(fname)
	if err != nil {
		return err
	}

	// UCL parses into map[string]interface{}
	fileBytes := bytes.NewBuffer([]byte(file))
	parser := ucl.NewParser(fileBytes)
	uclData, err := parser.Ucl()
	if err != nil {
		fmt.Fprintf(os.Stderr, "UCL error: %s\n", err.Error())
		os.Exit(1)
	}

	// take detour via JSON to load UCL into struct
	uclJSON, err := json.Marshal(uclData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "JSON marshal error: %s\n", err.Error())
		os.Exit(1)
	}
	json.Unmarshal([]byte(uclJSON), &c)

	return nil
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
