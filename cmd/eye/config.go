/*
 * Copyright (c) 2016, 1&1 Internet SE
 * Written by Jörg Pernfuß <joerg.pernfuss@1und1.de>
 * All rights reserved.
 */

package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/url"

	"github.com/asaskevich/govalidator"
	"github.com/nahanni/go-ucl"
)

type EyeConfig struct {
	Environment string     `json:"environment" valid:"alpha"`
	ReadOnly    bool       `json:"readonly,string" valid:"-"`
	Volatile    bool       `json:"volatile,string" valid:"-"`
	Daemon      EyeDaemon  `json:"daemon" valid:"required"`
	Database    DbConfig   `json:"database" valid:"required"`
	Soma        SomaConfig `json:"soma" valid:"required"`
	run         EyeRuntime
}

type DbConfig struct {
	Host    string `json:"host" valid:"dns"`
	User    string `json:"user" valid:"alphanum"`
	Name    string `json:"name" valid:"alphanum"`
	Port    string `json:"port" valid:"port"`
	Pass    string `json:"password" valid:"-"`
	Timeout string `json:"timeout" valid:"numeric"`
	TLSMode string `json:"tlsmode" valid:"alpha"`
}

type SomaConfig struct {
	url     *url.URL
	Address string `json:"address" valid:"requrl"`
}

type EyeDaemon struct {
	url    *url.URL
	Listen string `json:"listen" valid:"ip"`
	Port   string `json:"port" valid:"port"`
	TLS    bool   `json:"tls,string" valid:"-"`
	Cert   string `json:"cert-file" valid:"optional"`
	Key    string `json:"key-file" valid:"optional"`
}

type EyeRuntime struct {
	conn         *sql.DB
	checkItem    *sql.Stmt
	updateItem   *sql.Stmt
	checkLookup  *sql.Stmt
	insertLookup *sql.Stmt
	insertItem   *sql.Stmt
	deleteItem   *sql.Stmt
	deleteLookup *sql.Stmt
	getLookup    *sql.Stmt
	itemCount    *sql.Stmt
	getConfig    *sql.Stmt
	getItems     *sql.Stmt
	retrieve     *sql.Stmt
}

func (c *EyeConfig) readConfigFile(fname string) error {
	file, err := ioutil.ReadFile(fname)
	if err != nil {
		return err
	}

	log.Printf("Loading configuration from %s", fname)

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

	govalidator.SetFieldsRequiredByDefault(true)
	if ok, err := govalidator.ValidateStruct(c); !ok {
		return err
	}
	c.Soma.url, _ = url.Parse(c.Soma.Address)
	log.Printf("Configured SOMA base address: %s\n", c.Soma.url.String())
	return nil
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
