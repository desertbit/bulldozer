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
    	var space = "50px";
    	if (add === false) {
    		space = "0";
    	}

		$("body").css("margin-top", space);
		$(".bulldozer-topbar-auto-move").css("margin-top", space);
    };
};