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
        var e = $("#bud-loading-indicator");
        if (visible) {
            return;
        }
        visible = true;

        // Stop the timeout
        stopTimeout();

        // Remove the class first.
        e.removeClass('none-pointer-events');

        // Show the loading indicator and make it visible after 1 second
        e.css('opacity', '0').show();
        timeLoadingInd = setTimeout(function () {
            timeLoadingInd = false;
            e.css('opacity', '1').addClass('show');

            // Hide the loading indicator after 25 seconds and display an error message.
            timeLoadingInd = setTimeout(function () {
                timeLoadingInd = false;
                Bulldozer.loadingIndicator.hide();
                
                // Show an error message box
                Bulldozer.utils.showErrorMessageBox("Error", "Failed to perform the request. Timeout reached. Please try again...");  
            }, 25000);
        }, 1000);
    };

    this.hide = function () {
        var e = $("#bud-loading-indicator");
        if (!visible) {
            return;
        }
        visible = false;

        // Stop the timeout
        stopTimeout();

        // Remove the show class again and don't block the pointer events.
        e.removeClass('show').addClass('none-pointer-events');

         // Hide the loading indicator after 2 seconds.
        timeLoadingInd = setTimeout(function () {
            timeLoadingInd = false;
            e.hide();
        }, 2000);
    };

};