/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package controlpanel

import (
	"code.desertbit.com/bulldozer/bulldozer/log"
	"code.desertbit.com/bulldozer/bulldozer/template"
)

var (
	pages Pages = make(Pages)
)

func init() {
	p := &Page{
		ID:    "dashboard",
		Title: "Dashboard",
		Icon:  "fa-tachometer",
	}

	AddPage(p)
}

//#############//
//### Types ###//
//#############//

type Pages map[string]*Page

type Page struct {
	ID    string
	Group string

	Title    string
	Icon     string
	Template *template.Template
}

//##############//
//### Public ###//
//##############//

// AddPage adds a control panel page.
// Only add pages during application initialization!
func AddPage(page *Page) {
	// Check if invalid ID.
	if len(page.ID) == 0 {
		log.L.Error("failed to add control panel page with title '%s': emtpy ID!", page.Title)
		return
	}

	// Create the access ID
	id := page.ID
	if len(page.Group) > 0 {
		id = page.Group + "/" + id
	}

	// Check if already present with ID.
	if _, ok := pages[id]; ok {
		log.L.Error("failed to add control panel page with ID '%s' and Group '%s': a page with the same ID exists already!", page.ID, page.Group)
		return
	}

	// Add the page to the map.
	pages[id] = page
}
