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
	ClusterStatements = ``

	ClusterList = `
SELECT cluster_id,
       cluster_name,
       bucket_id
FROM   soma.clusters;`

	ClusterShow = `
SELECT cluster_id,
       bucket_id,
       cluster_name,
       object_state,
       organizational_team_id
FROM   soma.clusters
WHERE  cluster_id = $1::uuid;`

	ClusterMemberList = `
SELECT sn.node_id,
       sn.node_name,
       sc.cluster_name
FROM   soma.cluster_membership scm
JOIN   soma.nodes sn
ON     scm.node_id = sn.node_id
JOIN   soma.clusters sc
ON     scm.cluster_id = sc.cluster_id
WHERE  scm.cluster_id = $1::uuid;`

	ClusterBucketID = `
SELECT sc.bucket_id
FROM   soma.clusters sc
WHERE  sc.cluster_id = $1;`

	ClusterOncProps = `
SELECT op.instance_id,
       op.source_instance_id,
       op.view,
       op.oncall_duty_id,
       iot.name
FROM   soma.cluster_oncall_properties op
JOIN   inventory.oncall_team iot
  ON   op.oncall_duty_id = iot.id
WHERE  op.cluster_id = $1::uuid;`

	ClusterSvcProps = `
SELECT sp.instance_id,
       sp.source_instance_id,
       sp.view,
       sp.service_id
FROM   soma.cluster_service_property sp
WHERE  sp.cluster_id = $1::uuid;`

	ClusterSysProps = `
SELECT sp.instance_id,
       sp.source_instance_id,
       sp.view,
       sp.system_property,
       sp.value
FROM   soma.cluster_system_properties sp
WHERE  sp.cluster_id = $1::uuid;`

	ClusterCstProps = `
SELECT cp.instance_id,
       cp.source_instance_id,
       cp.view,
       cp.custom_property_id,
       cp.value,
       scp.custom_property
FROM   soma.cluster_custom_properties cp
JOIN   soma.custom_properties scp
  ON   cp.custom_property_id = scp.custom_property_id
WHERE  cp.cluster_id = $1::uuid;`

	ClusterSystemPropertyForDelete = `
SELECT view,
       system_property,
       value
FROM   soma.cluster_system_properties
WHERE  source_instance_id = $1::uuid
  AND  source_instance_id = instance_id;`

	ClusterCustomPropertyForDelete = `
SELECT sccp.view,
       sccp.custom_property_id,
       sccp.value,
       scp.custom_property
FROM   soma.cluster_custom_properties sccp
JOIN   soma.custom_properties scp
  ON   sccp.repository_id = scp.repository_id
 AND   sccp.custom_property_id = scp.custom_property_id
WHERE  sccp.source_instance_id = $1::uuid
  AND  sccp.source_instance_id = sccp.instance_id;`

	ClusterOncallPropertyForDelete = `
SELECT scop.view,
       scop.oncall_duty_id,
       iot.name,
       iot.phone_number
FROM   soma.cluster_oncall_properties scop
JOIN   inventory.oncall_team iot
  ON   scop.oncall_duty_id = iot.id
WHERE  scop.source_instance_id = $1::uuid
  AND  scop.source_instance_id = scop.instance_id;`

	ClusterServicePropertyForDelete = `
SELECT soma.cluster_service_property.view,
       soma.cluster_service_property.service_id
FROM   soma.cluster_service_property
JOIN   soma.service_property
  ON   soma.cluster_service_property.team_id = soma.service_property.team_id
 AND   soma.cluster_service_property.service_id = soma.service_property.id
WHERE  soma.cluster_service_property.source_instance_id = $1::uuid
  AND  soma.cluster_service_property.source_instance_id = soma.cluster_service_property.instance_id;`
)

func init() {
	m[ClusterBucketID] = `ClusterBucketID`
	m[ClusterCstProps] = `ClusterCstProps`
	m[ClusterCustomPropertyForDelete] = `ClusterCustomPropertyForDelete`
	m[ClusterList] = `ClusterList`
	m[ClusterMemberList] = `ClusterMemberList`
	m[ClusterOncProps] = `ClusterOncProps`
	m[ClusterOncallPropertyForDelete] = `ClusterOncallPropertyForDelete`
	m[ClusterServicePropertyForDelete] = `ClusterServicePropertyForDelete`
	m[ClusterShow] = `ClusterShow`
	m[ClusterSvcProps] = `ClusterSvcProps`
	m[ClusterSysProps] = `ClusterSysProps`
	m[ClusterSystemPropertyForDelete] = `ClusterSystemPropertyForDelete`
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
