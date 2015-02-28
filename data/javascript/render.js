/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */


/*
 * Bulldozer Render
 */

Bulldozer.fn.render = new function () {
	/*
     * Private Variables
     */

     var currentUrl;
     var manualHistoryChange = false;



    /*
     * Private Methods & Events
     */

    // Browser history change event
    $(window).on('statechange', function () {
        // Check if this state is pushed by a History call
        if (manualHistoryChange) {
            return;
        }

        // Get the url hash
        var hash = History.getState().hash;
        if (!hash) {
            hash = "/";
        }

        // Load the page
        Bulldozer.core.navigate(hash);
    });


    /*
     * Public Methods
     */

    this.updateTemplate = function (domId, body) {
        var obj = $("#" + domId);
        if (obj.length <= 0) {
            // Show an error message box
            Bulldozer.utils.showErrorMessageBox("Error", "Failed to update template: '" + domId + "'. Try to reload the page and please contact the site administrator!");
            return;
        }

        // Trigger the unload event
        Bulldozer.core.execJsUnload(domId);

        // Remove all the data attached to the object
        // and replace the template body with the new template body
        obj.removeData().replaceWith(body);

        // Execute the kepler init method to apply all new kepler changes
        Kepler.init();
    };



    this.page = function (body, title, url) {
        // Trigger the global js unload event
        $(document).triggerHandler('bulldozer.execJsUnload');

        // Clear all global server events.
        Bulldozer.core.clearGlobalServerEvents();

        var budBody = $("#bud-body");

        // Do some cleanup
        if (budBody && budBody.length > 0) {
            // Unbind all events of the current page and all its children
            budBody.off();
            budBody.find("*").off();
        }

        // Remove any element in the body tag that doesn't belong there
        var el, id;
        $('body').children().each(function() {
            el = $(this);
            id = el.attr('id');
            if (id === "bud-loading-indicator"
                || id === "bud-body"
                || id === "bud-connection-lost"
                || (el.is('noscript') && el.has('#bud-noscript')))
            {
                return;
            }
            
            $(this).remove();
        });

        // Create the new body.
        var newBudBody = $('<div id="bud-body"></div>');

        // Append the page body.
        newBudBody.append(body);

        // Replace the bulldozer body.
        budBody.replaceWith(newBudBody);

        // Scroll to the top of the page
        window.scrollTo(0, 0);

        // Push the new current url to the browser history, if it is not the same url as the current one
        if (url && currentUrl !== url) {
            // Update the current page url
            currentUrl = url;

            manualHistoryChange = true;
            History.pushState(null, null, url);
            manualHistoryChange = false;
        }

        // Set the new page title
        document.title = title;

        // Execute the kepler init method to apply all new kepler changes
        Kepler.init();
    };
};