/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */


/*
 * Bulldozer Utilities
 */
 
Bulldozer.utils = {
	// Replace all '\' to '\\' and '&' with '\&'
    escapeData : function(data) {
        return data.toString().replace(/\\|&/g, '\\$&');
    },


    showErrorMessageBox : function(title, text, error) {
        // First escape the strings
        title = Kepler.utils.escapeHTML(title);
        text = Kepler.utils.escapeHTML(text);

        // Prepare the messagebox body
        var body = '<div class="topbar alert"><div class="icon"></div><div class="title"><h3>' + title + '</h3></div></div><div class="kepler grid"><div class="large-12 column"><p>' + text + '</p>';

        if (error) {
            body += "<br><code>" + Kepler.utils.escapeHTML(error) + "</code>";
        }

        body += '</div><div class="large-12 column"><hr></hr></div><div class="large-12 column"><a class="kepler button expand close-modal">OK</a></div></div>';

        // Show the error messagebox
        this.addAndShowTmpModal(body,
            {
                closable: false,
                zIndex: 10001
            });
    },


    addAndShowTmpModal : function (body, options) {
        var settings = $.extend({
            domId: false,
            closable: true,
            class: "radius shadow",
            zIndex: 'auto'
        }, options);

        // Check if the body content is valid
        if (!body) {
            console.log("error: addAndShowTmpModal: body is invalid!");
            return;
        }

        // Create the dialog modal object
        var modal = $('<div class="kepler modal"></div>');
        if (settings.class) {
            modal.addClass(settings.class.toString());
        }
        if (settings.domId) {
            modal.attr("id", settings.domId.toString());
        }

        // Add the modal body
        modal.append(body);
        
        // Add the close button if required
        if (settings.closable) {
            var closeButton = '<a class="close-modal">&#215;</a>';
            var topbar = modal.find(".topbar:first");
            
            // Attach the close button into the topbar if present.
            // Otherwise add it to the modal context.
            if (topbar.length > 0) {
                topbar.append(closeButton);
            } else {
                modal.prepend(closeButton);
            }
        }
        
        // Add the modal to the end of the body tag
        modal.appendTo($('body'));

        // Open the modal
        Kepler.modal.open(modal, {
            closeOnBackdropClick: settings.closable,
            removeOnClose: true,
            zIndex: settings.zIndex
        });

        // Execute the kepler init method to apply all new kepler changes
        Kepler.init();
    }
};