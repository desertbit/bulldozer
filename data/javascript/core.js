/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */



/*
 * Bulldozer Core Methods
 */

Bulldozer.fn.core = new function () {
    /*
     * Private Variables
     */

    var documentReady = false;
    var scriptsToLoad = [];
    var pendingLoadJsTriggerIDs = [];
    var globalServerEvents = {};
    var exitMessage;



    /*
     * Private Events
     */

    $(document).on('bulldozer.ready', function() {
    	// Update the flag.
    	documentReady = true;

    	// Execute all js load functions.
    	Bulldozer.core.execJsLoad();
    });

    // Redirect all internal a href links to call the javascript loadPage method,
    // instead of redirecting to the page. This would kill the current socket session.
    $(document).on('click', 'a', function(e) {
        var url = String($(this).attr('href'));

        // Open mailto links in a new tab. Otherwise this might have sideeffects.
        // In Firefox the websocket connection is closed otherwise...
        if (url.slice(0, 7) === "mailto:") {
            e.preventDefault();
            window.open(url, '_blank');
            return false;
        }
        // Check if this is an internal link, if the url is not an anchor link
        // and if the url points not to a file on the server...
        else if (this.host === window.location.host
            && url.slice(0, 1) !== "#"
            && url.slice(0, 7) !== "public/"
            && url.slice(0, 8) !== "/public/")
        {
            e.preventDefault();

            if (url) {
                Bulldozer.core.navigate(url);
            }

            return false;
        }
    });



    /*
     * Public Methods
     */

    this.navigateToDefault = function () {
        this.navigate("/");
    };

    this.navigate = function (path) {
        // Show the loading indicator
        Bulldozer.loadingIndicator.show();

        // Construct the data object be send
        var data = {
            path: path
        };

        // Finally send the data
        Bulldozer.socket.send('route', data);
    };

    this.emit = function () {
        // Check if the DOM ID and key parameter is passed to this method
        if (arguments.length < 2) {
            console.log("Bulldozer.emit: Invalid arguments passed! The emit function requires a DOM ID and key parameter!");
            return;
        }

        // Construct the data object be send
        var data = {
            did: arguments[0],
            key: arguments[1]
        };

        // Append the arguments to the data string
        for (var i = 2; i < arguments.length; i++) {
            data['arg' + (i-1)] = arguments[i];
        }

        // Finally send the data
        Bulldozer.socket.send('emit', data);
    };



    /*
     * Load JS and Stylesheets
     */

    this.loadStyleSheet = function (url) {
        $('<link rel="stylesheet" type="text/css" href="' + url +'">').appendTo("head");
    };

    this.loadScript = function (url, callback) {
        // Create a new item and add it to the list
        var item = {
            url: url,
            callback: callback
        };

        scriptsToLoad.push(item);

        // Define the load function
        var ajaxLoadScript = function (url, callback) {
            var options = {
                dataType: "script",
                cache: true,
                url: url
            };

            var scriptLoadFinished = function () {
                if (scriptsToLoad.length > 0) {
                    // Load the pending scripts
                    ajaxLoadScript(scriptsToLoad[0].url, scriptsToLoad[0].callback);
                }
                else {
                    // Execute the template scripts if all script are loaded
                    Bulldozer.core.execJsLoad();
                }
            };

            // Use $.ajax() since it is more flexible than $.getScript
            jQuery.ajax(options)
                .done(function (script, textStatus) {
                    // Remove the first element
                    scriptsToLoad.shift();

                    // Call the callback if defined
                    if (callback)
                        callback();

                    scriptLoadFinished();
                })
                .fail(function (jqxhr, settings, exception) {
                    // Remove the first element
                    scriptsToLoad.shift();
                    
                    // Show an error message box
                    Bulldozer.utils.showErrorMessageBox("Error",
                        "Failed to load script '" + url + "'. Please contact the site administrator!",
                        "Error message: " + String(exception));

                    scriptLoadFinished();
                });
        };

        // Load the script if no other ajax loading script process is running
        if (scriptsToLoad.length <= 1) {
            ajaxLoadScript(url, callback);
        }
    };



    //
    // Exit Message
    //

    $(window).on('beforeunload', function () {
        // Only show the message if defined
        if (exitMessage) {
            return exitMessage;
        }
    });

    this.setExitMessage = function (msg) {
        exitMessage = String(msg);
    };

    this.resetExitMessage = function () {
        exitMessage = "";
    };



    //
    // JS Load & Unload
    //

    this.execJsLoad = function(id) {
        // Delay the execution for 10 ms to wait for other possible new calls on loadScript
        setTimeout(function () {
            // Only trigger the event if no more script is loading and only if the HTML document is loaded already.
            // Otherwise the event is triggered after all scripts are loaded or the document gets loaded.
            if (scriptsToLoad.length > 0 || !documentReady) {
                // Add the ID to the pending list
                if (id) {
                    pendingLoadJsTriggerIDs.push(id.toString());
                }
            }
            else {
                // Trigger the pending events
                $.each(pendingLoadJsTriggerIDs, function (i, id) {
                    $("#" + id).triggerHandler('bulldozer.execJsLoad');
                });

                // Clear the pending array
                pendingLoadJsTriggerIDs = [];

                // Trigger the event for the current id if defined
                if (id) {
                    $("#" + id).triggerHandler('bulldozer.execJsLoad'); 
                }

                // Hide the loading indicator after 50 ms. This will ensure, that the event is first executed...
                setTimeout(function () {
                    Bulldozer.loadingIndicator.hide();
                }, 50);  
            }
        }, 10);
    };

    this.onJsLoad = function (id, callback) {
        $("#" + id).one("bulldozer.execJsLoad", function() {
            try {
                callback();
            }
            catch(err) {
                console.log("execute js unload function error: " + err.message);
            }
        });
    };

    this.execJsUnload = function (id) {
        // Trigger the event
        $("#" + id).triggerHandler('bulldozer.execJsUnload');
    };

    this.onJsUnload = function (id, callback) {
        var cb = function() {
            try {
                callback();
            }
            catch(err) {
                console.log("execute js unload function error: " + err.message);
            }
        };

        $("#" + id).one("bulldozer.execJsUnload", cb);
        $(document).one("bulldozer.execJsUnload", cb);
    };



    //
    // Server events
    //

    this.addServerEvent = function (id, key, func) {
        var el = $("#" + id);
        if (el.length <= 0) {
            console.log("addServerEvent: element with id '" + id + "' does not exists!");
            return;
        }

        // Get the events array
        var events = el.data("bulldozerserverevents");

        // Create the object if it doesn't exists
        if (!events) {
            events = {};
        }

        // Add the function
        events[key] = func;

        // Set the data again
        el.data("bulldozerserverevents", events);
    };

    this.emitServerEvent = function (id, key) {
        var el = $("#" + id);
        if (el.length <= 0) {
            console.log("emitServerEvent: element with id '" + id + "' does not exists!");
            return;
        }

        // Get the events array
        var events = el.data("bulldozerserverevents");

        // Check if the object is defined
        if (!events) {
            console.log("emitServerEvent: event with key '" + key + "' does not exists!");
            return;
        }

        // Get the function
        func = events[key];
        if (!func) {
            console.log("emitServerEvent: event with key '" + key + "' does not exists!");
            return;
        }

        // Get the arguments and call the function.
        try {
            var args = Array.prototype.slice.call(arguments, 2);
            func.apply(el, args);
        }
        catch(err) {
            console.log("execute server event error: " + err.message);
        }
    };

    this.addGlobalServerEvent = function (key, func) {
        // Add the global function
        globalServerEvents[key] = func;
    };

    this.clearGlobalServerEvents = function () {
        globalServerEvents = {};
    };

    this.emitGlobalServerEvent = function (key) {
        // Get the function
        var func = globalServerEvents[key];
        if (!func) {
            console.log("emitGlobalServerEvent: event with key '" + key + "' does not exists!");
            return;
        }

        // Get the arguments and call the function
        try {
            var args = Array.prototype.slice.call(arguments, 1);
            func.apply(document, args);
        }
        catch(err) {
            console.log("execute global server event error: " + err.message);
        }
    };
};