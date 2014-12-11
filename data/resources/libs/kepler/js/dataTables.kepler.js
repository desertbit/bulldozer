/*
*  Kepler Frontend Framework
*  Copyright, DesertBit
*  Free to use under the GPL license.
*  http://www.gnu.org/copyleft/gpl.html
*/


/*
 * DataTables integration for Kepler. This requires Kepler 1 and
 * DataTables 1.10 or newer.
 * Based on Foundation 5 dataTables integration plug-in
 *
 * This file sets layout and style of datables to use Kepler elements.
 *
 */

(function(window, document, undefined){

    var factory = function( $, DataTable ) {
        "use strict";


        $.extend( DataTable.ext.classes, {
            sWrapper: "dataTables_wrapper dataTables-kepler"
        } );


        /* Set the defaults for DataTables initialisation */
        $.extend( true, DataTable.defaults, {
            dom:
            "<'kepler grid'<'small-12 medium-6 column'l><' small-12 medium-6 column'f>r>"+
            "t"+
            "<'kepler grid'<'small-12 medium-6 column context-wrap'i><'small-12 medium-6 column pagination-wrap'p>>",
            renderer: 'kepler'
        } );


        /* Page button renderer */
        DataTable.ext.renderer.pageButton.kepler = function ( settings, host, idx, buttons, page, pages ) {
            var api = new DataTable.Api( settings );
            var classes = settings.oClasses;
            var lang = settings.oLanguage.oPaginate;
            var btnDisplay, btnClass;

            var attach = function( container, buttons ) {
                var i, ien, node, button;
                var clickHandler = function ( e ) {
                    e.preventDefault();
                    if ( e.data.action !== 'ellipsis' ) {
                        api.page( e.data.action ).draw( false );
                    }
                };

                for ( i=0, ien=buttons.length ; i<ien ; i++ ) {
                    button = buttons[i];

                    if ( $.isArray( button ) ) {
                        attach( container, button );
                    }
                    else {
                        btnDisplay = '';
                        btnClass = '';

                        switch ( button ) {
                            case 'ellipsis':
                                btnDisplay = '&hellip;';
                                btnClass = 'unavailable';
                                break;

                            case 'first':
                                btnDisplay = lang.sFirst;
                                btnClass = button + (page > 0 ?
                                                     '' : ' unavailable');
                                break;

                            case 'previous':
                                btnDisplay = lang.sPrevious;
                                btnClass = button + (page > 0 ?
                                                     '' : ' unavailable');
                                break;

                            case 'next':
                                btnDisplay = lang.sNext;
                                btnClass = button + (page < pages-1 ?
                                                     '' : ' unavailable');
                                break;

                            case 'last':
                                btnDisplay = lang.sLast;
                                btnClass = button + (page < pages-1 ?
                                                     '' : ' unavailable');
                                break;

                            default:
                                btnDisplay = button + 1;
                                btnClass = page === button ?
                                    'current' : '';
                                break;
                        }

                        if (btnDisplay) {
                            node = $('<a>', {
                                'class': classes.sPageButton+' '+btnClass + ' pagelink',
                                'aria-controls': settings.sTableId,
                                'tabindex': settings.iTabIndex,
                                'id': idx === 0 && typeof button === 'string' ?
                                settings.sTableId +'_'+ button :
                                null
                            } )
                            .html(btnDisplay)
                            .appendTo(container);

                            settings.oApi._fnBindAction(
                                node, {action: button}, clickHandler
                            );
                        }
                    }
                }
            };

            attach(
                $(host).empty().html('<div class="pagination"/>').children('div'),
                buttons
            );
        };
    }; // /factory


    // Define as an AMD module if possible
    if ( typeof define === 'function' && define.amd ) {
        define( ['jquery', 'datatables'], factory );
    }
    else if ( typeof exports === 'object' ) {
        // Node/CommonJS
        factory( require('jquery'), require('datatables') );
    }
    else if ( jQuery ) {
        // Otherwise simply initialise as normal, stopping multiple evaluation
        factory( jQuery, jQuery.fn.dataTable );
    }


})(window, document);

