/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */


var Bulldozer = new function() {
    /*
     * Public Variables
     */

	this.fn = Object.getPrototypeOf(this);
	this.utils;


	/*
	 * Public
	 */

	 this.init = function(sessionID, socketAccessToken) {
	 	// Show the loading indicator.
	 	Bulldozer.loadingIndicator.show();

	 	// Initialize the socket session.
	 	Bulldozer.socket.init(sessionID, socketAccessToken);

	 	// Run Kepler init.
	 	Kepler.init();
	 };
};