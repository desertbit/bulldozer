/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */


/*
 * Bulldozer Socket
 */

Bulldozer.fn.socket = new function () {
    /*
     * Const
     */

    var reconnectAttempts = 3;

    var SocketKey = {
        Task: "tsk"
    };

    var SocketData = {
        InvalidRequest:     "invalid_request",
        Ping:               "ping",
        Pong:               "pong"
    };



    /*
     * Private Variables
     */

    var documentReady = false;
    var socket;
    var sid, instanceID, token;
    var connErrorLoadingIndShown = false;
    var timeoutShowLoadingIndicator = false;
    var reconnectCount = 0;



    /*
     * Private Methods
     */

    var prepareSendMsg = function(data) {
        return "sid=" + sid + "&tok=" + token + "&" + data;
    };

    var resetShowLoadingIndicatorTimeout = function() {
        // Stop the timeout timer
        if (timeoutShowLoadingIndicator !== false) {
            clearTimeout(timeoutShowLoadingIndicator);
        }
        
        // Start the timer again
        timeoutShowLoadingIndicator = setTimeout(function () {
            timeoutShowLoadingIndicator = false;

            // Show the loading indicator and set the flag
            connErrorLoadingIndShown = true;
// TODO
            //Bulldozer.loadingIndicator.show();
        }, 40000);
    };

    var handleReceivedData = function(data) {
        // Reset the timeout timer
        resetShowLoadingIndicatorTimeout();

        // Hide the loading indicator if shown by a connection error
        if (connErrorLoadingIndShown) {
            connErrorLoadingIndShown = false;
// TODO
            // Bulldozer.loadingIndicator.hide();
        }

        // Check if data is not empty
        if (data) {
            // Check if the server has send an invalid request notification
            if (data === SocketData.InvalidRequest) {
                console.log("The server replied with an invalid request notification! The previous request was invalid!");
                return;
            }

            // Split the new token from the rest of the data
            var i = data.indexOf('&');

            if (i < 0) {
                // Show an error message box
                Bulldozer.utils.showErrorMessageBox("Error",
                    "Warning! Invalid data received from server! Try to reload your browser and notify the site administrator!",
                    "Error data: '" + data + "'");
                return;
            }

            // Set the new token and the data variable
            token = data.substring(0, i);
            data = data.substr(i + 1);

            // Check if the server requests a pong reply
            if (data === SocketData.Ping) {
                socket.send(prepareSendMsg(SocketKey.Task + "=" + SocketData.Pong + "&"));
                return;
            }

            // Execute the received javascript code if not empty
            if (data) {
                jQuery.globalEval(data);
            }
        }
    };

    var handleInitializeSession = function(data) {
        if (!data) {
            // Show an error message box
            Bulldozer.utils.showErrorMessageBox("Error", "Failed to initialize socket session! Received data is emtpy!");
            return false;
        }

        // Check if the server has send an invalid request notification
        if (data === SocketData.InvalidRequest) {
            console.log("The server replied with an invalid request notification! The previous request was invalid!");
            return false;
        }

        // Split the received data
        var list = data.split('&');
        
        // Check if enough elements exist
        if (list.length < 2) {
            console.log("Failed to initialize socket session! Received list length is invalid: '" + data + "'");
            return false;
        }

        // Save the instance ID and the new token.
        instanceID = list[0];
        token = list[1];

        // Reset the reconnect count again
        reconnectCount = 0;

        return true;
    };

    var connectionError = function(whileInitialization, sessionID, socketAccessToken) {
        // Fallback to the ajax socket
        if (whileInitialization && socket.type() !== Bulldozer.AjaxSocket.type()) {
            // Reconnect with the ajax socket
            console.log("falling back to ajax socket...");
            Bulldozer.socket.init(sessionID, socketAccessToken, true);
            return;
        }

        // Set the flag
        connErrorLoadingIndShown = true;

        // Show the loading indicator
// TODO
        //Bulldozer.loadingIndicator.show();

        if (reconnectCount > reconnectAttempts) {
            // TODO: Show this in the loading indicator
            // Show an error message box
            Bulldozer.utils.showErrorMessageBox("Error",
                "Failed to establish a connection to the server. Please reload this page and try again.");
        }
        else {
            // Try to reconnect
            setTimeout(function() {
                Bulldozer.socket.init(sessionID, socketAccessToken);
            }, 5000);
        }
    };



    /*
     * Public Methods
     */

    this.sessionID = function() {
        return sid;
    };

    this.init = function (sessionID, socketAccessToken, forceFallback) {
        // Flags
        var isInitializing = true;

        if (!sessionID || !socketAccessToken) {
            console.log("empty session ID or socket access token!");
            return;
        }

        // Set the session ID and token
        sid = sessionID;
        token = socketAccessToken;

        // Show the loading indicator
// TODO
        //Bulldozer.loadingIndicator.show();

        // Increment the reconnect count
        reconnectCount += 1;

        // Reset the previous socket if set
        if (socket) {
            socket.onOpen = undefined;
            socket.onClose = undefined;
            socket.onMessage = undefined;
            socket.onError = undefined;

            if (socket.reset) {
                socket.reset();
            }
        }

        // Choose the socket layer depending on the browser support
        if (window["WebSocket"] && forceFallback !== true) {
            socket = Bulldozer.WebSocket;
        } else {
            socket = Bulldozer.AjaxSocket;
        }


        // Set the socket events
        socket.onOpen = function() {
            // Update the flags
            isInitializing = false;

            // Initialize the connection
            socket.send(prepareSendMsg(""));
        };



// TODO: Don't reconnect with init! The session ID changes always... Use reconnect() instead...


        socket.onClose = function() {
            connectionError(isInitializing, sessionID, socketAccessToken);
        };

        socket.onError = function() {
            console.log(socket.type() + ": a connection error occurred!");

            connectionError(isInitializing, sessionID, socketAccessToken);
        };

        socket.onMessage = function(data) {
            // Initialize the socket session
            if (!handleInitializeSession(data)) {
                // Invalidate the close and error methods.
                socket.onClose = socket.onError = null;

                // Reconnect and reinitialize this session.
                Bulldozer.socket.reconnect();
                return;
            }

            // Trigger the custom bulldozer ready event if this
            // is the first successfull socket connection.
            if (!documentReady) {
                documentReady = true;
                $(document).triggerHandler('bulldozer.ready');
            }

            // On success update the callback which handles incomming messages
            socket.onMessage = function(data) {
                handleReceivedData(data);
            };
        };


        // Connect to the server
        socket.open();
    };

    // Send the data object to the server. The data object is converted into a string...
    this.send = function(type, data) {
        var str = SocketKey.Task + '=' + String(type) + '&';
        for (var p in data) {
            if (data.hasOwnProperty(p)) {
                str += p + '=' + Bulldozer.utils.escapeData(data[p]) + '&';
            }
        }

        socket.send(prepareSendMsg(str));
    };

    this.reconnect = function() {
// TODO: On error maybe a full webpage refresh?

        $.ajax({
            url: "/bulldozer/reconnect",
            type: "POST",
            data: {
                id: instanceID
            },
            dataType: "text",
            timeout: 7000,
            success: function (data) {
                // Split the received data
                var list = data.split('&');
                
                // Check if enough elements exist
                if (list.length < 2) {
                    console.log("Failed to reconnect socket session! Received list length is invalid: '" + data + "'");
                    // TODO
                    return;
                }

                // Initialize the socket session.
                Bulldozer.socket.init(list[0], list[1]);
            },
            error: function () {
                console.log("failed to reconnect to server!");
                // TODO
            }
        });
    };
};