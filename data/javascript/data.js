/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */



/*
 * Bulldozer data Methods
 */

Bulldozer.fn.data = new function () {
    /*
     * Private Variables
     */

     var values = {};



    /*
     * Public Methods
     */

    this.set = function (key, data) {
        values[key] = data; 
    };

    this.delete = function (key) {
        delete values[key]; 
    };

    this.get = function (key) {
        return values[key]; 
    };

    this.getAndReply = function (key, randomKey) {
    	// Get the data.
    	var d = values[key];
    	if (!d) {
    		d = "";
    	}

        // Construct the data object be send
        var data = {
            key: key,
            rand: randomKey,
            data: d
        };

        // Finally send the data
        Bulldozer.socket.send('clientData', data);
    };
};