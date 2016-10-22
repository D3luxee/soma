/*-
 * Copyright (c) 2016, Jörg Pernfuß <joerg.pernfuss@1und1.de>
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package stmt

const UnitVerify = `
SELECT metric_unit
FROM   soma.metric_units
WHERE  metric_unit = $1::varchar;`

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix