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

		$("*").filter(function() {
			return $(this).css("position") === "fixed"
				&& $(this).attr("id") !== "bulldozer-topbar"
				&& $(this).attr("id") !== "bulldozer-loading-indicator"
				&& $(this).attr("id") !== "bulldozer-connection-lost";
		}).css("margin-top", space);
    };
};