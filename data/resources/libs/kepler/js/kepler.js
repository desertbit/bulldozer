/*!
 * Kepler v0.7.5 ()
 * Copyright 2014 
 * Licensed under    
 */
var Kepler = new function() {
    /*
     * Public Variables
     */

	this.module = Object.getPrototypeOf(this);
    this.utils;



    /*
     * Public Methods
     */

	// Parse the DOM structure and initialize all Kepler UI elements
	this.init = function() {
		// Trigger the init event
		$(this).triggerHandler("kepler.init");
		return this;
	};

	// Event which is triggered on each Kepler.init() call
	this.onInit = function(callback) {
		$(this).on("kepler.init", callback);
	};



    /*
     * JQuery Extensions
     */

    (function($){
        // Trigger kepler.element.remove if the element is removed from the DOM content
        $.cleanData = (function( orig ) {
            return function( elems ) {
                for ( var i = 0, elem; (elem = elems[i]) != null; i++ ) {
                    $( elem ).triggerHandler("kepler.element.remove");
                }
                orig( elems );
            };
        })( $.cleanData );
    })(jQuery);



    /*
     * Miscellaneous
     */

    // Private scan function to load code highlighting
    var loadCodeHighlighting = function() {
        // Scan if HighlightJS if present
        if (typeof hljs !== 'undefined') {
            $('code').each(function(i, block) {
                hljs.highlightBlock(block);
            });
        }
    };

    // Kepler init hook
    this.onInit(function() {
        // Scan all code tags if hightlight.js is present
        loadCodeHighlighting();
    });

    // On document ready
    $(document).ready(function() {
        // Enable FastClick if present
        if (typeof FastClick !== 'undefined') {
            // Don't attach to body if undefined
            if (typeof document.body !== 'undefined') {
                FastClick.attach(document.body);
            }
        }
    });
};

/*
 * Kepler Utilities
 */
Kepler.utils = {
    // Return a jquery object.
    // valid data values: object, JQeury object or DOM selector
    getJQueryObj : function(data) {
        // Check if valid
        if (!data) return false;

        // Get the jQuery object if it isn't one
        if ( !(data instanceof jQuery) ) {
            data = $(data);
        }

        // Check if the object exists
        if (data.length <= 0) return false;

        // Return
        return data;
    },

    escapeHTML : function(str) {
        return String(str)
            .replace(/&/g, '&amp;')
            .replace(/"/g, '&quot;')
            .replace(/'/g, '&#39;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;');
    },

    // Center an element on the screen
    centerElement : function(el, container) {
        el.css("position","absolute");
        el.css("top", Math.max(0, ((container.height() - el.outerHeight()) / 2) + container.scrollTop()) + "px");
        el.css("left", Math.max(0, ((container.width() - el.outerWidth()) / 2) + container.scrollLeft()) + "px");
        return this;
    },

    // Center an element vertical with the top and bottom margin
    centerElementVerticalMargin : function(el, container) {
        var elHeight = el.height();
        var containerHeight = container.height();

        // Only position if the element is smaller than the container
        if (elHeight <= containerHeight) {
            el.css("margin-top",(containerHeight - elHeight)/2 + 'px');
        }

        return this;
    },
    
    getScrollbarWidth : function() {
        var outer = document.createElement("div");
        outer.style.visibility = "hidden";
        outer.style.width = "100px";
        outer.style.msOverflowStyle = "scrollbar"; // needed for WinJS apps

        document.body.appendChild(outer);

        var widthNoScroll = outer.offsetWidth;
        // force scrollbars
        outer.style.overflow = "scroll";

        // add innerdiv
        var inner = document.createElement("div");
        inner.style.width = "100%";
        outer.appendChild(inner);        

        var widthWithScroll = inner.offsetWidth;

        // remove divs
        outer.parentNode.removeChild(outer);

        return widthNoScroll - widthWithScroll;
    },

    // Based on foundation.js utils
    //
    // Description:
    //    Executes a function a max of once every n milliseconds
    //
    // Arguments:
    //    Func (Function): Function to be throttled.
    //
    //    Delay (Integer): Function execution threshold in milliseconds.
    //
    // Returns:
    //    Lazy_function (Function): Function with throttling applied.
    throttle : function (func, delay) {
        var timer = null;

        return function () {
            var context = this, args = arguments;

            if (timer == null) {
                timer = setTimeout(function () {
                    func.apply(context, args);
                    timer = null;
                }, delay);
            }
        };
    },

    // Based on foundation.js utils
    //
    // Description:
    //    Executes a function when it stops being invoked for n seconds
    //    Modified version of _.debounce() http://underscorejs.org
    //
    // Arguments:
    //    Func (Function): Function to be debounced.
    //
    //    Delay (Integer): Function execution threshold in milliseconds.
    //
    //    Immediate (Bool): Whether the function should be called at the beginning
    //    of the delay instead of the end. Default is false.
    //
    // Returns:
    //    Lazy_function (Function): Function with debouncing applied.
    debounce : function (func, delay, immediate) {
        var timeout, result;
        return function () {
            var context = this, args = arguments;
            var later = function () {
                timeout = null;
                if (!immediate) result = func.apply(context, args);
            };
            var callNow = immediate && !timeout;
            clearTimeout(timeout);
            timeout = setTimeout(later, delay);
            if (callNow) result = func.apply(context, args);
                return result;
        };
    },

    // Based on foundation.js utils
    //
    // Description:
    //    Parses data-kepler-options attribute
    //
    // Arguments:
    //    El (jQuery Object): Element to be parsed.
    //
    // Returns:
    //    Options (Javascript Object): Contents of the element's data-options
    //    attribute.
    dataOptions : function (el, data_attr_name) {
        data_attr_name = data_attr_name || 'kepler-options';
        var opts = {}, ii, p, opts_arr;
        var cached_options = el.data(data_attr_name);

        if (typeof cached_options === 'object') {
            return cached_options;
        }

        opts_arr = (cached_options || ':').split(';');
        ii = opts_arr.length;

        function isNumber (o) {
            return ! isNaN (o-0) && o !== null && o !== "" && o !== false && o !== true;
        }

        function trim (str) {
            if (typeof str === 'string') return $.trim(str);
            return str;
        }

        while (ii--) {
            p = opts_arr[ii].split(':');
            p = [p[0], p.slice(1).join(':')];

            if (/true/i.test(p[1])) p[1] = true;
            if (/false/i.test(p[1])) p[1] = false;
            if (isNumber(p[1])) {
                if (p[1].indexOf('.') === -1) {
                    p[1] = parseInt(p[1], 10);
                } else {
                    p[1] = parseFloat(p[1]);
                }
            }

            if (p.length === 2 && p[0].length > 0) {
                opts[trim(p[0])] = trim(p[1]);
            }
        }

        return opts;
    },


    // Retuns the positions of all 4 sides of an element relative to its
    // offset parent including the parent's scroll position.
    // Use this to position elements independant of their parent.
    getRelativePosition : function(element) {
        //get jQuery object
        var $el = Kepler.utils.getJQueryObj(element);
        var offset = $el.offset();
        var offsetParent = $el.offsetParent();
        var offsetParentOffset = offsetParent.offset();
        var pos = {
            width: $el.outerWidth(),
            height: $el.outerHeight()
        };

        pos.left = offset.left - offsetParentOffset.left + offsetParent.scrollLeft();
        pos.right = offsetParent.innerWidth() - pos.left - pos.width;
        pos.top = offset.top - offsetParentOffset.top + offsetParent.scrollTop();
        pos.bottom =  offsetParent.innerHeight() - pos.top - pos.height;

        return pos;
    },

    
    // Returns a boolean if the element fits on screen
    fitsOnScreen : function(element) {
        var $el = Kepler.utils.getJQueryObj(element);
        var win = $(window);
        var viewport = {
            top: win.scrollTop(),
            left: win.scrollLeft()
        };

        viewport.right = viewport.left + win.width();
        viewport.bottom = viewport.top + win.height();

        var bounds = $el.offset();
        bounds.right = bounds.left + $el.outerWidth();
        bounds.bottom = bounds.top + $el.outerHeight();

        return viewport.right - bounds.right > 0
            && viewport.bottom - bounds.bottom > 0;
    },

    
    // Returns a boolean if the element is visible on the screen
    visibleOnScreen : function(element) {
        var $window = $(window);
        var viewport_top = $window.scrollTop();
        var viewport_height = $window.height();
        var viewport_bottom = viewport_top + viewport_height;
        var $elem = $(element);
        var top = $elem.offset().top;
        var height = $elem.height();
        var bottom = top + height;

        return (top >= viewport_top && top < viewport_bottom)
            || (bottom > viewport_top && bottom <= viewport_bottom)
            || (height > viewport_height && top <= viewport_top && bottom >= viewport_bottom);
    },
    
    // Scroll metod scrolls html page to position of an element
    // optional is an offset
    scrollTo : function(element, duration, offset){
        element = Kepler.utils.getJQueryObj(element);

        if (!$.isNumeric(duration) || duration < 0) duration = 650;
        if (!$.isNumeric(offset)) offset = 0;

        // Cheack if element is valid
        if(!element){
            console.log("Error: Unable to get scroll target. Invalid element passed (utils.scrollTo)");
            return false;
        }
        // Animate scroll to target position
        $('html,body').animate({
        scrollTop: element.offset().top + offset}, duration);
    },
    //This code loads the Script code asynchroniously.
    asyncScriptLoad : function(url){
        var tag = document.createElement('script');

        tag.src = "https://www.youtube.com/iframe_api";
        var firstScriptTag = document.getElementsByTagName('script')[0];
        firstScriptTag.parentNode.insertBefore(tag, firstScriptTag);

    }
};

/*! Copyright (c) 2013 Brandon Aaron (http://brandon.aaron.sh)
 * Licensed under the MIT License (LICENSE.txt).
 *
 * Version: 3.1.12
 *
 * Requires: jQuery 1.2.2+
 */

(function (factory) {
    if ( typeof define === 'function' && define.amd ) {
        // AMD. Register as an anonymous module.
        define(['jquery'], factory);
    } else if (typeof exports === 'object') {
        // Node/CommonJS style for Browserify
        module.exports = factory;
    } else {
        // Browser globals
        factory(jQuery);
    }
}(function ($) {

    var toFix  = ['wheel', 'mousewheel', 'DOMMouseScroll', 'MozMousePixelScroll'],
        toBind = ( 'onwheel' in document || document.documentMode >= 9 ) ?
                    ['wheel'] : ['mousewheel', 'DomMouseScroll', 'MozMousePixelScroll'],
        slice  = Array.prototype.slice,
        nullLowestDeltaTimeout, lowestDelta;

    if ( $.event.fixHooks ) {
        for ( var i = toFix.length; i; ) {
            $.event.fixHooks[ toFix[--i] ] = $.event.mouseHooks;
        }
    }

    var special = $.event.special.mousewheel = {
        version: '3.1.12',

        setup: function() {
            if ( this.addEventListener ) {
                for ( var i = toBind.length; i; ) {
                    this.addEventListener( toBind[--i], handler, false );
                }
            } else {
                this.onmousewheel = handler;
            }
            // Store the line height and page height for this particular element
            $.data(this, 'mousewheel-line-height', special.getLineHeight(this));
            $.data(this, 'mousewheel-page-height', special.getPageHeight(this));
        },

        teardown: function() {
            if ( this.removeEventListener ) {
                for ( var i = toBind.length; i; ) {
                    this.removeEventListener( toBind[--i], handler, false );
                }
            } else {
                this.onmousewheel = null;
            }
            // Clean up the data we added to the element
            $.removeData(this, 'mousewheel-line-height');
            $.removeData(this, 'mousewheel-page-height');
        },

        getLineHeight: function(elem) {
            var $elem = $(elem),
                $parent = $elem['offsetParent' in $.fn ? 'offsetParent' : 'parent']();
            if (!$parent.length) {
                $parent = $('body');
            }
            return parseInt($parent.css('fontSize'), 10) || parseInt($elem.css('fontSize'), 10) || 16;
        },

        getPageHeight: function(elem) {
            return $(elem).height();
        },

        settings: {
            adjustOldDeltas: true, // see shouldAdjustOldDeltas() below
            normalizeOffset: true  // calls getBoundingClientRect for each event
        }
    };

    $.fn.extend({
        mousewheel: function(fn) {
            return fn ? this.bind('mousewheel', fn) : this.trigger('mousewheel');
        },

        unmousewheel: function(fn) {
            return this.unbind('mousewheel', fn);
        }
    });


    function handler(event) {
        var orgEvent   = event || window.event,
            args       = slice.call(arguments, 1),
            delta      = 0,
            deltaX     = 0,
            deltaY     = 0,
            absDelta   = 0,
            offsetX    = 0,
            offsetY    = 0;
        event = $.event.fix(orgEvent);
        event.type = 'mousewheel';

        // Old school scrollwheel delta
        if ( 'detail'      in orgEvent ) { deltaY = orgEvent.detail * -1;      }
        if ( 'wheelDelta'  in orgEvent ) { deltaY = orgEvent.wheelDelta;       }
        if ( 'wheelDeltaY' in orgEvent ) { deltaY = orgEvent.wheelDeltaY;      }
        if ( 'wheelDeltaX' in orgEvent ) { deltaX = orgEvent.wheelDeltaX * -1; }

        // Firefox < 17 horizontal scrolling related to DOMMouseScroll event
        if ( 'axis' in orgEvent && orgEvent.axis === orgEvent.HORIZONTAL_AXIS ) {
            deltaX = deltaY * -1;
            deltaY = 0;
        }

        // Set delta to be deltaY or deltaX if deltaY is 0 for backwards compatabilitiy
        delta = deltaY === 0 ? deltaX : deltaY;

        // New school wheel delta (wheel event)
        if ( 'deltaY' in orgEvent ) {
            deltaY = orgEvent.deltaY * -1;
            delta  = deltaY;
        }
        if ( 'deltaX' in orgEvent ) {
            deltaX = orgEvent.deltaX;
            if ( deltaY === 0 ) { delta  = deltaX * -1; }
        }

        // No change actually happened, no reason to go any further
        if ( deltaY === 0 && deltaX === 0 ) { return; }

        // Need to convert lines and pages to pixels if we aren't already in pixels
        // There are three delta modes:
        //   * deltaMode 0 is by pixels, nothing to do
        //   * deltaMode 1 is by lines
        //   * deltaMode 2 is by pages
        if ( orgEvent.deltaMode === 1 ) {
            var lineHeight = $.data(this, 'mousewheel-line-height');
            delta  *= lineHeight;
            deltaY *= lineHeight;
            deltaX *= lineHeight;
        } else if ( orgEvent.deltaMode === 2 ) {
            var pageHeight = $.data(this, 'mousewheel-page-height');
            delta  *= pageHeight;
            deltaY *= pageHeight;
            deltaX *= pageHeight;
        }

        // Store lowest absolute delta to normalize the delta values
        absDelta = Math.max( Math.abs(deltaY), Math.abs(deltaX) );

        if ( !lowestDelta || absDelta < lowestDelta ) {
            lowestDelta = absDelta;

            // Adjust older deltas if necessary
            if ( shouldAdjustOldDeltas(orgEvent, absDelta) ) {
                lowestDelta /= 40;
            }
        }

        // Adjust older deltas if necessary
        if ( shouldAdjustOldDeltas(orgEvent, absDelta) ) {
            // Divide all the things by 40!
            delta  /= 40;
            deltaX /= 40;
            deltaY /= 40;
        }

        // Get a whole, normalized value for the deltas
        delta  = Math[ delta  >= 1 ? 'floor' : 'ceil' ](delta  / lowestDelta);
        deltaX = Math[ deltaX >= 1 ? 'floor' : 'ceil' ](deltaX / lowestDelta);
        deltaY = Math[ deltaY >= 1 ? 'floor' : 'ceil' ](deltaY / lowestDelta);

        // Normalise offsetX and offsetY properties
        if ( special.settings.normalizeOffset && this.getBoundingClientRect ) {
            var boundingRect = this.getBoundingClientRect();
            offsetX = event.clientX - boundingRect.left;
            offsetY = event.clientY - boundingRect.top;
        }

        // Add information to the event object
        event.deltaX = deltaX;
        event.deltaY = deltaY;
        event.deltaFactor = lowestDelta;
        event.offsetX = offsetX;
        event.offsetY = offsetY;
        // Go ahead and set deltaMode to 0 since we converted to pixels
        // Although this is a little odd since we overwrite the deltaX/Y
        // properties with normalized deltas.
        event.deltaMode = 0;

        // Add event and delta to the front of the arguments
        args.unshift(event, delta, deltaX, deltaY);

        // Clearout lowestDelta after sometime to better
        // handle multiple device types that give different
        // a different lowestDelta
        // Ex: trackpad = 3 and mouse wheel = 120
        if (nullLowestDeltaTimeout) { clearTimeout(nullLowestDeltaTimeout); }
        nullLowestDeltaTimeout = setTimeout(nullLowestDelta, 200);

        return ($.event.dispatch || $.event.handle).apply(this, args);
    }

    function nullLowestDelta() {
        lowestDelta = null;
    }

    function shouldAdjustOldDeltas(orgEvent, absDelta) {
        // If this is an older event and the delta is divisable by 120,
        // then we are assuming that the browser is treating this as an
        // older mouse wheel event and that we should divide the deltas
        // by 40 to try and get a more usable deltaFactor.
        // Side note, this actually impacts the reported scroll distance
        // in older browsers and can cause scrolling to be slower than native.
        // Turn this off by setting $.event.special.mousewheel.settings.adjustOldDeltas to false.
        return special.settings.adjustOldDeltas && orgEvent.type === 'mousewheel' && absDelta % 120 === 0;
    }

}));

/*
 * Kepler Module
 */
Kepler.module.alerts = new function() {
    /*
     * Private Variables
     */

    // Default settings
    var defaultSettings = {
        fadeDuration: 400,
        remove: false       // Remove the alert element from the DOM if the close button is clicked
    };



    /*
     * Private Methods
     */

    var getSettings = function(options) {
        // Combine the settings, but preserve the objects. Create a deep copy...
        return $.extend({}, defaultSettings, options);
    };

    // Handler which removes/hides an alert if clicked
    var removeOnClickHandler = function() {
        // Get the alert object
        var obj = Kepler.utils.getJQueryObj($(this).data("kepler-alert-object"));
        if (!obj) {
            console.log("error: data attribute kepler-alert-object is invalid!");
            return;
        }

        // Get the data options if set
        var settings = getSettings(Kepler.utils.dataOptions(obj, "kepler-alert-options"));

        // Fade out the alert
        obj.fadeOut(settings.fadeDuration, function() {
            if (settings.remove) {
                // Remove the alert
                obj.remove();
            }
            else {
                // Hide the alert
                obj.hide();
            }
        });
    };



    /*
     * Public Methods
     */

    // Sets the default options for the modal module
    this.defaultOptions = function(options) {
        defaultSettings = getSettings(options);

        // Return this to allow chaining
        return this;
    };



    /*
     * Kepler Events
     */

    // Scan the DOM structure for data elements
    Kepler.onInit(function() {
        // Find all objects with the data modal attribute
        $.each($("[data-kepler-alert]"), function(i, obj) {
            // Add the close click handler if an close button exists in the alert
            var closeButton = $(obj).find(".close");
            if (closeButton && closeButton.length > 0) {
                // Set the data attribute
                closeButton.data("kepler-alert-object", $(obj));

                // Detach the event first and attach the click handler again.
                // This way no dublicate events are bound if init is called mutliple times.
                closeButton.off('click', removeOnClickHandler);
                closeButton.on('click', removeOnClickHandler);
            }
        });
    });
};

/*
 * Kepler Module
 */
Kepler.module.collapse = new function() {
    /*
     * Private Variables
     */

    // Default settings
    var defaultSettings = {
        activeClass: 'active-content', //name of active class used to ref to active elements
        activeTriggerClass: 'active',
        linkedCollapse: true, // linked collapse only allows for one active collapsable in collapse area
        collapseOnLoad: true  // should collapse elements without active class be hidden onload
    };



    /*
     * Private Methods
     */

    var getSettings = function(options) {
        // Combine the settings, but preserve the objects. Create a deep copy...
        return $.extend({}, defaultSettings, options);
    };


    // Handler which shows a collapse
    var toggleCollapsableClickHandler = function() {
        // Get the collapse content object
        var content = Kepler.utils.getJQueryObj($(this).data("kepler-collapse-content"));
        if (!content) {
            console.log("error: data attribute kepler-collapse is invalid: collapse content object not found!");
            return;
        }

        // Get the parent
        var parent = Kepler.utils.getJQueryObj($(this).data("kepler-collapse-parent"));
        if (!parent) {
            console.log("error: collapses parent not found!");
            return;
        }

        // Get the data options if set
        var settings = getSettings(Kepler.utils.dataOptions(parent, "kepler-collapse-options"));

        // Check if elements in area are linked to allow opennig only one at a time
        if(settings.linkedCollapse){

            $.each($(parent).find("." + settings.activeTriggerClass), function(i, el) {
                // Get jQuery Onject
                el = $(el);
                el.removeClass(settings.activeTriggerClass);
            });
            // Search for elements with activeClass
            $.each($(parent).find("." + settings.activeClass), function(i, el) {
                // Get jQuery Onject
                el = $(el);

                // Check if content of this and foubd element is the same if so do nothing
                if(!content.is(el)){
                    el.hide();
                    el.removeClass(settings.activeClass);


                }
            });
        }

        // toggle Collapsable
        if(content.hasClass(settings.activeClass)){
            content.hide().removeClass(settings.activeClass);
            $(this).removeClass(settings.activeTriggerClass);
        }
        else{
            content.addClass(settings.activeClass).show();
            $(this).addClass(settings.activeTriggerClass);
        }
    };



    /*
     * Public Methods
     */

    // Sets the default options for the modal module
    this.defaultOptions = function(options) {
        defaultSettings = getSettings(options);

        // Return this to allow chaining
        return this;
    };



    /*
     * Kepler Events
     */

    // Scan the DOM structure for data elements
    Kepler.onInit(function() {
        // Init variables for later settings storage
        var options;
        var settings;
        // Find all objects with the data collapses attribute
        $.each($("[data-kepler-collapse-area]"), function(i, obj) {
            // Get the jQuery object
            obj = $(obj);

            // Iterate over all collapses of the current collapse object
            $.each(obj.find("[data-kepler-collapse]"), function(i, collapse) {
                // Get the jQuery objects
                collapse = $(collapse);
                // Get the trigger(button) object
                var collapseTrigger = $(collapse.find("[data-kepler-collapse-trigger]"));
                // Get the collapse target stored in data
                var triggerTargetData = $(collapseTrigger.data("kepler-collapse-trigger"));
                // Check if trigger has a set target else search for the data content element
                var collapseTarget = (triggerTargetData.length > 0) ? triggerTargetData : $(collapse.find("[data-kepler-collapse-content]"));

                // Get settings from data options
                options = Kepler.utils.dataOptions($(obj), "kepler-dropdown-options");
                settings = getSettings(options);

                // Check if element is displayed or is has no open preset class --> if so set active class
                if(collapseTarget.is(":visible") && !collapseTarget.hasClass(settings.activeClass) && settings.collapseOnLoad) collapseTarget.hide();

                // Save the parent reference
                collapseTrigger.data("kepler-collapse-parent", obj);
                // Save the content reference
                collapseTrigger.data("kepler-collapse-content", collapseTarget);
                // Detach the event first and attach the click handler again.
                // This way no dublicate events are bound if init is called mutliple times.
                collapseTrigger.off('click', toggleCollapsableClickHandler);
                collapseTrigger.on('click', toggleCollapsableClickHandler);
            });
        });
    });
};

/*
 * Kepler Module
 */
Kepler.module.exchange = new function() {
    /*
     * Private Variables
     */

    // Default settings
    var defaultSettings = {
        attribut: false,
        removeClass: false,
        style: false,
        transition: false,
        toggleClass: false,
        addClass: false,
        triggerEvent: "click"
    };

    /*
     * Private Methods
     */

    var getSettings = function(options) {
        // Combine the settings, but preserve the objects. Create a deep copy...
        var settings = $.extend({}, defaultSettings, options);

        return settings;
    };

    var applyValues = function(param, target){
        $.each(param, function(i){
            var attr = param[i].split('=');
            target.attr(attr[0], attr[1]);
        });
    };

    var applyStyles = function(param, target){
        //$.each(param, function(i){
            //var attr = param[i].split('=');
            console.log("CSS change is temporaly disabled");
            // console.log(attr[1]);
            //target.css(attr[0], attr[1]);
        //});
    };

    var applyExchange = function(target, settings){
       if(settings.removeClass){
           target.removeClass(settings.removeClass);
       }

       if(settings.addClass){
            target.addClass(settings.addClass);
       }

        if(settings.attribut){
            // Parse attributes from settings string
            var attribut = settings.attribut.split(' ');

            // Apply attribut values
            applyValues(attribut,target);
        }


        if(settings.toggleClass){
            target.toggleClass(settings.toggleClass);
        }

        if(settings.toggleVisible){
            target.toggle();
        }
       //if(settings.style){
            //var attribut = settings.style.split(' ');
            //applyStyles(attribut, target);
       //}
    };

    var exchangeEventHandler = function(e){
        // Prevent default button behaviour
        // Get the jQuery object
        e.preventDefault();

        var $el = Kepler.utils.getJQueryObj(this);

        // Get settings with default values
        var settings = getSettings(Kepler.utils.dataOptions($el, "kepler-exchange-options"));

       // Get linked element
       var dataTarget = $el.data("kepler-exchange-target");
       var linkedElement;

       // Check if target is this object or self
       if(dataTarget === "this" || dataTarget === "self"){
           linkedElement = $el;
       }
        // Else get the refered target
       else {
           // Get linkedEl jQuery object
           linkedElement = Kepler.utils.getJQueryObj(dataTarget);
           // Error console log
           if(!linkedElement) {
               console.log("Error: linked element reference invalid.");
               return;
           }
       }
       if(settings.transition){
           linkedElement.fadeOut('fast', function(){
               applyExchange(linkedElement, settings);
            });
        }
        else {
            applyExchange(linkedElement, settings);
        }

        if(settings.transition){
            linkedElement.fadeIn('fast');
        }
    };




    /*
     * Public Methods
     */

    // Sets the default options for the modal module
    this.defaultOptions = function(options) {
        defaultSettings = getSettings(options);

        // Return this to allow chaining
        return this;
    };



    /*
     * Kepler and window events
     */

    // Scan the DOM structure for data elements
    Kepler.onInit(function() {
        $.each($("[data-kepler-exchange-target]"), function(i, el){
            // Get jQuery object
            el = Kepler.utils.getJQueryObj(el);
            // Get settings with default values
            var settings = getSettings(Kepler.utils.dataOptions(el, "kepler-exchange-options"));
            // Bind click event on element
            el.off(settings.triggerEvent, exchangeEventHandler);
            el.on(settings.triggerEvent, exchangeEventHandler);
        });
    });
};

/*
 * Kepler Module
 */
Kepler.module.modal = new function() {
    /*
     * Private Variables
     */

    // Default settings
    var defaultSettings = {
        backdrop: true,
        closeOnBackdropClick: true,
        removeOnClose: false,
        zIndex: 'auto'
    };

    var lastModalZIndex = 1000;
    var openModals = [];
    var uniqueIndex = 0;



    /*
     * Private Methods
     */

    var getSettings = function(options) {
        // Combine the settings, but preserve the objects. Create a deep copy...
        return $.extend({}, defaultSettings, options);
    };

    var getBackdrop = function(modalObj) {
        var backdrop = modalObj.parent();

        if (!backdrop || backdrop.length <= 0) return false;

        backdrop = backdrop.parent();

		if (backdrop && backdrop.length > 0 && backdrop.hasClass('kepler') && backdrop.hasClass('modal-backdrop')) {
			return backdrop;
		}

        return false;
    };

    // Handler which shows a modal if clicked
    var openOnClickHandler = function() {
        // Get the modal object
        var modal = Kepler.utils.getJQueryObj($(this).data("kepler-modal"));
        if (!modal) {
            console.log("error: invalid kepler modal data attribute: invalid modal id: object does not exists!");
            return;
        }

        // Get the data otions if set.
        // Options set on the trigger element override options set on the modal element.
        var options = $.extend(Kepler.utils.dataOptions($(modal), "kepler-modal-options"),
                               Kepler.utils.dataOptions($(this), "kepler-modal-options"));

        // Show the modal
        Kepler.modal.open(modal, options);
    };

    // Handler which closes a modal if clicked
    var closeOnClickHandler = function() {
        // Get the modal object
        var modal = Kepler.utils.getJQueryObj($(this).data("kepler-modal-close"));
        if (!modal) {
            console.log("error: invalid kepler modal data attribute: invalid modal id: object does not exists!");
            return;
        }

        // Close the modal
        Kepler.modal.close(modal);
    };

    var removeActiveModal = function(modalObj) {
        // Return if the modal is hidden
        if (!modalObj.is(":visible")) return;

        // Get the modal id
        var id = modalObj.data("kepler-modal-id");
        if (!id) return;

        // Iterate over all open modals and find the current index
        $.each(openModals, function(i, o) {
            // Continue, if this isn't the right object
            if (o.id != id) {
                return;
            }

            // Remove the current modal object from the array
            openModals.splice(i, 1);

            // Break
            return false;
        });

        var length = openModals.length;

        // Remove the open class again from the new current backdrop to show the scroll bars again
        if (length > 0) {
            openModals[length - 1].backdrop.removeClass("modal-open");
        }
        // Reset the z-index value to the start value if all modals are closed
        else {
            lastModalZIndex = 1000;

            // Remove the style class again
            $('body').removeClass("kepler-modal-open");
        }
    };

    // Handler which is triggered when the modal is removed from the DOM content
    var removeHandler = function() {
        // Remove the modal from the active list
        removeActiveModal($(this));
    };



    /*
     * Public Methods
     */

    // Sets the default options for the modal module
    this.defaultOptions = function(options) {
        defaultSettings = getSettings(options);

        // Return this to allow chaining
        return this;
    };

	// Show the modal
	this.open = function(modalObj, options) {
        // Get the settings
        var settings = getSettings(options);

        // Get a jquery object if it isn't one
        modalObj = Kepler.utils.getJQueryObj(modalObj);

        // Get the body element
		var body = $('body');

        // Get the Backdrop if present
        var backdrop = getBackdrop(modalObj);

		// Check if the modal object is already attached to the backdrop window
		if (backdrop) {
			// Just return if the modal is already visible
			if (backdrop.is(":visible")) return;

			// Show the backdrop and the modal
			backdrop.show();
		} else {
            var scrollArea = $('<div class="scroll-area"></div>');

			// Append the backdrop
			if (settings.backdrop) {
				backdrop = $('<div class="kepler modal-backdrop"></div>').appendTo(body);
			}
			// Append an invisible backdrop
			else {
				backdrop = $('<div class="kepler invisible modal-backdrop"></div>').appendTo(body);
			}

			if (settings.closeOnBackdropClick) {
				// Close the modal on click
				backdrop.click(function(e) {
                    var t = $(e.target);

                    // Don't close the modal on click events of modal children elements
                    if (!t.is(this) && !t.is(scrollArea)) return;
					Kepler.modal.close(modalObj);
				});
			}

            // Attach the scroll area to the backdrop
            scrollArea.appendTo(backdrop);

			// Detach the modal div and attach to the scroll area
			modalObj.detach().appendTo(scrollArea);

            // Show the modal, because it is hidden by default
            modalObj.show();
		}

        // Center the element
        Kepler.utils.centerElementVerticalMargin(modalObj, backdrop);

        // Set the new z-index by incrementing the global z-index
        var zIndex;
        if (settings.zIndex === 'auto') {
            zIndex = lastModalZIndex++;
        }
        else {
            zIndex = settings.zIndex;
        }
        backdrop.css("z-index", zIndex);

        // Focus the modal
        modalObj.attr("tabindex",-1).focus();

        // Add the class to the body and all open modal scroll areas to hide the scrollbars if present
        body.addClass("kepler-modal-open");
        $('.kepler.modal-backdrop').not(backdrop).addClass("modal-open");

        // Attach the remove event, which is called when the modal is removed from the DOM structure.
        // This event is used to cleanup and free unused data.
        modalObj.off('kepler.element.remove', removeHandler);
        modalObj.on('kepler.element.remove', removeHandler);

        // Create a new unique index and set it to the modal object
        uniqueIndex++;
        if (uniqueIndex > 1000000) uniqueIndex = 0;
        modalObj.data("kepler-modal-id", uniqueIndex);

        // Add the current modal and backdrop to the array
        openModals.push({
            id: uniqueIndex,
            backdrop: backdrop,
            modal: modalObj,
            settings: settings
        });

        // Add the close click handler if an close modal button exists in the modal
        var closeModalButtons = modalObj.find(".close-modal");
        if (closeModalButtons && closeModalButtons.length > 0) {
            $.each(closeModalButtons, function(i, obj) {
                // Set the data attribute
                $(obj).data("kepler-modal-close", modalObj);

                // Detach the event first and attach the click handler again.
                // This way no dublicate events are bound.
                $(obj).off('click', closeOnClickHandler);
                $(obj).on('click', closeOnClickHandler);
            });
        }

        // CSS transitions don't take effect if the element is hidden.
        // A quick solution is to force the browser to process the display change first, then the transition.
        setTimeout(function() {
            // Add the modal open class
            modalObj.addClass("open");
        }, 0);


        // Finally trigger the event
        modalObj.trigger("kepler.modal.opened");

        // Return this to allow chaining
        return this;
	};

	// Close the modal.
    // If remove is true, the modal with the backdrop is removed from the DOM structure.
	this.close = function(modalObj, remove) {
        // Get a jquery object if it isn't one
        modalObj = Kepler.utils.getJQueryObj(modalObj);

        // Detach the remove event again
        modalObj.off('kepler.element.remove', removeHandler);

        // Remove the modal open class again
        modalObj.removeClass("open");

        // Get the Backdrop if present
        var backdrop = getBackdrop(modalObj);

		// Check if the backdrop was created
		if (backdrop) {
            // Check if already hidden
			if (!backdrop.is(":visible")) return;

            // If remove is not set, then get the value from the settings
            if (jQuery.type(remove) !== "boolean") {
                var i = openModals.length - 1;
                if (i >= 0) {
                    remove = openModals[i].settings.removeOnClose;
                }
            }

            // Remove the current modal again from the active list
            removeActiveModal(modalObj);

            // Hide the backdrop and the modal
            backdrop.hide();

            // Remove the z-index again
            backdrop.css("z-index", "");

            // Finally trigger the event
            modalObj.trigger("kepler.modal.closed");

            // Remove the backdrop and the modal if required
			if (remove === true) {
				backdrop.remove();
			}
        }

        // Return this to allow chaining
        return this;
	};

    // Close all modals
    // If remove is true, all open modals are removed from the DOM structure.
    this.closeAll = function(remove) {
        // Iterate over all open modals
        while(openModals.length > 0) {
            var o = openModals[openModals.length - 1];

            // Close the modal and remove it from the array
            Kepler.modal.close(o.modal, remove);
        }

        // Return this to allow chaining
        return this;
    };



    /*
     * Public Events
     */

    // Triggered when the modal is opened
	this.opened = function(modalObj, callback) {
        // Get a jquery object if it isn't one
        modalObj = Kepler.utils.getJQueryObj(modalObj);

        // Attach the event
		modalObj.on("kepler.modal.opened", callback);

        // Return this to allow chaining
        return this;
	};

    // Triggered when the modal is closed
    this.closed = function(modalObj, callback) {
        // Get a jquery object if it isn't one
        modalObj = Kepler.utils.getJQueryObj(modalObj);

        // Attach the event
		modalObj.on("kepler.modal.closed", callback);

        // Return this to allow chaining
        return this;
	};



    /*
     * Kepler Events
     */

    // Scan the DOM structure for data elements
    Kepler.onInit(function() {
        // Find all objects with the data modal attribute
        $.each($("[data-kepler-modal]"), function(i, obj) {
            // Detach the event first and attach the click handler again.
            // This way no dublicate events are bound if init is called mutliple times.
            $(obj).off('click', openOnClickHandler);
            $(obj).on('click', openOnClickHandler);
        });

        // Find all objects with the close data modal attribute
        $.each($("[data-kepler-modal-close]"), function(i, obj) {
            // Detach the event first and attach the click handler again.
            // This way no dublicate events are bound if init is called mutliple times.
            $(obj).off('click', closeOnClickHandler);
            $(obj).on('click', closeOnClickHandler);
        });
    });
};

/*
 * Kepler Module
 */
Kepler.module.nebula = new function() {
    /*
     * Private Variables
     */

    // Default settings
    var defaultSettings = {
        activeClass: 'active',
        throttleDelay: 50, // calculation throttling to increase framerate
        fixedTop: 0, // top distance in pixels assigend to the fixed element on scroll
        autoWidth: true, // Automatically adjust the nebula object width
        slideOutDuration: 300
    };

    var State = {
        NORMAL : 0,
        FIXEDTOP : 1,
        ENDLINEPASSED : 2
    };

    var nebulaObjects = [];
    var uniqueIndex = 0;



    /*
     * Private Methods
     */

    var getSettings = function(options) {
        // Combine the settings, but preserve the objects. Create a deep copy...
        var settings = $.extend({}, defaultSettings, options);

        // Be sure the fixed top variable is an integer
        settings.fixedTop = parseInt(settings.fixedTop, 10);

        return settings;
    };

    // Handler which is triggered when the nebula object is removed from the DOM content
    var removeHandler = function() {
        // Get the array index
        var id = $(this).data("kepler-nebula-id");
        if (!id) return;

        // Iterate over all nebula objects and find the current index
        $.each(nebulaObjects, function(i, o) {
            // Continue, if this isn't the right object
            if (o.id != id) {
                return;
            }

            // Remove the object from the array
            nebulaObjects.splice(i, 1);

            // Break
            return false;
        });
    };

    var updateNebulaPositions = Kepler.utils.throttle(function() {
        var scrollTop = $(document).scrollTop();

        // Iterate over all active nebula objects
        $.each(nebulaObjects, function(i, o) {
            // Just continue the loop if the nebula object is not visible
            if (!o.object.is(":visible")) return;

            var endLinePassed = false;

            // Check if the endline is passed
            if (o.endLine && (o.endLine.offset().top - scrollTop - o.object.outerHeight(true)) <= 0) {
                endLinePassed = true;
            }

            var offset;

            // Calculate the offset
            if (o.dummy.is(":visible")) {
                offset = o.dummy.offset().top - o.settings.fixedTop - scrollTop;
            }
            else {
                offset = o.object.offset().top - o.settings.fixedTop - scrollTop;
            }

            // Check if it is required to attach or detach a nebula object
            if (offset <= 0 && !endLinePassed) {
                // Check if already attached to the top
                if (o.state == State.FIXEDTOP) {
                    // Resize the nebula object to the dummy width
                    if (o.settings.autoWidth) o.object.css("width", o.dummy.width() + "px");

                    return;
                }

                // Stop any animation
                o.object.stop();

                // Set the active class
                o.object.addClass(o.settings.activeClass);

                // Save the current position and top values
                if (o.state != State.ENDLINEPASSED) {
                    o.settings.previousPosition = o.object.css("position");
                    o.settings.previousTop = o.object.css("top");
                }

                // Set the dummy height to the same height as the original object
                o.dummy.css("height", o.object.outerHeight(true) + "px");

                // Show the dummy
                o.dummy.show();

                // Set the position of the nebula object to fixed
                o.object.css({
                    "position": "fixed",
                    "top": o.settings.fixedTop + "px"
                });

                // Resize the nebula object to the dummy width
                if (o.settings.autoWidth) o.object.css("width", o.dummy.width() + "px");

                // Slide in the nebula object if it was previously hidden because it passed the endline
                if (o.state == State.ENDLINEPASSED) {
                    var height = o.object.outerHeight(true);

                    // Reposition the object
                    o.object.css("top", (-height - 5) + "px");

                    // Slide in the object
                    o.object.animate({ "top": "+=" + (height + o.settings.fixedTop + 5) + "px" }, o.settings.slideOutDuration);
                }

                // Update the state
                o.state = State.FIXEDTOP;
            }
            else {
                if (endLinePassed) {
                    // Just return if the nebula object is already positioned right.
                    if (o.state == State.ENDLINEPASSED) return;

                    // Slide out the nebula object if it is currently displayed at the fixed top
                    if (o.state == State.FIXEDTOP) {
                        o.object.stop().animate({ "top": "-=" + (o.object.outerHeight(true) + o.settings.fixedTop + 5) + "px" }, o.settings.slideOutDuration);
                    }

                    // Update the state
                    o.state = State.ENDLINEPASSED;
                }
                else {
                    // Check if already detached from the top
                    if (o.state == State.NORMAL) return;

                    // Stop any animation
                    o.object.stop();

                    // Remove the active class again
                    o.object.removeClass(o.settings.activeClass);

                    // Hide the dummy
                    o.dummy.hide();

                    // Reset the previous values
                    o.object.css({
                        "position": o.settings.previousPosition,
                        "top": o.settings.previousTop
                    });

                    // Reset the width to auto if the auto width option is set
                    if (o.settings.autoWidth) o.object.css("width", "auto");

                    // Update the state
                    o.state = State.NORMAL;
                }
            }
        });
    }, defaultSettings.throttleDelay);



    /*
     * Public Methods
     */

    // Sets the default options for the modal module
    this.defaultOptions = function(options) {
        defaultSettings = getSettings(options);

        // Return this to allow chaining
        return this;
    };

    this.set = function(obj, options) {
        // Get the settings
        var settings = getSettings(options);

        // Get the object
        obj = Kepler.utils.getJQueryObj(obj);
        if (!obj) {
            console.log("error: invalid kepler nebula object!");
            return;
        }

        // Attach the remove event, which is called when the element is removed from the DOM structure.
        // This event is used to cleanup and free unused data.
        obj.off('kepler.element.remove', removeHandler);
        obj.on('kepler.element.remove', removeHandler);

        // Create the dummy element
        var dummy = $('<div class="kepler nebula-dummy"></div>');

        // Save the current position and top
        settings.previousPosition = obj.css("position");
        settings.previousTop = obj.css("top");

        // Create a new unique index
        uniqueIndex++;
        if (uniqueIndex > 1000000) uniqueIndex = 0;

        var nebulaObj = {
            id: uniqueIndex,
            object: obj,
            dummy: dummy,
            endLine: false,
            state: -1,
            settings: settings
        };

        // Get the nebula object id
        var id = "#" + obj.attr("id");
        if (id) {
            // Get all end lines
            var endLines = $('[data-kepler-nebula-end]');

            // Iterate through all and get the ids. Check if the id matches...
            $.each(endLines, function(i, endLine) {
                var ids = $(endLine).data("kepler-nebula-end").toString().split(' ');

                // Check if the current id exists in the array
                if (jQuery.inArray(id, ids) >= 0) {
                    // Set the endline
                    nebulaObj.endLine = $(endLine);

                    // Break the each loop
                    return false;
                }
            });
        }

        // Add the nebula object to the array
        nebulaObjects.push(nebulaObj);

        // Save the unique id
        obj.data("kepler-nebula-id", nebulaObj.id);

        // Prepend the dummy to the DOM structure
        obj.before(dummy);
    };



    /*
     * Kepler and window events
     */

    // Scan the DOM structure for data elements
    Kepler.onInit(function() {
        // Save the previous
        var previousNebulaObjects = nebulaObjects;

        // Clear the array
        nebulaObjects = [];

        var options;
        var found;

        // Find all objects with the data modal attribute
        $.each($("[data-kepler-nebula]"), function(i, obj) {
            // Get the data otions if set
            options = Kepler.utils.dataOptions($(obj), "kepler-nebula-options");

            // Set the found flag
            found = false;

            // Iterate over all the previous nebula objects and try to find the current
            $.each(previousNebulaObjects, function(i, o) {
                if (!o.object.is(obj)) return;

                // Add the previous nebula object to the new list
                nebulaObjects.push(o);

                // Remove the current nebula object from the previous list
                previousNebulaObjects.splice(i, 1);

                // Set the flag
                found = true;

                // Break
                return false;
            });

            // Create the nebula object if not found
            if (!found) {
                Kepler.nebula.set(obj, options);
            }
        });

        // Iterate over all previous nebula objects and detach the remove event
        $.each(previousNebulaObjects, function(i, o) {
            o.object.off('kepler.element.remove', removeHandler);
        });

        // Unbind the scroll and resize events first
        var w = $(window);
        w.off('scroll', updateNebulaPositions);
        w.off('resize', updateNebulaPositions);

        // Only bind the events and call the function if the array is not empty
        if (nebulaObjects.length > 0) {
            // Bind the scroll and resize events
            w.on('scroll', updateNebulaPositions);
            w.on('resize', updateNebulaPositions);

            // Update the nebula object positions
            updateNebulaPositions();
        }
    });
};

/*
 * Kepler Module
 */
Kepler.module.odin = new function() {
    /*
     * Private Variables
     */

    // Default settings
    var defaultSettings = {
        html: false,            // Insert HTML into the error box. If false, jQuery's text method will be used to insert content into the DOM.
                                // Use text if you're worried about XSS attacks.
    };

    // Default patterns
    var patterns = {
        // String containing only whitespaces is also considered as emtpy
        notEmptyNoWhitespace: /([^\s])/,

        alpha: /^[a-zA-Z]+$/,
        alpha_numeric : /^[a-zA-Z0-9]+$/,
        integer: /^[-+]?\d+$/,
        number: /^[-+]?\d*(?:[\.\,]\d+)?$/,

        // At least one number, one lowercase and one uppercase letter and at least 8 characters
        password: /(?=.*\d)(?=.*[a-z])(?=.*[A-Z]).{8,}/,

        // amex, visa, diners
        card : /^(?:4[0-9]{12}(?:[0-9]{3})?|5[1-5][0-9]{14}|6(?:011|5[0-9][0-9])[0-9]{12}|3[47][0-9]{13}|3(?:0[0-5]|[68][0-9])[0-9]{11}|(?:2131|1800|35\d{3})\d{11})$/,
        cvv : /^([0-9]){3,4}$/,

        // http://www.whatwg.org/specs/web-apps/current-work/multipage/states-of-the-type-attribute.html#valid-e-mail-address
        email: /^[a-zA-Z0-9.!#$%&'*+\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$/,

        // http://desertbit.com
        url: /^(https?|ftp|file|ssh):\/\/(((([a-zA-Z]|\d|-|\.|_|~|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])|(%[\da-f]{2})|[!\$&'\(\)\*\+,;=]|:)*@)?(((\d|[1-9]\d|1\d\d|2[0-4]\d|25[0-5])\.(\d|[1-9]\d|1\d\d|2[0-4]\d|25[0-5])\.(\d|[1-9]\d|1\d\d|2[0-4]\d|25[0-5])\.(\d|[1-9]\d|1\d\d|2[0-4]\d|25[0-5]))|((([a-zA-Z]|\d|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])|(([a-zA-Z]|\d|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])([a-zA-Z]|\d|-|\.|_|~|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])*([a-zA-Z]|\d|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])))\.)+(([a-zA-Z]|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])|(([a-zA-Z]|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])([a-zA-Z]|\d|-|\.|_|~|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])*([a-zA-Z]|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])))\.?)(:\d*)?)(\/((([a-zA-Z]|\d|-|\.|_|~|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])|(%[\da-f]{2})|[!\$&'\(\)\*\+,;=]|:|@)+(\/(([a-zA-Z]|\d|-|\.|_|~|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])|(%[\da-f]{2})|[!\$&'\(\)\*\+,;=]|:|@)*)*)?)?(\?((([a-zA-Z]|\d|-|\.|_|~|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])|(%[\da-f]{2})|[!\$&'\(\)\*\+,;=]|:|@)|[\uE000-\uF8FF]|\/|\?)*)?(\#((([a-zA-Z]|\d|-|\.|_|~|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])|(%[\da-f]{2})|[!\$&'\(\)\*\+,;=]|:|@)|\/|\?)*)?$/,

        // desertbit.com
        domain: /^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,6}$/,

        //YYYY-MM-DDThh:mm:ssTZD
        datetime: /^([0-2][0-9]{3})\-([0-1][0-9])\-([0-3][0-9])T([0-5][0-9])\:([0-5][0-9])\:([0-5][0-9])(Z|([\-\+]([0-1][0-9])\:00))$/,
        // YYYY-MM-DD
        date: /(?:19|20)[0-9]{2}-(?:(?:0[1-9]|1[0-2])-(?:0[1-9]|1[0-9]|2[0-9])|(?:(?!02)(?:0[1-9]|1[0-2])-(?:30))|(?:(?:0[13578]|1[02])-31))$/,
        // HH:MM:SS
        time : /^([01]\d|2[0-3]):?([0-5]\d)$/,
        dateISO: /^\d{4}[\/\-]\d{1,2}[\/\-]\d{1,2}$/,
        // MM/DD/YYYY
        month_day_year : /^(0[1-9]|1[012])[- \/.](0[1-9]|[12][0-9]|3[01])[- \/.]\d{4}$/,

        // #FFF or #FFFFFF
        color: /^#?([a-fA-F0-9]{6}|[a-fA-F0-9]{3})$/
    };

    // Custom validators
    var validators = {
        notEmpty: function(el, value, required, validationObj) {
            return value !== "";
        },
        equalTo: function(el, value, required, validationObj) {
            // Check if the object has the equalto attribute
            var equalTo = el.attr("equalto");
            if (equalTo) {
                // Get the equalTo object
                equalTo = $("#" + equalTo);

                // Check if the object exists
                if (equalTo.length <= 0) {
                    console.log("error: odin equalto object is invalid");
                    return false;
                }

                // Perform the match
                return (value === equalTo.val().toString());
            } else {
                console.log("error: no odin equalto attribute set");
                return false;
            }
        }
    }



    /*
     * Private Methods
     */

    var getSettings = function(options) {
        // Combine the settings, but preserve the objects. Create a deep copy...
        var settings = $.extend({}, defaultSettings, options);

        return settings;
    };

    var getValidationObj = function(validation) {
        // Check if validation is already an object
        if (validation instanceof jQuery) {
            return validation;
        }

        // Get the validation object
        var validationObj = $('[data-kepler-odin="' + String(validation) + '"]');
        if (validationObj.length <= 0) {
            console.log("Error: failed to get validation object by id: " + String(validation));
            return false;
        }

        return validationObj;
    };

    var validateHandler = function() {
        // Get the odin reference
        var odin = $(this).data("kepler-odin-ref");
        if (!odin) {
            console.log("error: no odin reference!");
            return;
        }

        // Validate
        Kepler.odin.validate(odin);
    };



    /*
     * Public Methods
     */

    // Sets the default options for the popover module
    this.defaultOptions = function(options) {
        defaultSettings = getSettings(options);

        // Return this to allow chaining
        return this;
    };

    // Extend the default patterns
    this.extendPatterns = function(additionalPatterns) {
        // Extend the patterns
        patterns = $.extend(patterns, additionalPatterns);

        // Return this to allow chaining
        return this;
    };

    // Extend the default custom validators
    this.extendValidators = function(additionalValidators) {
        // Extend the patterns
        validators = $.extend(validators, additionalValidators);

        // Return this to allow chaining
        return this;
    };

    this.validate = function(validation, successCallback, errorCallback) {
        // Get the validation object
        var validationObj = getValidationObj(validation);
        if (!validationObj) {
            // Return this to allow chaining
            return this;
        }

        // Get settings with default values
        var settings = getSettings(Kepler.utils.dataOptions(validationObj, "kepler-odin-options"));

        var state = true;
        var str;

        // Iterate over all objects with the match or required attribute
        $.each(validationObj.find("[match], [required]"), function(i, obj) {
            // Get the jQuery object
            obj = $(obj);

            // Get the error box if present
            var errorBox = obj.data("kepler-odin-err");

            // Get the string value
            str = obj.val().toString();

            // Get the match attribute
            var match = obj.attr("match");
            if (!match) {
                // Fallback to default match validator
                match = validators.notEmpty;
            }
            else {
                // Check if a preset pattern is required
                var data = patterns[match];
                if (data) {
                    // Set the preset regexp
                    match = data;
                }
                // Check if a preset validator is required
                else {
                    data = validators[match];
                    if (data) {
                        match = data;
                    }
                    // Nothing found. Assume this is a custom regexp string...
                    else {
                        // Tranform to a valid regexp
                        match = new RegExp(match);
                    }
                }
            }

            // Set the equalto validator if the equalto attribute is set
            if (obj.attr("equalto")) {
                match = validators.equalTo;
            }

            // Check whenever the match is required
            var required = obj.attr("required");
            if (required) required = true; // Be sure this is a boolean value

            // Boolean value if the match variable is a function
            var matchIsFunction = typeof(match) == "function";

            // Perform the actual match
            if (!required && str === ""
                || (matchIsFunction && match(obj, str, required, validationObj))
                || (!matchIsFunction && str.match(match)))
            {
                // Remove the failed class
                obj.removeClass("kepler-odin-failed kepler-odin-failed-border");

                // Add the success class
                obj.addClass("kepler-odin-success");

                // Remove the error box if defined
                if (errorBox) {
                    errorBox.remove();

                    // Reset the data reference
                    obj.data("kepler-odin-err", false);
                }

                // Continue the loop
                return;
            }

            // Remove the success class
            obj.removeClass("kepler-odin-success kepler-odin-failed-border");

            // Add the failed class
            obj.addClass("kepler-odin-failed");

            // Get the error text
            data = obj.attr("error");

            // Create the error box if not present
            if (data && !errorBox) {
                // Create a new error box
                errorBox = $('<small class="kepler-odin-error"></small>');

                // Insert HTML to the error box. Use text if you're worried about XSS attacks.
                if (settings.html) {
                    errorBox.html(data.toString());
                }
                else {
                    errorBox.text(data.toString());
                }

                // Save the error box reference
                obj.data("kepler-odin-err", errorBox);

                // Insert the error box
                errorBox.insertAfter(obj);
            } else if (!data) {
                obj.addClass("kepler-odin-failed-border");
            }

            // Set the state variable to false
            state = false;
        });

        if (state) {
            // Call the callback if defined
            if (successCallback) successCallback();

            // Trigger the event
            validationObj.trigger("kepler.odin.valid");
        }
        else  {
            // Call the error callback if defined
            if (errorCallback) errorCallback();

            // Trigger the event
            validationObj.trigger("kepler.odin.invalid");
        }

        // Return this to allow chaining
        return this;
    };

    // This event is fired when the validation was validated successful
	this.valid = function(validation, callback) {
        // Get the validation object
        var validationObj = getValidationObj(validation);
        if (!validationObj) {
            // Return this to allow chaining
            return this;
        }

        // Attach the event
		validationObj.on("kepler.odin.valid", callback);

        // Return this to allow chaining
        return this;
	};

    // This event is fired when the validation was validated with an error
	this.invalid = function(validation, callback) {
        // Get the validation object
        var validationObj = getValidationObj(validation);
        if (!validationObj) {
            // Return this to allow chaining
            return this;
        }

        // Attach the event
		validationObj.on("kepler.odin.invalid", callback);

        // Return this to allow chaining
        return this;
	};



    /*
     * Kepler events
     */

    // Scan the DOM structure for data elements
    Kepler.onInit(function() {
        $.each($("[data-kepler-odin]"), function(i, odin) {
            // Get jQuery object
            odin = $(odin);

            // Iterate over all objects with the validate attribute
            $.each(odin.find("[validate]"), function(i, obj) {
                // Get the jQuery object
                obj = $(obj);

                // Save the odin reference
                obj.data("kepler-odin-ref", odin);

                // Attach the event
                obj.off('click', validateHandler);
                obj.on('click', validateHandler);
            });
        });
    });
};

/*
 * Kepler Module
 */
Kepler.module.pagescroll = new function() {
    /*
     * Private Variables
     */

    // Default settings
    var defaultSettings = {
        scrollDuration : 600,
        activeClass : 'active'
    };



    /*
     * Private Methods
     */

    var getSettings = function(options) {
        // Combine the settings, but preserve the objects. Create a deep copy...
        return $.extend({}, defaultSettings, options);
    };

    // Scroll the page.
    var scrollPage = function(el, deltaY) {
        // Get the jQuery element
        el =  Kepler.utils.getJQueryObj(el);

        // Get the pagescroll parent
        var parent = el.data("kepler-pagescroll-parent");

        // Return if currently animating...
        if (parent.hasClass("isScrolling")) {
            // Return true to prevent any mouse scrolling during the animation
            return true;
        }

        // Define the scollPage variable
        var scrollPage = false;

        // The deltaY is used to define the scroll direction.
        if(deltaY > 0) {
            // Get the previous page and save the reference
            scrollPage = el.prev();
        }
        else if (deltaY < 0) {
            // Get the next page and save the reference
            scrollPage = el.next();
        }

        // If the previous element does not exists, then
        // return false to not block the scroll request.
        if (!scrollPage || scrollPage.length <= 0) {
            return false;
        }

        // Get settings from slider
        var settings = getSettings(Kepler.utils.dataOptions(parent, "kepler-pagescroll-options"));

        // Get the scroll range
        var scroll = scrollPage.offset().top;

        // TODO: Remove me if bug is fixed
        console.log("size: " + scrollPage.outerHeight(true));
        console.log("scroll: " + scroll);

        // Set the animating class
        parent.addClass('isScrolling');

        // animate scroll
        $('html, body').animate({
            scrollTop: scroll
        }, settings.scrollDuration, function() {
            setTimeout(function() {
                // Remove the active scrolling class again
                parent.removeClass('isScrolling');

                // Remove the previous active class
                parent.children().removeClass(settings.activeClass);

                // Set the active class
                scrollPage.addClass(settings.activeClass);
            }, 350);
        });

        // Return true to prevent any mouse scrolling during the animation
        return true;
    };

    // Event that gets called on mouse wheel events
    var mouseWheelHandler = function(event) {
        // Scroll the page
        if (scrollPage(this, event.deltaY)) {
            // Prevent the default behavior if the scrollPage method returns true
            event.preventDefault();
            return false;
        }
    };

    var linkClickHandler = function() {
        // Get the parent
        var parent = $(this).data("kepler-pagescroll-parent");

        // Get settings from slider
        var settings = getSettings(Kepler.utils.dataOptions(parent, "kepler-pagescroll-options"));

        // Get the current active page
        var curPage = parent.children('.' + settings.activeClass);

        // Set the first page if no active class is set
        if (curPage.length <= 0) {
            var c = parent.children();

            if (c.length <= 0) {
                return;
            }

            curPage = c[0];
        }

        // Obtain the scroll direction
        var direction = -1;
        if ($(this).is("[data-kepler-pagescroll-prev]")) {
            direction = 1;
        }

        // Scroll to the next page
        scrollPage(curPage, direction);
    };



    /*
     * Public Methods
     */

    // Sets the default options for the modal module
    this.defaultOptions = function(options) {
        defaultSettings = getSettings(options);

        // Return this to allow chaining
        return this;
    };



    /*
     * Kepler and window events
     */

    // Scan the DOM structure for data elements
    Kepler.onInit(function() {
        $.each($("[data-kepler-pagescroll]"), function(x, pagescroll) {
            // Get the jQuery object
            pagescroll = $(pagescroll);

            // Go through all direct children of the pagescroll wrapper element
            $.each(pagescroll.children(), function(i, el) {
                // Get the jQuery element
                el = $(el);

                // Save the parent reference to the child object
                el.data("kepler-pagescroll-parent", pagescroll);

                // Attach the clicik events for the next and previous scroll links
                $.each(el.find("[data-kepler-pagescroll-next],[data-kepler-pagescroll-prev]"), function(y, linkEl) {
                    // Get the jQuery object
                    linkEl = $(linkEl);

                    // Save the pagescroll wrapper reference
                    linkEl.data("kepler-pagescroll-parent", pagescroll);

                    // Attach the click event
                    linkEl.click(linkClickHandler);
                });

                // Add the mousewheel event
                el.off('mousewheel', mouseWheelHandler);
                el.on('mousewheel', mouseWheelHandler);
            });
        });
    });
};

/*
 * Kepler Module
 */
Kepler.module.popover = new function() {
    /*
     * Private Variables
     */

    // Default settings
    var defaultSettings = {
        animation: true,                    // Apply a simple fade in and out animation
        animationDuration: 150,             // The fade animation duration
        activeZIndex: 100,                  // The active z-indes value
        trigger: 'click',                   // Valid trigger values are 'click', 'hover' or 'manual'
        offset: 10,                         // The popover offset
        placement: 'bottom',                // The popover placement: 'top', 'bottom', 'left', 'right', 'auto' or 'auto right'... (suggested placment for auto)
        mobileBreakPoint: 640,              // The mobile break point. The popover is positioned auto if reached.
        arrow: true,                        // Whenever an arrow should be shown
        closeOnFocusLost: true,             // Whenever the popover should be closed if it loses focus
        closeOnFocusLostExceptions: false   // A string with selectors of elements which should be skipped.
                                            // The popover isn't closed if these exceptions get a focus.
                                            // Seperate multiple exception selectors with a space.
    };

    var globalClickEventBound = false;
    var globalResizeEventBound = false;


    /*
     * Private Methods
     */

    var getSettings = function(options) {
        // Combine the settings, but preserve the objects. Create a deep copy...
        var settings = $.extend({}, defaultSettings, options);

        return settings;
    };

    var closeButtonTriggerHandler = function() {
        // Get the popover object
        var popover = $(this).data("kepler-popover-object");
        if (!popover) {
            console.log("Error: popover object is invalid!");
            return;
        }

        // Close the popover
        Kepler.popover.close(popover);
    };

    var mouseLeaveTriggerHandler = function() {
        // Get jQuery object
        var el = Kepler.utils.getJQueryObj(this);

        // Get the popover object
        var popover = Kepler.utils.getJQueryObj(el.data("kepler-popover"));
        if (!popover) {
            console.log("error: invalid popover object: data-kepler-popover selector is invalid!");
            return;
        }

        // Close the popover
        Kepler.popover.close(popover);
    };

    var closeTriggerHandler = function (e) {
        // Get all popovers which should be closed on focus lost
        var popovers = $(".kepler-popover-close-focus-lost");

        // Iterate through all popovers
        $.each(popovers, function(i, popover) {
            popover = $(popover);

            // Get the popover data
            var data = popover.data("kepler-popover-active-data");
            if (!data) {
                console.log("Error: failed to get active popover data!");
                return;
            }

            // Check if event was triggered by the popover itself, its siblings, the trigger target or the trigger target siblings.
            if (popover.is(e.target) || popover.has(e.target).length > 0
                || data.target.is(e.target) || data.target.has(e.target).length > 0) {
                return;
            }

            // Chech if exception elements are defined
            if (data.settings.closeOnFocusLostExceptions) {
                // Get exception array
                var closeExceptions = data.settings.closeOnFocusLostExceptions.toString().split(' ');

                // Flag
                var foundException = false;

                // Check if the trigger element is part of the exceptions
                $.each(closeExceptions, function(i, element) {
                    if ($(element).is(e.target)) {
                        foundException = true;
                        return false;
                    }
                });

                // Don't close the popover if an exception matches
                if (foundException) return;
            }

            // Close this popover
            Kepler.popover.close(popover);
        });
    };

    var openTriggerHandler = function (e) {
        // Get jQuery object
        var el = Kepler.utils.getJQueryObj(this);

        // Trigger the event and stop the execution if false is returned
        if (el.triggerHandler("kepler.popover.openTriggered") === false) {
            return;
        }

        // Get the popover object
        var popover = Kepler.utils.getJQueryObj(el.data("kepler-popover"));
        if (!popover) {
            console.log("error: invalid popover object: data-kepler-popover selector is invalid!");
            return;
        }

        // Get the data options if set.
        // Options set on the trigger element override options set on the popover element.
        var options = $.extend(Kepler.utils.dataOptions(popover, "kepler-popover-options"),
                               Kepler.utils.dataOptions(el, "kepler-popover-options"));

        // Get settings with default values
        var settings = getSettings(options);

        // Get the popover data
        var data = popover.data("kepler-popover-active-data");

        if (settings.trigger === "click" && data && el.is(data.target)) {
            // Toggle the popover if the trigger is a click event
            // and if the previous target and the current are the same.
            Kepler.popover.toggle(popover, el, settings);
        }
        else {
             // Open the popover
            Kepler.popover.open(popover, el, settings);
        }
    };

    var _positionPopover = function(popoverObj, position) {
        // Position the popover manual.
        // Set unused positions to auto.
        popoverObj.css({
            "left" : ("left" in position) ? position.left + "px" : "auto",
            "right" : ("right" in position) ? position.right + "px" : "auto",
            "bottom" : ("bottom" in position) ? position.bottom + "px" : "auto",
            "top" : ("top" in position) ? position.top + "px" : "auto"
        });
    };

    var _positionPopoverToTarget = function(popoverObj, target, placement, offset){
        // Get the relative target position
        var targetPos = Kepler.utils.getRelativePosition(target);

        // Obtain final position for the popover
        var pos = {};
        switch(placement) {
            case "bottom":
                pos.left    = targetPos.left;
                pos.top     = targetPos.top + offset + targetPos.height;
                break;

            case "top":
                pos.left    = targetPos.left;
                pos.bottom  = targetPos.bottom + offset + targetPos.height;
                break;

            case "right":
                pos.left    = targetPos.left + offset + targetPos.width;
                pos.top     = targetPos.top;
                break;

            case "left":
                pos.right   = targetPos.right + offset + targetPos.width;
                pos.top     = targetPos.top;
                break;

            default:
                console.log("warning: invalid popover placement value: '" + String(placement) + "'");
                console.log("falling back to default top popover placement...");

                pos.left = targetPos.left;
                pos.bottom   = targetPos.bottom + offset + targetPos.height;
        }

        // Set the popover position
        _positionPopover(popoverObj, pos);
    };

    var _autoPositionPopover = function(popoverObj, target, suggestedPlacement, offset) {
        // Create an array with all possible placements
        var placements = [suggestedPlacement, "bottom", "top", "right", "left"];
        var triedPlacements = [];
        var bestPlacement = false;

        // Show the popover with an transparent opacity to get the dimensions
        var isHidden = !popoverObj.is(":visible");
        if (isHidden) {
            // Set opacity to invisible
            popoverObj.css("opacity", "0");

            // Show the popover to obtain the dimensions
            popoverObj.show();
        }

        // Try all possible placements
        $.each(placements, function(i, p) {
            // Check if the placement was already tested
            if($.inArray(p, triedPlacements) !== -1) return;

            // Add the current placement to the array
            triedPlacements.push(p);

            // Position the popover to the target
            _positionPopoverToTarget(popoverObj, target, p, offset);

            // Check if the popover fits on the screen
            if (Kepler.utils.fitsOnScreen(popoverObj)) {
                // Save the current placement
                bestPlacement = p;

                // Break the loop
                return false;
            }
        });

        // If no placement fits on the screen, just set the suggested placement
        if (!bestPlacement) {
            bestPlacement = suggestedPlacement;

            // Position the popover to the target
            _positionPopoverToTarget(popoverObj, target, suggestedPlacement, offset);
        }

        if (isHidden) {
            // Hide the popover again
            popoverObj.hide();

            // Set opacity back to normal
            popoverObj.css("opacity", "1");
        }

        return bestPlacement;
    };

    // Position the popover object.
    // The target value can either be a target object or a position object holding the new popover position.
    // The placement parameter is optional and is used only, if the target is not a position object.
    var positionPopover = function(popoverObj, target, offset, placement) {
        // Check if the target is a position object
        if(target && target.left && target.top) {
            // Position the popover manual.
            _positionPopover(popoverObj, target);
            return false;
        }

        // Set the placement if not defined
        if (!placement) placement = "bottom";

        var placements = placement.split(" ");
        if (placements[0] === "auto") {
            // Set the suggested placement to bottom if it isn't set
            if (placements.length == 1) {
                return _autoPositionPopover(popoverObj, target, "bottom", offset);
            }
            else {
                return _autoPositionPopover(popoverObj, target, placements[1], offset);
            }
        }
        else {
            // Position the popover to the target with the given placement
            _positionPopoverToTarget(popoverObj, target, placement, offset);
        }

        return placement;
    };

    var setPopoverArrow = function(popoverObj, direction) {
        // Remove all previous arrows
        popoverObj.removeClass("arrow-left arrow-right arrow-top arrow-bottom");

        // Check if the direction is valid
        if (direction === "left" || direction === "right" || direction === "top" || direction === "bottom") {
            popoverObj.addClass("arrow-" + direction);
        } else {
            console.log("error: invalid popover arrow direction: " + String(direction));
        }
    };

    var resizeHandler = Kepler.utils.throttle(function() {
        // Get all active popovers
        var activePopovers = $(".kepler-popover-active");

        // Get the window width
        var windowWidth = $(window).width();

        // Iterate through all active popovers
        $.each(activePopovers, function(i, popover) {
            popover = $(popover);

            // Get the popover data
            var data = popover.data("kepler-popover-active-data");
            if (!data) {
                console.log("Error: failed to get active popover data!");
                return;
            }

            // Get the placement
            var placement = data.settings.placement;

            // Set the placement to auto if the mobile breakpoint is reached
            if (windowWidth <= data.settings.mobileBreakPoint) {
                placement = "auto";
            }

            // Position the popover
            placement = positionPopover(popover, data.target, data.settings.offset, placement);

            // Set the popover arrow if enabled
            if (data.settings.arrow && placement) {
                setPopoverArrow(popover, placement);
            }
        });
    }, 50);



    /*
     * Public Methods
     */

    // Sets the default options for the popover module
    this.defaultOptions = function(options) {
        defaultSettings = getSettings(options);

        // Return this to allow chaining
        return this;
    };

    // Initialize a target object and bind the events to show a popover on the trigger event
    this.init = function(target, options) {
        // Get the settings
        var settings = getSettings(options);

        // Get a jQuery target object
        target = Kepler.utils.getJQueryObj(target);
        if (!target) {
            console.log("Error: invalid target object!");

            // Return this to allow chaining
            return this;
        }

        // First detach all possible previous event handlers
        target.off('click', openTriggerHandler);
        target.off('mousehover', openTriggerHandler);

        // Attach the new event handler to show the popover
        if (settings.trigger === "click") {
            target.on('click', openTriggerHandler);
        }
        else if (settings.trigger === "hover") {
            target.on('mouseover', openTriggerHandler);
        }
        else if (settings.trigger !== "manual") {
            console.log("warning: invalid trigger setting: '" + String(settings.trigger) + "'");
        }

        // Return this to allow chaining
        return this;
    };

    // Show the popover.
    // The target value can either be a target object or a position object holding the new popover position.
    // A position object requires a top and left integer and can contain optional a right and bottom integer.
	this.open = function(popoverObj, target, options) {
        // Get the settings
        var settings = getSettings(options);

        // Get a jQuery object
        popoverObj = Kepler.utils.getJQueryObj(popoverObj);
        if (!popoverObj) {
            console.log("Error: invalid popover object!");

            // Return this to allow chaining
            return this;
        }

        // Get the popover data if already present
        var data = popoverObj.data("kepler-popover-active-data");

        // Check if the popover is already active and the previous target is the same.
        // So just return, because nothing is to do...
        if (popoverObj.hasClass("kepler-popover-active") && data && target.is(data.target)) {
            // Return this to allow chaining
            return this;
        }

        // Trigger the requesting event and stop the execution if false is returned
        if (popoverObj.triggerHandler("kepler.popover.openRequested") === false) {
            // Return this to allow chaining
            return this;
        }

        // Set the CSS position to absolute
        popoverObj.css({"position":"absolute"});

        // Move the popover just below the target object.
        // This way, the popover is positioned right, also if the
        // target object is nested in multiple relative containers...
        popoverObj.insertAfter(target);

        // Get the placement
        var placement = settings.placement;

        // Set the placement to auto if the mobile breakpoint is reached
        if ($(window).width() <= settings.mobileBreakPoint) {
            placement = "auto";
        }

        // Position the popover
        placement = positionPopover(popoverObj, target, settings.offset, placement);

        // Set the popover arrow if enabled
        if (settings.arrow && placement) {
            setPopoverArrow(popoverObj, placement);
        }

        // Add the close click handler if an close button exists in the popover
        var closeButton = popoverObj.find(".close");
        if (closeButton && closeButton.length > 0) {
            // Set the data attribute
            closeButton.data("kepler-popover-object", popoverObj);

            // Detach the event first and attach the click handler again.
            // This way no dublicate events are bound if init is called mutliple times.
            closeButton.off('click', closeButtonTriggerHandler);
            closeButton.one('click', closeButtonTriggerHandler);
        }

        // If data is set, then remove the previous target active class
        if (data) {
            data.target.removeClass("active");
        }
        // If no data is set, create one and add it to the popover object
        else {
            data = {};
            popoverObj.data("kepler-popover-active-data", data);
        }

        // Set the new data values
        data.target = target;
        data.settings = settings;

        // Set the z-index
        popoverObj.css("z-index", settings.activeZIndex);

        if (settings.closeOnFocusLost) {
            // If the trigger is the mouse over event, attach an mouseleave event instead of the default click event
            if (settings.trigger === "hover") {
                // Attach the mouseleave event
                target.off('mouseleave');
                target.one('mouseleave', mouseLeaveTriggerHandler);
            }
            else {
                // Add the class
                popoverObj.addClass("kepler-popover-close-focus-lost");

                // Bind the global events if not already
                if (!globalClickEventBound) {
                    $(document).off('click', closeTriggerHandler);
                    $(document).on('click', closeTriggerHandler);

                    // Set the flag
                    globalClickEventBound = true;
                }
            }
        }
        else {
            // Add the class
            popoverObj.removeClass("kepler-popover-close-focus-lost");
        }

        // Bind the resize event if not already
        if (!globalResizeEventBound) {
            $(window).off('resize', resizeHandler);
            $(window).on('resize', resizeHandler);

            // Update the flag
            globalResizeEventBound = true;
        }

        // Add the active class to the popover and the target
        popoverObj.addClass("kepler-popover-active");
        target.addClass("active");

        if (settings.animation) {
            // Fade in the popover
            popoverObj.stop().fadeIn(settings.animationDuration, function() {
                // Finally trigger the event
                popoverObj.trigger("kepler.popover.opened");
            });
        }
        else {
            // Remove any transparent settings
            popoverObj.css("opacity", "1");

            // Show the popover
            popoverObj.stop().show();

            // Finally trigger the event
            popoverObj.trigger("kepler.popover.opened");
        }

        // Return this to allow chaining
        return this;
    };

    // Toggle the popover.
    // The target and options values are optional, if the popover open method was called at least once!
    this.toggle = function(popoverObj, target, options) {
        // Get a jQuery object
        popoverObj = Kepler.utils.getJQueryObj(popoverObj);
        if (!popoverObj) {
            console.log("Error: invalid popover object!");

            // Return this to allow chaining
            return this;
        }

        // Check if no target and options are passed. So try to retrieve the previous values.
        if (!target || !options) {
            // Get the popover data
            var data = popoverObj.data("kepler-popover-active-data");

            // Check if defined
            if (!data) {
                console.log("error: popover toggle called without target and options values, but no previous target and options values set!");

                // Return this to allow chaining
                return this;
            }

            // Set the previous values
            target = data.target;
            options = data.settings;
        }

        // Check if the popover is active
        if (popoverObj.hasClass("kepler-popover-active")) {
            // Close the popover
            Kepler.popover.close(popoverObj);
        }
        else {
            // Open the popover
            Kepler.popover.open(popoverObj, target, options);
        }

         // Return this to allow chaining
        return this;
    };

    // Close the popover
	this.close = function(popoverObj) {
        // Get a jQuery object
        popoverObj = Kepler.utils.getJQueryObj(popoverObj);
        if (!popoverObj) {
            console.log("Error: invalid popover object!");

            // Return this to allow chaining
            return this;
        }

        // Check if the popover is not active. So just return.
        if (!popoverObj.hasClass("kepler-popover-active")) {
            // Return this to allow chaining
            return this;
        }

        // Trigger the requesting event and stop the execution if false is returned
        if (popoverObj.triggerHandler("kepler.popover.closeRequested") === false) {
            // Return this to allow chaining
            return this;
        }

        // Get the popover data
        var data = popoverObj.data("kepler-popover-active-data");

        // Check if defined
        if (!data) {
            console.log("error: failed to get popover active data!");
            // Return this to allow chaining
            return this;
        }

        // Remove the classes again
        popoverObj.removeClass("kepler-popover-active kepler-popover-close-focus-lost");
        data.target.removeClass("active");

        if (data.settings.animation) {
            // Fade out the popover
            popoverObj.stop().fadeOut(data.settings.animationDuration, function() {
                // Hide the popover as soon as the animation finished
                popoverObj.stop().hide();

                // Finally trigger the event
                popoverObj.trigger("kepler.popover.closed");
            });
        }
        else {
            // Hide the popover again
            popoverObj.stop().hide();

            // Finally trigger the event
            popoverObj.trigger("kepler.popover.closed");
        }

        // Check if there are no more popovers which should be closed on focus lost.
        if (globalClickEventBound && $(".kepler-popover-close-focus-lost").length <= 0) {
            // Unbind the global click event to remove overhead
            $(document).off('click', closeTriggerHandler);

            // Update the flag
            globalClickEventBound = false;
        }

        // Unbind the resize event
        if (globalResizeEventBound && $(".kepler-popover-active").length <= 0) {
            $(window).off('resize', resizeHandler);

            // Update the flag
            globalResizeEventBound = false;
        }

        // Return this to allow chaining
        return this;
    };

    // Close all active popovers
    this.closeAll = function() {
        // Get all active popovers
        var activePopovers = $(".kepler-popover-active");

        // Close each active popover
        $.each(activePopovers, function(i, popover) {
            Kepler.popover.close(popover);
        });

        // Return this to allow chaining
        return this;
    };




    /*
     * Public Events
     */

    // This event fires immediately when the open instance method is called.
    // Return false, to prevent the execution...
	this.openRequested = function(popoverObj, callback) {
        // Get a jquery object if it isn't one
        popoverObj = Kepler.utils.getJQueryObj(popoverObj);

        // Attach the event
		popoverObj.on("kepler.popover.openRequested", callback);

        // Return this to allow chaining
        return this;
	};

    // This event is fired immediately when the close instance method has been called.
    // Return false, to prevent the execution...
	this.closeRequested = function(popoverObj, callback) {
        // Get a jquery object if it isn't one
        popoverObj = Kepler.utils.getJQueryObj(popoverObj);

        // Attach the event
		popoverObj.on("kepler.popover.closeRequested", callback);

        // Return this to allow chaining
        return this;
	};

    // This event is fired when the popover has been made visible to the user (will wait for the animation to complete).
	this.opened = function(popoverObj, callback) {
        // Get a jquery object if it isn't one
        popoverObj = Kepler.utils.getJQueryObj(popoverObj);

        // Attach the event
		popoverObj.on("kepler.popover.opened", callback);

        // Return this to allow chaining
        return this;
	};

    // This event is fired when the popover has finished being hidden from the user (will wait for the animation to complete).
    this.closed = function(popoverObj, callback) {
        // Get a jquery object if it isn't one
        popoverObj = Kepler.utils.getJQueryObj(popoverObj);

        // Attach the event
		popoverObj.on("kepler.popover.closed", callback);

        // Return this to allow chaining
        return this;
	};



    /*
     * Kepler init
     */

    // Scan the DOM structure for data elements
    Kepler.onInit(function() {
        var options;

        // Find all objects with the data popover attribute
        $.each($("[data-kepler-popover]"), function(i, obj) {
            // Make obj to a jQuery element
            obj = $(obj);

            // Get settings with default values
            options = Kepler.utils.dataOptions(obj, "kepler-popover-options");

            // Initialize and bind the target trigger event
            Kepler.popover.init(obj, options);
        });
    });
};

/*
 * Kepler Module
 */
Kepler.module.slider = new function() {

    /*
     * Private Variables
     */

    // Default settings
    var defaultSettings = {
        transitionEffect : 'slide',
        transitionMethod : 'css',
        autoplay : true,
        sliderInterval : 2000,
        useYouTubeAPI : false,
        animationMode : 'css',
        activeClass: "activeSlide",
        // Navigation settings
        navigation : true,
        dotsNav : 'create', // Set to false to disable
        prevNav : 'create', // Set to false to disable
        nextNav : 'create', // Set to false to disable
        playNav : 'create', // Set to false to disable
        dotClass : 'kepler slider-dot',
        dotsNavClass: 'kepler slider-nav',
        dotActiveClass: 'active',
        prevNavClass : 'kepler slider-prev',
        nextNavClass : 'kepler slider-next',
        playNavClass : 'kepler slider-play',
        timerBar : false,
        timerBarClass : 'kepler slider-time',
        timerBarInterval : 10
    };

    var globalResizeEventBound = false;


    /*
     * Private Methods
     */


    var getSettings = function(options) {
        // Combine the settings, but preserve the objects. Create a deep copy...
        var settings = $.extend({}, defaultSettings, options);

        return settings;
    }; 

    // Sets the default options for the slider module
    this.defaultOptions = function(options) {
        defaultSettings = getSettings(options);

        // Return this to allow chaining
        return this;
    };

    // --------------
    // VIDEO PLAYBACK
    // --------------

    var parseVideoURL = function(url){
        url.match(/^http:\/\/(?:.*?)\.?(youtube|vimeo)\.com\/((watch\?[^#]*v=(\w+)|(\d+))||(embed\?[^#]*v=(\w+)|(\d+))).+$/);

        return RegExp.$1;
    };

    var startVideoSlide = function(type, videoEl){
        if(type === 'youtube'){
            // TODO do something when video starts
        }
        if(type === 'vimeo'){
            // TODO do something when video starts
        }
    };

    /*
     * Public Methods
     */



    // With this function we do the actual sliding
    // The idea behind the slider is to simply move the wrap element inside the [data-kepler-slider] element (need overflow:hidden to work)
    var slide = function(target, el){
        //TODO parse el to to $(el) and save with utils method
        // Get settings from slider
        var settings = getSettings(Kepler.utils.dataOptions(el, "kepler-slider-options"));
        var targetNum;
        var $wrap = $($(el).find("> .wrap"));
        var slideLength = $wrap.children().length -1;

        // Cheack if target is a command
        if(target === 'next'){
            targetNum = $wrap.find("> ." + settings.activeClass).index() + 1;
        }

        else if(target === 'prev'){
            targetNum = $wrap.find("> ." + settings.activeClass).index() - 1;
        }

        else {
            targetNum = target;
        }
        if(targetNum > slideLength){
            targetNum = 0;
        }
        if(targetNum < 0){
            targetNum = slideLength -1;
        }
        target = $($wrap.find("> [data-kepler-slide]").get(targetNum));

        if(settings.transitionMethod === 'css'){
            //var afterTransition = function(){
                // Timing critical functions here
            //    $wrap.off(afterTransition);
            //};
            //$wrap.on("transitionend webkitTransitionEnd oTransitionEnd MSTransitionEnd", function(){afterTransition();});
            $wrap.css('transform','translateX(-' + target.position().left  + 'px)');
        }
        // Animation based on jQuery animate
        if(settings.transitionMethod === 'java'){
            $wrap.animate({ left: - targetNum*100 + '%'}, settings.animateDuration);
        }
        // If animation is disabled simply jump to next slide without transition
        if(settings.transitionMethod === false){
            $wrap.css("left",- targetNum*100 + '%');
        }

        // Set new active class

        $wrap.find("." + settings.activeClass).removeClass(settings.activeClass);
        target.addClass(settings.activeClass);
        // Get nav element from slider
        var $nav = el.data("kepler-slider-nav");
        // Update active state of navigation
        $nav.children("." + settings.dotActiveClass).removeClass(settings.dotActiveClass);
        $nav.children().eq(targetNum).addClass(settings.dotActiveClass);
    };


    // This function manages the window resize event and corrects slide positions by calling
    // the slide function on each slider
    var resizeHandler = Kepler.utils.throttle(function() {
        // Get all active popovers
        var sliderList = $("[data-kepler-slider]");

        // Get the window width
        var windowWidth = $(window).width();

        // Iterate through all active popovers
        $.each(sliderList, function(i, slider) {
            slider = $(slider);
            // Get settings from slider
            var settings = getSettings(Kepler.utils.dataOptions(slider, "kepler-slider-options"));

            // Get active slide
            var slideIndex = slider.find('.' + settings.activeClass).index();
            // Slide to the slide with active class
            slide(slideIndex, slider);
        });
    }, 50);

    // --------------------
    // NAVIGATION FUNCTIONS
    // --------------------
    // The folowing functions add navigation elements and provide functions

    // createElement : a simple function to create div element
    var createElement = function(elClass){
        var el = '<div></div>';
        el = $(el);

        el.addClass(elClass);

        return el;
    };

    // addDotNav : adds nav dots for defined target
    // inputs    :  - target a numeric index for the linked slide. Index : [data-kepler-slider] > wrap > [data-kepler-slide]
    //              - $el reference to slider
    //              - dotClass class to add to dot element

    var addDotNav = function(target, $el, dotClass){
        var dot = createElement(dotClass);
        $el.data('kepler-slider-nav').append(dot);

        //Bind click event to dot
        dot.click(function(){
            slide(target, $el);
        });
    };

    // timerBarInterval : updates timer bar to represent how much time is left for slide
    var timerBarInterval = function(){
    // TODO
    };

    // play: this function starts auto slide timer set with sliderInterval
    // Input : needs slider element reference as jQuery object
    var play = function($el){
        var settings = getSettings(Kepler.utils.dataOptions($el, "kepler-slider-options"));
        // init timer var
        var timer;

        if(!$el.data("kepler-slider-timer")){
            // If timer bar is active update bar state
            if(settings.timerBar!==false){
                timer = setInterval(function(){timerBarInterval();}, settings.sliderInterval / settings.timerBarInterval);
            }
            // Default Interval function to slide each sliderInterval
            else {
                timer = setInterval(function(){slide('next', $el);}, settings.sliderInterval);
            }
            // Store interval ref to slider div
            $el.data("kepler-slider-timer", timer);
        }
    };

    // play: this function stops auto slide timer and removes ref from slider element
    // Input : needs slider element reference as jQuery object
    var pause = function($el){
        var timer = $el.data("kepler-slider-timer");
        clearInterval(timer);
        $el.data("kepler-slider-timer", false);
    };

    // playToggle: toggles play pause state.
    // Input : needs slider element reference as jQuery object
    var playToggle = function($el){
        if(!$el.data("kepler-slider-timer")){
            play($el);
        }
        else {
            pause($el);
        }
    };

    var destroy = function($el){
        // Pause and remove Interval
        pause($el);

        //TODO finish removal of classes and wraps as well as default navigation elements
    };

    // --------------
    // INIT FUNCTIONS
    // --------------

    // initNavigation: adds navigation and binds events to buttons
    // Components:
    //      - Dots navigation : toggle dotsNav to disable or set. Class settings stored in dotsNavClass
    //      - Previous button : toggle prevNav to disable or set. Class settings stored in prevNavClass
    //      - Next button     : toggle nextNav to disable or set. Class settings stored in nextNavClass
    //      - play button     : toggle playNav to disable or set. Class settings stored in playNavClass
    var initNavigation = function($el){
        // Get settings from slider
        var settings = getSettings(Kepler.utils.dataOptions($el, "kepler-slider-options"));

        var dotsNav;

        if(settings.dotsNav === 'create'){
            dotsNav = createElement(settings.dotsNavClass);
            $el.append(dotsNav);
        }
        else {
            dotsNav = Kepler.utils.getJQueryObj(settings.dotsNav);
            // Check if element exists
            if(!dotsNav){
                console.log("Error: Slider Dots navigation target does not exist.");
            }
        }

        // Save dots nav reference to $el

        $el.data("kepler-slider-nav", dotsNav);
        if(settings.prevNav!==false){

            var prevNav;

            if(settings.prevNav === 'create'){
                prevNav = createElement(settings.prevNavClass);
                $el.append(prevNav);
            }
            else {
                prevNav = Kepler.utils.getJQueryObj(settings.prevNav);
                // Check if element exists
                if(!prevNav){
                    console.log("Error: Slider previous button target does not exist.");
                    return false;
                }
            }

            //Add event listener to prev button
            prevNav.click(function(){
                pause($el);
                slide('prev', $el);
            });
        }
        
        // Add navigation next button if enabled
        if(settings.nextNav!==false){
            var nextNav;

            // Create a new button if no existing is passed to the slider via settings
            if(settings.nextNav === 'create'){
                nextNav = createElement(settings.nextNavClass);
                $el.append(nextNav);
            }
            // If button passed in settings set it as next button
            else {
                nextNav = Kepler.utils.getJQueryObj(settings.nextNav);
                // Check if element exists
                if(!nextNav){
                    console.log("Error: Slider next button target does not exist.");
                }
            }

            //Add event listener to next button
            nextNav.click(function(){
                pause($el);
                slide('next', $el);
            });
        }
        // Add play pause button if set
        if(settings.playNav!==false){
            var playNav;

            if(settings.playNav === 'create'){
                playNav = createElement(settings.playNavClass);
                $el.append(playNav);
            }
            else {
                playNav = Kepler.utils.getJQueryObj(settings.playNav);
                // Check if element exists
                if(!playNav){
                    console.log("Error: Slider play button target does not exist.");
                }
            }

            //Add event listener to next button
            playNav.click(function(){
                playToggle($el);
            });
        }
    };


    var initVideoApiSlide = function(type, videoFrame){
        return new YT.Player(videoFrame);
    };

    // initSlider : this function inits slider
    //      inputs: - $el selected slider element
    //              - settings : settings can be passed here if no settings are passed using settings set in $el
    //                           if no settings set in $el using default settings
    var initSlider = function($el, settings) {

        var slides = $el.find("> [data-kepler-slide]");
        var slideCount = slides.length;

        // Wrap slider content inside a div
        $el.wrapInner('<div class="wrap" />');
        var $wrap = $($el.find(".wrap"));
        // Set wrap width to contain all slides
        $wrap.css({
            "width" : 100 * slideCount + "%",
            "height" : "100%"
        });
        // If no settings passed use settings stored in $el or the defaults
        if(!settings){
            // Get settings with default values
            settings = getSettings(Kepler.utils.dataOptions($el, "kepler-slider-options"));
        }

        // Navigation: Add dots navigation element
        if(settings.navigation){
            initNavigation($el);
        }
        // Flag var to cheack if manually set start slide is defined by activeClass
        var initSlide = false;
        // Traverse data-kepler-slide inside the slider and adjust size as well as init dots
        $.each(slides, function(){
            var $slide = $(this);
            var index = $slide.index();
            // Adjust element size
            $slide.css({
                "width" : 100 / slideCount + "%",
                "height" : "100%"
            });
            // Add dots to dot nav
            if(settings.navigation && $el.data('kepler-slider-nav')){
                addDotNav(index, $el, settings.dotClass);
            }
            // Slide to this element if element has activeClass
            if($slide.hasClass(settings.activeClass)){
                slide(index, $el);
            }
            // If slide is an iframe check for youtube or vimeo src to init video functions
            /*if($slide.prop('tagName') === 'IFRAME'){
                var url = parseVideoURL($slide.prop('src'));
                if(url === 'youtube'){
                    load API if set to true
                    $slide.data('kepler-slide', url);
                    if(settings.useYouTubeAPI){
                         load youtube api if not loaded
                        if(!window['YT']){
                            Kepler.utils.asyncScriptLoad("https://www.youtube.com/iframe_api");
                        }
                        var player = new YT.Player($slide.get(0));
                        $slide.data('kepler-slide', player);
                        player.playVideo();
                    }
                }
            }*/

        });

        // If no slide is sat as active init the first one as active
        if(!initSlide && $wrap.find("> [data-kepler-slide]").length>0){
            slide(0, $el);
        }

        // Set init flag to signal complete init of slider
        $el.data("kepler-slider", true);

        // Start Interval for slider if autoplay is set to true
        if(settings.autoplay){
            play($el);
        }

        // Bind the resize event if not already
        // Only bind if using css transitions since it is not posible to use percent
        // in combination with translate and therefore the slider won't auto resize
        if (!globalResizeEventBound && settings.transitionMethod === 'css') {
            $(window).off('resize', resizeHandler);
            $(window).on('resize', resizeHandler);

            // Update the flag
            globalResizeEventBound = true;
        }
    };

    /*
     * Kepler and window events
     */
 
    // Scan the DOM structure for data elements
    Kepler.onInit(function() {
        $.each($("[data-kepler-slider]"), function(){
           var $el = $(this);
            // Only init slider if not already done
            if($el.data("kepler-slider")!==true){
                initSlider($el);
            }
        });
    });
};

/*
 * Kepler Module
 */
Kepler.module.sputnik = new function() {
    /*
     * Private Variables
     */

    // Default settings
    var defaultSettings = {
        activeClass: "active",
        group: false    // If a group name string is set, then there is maximal only one active sputnik tracker in this group.
    };

    // Linked elements array
    var sputnikSections = [];



    /*
     * Private Methods
     */

    var getSettings = function(options) {
        // Combine the settings, but preserve the objects. Create a deep copy...
        var settings = $.extend({}, defaultSettings, options);

        // Be sure this is a string
        settings.group = String(settings.group);

        return settings;
    };

    var updateSputnik = Kepler.utils.throttle(function() {
        var groups = {};

        jQuery.each(sputnikSections, function(i, entry) {
            entry = $(entry);

            // Get settings with default values
            var settings = getSettings(Kepler.utils.dataOptions(entry, "kepler-sputnik-options"));

            // Get the dispatcher
            var dispatcher = Kepler.utils.getJQueryObj(entry.data("kepler-sputnik-dispatcher"));
            if (!dispatcher) return;

            // If this entry belongs to a group, then add it to the groups array
            if (settings.group) {
                // Get the group list array
                var list = groups[settings.group];
                if (!list) {
                    // Create the array
                    list = [];

                    // Add the list array to the groups
                    groups[settings.group] = list;
                }

                // Add the current tracker with options to the array
                list.push({
                    entry: entry,
                    settings: settings,
                    dispatcher: dispatcher
                });

                // Continue the loop
                return;
            }

            if (Kepler.utils.visibleOnScreen(entry)) {
                if (!dispatcher.hasClass(settings.activeClass)) {
                    dispatcher.addClass(settings.activeClass);
                }
            }
            else {
                if (dispatcher.hasClass(settings.activeClass)) {
                    dispatcher.removeClass(settings.activeClass);
                }
            }
        });

        // TODO: Find the group item which is nearest to the viewport center...

        // Now handle the group items. Maximal one element of each group should be active...
        jQuery.each(groups, function(i, list) {
            jQuery.each(list, function(i, item) {
                if (Kepler.utils.visibleOnScreen(item.entry)) {
                    // Break the loop if the current item has already the active class
                    if (item.dispatcher.hasClass(item.settings.activeClass)) {
                        // Break the loop
                        return false;
                    }

                    // First remove all active classes
                    jQuery.each(list, function(i, item) {
                        if (item.dispatcher.hasClass(item.settings.activeClass)) {
                            item.dispatcher.removeClass(item.settings.activeClass);
                        }
                    });

                    // Set the new active class
                    item.dispatcher.addClass(item.settings.activeClass);

                    // Break the loop
                    return false;
                }
            });
        });
    }, 100);



    /*
     * Public Methods
     */

    // Sets the default options for the modal module
    this.defaultOptions = function(options) {
        defaultSettings = getSettings(options);

        // Return this to allow chaining
        return this;
    };



    /*
     * Kepler events
     */

    // Scan the DOM structure for data elements
    Kepler.onInit(function() {
        // Reset the sections array
        sputnikSections = [];

        $.each($("[data-kepler-sputnik-tracker]"), function(i, el){
            // Get jQuery object
            el = Kepler.utils.getJQueryObj(el);

            // Get the linked element
            var linkedElement;
            var data = el.data("kepler-sputnik-tracker");
            if (data === "self" || data === "this") {
                linkedElement = el;
            }
            else {
                linkedElement = Kepler.utils.getJQueryObj(data);
            }

            // Check if the linked element is valid
            if(!linkedElement) {
                console.log("Error: linked element reference invalid.");
                return;
            }

            // Save the dispatcher
            linkedElement.data("kepler-sputnik-dispatcher", el);

            // Add the linked element to the array
            sputnikSections.push(linkedElement);
        });

        // Unbind the scroll and resize events first
        var w = $(window);
        w.off('scroll', updateSputnik);
        w.off('resize', updateSputnik);

        // Only bind the events and call the function if the array is not empty
        if (sputnikSections.length > 0) {
            // Bind the scroll and resize events
            w.on('scroll', updateSputnik);
            w.on('resize', updateSputnik);

            // Call the method once to set the classes once on init
            updateSputnik();
        }
    });
};

/*
 * Kepler Module
 */
Kepler.module.tabs = new function() {
    /*
     * Private Variables
     */

    // Default settings
    var defaultSettings = {
        activeClass: 'active'
    };



    /*
     * Private Methods
     */

    var getSettings = function(options) {
        // Combine the settings, but preserve the objects. Create a deep copy...
        return $.extend({}, defaultSettings, options);
    };


    // Handler which shows a tab
    var showTabClickHandler = function() {
        // Get the tab content object
        var content = Kepler.utils.getJQueryObj($(this).data("kepler-tab"));
        if (!content) {
            console.log("error: data attribute kepler-tab is invalid: tab content object not found!");
            return;
        }

        // Get the parent
        var parent = Kepler.utils.getJQueryObj($(this).data("kepler-tabs-parent-obj"));
        if (!parent) {
            console.log("error: tabs parent not found!");
            return;
        }

        // Get the data options if set
        var settings = getSettings(Kepler.utils.dataOptions(parent, "kepler-tabs-options"));

        // Iterate over all tabs of the current parent and hide the active previous content
        $.each(parent.find("[data-kepler-tab]"), function(i, tab) {
            // Get the jQuery object
            tab = $(tab);

            // Get the tab content object and hide it if visible
            var content = Kepler.utils.getJQueryObj(tab.data("kepler-tab"));
            if (content && content.is(':visible')) {
                // Hide the content
                content.hide();

                // Remove the active class again
                content.removeClass(settings.activeClass);
                tab.parent('.tab-button').removeClass(settings.activeClass);
            }
        });

        // Show the new current content
        content.show();

        // Set the active class
        $(this).parent('.tab-button').addClass(settings.activeClass);
        content.addClass(settings.activeClass);
    };



    /*
     * Public Methods
     */

    // Sets the default options for the modal module
    this.defaultOptions = function(options) {
        defaultSettings = getSettings(options);

        // Return this to allow chaining
        return this;
    };



    /*
     * Kepler Events
     */

    // Scan the DOM structure for data elements
    Kepler.onInit(function() {
        // Find all objects with the data tabs attribute
        $.each($("[data-kepler-tabs]"), function(i, obj) {
             // Get the jQuery object
            obj = $(obj);

            // Iterate over all tabs of the current tab object
            $.each(obj.find("[data-kepler-tab]"), function(i, tab) {
                // Get the jQuery object
                tab = $(tab);

                // Save the parent reference
                tab.data("kepler-tabs-parent-obj", obj);

                // Detach the event first and attach the click handler again.
                // This way no dublicate events are bound if init is called mutliple times.
                tab.off('click', showTabClickHandler);
                tab.on('click', showTabClickHandler);
            });
        });
    });
};

/*
 * Kepler Module
 */
Kepler.module.tooltip = new function() {
    /*
     * Private Variables
     */

    // Default settings
    var defaultSettings = {
        html: false,            // Insert HTML into the popover. If false, jQuery's text method will be used to insert content into the DOM.
                                // Use text if you're worried about XSS attacks.
        trigger: 'hover',       // See the popover settings
        placement: 'auto top'   // See the popover settings

    };

    var uniqueTooltipId = 0;



    /*
     * Private Methods
     */

    var getSettings = function(options) {
        // Combine the settings, but preserve the objects. Create a deep copy...
        var settings = $.extend({}, defaultSettings, options);

        return settings;
    };

    var newUniqueTooltipID = function() {
        uniqueTooltipId++;
        if (uniqueTooltipId > 1000000) uniqueTooltipId = 1;
        return uniqueTooltipId;
    };

    var closeTriggerHandler = function() {
        // Remove the popover from the DOM structure again
        $(this).remove();
    };

    var openTriggerHandler = function() {
        // Get the target jQuery object
        var target = $(this);

        // Just exit, if the target hasn't any tooltip data attribute.
        // It might be removed by a custom javascript call.
        if (!target.attr("data-kepler-tooltip")) return false;

        var settings = target.data("kepler-popover-options");
        if (!settings) {
            console.log("error: no tooltip settings set!");
            return;
        }

        // Try to obtain the tooltip Id
        var popoverId = target.data("kepler-popover");

        // Check if there is already a popover object present
        if (popoverId && $(popoverId).length > 0) return;

        // Create a new unique popover ID
        popoverId = "kepler-tooltip-popover-" + newUniqueTooltipID();

        // Get the popover content
        var popover = target.data("kepler-tooltip");
        if (!popover) {
            console.log("error: no tooltip attribute set!");
            return;
        }

        // Create the popover object
        var popoverObj = $('<div id="' + popoverId + '" class="kepler tooltip radius shadow"></div>');

        // Insert HTML to the popover object. Use text if you're worried about XSS attacks.
        if (settings.html) {
            popoverObj.html(popover.toString());
        }
        else {
            popoverObj.text(popover.toString());
        }

        // Add the popover to the DOM content
        popoverObj.insertAfter(target);

        // Bind the closed event
        Kepler.popover.closed(popoverObj, closeTriggerHandler);

        // Add the popover Id to the target object
        target.data("kepler-popover", "#" + popoverId);
    };



    /*
     * Public Methods
     */

    // Sets the default options for the popover module
    this.defaultOptions = function(options) {
        defaultSettings = getSettings(options);

        // Return this to allow chaining
        return this;
    };



    /*
     * Kepler init
     */

    // Scan the DOM structure for data elements
    Kepler.onInit(function() {
        var settings;

        // Find all objects with the data tooltip attribute
        $.each($("[data-kepler-tooltip]"), function(i, obj) {
            // Make obj to a jQuery element
            obj = $(obj);

            // Get settings with default values
            settings = getSettings(Kepler.utils.dataOptions(obj, "kepler-tooltip-options"));

            // Set the popover options. These are the tooltip settings...
            obj.data("kepler-popover-options", settings);

            // Attach the open trigger event
            obj.off("kepler.popover.openTriggered", openTriggerHandler);
            obj.on("kepler.popover.openTriggered", openTriggerHandler);

            // Initialize the popover trigger event binding
            Kepler.popover.init(obj, settings);
        });
    });
};
