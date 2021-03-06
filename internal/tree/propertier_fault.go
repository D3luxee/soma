/*-
 * Copyright (c) 2016, 1&1 Internet SE
 * Copyright (c) 2016, Jörg Pernfuß
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package tree

func (tef *Fault) SetProperty(p Property) {
}

func (tef *Fault) setPropertyInherited(p Property) {
}

func (tef *Fault) setPropertyOnChildren(p Property) {
}

func (tef *Fault) addProperty(p Property) {
}

func (tef *Fault) UpdateProperty(p Property) {
}

func (tef *Fault) updatePropertyInherited(p Property) {
}

func (tef *Fault) updatePropertyOnChildren(p Property) {
}

func (tef *Fault) switchProperty(p Property) bool {
	return false
}

func (tef *Fault) getCurrentProperty(p Property) Property {
	return nil
}

func (tef *Fault) DeleteProperty(p Property) {
}

func (tef *Fault) deletePropertyInherited(p Property) {
}

func (tef *Fault) deletePropertyOnChildren(p Property) {
}

func (tef *Fault) deletePropertyAllInherited() {
}

func (tef *Fault) deletePropertyAllLocal() {
}

func (tef *Fault) rmProperty(p Property) bool {
	return false
}

func (tef *Fault) verifySourceInstance(id, prop string) bool {
	return false
}

func (tef *Fault) findIDForSource(source, prop string) string {
	return ``
}

func (tef *Fault) syncProperty(childID string) {
}

func (tef *Fault) checkProperty(propType, propID string) bool {
	return false
}

func (tef *Fault) checkDuplicate(p Property) (bool, bool, Property) {
	return true, false, nil
}

func (tef *Fault) resyncProperty(srcID, pType, childID string) {
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
