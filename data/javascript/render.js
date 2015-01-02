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

     var currentPage;
     var currentUrl;
     var manualHistoryChange = false;
     var isInitialized = false;



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
        Bulldozer.core.loadPage(hash);
    });

    var createMissingWrappers = function() {
        var bulldozerBody = $("#bulldozer-body");

        if (!isInitialized) {
            // Empty the bulldozer body by simply replacing it.
            bulldozerBody.replaceWith('<div id="bulldozer-body"></div>');

            // Update the object
            bulldozerBody = $("#bulldozer-body");

            // Update the flag
            isInitialized = true;
        }

        // Create the top template if not present
        if ($("#bulldozer-top").length <= 0) {
            bulldozerBody.append('<div id="bulldozer-top"></div>');
        }

        // Create the page body if not present
        if ($("#bulldozer-pages-body").length <= 0) {
            bulldozerBody.append('<div id="bulldozer-pages-body"></div>');
        }

        // Create the bottom template if not present
        if ($("#bulldozer-bottom").length <= 0) {
            bulldozerBody.append('<div id="bulldozer-bottom"></div>');
        }
    };


    /*
     * Public Methods
     */

    this.currentPageUrl = function() {
        return currentUrl;
    };

    this.updateTemplate = function (domId, body) {
        var obj = $("#" + domId);
        if (obj.length <= 0) {
            // Show an error message box
            Bulldozer.utils.showErrorMessageBox("Error", "Failed to update template: '" + domId + "'. Try to reload the page and please contact the site administrator!");
            return;
        }

        // Trigger the unload event
        Bulldozer.core.execJsUnload(domId);

        // Replace the template body with the new template body
        obj.replaceWith(body);

        // Execute the kepler init method to apply all new kepler changes
        Kepler.init();
    };

    this.top = function (body) {
        // Get the pages body
        var t = $("#bulldozer-top");

        // Check if it exists. Otherwise create it.
        if (t.length <= 0) {
            createMissingWrappers();
            t = $("#bulldozer-top");
        }

        // Update the top template
        t.html(body);
    };

    this.bottom = function (body) {
        // Get the pages body
        var t = $("#bulldozer-bottom");

        // Check if it exists. Otherwise create it.
        if (t.length <= 0) {
            createMissingWrappers();
            t = $("#bulldozer-bottom");
        }

        // Update the bottom template
        t.html(body);
    };

    this.page = function (pageId, pageBody, title, url) {
        // Trigger the global js unload event
        $(document).triggerHandler('bulldozer.execJsUnload');

        // Hide all pages
        //$(".bulldozer-page").hide();
        // Remove all other pages. The dynamic hide is currently deactivated
        $(".bulldozer-page").remove();

        // Do some cleanup if the current page is set
        if (currentPage && currentPage.length > 0) {
            // Unbind all events of the current page and all its children
            currentPage.off();
            currentPage.find("*").off();
        }

        // Remove any element in the body tag that doesn't belong there
        var id;
        $('body').children().each(function () {
            id = $(this).attr('id');
            if (id === "bulldozer-loading-indicator" || id === "bulldozer-body") { return; }
            $(this).remove();
        });

        // Check if the page div is already present
        if($("#bulldozer-pages-body > #" + pageId).length > 0) {
            // Replace the page with the new page body
            $("#" + pageId).replaceWith(pageBody);
        }
        else {
            // Get the pages body
            var pagesBody = $("#bulldozer-pages-body");

            // Check if it exists. Otherwise create it.
            if (pagesBody.length <= 0) {
                createMissingWrappers();
                pagesBody = $("#bulldozer-pages-body");
            }

            // Add the page body
            pagesBody.append(pageBody);
        }

        // Set the new current page
        currentPage = $("#" + pageId);

        // Finally show the current page
        currentPage.show();

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