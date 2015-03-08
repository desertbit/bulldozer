/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package auth

import (
	tr "code.desertbit.com/bulldozer/bulldozer/translate"

	"fmt"
)

const (
	// The internal groups:
	// --------------------

	// System operator (system administrator): full access to the system.
	GroupSysOp = "sysop"

	// Administrator: client of the application. Administration of the page content and some basic stuff...
	GroupAdmin = "admin"
)

var (
	groups Groups
)

func init() {
	// Register the internal groups.
	RegisterGroup(GroupSysOp, tr.S("bud.auth.groupSysOpDescription"))
	RegisterGroup(GroupAdmin, tr.S("bud.auth.groupAdminDescription"))
}

//####################//
//### Group Struct ###//
//####################//

type Groups []*Group

type Group struct {
	name        string
	description string
}

func newGroup(name, desc string) *Group {
	return &Group{
		name:        name,
		description: desc,
	}
}

func (g *Group) Name() string {
	return g.name
}

func (g *Group) Description() string {
	return g.description
}

//##############//
//### Public ###//
//##############//

// RegisterGroup registeres a group.
// It is suggested to only use lower characters.
func RegisterGroup(name string, description string) error {
	// Check if the group already exists
	if groupExists(name) {
		return fmt.Errorf("failed to register new group '%s': group was already registered!", name)
	}

	// Add the group to the groups slice.
	groups = append(groups, newGroup(name, description))

	return nil
}

// GetGroups returns a slice of all available groups
func GetGroups() Groups {
	return groups
}

//###############//
//### Private ###//
//###############//

func groupExists(name string) bool {
	for _, g := range groups {
		if name == g.name {
			return true
		}
	}

	return false
}
