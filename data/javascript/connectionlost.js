/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

Bulldozer.fn.connectionLost = new function () {
	/*
     * Private Variables
     */

	var timeConnLost = false;
	var timeRemoveClassess = false;
    var visible = false;


    /*
     * Private Methods
     */

    var stopTimeout = function() {
        if (timeConnLost !== false) {
            clearTimeout(timeConnLost);
            timeConnLost = false;
        }
    };

    var resetTimeoutRemoveClasses = function() {
        if (timeRemoveClassess !== false) {
            clearTimeout(timeRemoveClassess);
        }

        timeRemoveClassess = setTimeout(function () {
            timeRemoveClassess = false;
            $("#bud-connection-lost .click-to-reconnect").removeClass("connecting fail success");
        }, 1500);
    };



    /*
     * Public Methods
     */

    this.show = function () {
        var e = $("#bud-connection-lost");
        if (visible) {
            return;
        }
        visible = true;

        // Stop the timeout
        stopTimeout();

        // Show the connection lost widget after a short timeout.
        timeConnLost = setTimeout(function () {
            timeConnLost = false;
        	e.show().addClass('show');
        }, 700);
    };

    this.hide = function () {
        var e = $("#bud-connection-lost");
        if (!visible) {
            return;
        }
        visible = false;

        // Stop the timeout
        stopTimeout();

        // Remove the show class again.
        e.removeClass('show');

         // Hide the connection lost widget after 3 seconds.
        timeConnLost = setTimeout(function () {
            timeConnLost = false;
            e.hide();
        }, 3000);
    };

    this.connectionLost = function() {
    	return visible;
    };

    this.reconnectFailed = function() {
    	var el = $("#bud-connection-lost .click-to-reconnect");

    	// Ignore if the class is already set or the connecting class is missing.
    	// We only want to set the fail class if the reconnect was triggered from the user.
		if (!el.hasClass("connecting") || el.hasClass("fail")) {
			return;
		}

		// Add the class.
		el.addClass("fail");

		// Reset the remove classes timeout.
		resetTimeoutRemoveClasses();
    };

    this.reconnectSuccess = function() {
    	var el = $("#bud-connection-lost .click-to-reconnect");

    	// Ignore if the class is already set.
		if (el.hasClass("success")) {
			return;
		}

		// Add the class.
		el.addClass("success");

		// Reset the remove classes timeout.
		resetTimeoutRemoveClasses();
    };



    /*
     * Private
     */

	// On document ready.
    $(function() {
	    // Add the click handler.
		$("#bud-connection-lost .click-to-reconnect").click(function() {
			var el = $(this);

			// Ignore if already reconnecting.
			if (el.hasClass("connecting")) {
				return;
			}

			// Add the class.
			el.addClass("connecting");

			// Reconnect the socket.
			Bulldozer.socket.reconnect();

			// Reset the remove classes timeout.
			resetTimeoutRemoveClasses();
		});
	});
};