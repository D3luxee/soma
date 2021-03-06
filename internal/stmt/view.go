/*-
 * Copyright (c) 2016, 1&1 Internet SE
 * Copyright (c) 2016, Jörg Pernfuß <joerg.pernfuss@1und1.de>
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package stmt

const (
	ViewStatements = ``

	ViewVerify = `
SELECT view
FROM   soma.views
WHERE  view = $1::varchar;`

	ViewList = `
SELECT view
FROM   soma.views;`

	ViewShow = ViewVerify

	ViewAdd = `
INSERT INTO soma.views (
            view)
SELECT   $1::varchar
WHERE    NOT EXISTS (
    SELECT  view
    FROM    soma.views
    WHERE   view = $1::varchar);`

	ViewDel = `
DELETE FROM soma.views
WHERE  view = $1::varchar;`

	ViewRename = `
UPDATE soma.views
SET    view = $1::varchar
WHERE  view = $2::varchar;`
)

func init() {
	m[ViewAdd] = `ViewAdd`
	m[ViewDel] = `ViewDel`
	m[ViewList] = `ViewList`
	m[ViewRename] = `ViewRename`
	m[ViewShow] = `ViewShow`
	m[ViewVerify] = `ViewVerify`
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
