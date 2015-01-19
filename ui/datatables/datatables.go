/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package datatables

import (
	tr "code.desertbit.com/bulldozer/bulldozer/translate"

	"code.desertbit.com/bulldozer/bulldozer/sessions"
	"code.desertbit.com/bulldozer/bulldozer/settings"
)

// Load loads the required files for the datatables plugin.
func Load(s *sessions.Session) {
	// Load the DataTables javascripts
	s.LoadJavaScript(settings.UrlBulldozerResources+"libs/kepler/js/vendors/datatables/jquery.dataTables.min.js", `
$.extend( $.fn.dataTable.defaults, {
    "searching": true,
    "ordering": true,
	"paging": true,
	"lengthChange": true,
	"info": true,
	"language": {
		"emptyTable":     "`+tr.S("blz.datatables.emptyTable")+`",
		"info":           "`+tr.S("blz.datatables.info")+`",
		"infoEmpty":      "`+tr.S("blz.datatables.infoEmpty")+`",
		"infoFiltered":   "`+tr.S("blz.datatables.infoFiltered")+`",
		"infoPostFix":    "`+tr.S("blz.datatables.infoPostFix")+`",
		"thousands":      "`+tr.S("blz.datatables.thousands")+`",
		"lengthMenu":     "`+tr.S("blz.datatables.lengthMenu")+`",
		"loadingRecords": "`+tr.S("blz.datatables.loadingRecords")+`",
		"processing":     "`+tr.S("blz.datatables.processing")+`",
		"search":         "`+tr.S("blz.datatables.search")+`",
		"zeroRecords":    "`+tr.S("blz.datatables.zeroRecords")+`",
		"paginate": {
			"first":      "`+tr.S("blz.datatables.Pagination.first")+`",
			"last":       "`+tr.S("blz.datatables.Pagination.last")+`",
			"next":       "`+tr.S("blz.datatables.Pagination.next")+`",
			"previous":   "`+tr.S("blz.datatables.Pagination.previous")+`"
		},
		"aria": {
			"sortAscending":  "`+tr.S("blz.datatables.sortAscending")+`",
			"sortDescending": "`+tr.S("blz.datatables.sortDescending")+`"
		}
	}
} );`)

	s.LoadJavaScript(settings.UrlBulldozerResources+"libs/kepler/js/dataTables.kepler.min.js", "")
}
