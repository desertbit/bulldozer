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
        RefreshRequest:     "req_refresh",
        Ping:               "ping",
        Pong:               "pong"
    };



    /*
     * Private Variables
     */

    var documentReady = false;
    var socket;
    var sid, instanceID, token;
    var timeoutConnectionLost = false;
    var reconnectCount = 0;




    /*
     * Private Methods
     */

    var prepareSendMsg = function(data) {
        return "sid=" + sid + "&tok=" + token + "&" + data;
    };

    var stopConnectionLostTimeout = function() {
        // Stop the timeout timer
        if (timeoutConnectionLost !== false) {
            clearTimeout(timeoutConnectionLost);
            timeoutConnectionLost = false;
        }
    };

    var resetConnectionLostTimeout = function() {
        // Stop the timeout timer
        if (timeoutConnectionLost !== false) {
            clearTimeout(timeoutConnectionLost);
        }
        
        // Hide the connection lost widget.
        Bulldozer.connectionLost.hide();

        // Start the timer again.
        // A Ping message should arrive each 30 seconds from the server.
        // If nothing happens for 40 seconds, then show the connection lost widget.
        timeoutConnectionLost = setTimeout(function () {
            // Update the flag.
            timeoutConnectionLost = false;

            // Show the connection lost widget.
            Bulldozer.connectionLost.show();
        }, 40000);
    };

    var handleReceivedData = function(data) {
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
                    "Warning! Invalid data received from server! Please reload this webpage and notify the site administrator!",
                    "Error data: '" + data + "'");
                return;
            }

            // Set the new token and the data variable
            token = data.substring(0, i);
            data = data.substr(i + 1);

            // Check if the server requests a pong reply
            if (data === SocketData.Ping) {
                // Send the pong message.
                socket.send(prepareSendMsg(SocketKey.Task + "=" + SocketData.Pong + "&"));

                // Reset the timeout timer
                resetConnectionLostTimeout();
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
            console.log("Failed to initialize socket session! Received emtpy data from server!");
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

        // Reset the reconnect count.
        reconnectCount = 0;

        // Notify, that the reconnect was successfully.
        Bulldozer.connectionLost.reconnectSuccess();

        // Hide the connection lost widget.
        Bulldozer.connectionLost.hide();

        // Reset the timeout timer
        resetConnectionLostTimeout();

        return true;
    };

    // The callback function for connection errors.
    var connectionError = function() {
        // Show the connection lost widget.
        Bulldozer.connectionLost.show();

        // Notify, that the connection failed.
        Bulldozer.connectionLost.reconnectFailed();

        // Stop the reset timeout timer.
        stopConnectionLostTimeout();

        // Force fallback flag.
        var fallback = false;

        // Increment the count.
        reconnectCount += 1;

        // Fallback to the ajax socket if this is the last attempt.
        if (reconnectCount == reconnectAttempts
            && socket.type() !== Bulldozer.AjaxSocket.type()) {
            // Reconnect with the ajax socket
            console.log("falling back to ajax socket...");
            fallback = true;
        }

        if (reconnectCount <= reconnectAttempts) {
            // Try to reconnect
            setTimeout(function() {
                Bulldozer.socket.reconnect(fallback);
            }, 1500);
        }
        else {
            console.log("giving up...");

            // Show the connection lost widget.
            Bulldozer.connectionLost.show();
        }
    };



    /*
     * Public Methods
     */

    this.init = function (sessionID, socketAccessToken, forceFallback) {
        if (!sessionID || !socketAccessToken) {
            console.log("empty session ID or socket access token!");
            return;
        }

        // Stop the reset timeout timer.
        stopConnectionLostTimeout();

        // Set the session ID and token
        sid = sessionID;
        token = socketAccessToken;

        var waitDuration = 0;

        // Reset the previous socket if set
        if (socket) {
            socket.onOpen = undefined;
            socket.onClose = undefined;
            socket.onMessage = undefined;
            socket.onError = undefined;

            // Reset the socket.
            socket.reset();

            // Set the wait duration to a short timeout.
            waitDuration = 300;
        }

        // Wait for a short timeout, if set.
        setTimeout(function() {
            // Choose the socket layer depending on the browser support
            if (window["WebSocket"] && forceFallback !== true) {
                socket = Bulldozer.WebSocket;
            } else {
                socket = Bulldozer.AjaxSocket;
            }


            // Set the socket events
            socket.onOpen = function() {
                // Initialize the connection
                socket.send(prepareSendMsg(""));
            };

            socket.onClose = function() {
                connectionError();
            };

            socket.onError = function() {
                console.log(socket.type() + ": a connection error occurred!");
                connectionError();
            };

            socket.onMessage = function(data) {
                // Initialize the socket session
                if (!handleInitializeSession(data)) {
                    // Show an error message box
                    Bulldozer.utils.showErrorMessageBox("Error", "Failed to initialize web session! Please reload this webpage and try again...");
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
        }, waitDuration);
    };

    // Send the data object to the server. The data object is converted into a string.
    // A boolean is returned, indicating if the data has been send to the server.
    this.send = function(type, data) {
        // Reconnect the socket session if the connection is lost.
        if (Bulldozer.connectionLost.connectionLost()) {
            Bulldozer.socket.reconnect();
            return false;
        }

        var str = SocketKey.Task + '=' + String(type) + '&';
        for (var p in data) {
            if (data.hasOwnProperty(p)) {
                str += p + '=' + Bulldozer.utils.escapeData(data[p]) + '&';
            }
        }

        socket.send(prepareSendMsg(str));
        return true;
    };

    this.reconnect = function(forceFallback) {
        $.ajax({
            url: "/bulldozer/reconnect",
            type: "POST",
            data: {
                id: instanceID
            },
            dataType: "text",
            timeout: 7000,
            success: function (data) {
                if (data === SocketData.RefreshRequest) {
                    // Reload the page.
                    window.location.reload();
                    return;
                }

                // Split the received data
                var list = data.split('&');
                
                // Check if enough elements exist
                if (list.length < 2) {
                    console.log("Failed to reconnect socket session! Received list length is invalid: '" + data + "'");
                    // Show an error message box
                    Bulldozer.utils.showErrorMessageBox("Error", "Failed to reconnect to server! Please reload this webpage and try again...");
                    return;
                }

                // Initialize the socket session.
                Bulldozer.socket.init(list[0], list[1], forceFallback);
            },
            error: function () {
                console.log("failed to reconnect to server!");
                connectionError();
            }
        });
    };
};