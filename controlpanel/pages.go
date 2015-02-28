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
	pages Pages
)

//#############//
//### Types ###//
//#############//

type Pages []*Page

type Page struct {
	ID string

	Title string
	Icon  string

	AuthGroups []string
	Template   *template.Template
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

	// Check if already present with ID.
	for _, p := range pages {
		if p.ID == page.ID {
			log.L.Error("failed to add control panel page with ID '%s': a page with the same ID exists already!", page.ID)
			return
		}
	}

	// Add the page to the slice.
	pages = append(pages, page)
}
