/*-
 * Copyright (c) 2016, Jörg Pernfuß <joerg.pernfuss@1und1.de>
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package msg

import "github.com/1and1/soma/lib/proto"

// Authorization struct
type Authorization struct {
	User         string
	RemoteAddr   string
	Section      string
	Action       string
	TeamID       string
	OncallID     string
	MonitoringID string
	CapabilityID string
	RepositoryID string
	BucketID     string
	GroupID      string
	ClusterID    string
	NodeID       string
	Grant        *proto.Grant
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
