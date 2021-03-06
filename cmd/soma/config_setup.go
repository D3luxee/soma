package main

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/codegangsta/cli"
	"github.com/mitchellh/go-homedir"
)

func configSetup(c *cli.Context) error {

	home, err := homedir.Dir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not determine home directory: %s\n", err.Error())
		os.Exit(1)
	}
	var confPath string
	if c.GlobalIsSet("config") {
		confPath = path.Join(home, ".soma", "adm", c.GlobalString("config"))
	} else {
		confPath = path.Join(home, ".soma", "adm", "somaadm.conf")
	}

	// try loading a configuration file
	err = Cfg.populateFromFile(confPath)
	if err != nil {
		if c.GlobalIsSet("config") {
			// missing configuration file is only a error if set via cli
			fmt.Fprintf(os.Stderr, "Error opening config: %s\n", confPath)
			os.Exit(1)
		}
	}

	// check account activation mode
	if Cfg.Activation == `` {
		Cfg.Activation = `ldap`
	} else if Cfg.Activation != `ldap` && Cfg.Activation != `mailtoken` {
		fmt.Fprintln(os.Stderr, `Config setting activation.mode must be 'ldap' or 'mailtoken'.`)
	}

	// finish setting up runtime configuration
	params := []string{"api", "timeout", "user", "logdir", "dbdir"}

	for p := range params {
		// update configuration with cli argument overrides
		if c.GlobalIsSet(params[p]) {
			switch params[p] {
			case "user":
				Cfg.Auth.User = c.GlobalString(params[p])
			case "timeout":
				Cfg.Timeout = uint(c.GlobalInt(params[p]))
			case "host":
				Cfg.API = c.GlobalString(params[p])
			case "dbdir":
				Cfg.BoltDB.Path = c.GlobalString(params[p])
			case "logdir":
				Cfg.LogDir = c.GlobalString(params[p])
			}
			continue
		}
		// set default values for unset configuration parameters
		switch params[p] {
		case "timeout":
			if Cfg.Timeout == 0 {
				Cfg.Timeout = 2
			}
		case "logdir":
			if Cfg.LogDir == `` {
				Cfg.LogDir = "logs"
			}
		case "dbdir":
			if Cfg.BoltDB.Path == `` {
				Cfg.BoltDB.Path = "db"
			}
		}
	}

	Cfg.Run.PathLogs = path.Join(home, ".soma", "adm",
		Cfg.LogDir)

	Cfg.Run.PathBoltDB = path.Join(home, ".soma", "adm",
		Cfg.BoltDB.Path, Cfg.BoltDB.File)
	Cfg.Run.ModeBoltDB, err = strconv.ParseUint(Cfg.BoltDB.Mode, 8, 32)
	if err != nil {
		return fmt.Errorf(
			"Failed to parse configuration field boltdb.mode: "+
				"%s\n", err.Error())
	}
	Cfg.Run.TimeoutBoltDB = time.Duration(Cfg.BoltDB.Timeout) * time.Second
	Cfg.Run.TimeoutResty = time.Duration(Cfg.Timeout) * time.Second

	Cfg.Run.SomaAPI, err = url.Parse(Cfg.API)
	if err != nil {
		return fmt.Errorf(
			"Failed to parse SOMA API address: %s\n", err.Error())
	}

	if Cfg.Run.SomaAPI.Scheme == `https` {
		if Cfg.Cert == `` {
			return fmt.Errorf(
				`HTTPS API endpoint requires configured CA file`)
		}
		Cfg.Run.CertPath = path.Join(home, ".soma", "adm", Cfg.Cert)
	}

	return nil
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
