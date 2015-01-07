/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

Bulldozer.fn.WebSocket = new function () {
    /*
     * Private Variables
     */

    var ws;


    /*
     * Public Methods
     */

    this.onOpen;
    this.onClose;
    this.onMessage;
    this.onError;

    this.type = function () {
        return "websocket";
    };

    this.open = function () {
        try {
            // Generate the websocket url
            var url = "ws://";
            if (window.location.protocol === 'https:') {
                url = "wss://";
            }
            url += window.location.host + "/bulldozer/ws";

            // Open the websocket connection
            ws = new WebSocket(url);

            // Set the callback handlers
            ws.onmessage = function(event) {
                Bulldozer.WebSocket.onMessage(event.data.toString());
            };

            ws.onerror = function() {
                if (Bulldozer.WebSocket.onError) {
                    Bulldozer.WebSocket.onError();
                }
            };

            ws.onclose = function() {
                if (Bulldozer.WebSocket.onClose) {
                    Bulldozer.WebSocket.onClose();
                }
            };

            ws.onopen = function() {
                Bulldozer.WebSocket.onOpen();
            };
        } catch (e) {
            if (Bulldozer.WebSocket.onError) {
                Bulldozer.WebSocket.onError();
            }
        }
    };

    this.send = function (data) {
        // Send the data to the server
        ws.send(data);     
    };

    this.reset = function() {
        // Close the websocket if defined.
        if (ws) {
            ws.close();
        }

        ws = undefined;
    };
};