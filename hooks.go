/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package bulldozer

var (
	initFuncs []OnInitFunc
)

//#############//
//### Types ###//
//#############//

type OnInitFunc func() error

//##############//
//### Public ###//
//##############//

// OnInit adds a hook functions which is triggered during
// the Bulldozer initialization.
func OnInit(f OnInitFunc) {
	initFuncs = append(initFuncs, f)
}

//###############//
//### Private ###//
//###############//

func triggerOnInit() (err error) {
	for _, f := range initFuncs {
		err = f()
		if err != nil {
			return err
		}
	}

	return nil
}
