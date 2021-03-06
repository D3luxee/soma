/*-
 * Copyright (c) 2016, 1&1 Internet SE
 * Copyright (c) 2016, Jörg Pernfuß
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package tree

import "fmt"

type Error struct {
	Action string `json:",omitempty"`
	Text   string `json:",omitempty"`
}

func (e Error) String() string {
	return fmt.Sprintf("Tree error during %s action", e.Action)
}

func (e Error) Error() error {
	return fmt.Errorf("Tree error during %s action", e.Action)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
