/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package topbar

var (
	leftMenuItems     Items
	rightMenuItems    Items
	editmodeMenuItems Items
	userMenuItems     Items
)

//###################//
//### Item struct ###//
//###################//

type Items []*Item

type Item struct {
	Title      string
	Icon       string
	AuthGroups []string

	// The loading indicator is always shown on trigger.
	EventFunc string

	// Sub Menu items.
	SubItems Items
}

//##############//
//### Public ###//
//##############//

// AddRightMenuItem adds a menu item to the right menu.
// Only call this during application initialization.
func AddRightMenuItem(item *Item) {
	rightMenuItems = append(rightMenuItems, item)
}

// AddLeftMenuItem adds a menu item to the left menu.
// Only call this during application initialization.
func AddLeftMenuItem(item *Item) {
	leftMenuItems = append(leftMenuItems, item)
}

// AddEditmodeMenuItem adds a menu item to the editmode menu.
// Only call this during application initialization.
func AddEditmodeMenuItem(item *Item) {
	editmodeMenuItems = append(editmodeMenuItems, item)
}

// AddUserMenuItem adds a menu item to the user menu.
// Only call this during application initialization.
func AddUserMenuItem(item *Item) {
	userMenuItems = append(userMenuItems, item)
}
