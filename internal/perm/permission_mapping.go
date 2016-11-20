/*-
 * Copyright (c) 2016, Jörg Pernfuß
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package perm

import "github.com/1and1/soma/lib/proto"

// permissionMapping is the cache datastructure that keep track of
// sections and actions mapped to a permission
type permissionMapping struct {
	// what the unmap functions do with the values is more akin to
	// playing peek-a-boo with a toddler than actually deleting things.
	// This counter is incremented whenever a value is hidden from the
	// slice; it is still referenced by the underlying array, thus
	// leaked.
	// Once enough elements have been leaked, a compaction can occur
	// to free the values.
	compactionCounter int64
	// sectionID -> []permissionID
	section map[string][]string
	// sectionID -> actionID -> []permissionID
	action map[string]map[string][]string
	// permissionID -> []sectionID
	permSection map[string][]string
	// permissionID -> []protoAction{Id:, SectionId:}
	permAction map[string][]proto.Action
}

// newPermissionMapping returns an initialized permissionMapping
func newPermissionMapping() *permissionMapping {
	p := permissionMapping{}
	p.compactionCounter = 0
	// the following only initialize the first map
	p.section = map[string][]string{}
	p.action = map[string]map[string][]string{}
	p.permSection = map[string][]string{}
	p.permAction = map[string][]proto.Action{}
	return &p
}

// mapSection records that a section has been mapped to a permission
func (m *permissionMapping) mapSection(sectionID,
	permissionID string) {
	// ensure map for the section is initialized
	if _, ok := m.section[sectionID]; !ok {
		m.section[sectionID] = []string{}
	}
	// ensure map for the permission is initialized
	if _, ok := m.permSection[permissionID]; !ok {
		m.permSection[permissionID] = []string{}
	}

	// record mapping
	m.section[sectionID] = append(m.section[sectionID], permissionID)
	m.permSection[permissionID] = append(m.permSection[permissionID],
		sectionID)
}

// unmapSection removes the mapping of a section
func (m *permissionMapping) unmapSection(sectionID,
	permissionID string) {
	var found bool

	// section should not be unmapped, but might be
	if _, ok := m.section[sectionID]; !ok {
		return
	}
	for i, p := range m.section[sectionID] {
		if p != permissionID {
			continue
		}
		found = true
		m.section[sectionID] = append(m.section[sectionID][:i],
			m.section[sectionID][i+1:]...)
		m.compactionCounter++
		break
	}
	if !found {
		return
	}

	// classified as `should never happen`
	if _, ok := m.permSection[sectionID]; !ok {
		return
	}
	for i, s := range m.permSection[permissionID] {
		if s != sectionID {
			continue
		}
		m.permSection[permissionID] = append(
			m.permSection[permissionID][:i],
			m.permSection[permissionID][i+1:]...,
		)
		m.compactionCounter++
		break
	}
}

// mapAction records that an action has been mapped to a permission
func (m *permissionMapping) mapAction(sectionID, actionID,
	permissionID string) {
	// ensure map for the section is initialized
	if _, ok := m.action[sectionID]; !ok {
		m.action[sectionID] = map[string][]string{}
	}
	// ensure map for the action is initialized
	if _, ok := m.action[sectionID][actionID]; !ok {
		m.action[sectionID][actionID] = []string{}
	}
	// ensure map for the permission is initialized
	if _, ok := m.permAction[permissionID]; !ok {
		m.permAction[permissionID] = []proto.Action{}
	}

	// record mapping
	m.action[sectionID][actionID] = append(
		m.action[sectionID][actionID], permissionID)
	m.permAction[permissionID] = append(m.permAction[permissionID],
		proto.Action{
			Id:        actionID,
			SectionId: sectionID,
		})
}

// unmapAction removes the mapping of an action
func (m *permissionMapping) unmapAction(sectionID, actionID,
	permissionID string) {
	var found bool

	// section should not be unmapped, but might be
	if _, ok := m.action[sectionID]; !ok {
		return
	}
	// action should not be unmapped, but might be
	if _, ok := m.action[sectionID][actionID]; !ok {
		return
	}
	for i, p := range m.action[sectionID][actionID] {
		if p != permissionID {
			continue
		}
		found = true
		m.action[sectionID][actionID] = append(
			m.action[sectionID][actionID][:i],
			m.action[sectionID][actionID][i+1:]...)
		m.compactionCounter++
		break
	}
	if !found {
		return
	}

	// classified as `should never happen`
	if _, ok := m.permAction[permissionID]; !ok {
		return
	}
	for i, a := range m.permAction[permissionID] {
		if a.Id != actionID {
			continue
		}
		m.permAction[permissionID] = append(
			m.permAction[permissionID][:i],
			m.permAction[permissionID][i+1:]...,
		)
		m.compactionCounter++
		break
	}
}

// removePermission removes a permission from the mapping
func (m *permissionMapping) removePermission(permissionID string) {

	// check for mapped sections
	if _, ok := m.permSection[permissionID]; ok {
		// copy out the sections to avoid modifying the map/slice
		// while iterating over it
		sections := []string{}
		for _, s := range m.permSection[permissionID] {
			sections = append(sections, s)
		}
		for _, s := range sections {
			m.unmapSection(s, permissionID)
		}
	}

	// check for mapped actions
	if _, ok := m.permAction[permissionID]; ok {
		// copy out the actions to avoid modifying the map/slice
		// while iterating over it
		actions := [][2]string{}
		for _, a := range m.permAction[permissionID] {
			actions = append(actions, [2]string{a.SectionId, a.Id})
		}
		for _, a := range actions {
			m.unmapAction(a[0], a[1], permissionID)
		}
	}
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix