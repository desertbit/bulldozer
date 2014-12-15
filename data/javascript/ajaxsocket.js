/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

Bulldozer.fn.AjaxSocket = new function () {
    /*
     * Private Variables
     */

     var uid, pollToken;
     var sendTimeout = 7000;
     var pollTimeout = 45000;

     var pollXhr = false;
     var sendXhr = false;

     var Type = {
        Init: "init"
     };

    /*
     * Private Methods
     */

    var triggerError = function() {
        // Kill the ajax requests
        if (pollXhr) {
            pollXhr.abort();
        }
        if (sendXhr) {
            sendXhr.abort();
        }

        // Trigger the event
        Bulldozer.AjaxSocket.onError();
    };

    var poll = function () {
        pollXhr = $.ajax({
            url: "/bulldozer/ajax/poll",
            success: function (data) {
                pollXhr = false;

                // Split the new token from the rest of the data
                var i = data.indexOf('&');
                if (i < 0) {
                    console.log("ajaxsocket: failed to split poll token from data! '&' not found! data: " + data);
                    triggerError();
                    return;
                }

                // Set the new token and the data variable
                pollToken = data.substring(0, i);
                data = data.substr(i + 1);

                // Start the next poll request
                poll();

                // Call the event
                Bulldozer.AjaxSocket.onMessage(data);
            },
            error: function () {
                pollXhr = false;
                triggerError();
            },
            type: "POST",
            data: uid + "&" + pollToken,
            dataType: "text",
            timeout: pollTimeout
        });
    };


    var send = function (data, callback) {
        sendXhr = $.ajax({
            url: "/bulldozer/ajax",
            success: function (data) {
                sendXhr = false;

                if (callback) {
                    callback(data);
                }
            },
            error: function () {
                sendXhr = false;
                triggerError();
            },
            type: "POST",
            data: data,
            dataType: "text",
            timeout: sendTimeout
        });
    };



    /*
     * Public Methods
     */

    this.onOpen;
    this.onClose;
    this.onMessage;
    this.onError;
    
    this.type = function () {
        return "ajaxsocket";
    };

    this.open = function () {
        // Initialize the ajax socket session
        send(Type.Init, function (data) {
            // Get the uid and token string
            var i = data.indexOf('&');
            if (i < 0) {
                console.log("ajaxsocket: failed to split uid and poll token from data! '&' not found! data: " + data);
                triggerError();
                return;
            }

            // Set the unique id and token
            uid = data.substring(0, i);
            pollToken = data.substr(i + 1);

            // Start the long polling process
            poll();

            // Trigger the event
            Bulldozer.AjaxSocket.onOpen();
        });
    };

    this.send = function (data) {
        // Always prepend the uid to the data
        send(uid + "&" + data);
    };
};