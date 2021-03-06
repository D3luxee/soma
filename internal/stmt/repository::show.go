/*-
 * Copyright (c) 2018, Jörg Pernfuß
 * Copyright (c) 2018, 1&1 IONOS SE
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package stmt // import "github.com/mjolnir42/soma/internal/stmt"

const AuthorizedRepositoryShow = `
-- $1 section.name       ::varchar
-- $2 action.name        ::varchar
-- $3 user.uid           ::varchar
-- $4 repository.id      ::uuid
-- $5 repository.team_id ::uuid
-------------------------------
-- CASE1: root user has omnipotence permission
SELECT      soma.repository.id,
            soma.repository.name,
            soma.repository.is_deleted,
            soma.repository.is_active,
            soma.repository.team_id,
            creator.uid,
            soma.repository.created_at
FROM        inventory.user
JOIN        soma.authorizations_global
  ON        inventory.user.id = soma.authorizations_global.user_id
JOIN        soma.permission
  ON        soma.authorizations_global.permission_id = soma.permission.id
            -- unscoped, use carthesian product on all repositories
CROSS JOIN  soma.repository
JOIN        inventory.user AS creator
  ON        soma.repository.created_by = creator.id
WHERE       inventory.user.uid = $3::varchar
  AND       soma.authorizations_global.category = 'omnipotence'
  AND       soma.permission.name = 'omnipotence'
  AND       (   $1::varchar = 'repository'
             OR $1::varchar = 'repository-config')
  AND       $2::varchar = 'show'
  AND       soma.repository.id = $4::uuid
  AND       soma.repository.team_id = $5::uuid
UNION
-- CASE2: admin user has suitable system permission for requested section::action
SELECT      soma.repository.id,
            soma.repository.name,
            soma.repository.is_deleted,
            soma.repository.is_active,
            soma.repository.team_id,
            creator.uid,
            soma.repository.created_at
FROM        auth.admin
JOIN        soma.authorizations_global
  ON        auth.admin.id = soma.authorizations_global.admin_id
JOIN        soma.permission
  ON        soma.authorizations_global.permission_id = soma.permission.id
JOIN        soma.section
            -- system permissions have the category they grant as permission name
  ON        soma.permission.name = soma.section.category
JOIN        soma.action
  ON        soma.section.id = soma.action.section_id
            -- unscoped, use carthesian product on all repositories
CROSS JOIN  soma.repository
JOIN        inventory.user AS creator
  ON        soma.repository.created_by = creator.id
WHERE       auth.admin.uid = $3::varchar
  AND       auth.admin.is_active
  AND       soma.authorizations_global.category = 'system'
  AND       soma.section.name = $1::varchar
  AND       soma.action.name  = $2::varchar
  AND       (   $1::varchar = 'repository'
             OR $1::varchar = 'repository-config')
  AND       $2::varchar = 'show'
  AND       soma.repository.id = $4::uuid
  AND       soma.repository.team_id = $5::uuid
UNION
-- CASE 3:  regular user has repository scoped repository-config::search, which allows to find
--          that one repository
SELECT      soma.repository.id,
            soma.repository.name,
            soma.repository.is_deleted,
            soma.repository.is_active,
            soma.repository.team_id,
            creator.uid,
            soma.repository.created_at
FROM        inventory.user
JOIN        soma.authorizations_repository
            -- authorization could be on the user or inherited from the team
  ON        (   inventory.user.id      = soma.authorizations_repository.user_id
             OR inventory.user.team_id = soma.authorizations_repository.team_id)
JOIN        soma.permission_map
  ON        soma.authorizations_repository.permission_id = soma.permission_map.permission_id
JOIN        soma.section
  ON        soma.permission_map.section_id = soma.section.id
JOIN        soma.action
  ON        soma.section.id = soma.action.section_id
            -- grant must be scoped on target repository
JOIN        soma.repository
  ON        soma.authorizations_repository.repository_id = soma.repository.id
JOIN        inventory.user AS creator
  ON        soma.repository.created_by = creator.id
WHERE       inventory.user.uid = $3::varchar
  AND       inventory.user.is_active
  AND NOT   inventory.user.is_deleted
  AND       soma.section.name = $1::varchar
  AND       soma.action.name  = $2::varchar
  AND       $1::varchar = 'repository-config'
  AND       $2::varchar = 'show'
            -- section grant for all actions has soma.permission_map.action_id as NULL
  AND       (   soma.permission_map.action_id = soma.action.id
             OR soma.permission_map.action_id IS NULL                 )
  AND       soma.repository.id = $4::uuid
  AND       soma.repository.team_id = $5::uuid
UNION
-- CASE 4:  regular user has team scoped repository::search, which allows to find all
--          repositories owned by a team
SELECT      soma.repository.id,
            soma.repository.name,
            soma.repository.is_deleted,
            soma.repository.is_active,
            soma.repository.team_id,
            creator.uid,
            soma.repository.created_at
FROM        inventory.user
JOIN        soma.authorizations_team
            -- authorization could be on the user or inherited from the team
  ON        (   inventory.user.id       = soma.authorizations_team.user_id
             OR inventory.user.team_id  = soma.authorizations_team.team_id)
JOIN        soma.permission_map
  ON        soma.authorizations_team.permission_id = soma.permission_map.permission_id
JOIN        soma.section
  ON        soma.permission_map.section_id = soma.section.id
JOIN        soma.action
  ON        soma.section.id = soma.action.section_id
            -- grant must be scoped on target repository owner team
JOIN        soma.repository
  ON        soma.authorizations_team.authorized_team_id = soma.repository.team_id
JOIN        inventory.user AS creator
  ON        soma.repository.created_by = creator.id
WHERE       inventory.user.uid = $3::varchar
  AND       inventory.user.is_active
  AND NOT   inventory.user.is_deleted
  AND       soma.section.name = $1::varchar
  AND       soma.action.name  = $2::varchar
  AND       $1::varchar = 'repository'
  AND       $2::varchar = 'show'
            -- section grant for all actions has soma.permission_map.action_id as NULL
  AND       (   soma.permission_map.action_id = soma.action.id
             OR soma.permission_map.action_id IS NULL                 )
  AND       soma.repository.id = $4::uuid
  AND       soma.repository.team_id = $5::uuid;`

func init() {
	m[AuthorizedRepositoryShow] = `AuthorizedRepositoryShow`
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
