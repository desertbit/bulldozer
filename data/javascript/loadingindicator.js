/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

Bulldozer.fn.loadingIndicator = new function () {
	/*
     * Private Variables
     */

	var timeLoadingInd = false;
    var visible = false;


    /*
     * Private Methods
     */

    var stopTimeout = function() {
        if (timeLoadingInd !== false) {
            clearTimeout(timeLoadingInd);
            timeLoadingInd = false;
        }
    };



    /*
     * Public Methods
     */

    this.show = function () {
        var e = $("#bulldozer-loading-indicator");
        if (visible) {
            return;
        }
        visible = true;

        // Stop the timeout
        stopTimeout();

        // Show the loading indicator and make it visible after 1 second
        e.css('opacity', '0').show();
        timeLoadingInd = setTimeout(function () {
            timeLoadingInd = false;
            e.css('opacity', '1').addClass('show');
        }, 1000);
    };

    this.hide = function () {
        var e = $("#bulldozer-loading-indicator");
        if (!visible) {
            return;
        }
        visible = false;

        // Stop the timeout
        stopTimeout();

        // Remove the show class again.
        e.removeClass('show');

         // Hide the loading indicator after 2 seconds.
        timeLoadingInd = setTimeout(function () {
            timeLoadingInd = false;
            e.hide();
        }, 2000);
    };
};