/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package sessions

var (
	backendI Interface
)

//#################//
//### Interface ###//
//#################//

type Interface interface {
	// The NavigateFunc function
	// is executed on each session navigate request.
	// You have to set the new current session path manually.
	// The string parameter specifies the route path.
	NavigateFunc(*Session, string)

	ShowErrorPage(*Session, string, ...bool)
	ShowNotFoundPage(*Session)
}
