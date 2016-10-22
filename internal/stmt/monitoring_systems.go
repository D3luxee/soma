/*-
 * Copyright (c) 2016, Jörg Pernfuß <joerg.pernfuss@1und1.de>
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package stmt

const ListAllMonitoringSystems = `
SELECT monitoring_id,
       monitoring_name
FROM   soma.monitoring_systems;`

const ListScopedMonitoringSystems = `
SELECT sms.monitoring_id,
       sms.monitoring_name
FROM   inventory.users iu
JOIN   soma.monitoring_system_users smsu
  ON   iu.organizational_team_id = smsu.organizational_team_id
JOIN   soma.monitoring_systems sms
  ON   smsu.monitoring_id = sms.monitoring_id
WHERE  iu.user_uid = $1::varchar
  AND  sms.monitoring_system_mode = 'private'
UNION
SELECT sms.monitoring_id,
       sms.monitoring_name
FROM   inventory.users iu
JOIN   soma.monitoring_systems sms
  ON   iu.organizational_team_id = sms.monitoring_owner_team
WHERE  iu.user_uid = $1::varchar
  AND  sms.monitoring_system_mode = 'private'
UNION
SELECT sms.monitoring_id,
       sms.monitoring_name
FROM   soma.monitoring_systems sms
WHERE  sms.monitoring_system_mode = 'public';`

const ShowMonitoringSystem = `
SELECT monitoring_id,
       monitoring_name,
       monitoring_system_mode,
       monitoring_contact,
       monitoring_owner_team,
       monitoring_callback_uri
FROM   soma.monitoring_systems
WHERE  monitoring_id = $1::uuid;`

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix