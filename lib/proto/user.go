/*-
 * Copyright (c) 2015-2016, 1&1 Internet SE
 * Copyright (c) 2015-2016, Jörg Pernfuß <joerg.pernfuss@1und1.de>
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package proto // import "github.com/mjolnir42/soma/lib/proto"

// User describes a user
type User struct {
	ID             string           `json:"id,omitempty"`
	UserName       string           `json:"userName,omitempty"`
	FirstName      string           `json:"firstName,omitempty"`
	LastName       string           `json:"lastName,omitempty"`
	EmployeeNumber string           `json:"employeeNumber,omitempty"`
	MailAddress    string           `json:"mailAddress,omitempty"`
	IsActive       bool             `json:"isActive,omitempty"`
	IsSystem       bool             `json:"isSystem,omitempty"`
	IsDeleted      bool             `json:"isDeleted,omitempty"`
	TeamID         string           `json:"teamId,omitempty"`
	Details        *UserDetails     `json:"details,omitempty"`
	Credentials    *UserCredentials `json:"credentials,omitempty"`
}

type UserCredentials struct {
	Reset          bool   `json:"reset,omitempty"`
	ForcedPassword string `json:"forcedPassword,omitempty"`
}

type UserDetails struct {
	Creation       *DetailsCreation `json:"creation,omitempty"`
	DictionaryID   string           `json:"dictionaryID,omitempty"`
	DictionaryName string           `json:"dictionaryName,omitempty"`
}

type UserFilter struct {
	UserName  string `json:"userName,omitempty"`
	IsActive  bool   `json:"isActive,omitempty"`
	IsSystem  bool   `json:"isSystem,omitempty"`
	IsDeleted bool   `json:"isDeleted,omitempty"`
}

func NewUserRequest() Request {
	return Request{
		Flags: &Flags{},
		User:  &User{},
	}
}

func NewUserFilter() Request {
	return Request{
		Filter: &Filter{
			User: &UserFilter{},
		},
	}
}

func NewUserResult() Result {
	return Result{
		Errors: &[]string{},
		Users:  &[]User{},
	}
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
