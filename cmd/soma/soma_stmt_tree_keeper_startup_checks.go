package main

const tkStmtLoadChecks = `
SELECT check_id,
       bucket_id,
       source_check_id,
       source_object_type,
       source_object_id,
       configuration_id,
       capability_id,
       object_id,
       object_type
FROM   soma.checks
WHERE  repository_id = $1::uuid
AND    check_id = source_check_id
AND    source_object_type = $2::varchar
AND    NOT deleted;`

const tkStmtLoadInheritedChecks = `
SELECT check_id,
       object_id,
       object_type
FROM   soma.checks
WHERE  repository_id = $1::uuid
AND    source_check_id = $2::uuid
AND    source_check_id != check_id
AND    NOT deleted;`

const tkStmtLoadChecksForType = `
SELECT check_id,
       object_id
FROM   soma.checks
WHERE  repository_id = $1::uuid
AND    object_type = $2::varchar
AND    NOT deleted;`

const tkStmtLoadCheckConfiguration = `
SELECT bucket_id,
       configuration_name,
       configuration_object,
       configuration_object_type,
       configuration_active,
       inheritance_enabled,
       children_only,
       capability_id,
       interval,
       enabled,
       external_id
FROM   soma.check_configurations
WHERE  configuration_id = $1::uuid
AND    repository_id = $2::uuid
AND    NOT deleted;`

const tkStmtLoadAllCheckConfigurationsForType = `
SELECT configuration_id,
       bucket_id,
       configuration_name,
       configuration_object,
       inheritance_enabled,
       children_only,
       capability_id,
       interval,
       enabled,
       external_id
FROM   soma.check_configurations
WHERE  configuration_object_type = $1::varchar
AND    repository_id = $2::uuid
AND    NOT deleted;`

const tkStmtLoadCheckThresholds = `
SELECT sct.predicate,
       sct.threshold,
       snl.level_name,
       snl.level_shortname,
       snl.level_numeric
FROM   soma.configuration_thresholds sct
JOIN   soma.notification_levels snl
ON     sct.notification_level = snl.level_name
WHERE  configuration_id = $1::uuid;`

const tkStmtLoadCheckConstraintCustom = `
SELECT sccp.custom_property_id,
       scp.custom_property,
       sccp.property_value
FROM   soma.constraints_custom_property sccp
JOIN   soma.custom_properties scp
ON     sccp.custom_property_id = scp.custom_property_id
AND    sccp.repository_id = scp.repository_id
WHERE  configuration_id = $1::uuid;`

// do not get distracted by the squirrels! All constraint
// statements are constructed to use three result variables,
// so they can be loaded in one unified loop.
const tkStmtLoadCheckConstraintNative = `
SELECT native_property,
       property_value,
       'squirrel'
FROM   soma.constraints_native_property
WHERE  configuration_id = $1::uuid;`

// return configuration id: every constraint query has 2 columns
const tkStmtLoadCheckConstraintOncall = `
SELECT scop.oncall_duty_id,
       oncall_duty_name,
       oncall_duty_phone_number
FROM   soma.constraints_oncall_property scop
JOIN   inventory.oncall_duty_teams iodt
ON     scop.oncall_duty_id = iodt.oncall_duty_id
WHERE  scop.configuration_id = $1::uuid;`

const tkStmtLoadCheckConstraintAttribute = `
SELECT service_property_attribute,
       attribute_value,
       'squirrel'
FROM   soma.constraints_service_attribute
WHERE  configuration_id = $1::uuid;`

const tkStmtLoadCheckConstraintService = `
SELECT organizational_team_id,
       service_property,
       'squirrel'
FROM   soma.constraints_service_property
WHERE  configuration_id = $1::uuid;`

const tkStmtLoadCheckConstraintSystem = `
SELECT system_property,
       property_value,
       'squirrel'
FROM   soma.constraints_system_property
WHERE  configuration_id = $1::uuid;`

const tkStmtLoadCheckInstances = `
SELECT check_instance_id,
       check_configuration_id
FROM   soma.check_instances
WHERE  check_id = $1::uuid
AND    NOT deleted;`

// load the most recent configuration for this instance, which is
// not always the current one, since a newer version could be blocked
// by the current versions rollout
const tkStmtLoadCheckInstanceConfiguration = `
SELECT check_instance_config_id,
       version,
       monitoring_id,
       constraint_hash,
       constraint_val_hash,
       instance_service,
       instance_service_cfg_hash,
       instance_service_cfg
FROM   soma.check_instance_configurations
WHERE  check_instance_id = $1::uuid
ORDER  BY created DESC
LIMIT  1;`

const tkStmtLoadCheckGroupState = `
SELECT sg.group_id,
       sg.object_state
FROM   soma.buckets sb
JOIN   soma.groups  sg
ON     sb.bucket_id = sg.bucket_id
WHERE  sb.repository_id = $1::uuid;`

const tkStmtLoadCheckGroupRelations = `
SELECT sgmg.group_id,
       sgmg.child_group_id
FROM   soma.buckets sb
JOIN   soma.group_membership_groups sgmg
ON     sb.bucket_id = sgmg.bucket_id
WHERE  sb.repository_id = $1::uuid;`

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix