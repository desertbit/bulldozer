/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

Bulldozer.fn.loadingIndicator = new function () {
	/*
     * Private Variables
     */

	var timeLoadingInd = false;



    /*
     * Public Methods
     */

    this.show = function () {
        // Show the loading indicator and make it visible after 1 second
        var e = $("#bulldozer-loading-indicator");
        if (!e.is(":visible")) {
            e.css('opacity', '0').show();
            timeLoadingInd = setTimeout(function () {
                timeLoadingInd = false;
                e.stop().fadeTo(300, 1);
            }, 1000);
        }
    };

    this.hide = function () {
        // Stop the timeout
        if (timeLoadingInd !== false) {
            clearTimeout(timeLoadingInd);
        }

        // Hide the loading indicator
        var e = $("#bulldozer-loading-indicator");
        if (e.is(":visible")) {
            e.stop().fadeOut(300);
        }
    };
};