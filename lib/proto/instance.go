/*-
 * Copyright (c) 2016, 1&1 Internet SE
 * Copyright (c) 2016, Jörg Pernfuß <joerg.pernfuss@1und1.de>
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package proto

type Instance struct {
	ID               string               `json:"id,omitempty"`
	Version          uint64               `json:"version"`
	CheckID          string               `json:"checkId,omitempty"`
	ConfigID         string               `json:"configId,omitempty"`
	InstanceConfigID string               `json:"instanceConfigID,omitempty"`
	RepositoryID     string               `json:"repositoryId,omitempty"`
	BucketID         string               `json:"bucketId,omitempty"`
	ObjectID         string               `json:"objectId,omitempty"`
	ObjectType       string               `json:"objectType,omitempty"`
	CurrentStatus    string               `json:"currentStatus,omitempty"`
	NextStatus       string               `json:"nextStatus,omitempty"`
	IsInherited      bool                 `json:"isInherited"`
	Info             *InstanceVersionInfo `json:"instanceVersionInfo,omitempty"`
	Deployment       *Deployment          `json:"deployment,omitempty"`
}

type InstanceVersionInfo struct {
	CreatedAt           string `json:"createdAt"`
	ActivatedAt         string `json:"activatedAt"`
	DeprovisionedAt     string `json:"deprovisionedAt"`
	StatusLastUpdatedAt string `json:"statusLastUpdatedAt"`
	NotifiedAt          string `json:"notifiedAt"`
}

func NewInstanceResult() Result {
	return Result{
		Errors:    &[]string{},
		Instances: &[]Instance{},
	}
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
