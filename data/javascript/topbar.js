/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */


/*
 * Bulldozer TopBar
 */

Bulldozer.fn.topbar = new function () {
    var currentSpace = "0";

    /*
     * Public Methods
     */

     // space adds or removes the required topbar space.
     // If no boolean value is passed, then the previous state is restored.
    this.space = function (add) {
    	if (add === true) {
            currentSpace = "45px";
        }
        else if (add === false) {
    		currentSpace = "0";
    	}

		$("body").css("margin-top", currentSpace);
		$(".bud-topbar-auto-move").css("margin-top", currentSpace);
    };
};