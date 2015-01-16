/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package auth

//####################//
//### Group Struct ###//
//####################//

type Groups []*Group

type Group struct {
	g *dbGroup
}

func newGroup(g *dbGroup) *Group {
	return &Group{
		g: g,
	}
}

func (g *Group) Name() string {
	return g.g.Name
}

func (g *Group) Description() string {
	return g.g.Description
}

//##############//
//### Public ###//
//##############//

// GetGroups returns a slice of all groups
func GetGroups() (Groups, error) {
	// Get all groups from the database.
	dbGroups, err := dbGetGroups()
	if err != nil {
		return nil, err
	}

	// Create an emtpy slice
	groups := make(Groups, len(dbGroups))

	// Fill the groups slice
	for i, g := range dbGroups {
		groups[i] = newGroup(g)
	}

	return groups, nil
}

// Create the group if it doesn't exists.
func AddGroupIfNotExists(name string, description string) error {
	// Check if the group already exists
	exist, err := dbGroupExists(name)
	if err != nil {
		return err
	} else if exist {
		return nil
	}

	// Create the group
	_, err = dbAddGroup(name, description)
	if err != nil {
		return err
	}

	return nil
}
