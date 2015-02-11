/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package sessions

var (
	navigateFunc NavigateFunc
)

//#############//
//### Types ###//
//#############//

// NavigateFunc callback:
// The string parameter specifies the route path.
type NavigateFunc func(*Session, string)

//##############//
//### Public ###//
//##############//

// SetNavigateFunc sets the navigate function callback
// which is executed on each session navigate request.
// You have to set the new current session path manually.
func SetNavigateFunc(f NavigateFunc) {
	navigateFunc = f
}
