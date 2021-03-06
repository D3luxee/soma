/*
Copyright (c) 2016, Jörg Pernfuß <code.jpe@gmail.com>
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

* Redistributions of source code must retain the above copyright notice, this
  list of conditions and the following disclaimer.

* Redistributions in binary form must reproduce the above copyright notice,
  this list of conditions and the following disclaimer in the documentation
  and/or other materials provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/asaskevich/govalidator"
	"github.com/julienschmidt/httprouter"
)

// global variables
var conn *sql.DB
var Eye EyeConfig
var somaVersion string

func main() {
	var (
		configFlag, configFile string
		err                    error
		versionFlag            bool
	)
	flag.StringVar(&configFlag, "config", "/srv/eye/conf/eye.conf", "Configuration file location")
	flag.BoolVar(&versionFlag, `version`, false, `Print version information`)
	flag.Parse()

	if versionFlag {
		version()
	}

	log.Printf("Starting runtime config initialization, Eye v%s", somaVersion)
	/*
	 * Read configuration file
	 */
	if configFile, err = filepath.Abs(configFlag); err != nil {
		log.Fatal(err)
	}
	if configFile, err = filepath.EvalSymlinks(configFile); err != nil {
		log.Fatal(err)
	}
	err = Eye.readConfigFile(configFile)
	if err != nil {
		log.Fatal(err)
	}

	/*
	 * Construct listen address
	 */
	Eye.Daemon.url = &url.URL{}
	Eye.Daemon.url.Host = fmt.Sprintf("%s:%s", Eye.Daemon.Listen, Eye.Daemon.Port)
	if Eye.Daemon.TLS {
		Eye.Daemon.url.Scheme = "https"
		if ok, ptype := govalidator.IsFilePath(Eye.Daemon.Cert); !ok {
			log.Fatal("Missing required certificate configuration config/daemon/cert-file")
		} else {
			if ptype != govalidator.Unix {
				log.Fatal("config/daemon/cert-File: valid Windows paths are not helpful")
			}
		}
		if ok, ptype := govalidator.IsFilePath(Eye.Daemon.Key); !ok {
			log.Fatal("Missing required key configuration config/daemon/key-file")
		} else {
			if ptype != govalidator.Unix {
				log.Fatal("config/daemon/key-file: valid Windows paths are not helpful")
			}
		}
	} else {
		Eye.Daemon.url.Scheme = "http"
	}

	/*
	 * Initialize database
	 */
	connectToDatabase()
	prepareStatements()
	// Close() must be deferred here since it triggers on function exit
	defer Eye.run.checkItem.Close()
	defer Eye.run.checkLookup.Close()
	defer Eye.run.deleteItem.Close()
	defer Eye.run.deleteLookup.Close()
	defer Eye.run.getConfig.Close()
	defer Eye.run.getItems.Close()
	defer Eye.run.getLookup.Close()
	defer Eye.run.insertItem.Close()
	defer Eye.run.insertLookup.Close()
	defer Eye.run.itemCount.Close()
	defer Eye.run.retrieve.Close()
	defer Eye.run.updateItem.Close()
	go pingDatabase()

	/*
	 * Register http handlers
	 */
	router := httprouter.New()
	router.GET("/api/v1/configuration/:lookup", RetrieveConfigurationItems)
	router.GET("/api/v1/item/", ListConfigurationItems)
	router.POST("/api/v1/item/", AddConfigurationItem)
	router.GET("/api/v1/item/:item", GetConfigurationItem)
	router.PUT("/api/v1/item/:item", UpdateConfigurationItem)
	router.DELETE("/api/v1/item/:item", DeleteConfigurationItem)
	router.POST("/api/v1/notify/", FetchConfigurationItems)
	router.POST("/api/v1/notify", FetchConfigurationItems)

	if Eye.Daemon.TLS {
		log.Fatal(http.ListenAndServeTLS(
			Eye.Daemon.url.Host,
			Eye.Daemon.Cert,
			Eye.Daemon.Key,
			router))
	} else {
		log.Fatal(http.ListenAndServe(Eye.Daemon.url.Host, router))
	}
}

func version() {
	fmt.Fprintln(os.Stderr, `Eye Configuration Lookup Service`)
	fmt.Fprintf(os.Stderr, "Version: %s\n", somaVersion)
	os.Exit(0)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
