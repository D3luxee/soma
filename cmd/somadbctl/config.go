package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/nahanni/go-ucl"
)

type Config struct {
	Environment string   `json:"environment"`
	Timeout     string   `json:"timeout"`
	TLSMode     string   `json:"tlsmode"`
	Database    DbConfig `json:"database"`
}

type DbConfig struct {
	Host string `json:"host"`
	User string `json:"user"`
	Name string `json:"dbname"`
	Port string `json:"port"`
	Pass string `json:"password"`
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
		log.Fatal("UCL error: ", err)
	}

	// take detour via JSON to load UCL into struct
	uclJSON, err := json.Marshal(uclData)
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal([]byte(uclJSON), &c)

	return nil
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
