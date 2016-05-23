package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"strings"

	"gopkg.in/ldap.v2"
)

func validateLdapCredentials(user, password string) (bool, error) {
	var (
		conn *ldap.Conn
		err  error
		pem  []byte
	)

	addr := fmt.Sprintf("%s:%d", SomaCfg.Ldap.Address, SomaCfg.Ldap.Port)
	bindDN := strings.Join(
		[]string{
			strings.Join(
				[]string{
					SomaCfg.Ldap.Attribute,
					user,
				},
				`=`,
			),
			SomaCfg.Ldap.UserDN,
			SomaCfg.Ldap.BaseDN,
		},
		`,`,
	)

	if SomaCfg.Ldap.Tls {
		conf := &tls.Config{
			InsecureSkipVerify: SomaCfg.Ldap.SkipVerify,
			ServerName:         SomaCfg.Ldap.Address,
		}
		if SomaCfg.Ldap.Cert != "" {
			if pem, err = ioutil.ReadFile(SomaCfg.Ldap.Cert); err != nil {
				return false, err
			}
			conf.RootCAs = x509.NewCertPool()
			conf.RootCAs.AppendCertsFromPEM(pem)
		}
		conn, err = ldap.DialTLS(`tcp`, addr, conf)
	} else {
		conn, err = ldap.Dial(`tcp`, addr)
	}
	if err != nil {
		return false, err
	}
	defer conn.Close()

	// attempt bind
	err = conn.Bind(bindDN, password)
	if err != nil && ldap.IsErrorWithCode(err,
		ldap.LDAPResultInvalidCredentials) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
