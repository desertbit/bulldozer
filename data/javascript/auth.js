/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */


/*
 * Bulldozer Authentication
 */

Bulldozer.fn.auth = new function () {
    /*
     * Public Methods
     */

    this.hashPassword = function (pw, token) {
        return CryptoJS.SHA256(CryptoJS.SHA256(pw) + Bulldozer.socket.sessionID() + token).toString();
    };
};