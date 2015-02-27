/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */


/*
 * Bulldozer TopBar
 */

Bulldozer.fn.topbar = new function () {
    /*
     * Public Methods
     */

    this.space = function (add) {
    	var space = "45px";
    	if (add === false) {
    		space = "0";
    	}

		$("body").css("margin-top", space);
		$(".bud-topbar-auto-move").css("margin-top", space);
    };
};