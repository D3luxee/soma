/*-
 * Copyright (c) 2016-2017, Jörg Pernfuß
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package perm // import "github.com/mjolnir42/soma/internal/perm"

import "github.com/mjolnir42/soma/lib/proto"

// actionLookup is the cache data structure for permission actions,
// allowing lookup by ID or name
type actionLookup struct {
	byName map[string]map[string]*proto.Action
	byID   map[string]map[string]*proto.Action
}

// newActionLookup returns an initialized actionLookup
func newActionLookup() *actionLookup {
	a := actionLookup{}
	a.byName = map[string]map[string]*proto.Action{}
	a.byID = map[string]map[string]*proto.Action{}
	return &a
}

// add inserts an action into the cache
func (m *actionLookup) add(sID, sName, aID, aName, category string) {
	a := &proto.Action{
		ID:          aID,
		Name:        aName,
		SectionID:   sID,
		SectionName: sName,
		Category:    category,
	}
	if _, ok := m.byName[sName]; !ok {
		m.byName[sName] = map[string]*proto.Action{}
	}
	if _, ok := m.byID[sID]; !ok {
		m.byID[sID] = map[string]*proto.Action{}
	}
	m.byName[sName][aName] = a
	m.byID[sID][aID] = a
}

// getByID returns an action from the cache. The action is identified
// by its section and action ID. Returns nil if the action was
// not found.
func (m *actionLookup) getByID(sID, aID string) *proto.Action {
	if _, ok := m.byID[sID]; !ok {
		return nil
	}
	return m.byID[sID][aID]
}

// getByName returns an action from the cache. The action is identified
// by its section and action name. Returns nil if the action was
// not found.
func (m *actionLookup) getByName(sName, aName string) *proto.Action {
	if _, ok := m.byName[sName]; !ok {
		return nil
	}
	return m.byName[sName][aName]
}

// rmActionByID removes an action from the cache. The action is
// identified by section and action ID.
func (m *actionLookup) rmActionByID(sID, aID string) {
	if _, ok := m.byID[sID]; !ok {
		return
	}
	a, ok := m.byID[sID][aID]
	if !ok {
		return
	}
	if sID != a.SectionID || aID != a.ID {
		return
	}

	delete(m.byName[a.SectionName], a.Name)
	delete(m.byID[a.SectionID], a.ID)
}

// rmActionByName removes and action from the cache. The action is
// identified by section and action name.
func (m *actionLookup) rmActionByName(sName, aName string) {
	if _, ok := m.byName[sName]; !ok {
		return
	}
	a, ok := m.byName[sName][aName]
	if !ok {
		return
	}
	if sName != a.SectionName || aName != a.Name {
		return
	}

	delete(m.byID[a.SectionID], a.ID)
	delete(m.byName[a.SectionName], a.Name)
}

// rmSectionByID removes all actions from the cache that belong to
// a section identified by its ID.
func (m *actionLookup) rmSectionByID(sID string) {
	if _, ok := m.byID[sID]; !ok {
		return
	}
	var sName string
	for _, a := range m.byID[sID] {
		// get the section name from the first found action
		sName = a.SectionName
		break
	}
	delete(m.byName, sName)
	delete(m.byID, sID)
}

// rmSectionByName removes all actions from the cache that belong to
// a section identified by its name.
func (m *actionLookup) rmSectionByName(sName string) {
	if _, ok := m.byName[sName]; !ok {
		return
	}
	var sID string
	for _, a := range m.byName[sName] {
		// get the section name from the first found action
		sID = a.SectionID
		break
	}
	delete(m.byID, sID)
	delete(m.byName, sName)
}

// getActionsBySectionID returns all actionIDs for a specific section
func (m *actionLookup) getActionsBySectionID(sID string) []string {
	if _, ok := m.byID[sID]; !ok {
		return []string{}
	}
	res := make([]string, len(m.byID[sID]))
	i := 0
	for aID := range m.byID[sID] {
		res[i] = aID
		i++
	}
	return res
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
