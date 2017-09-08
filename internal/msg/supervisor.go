/*-
 * Copyright (c) 2016, Jörg Pernfuß <joerg.pernfuss@1und1.de>
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package msg

import (
	"github.com/mjolnir42/soma/lib/auth"
	"github.com/mjolnir42/soma/lib/proto"
)

type Supervisor struct {
	Verdict            uint16
	RestrictedEndpoint bool
	// KeyExchange Data
	Kex auth.Kex
	// Fields for encrypted requests
	Encrypted struct {
		KexID string
		Data  []byte
	}
	// Fields for basic authentication requests
	BasicAuth struct {
		User  string
		Token string
	}
	// XXX Everything below is deprecated
	// Fields for permission authorization requests
	Request        *Authorization
	PermAction     string //XXX
	PermRepository string //XXX
	PermMonitoring string //XXX
	PermNode       string //XXX
	// Fields for map update notifications
	Object string
	User   proto.User
	Team   proto.Team
	// Fields for Grant revocation
	GrantId string
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
