/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package datatables

import (
	tr "github.com/desertbit/bulldozer/translate"

	"github.com/desertbit/bulldozer/sessions"
	"github.com/desertbit/bulldozer/settings"
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
		"emptyTable":     "`+tr.S("bud.datatables.emptyTable")+`",
		"info":           "`+tr.S("bud.datatables.info")+`",
		"infoEmpty":      "`+tr.S("bud.datatables.infoEmpty")+`",
		"infoFiltered":   "`+tr.S("bud.datatables.infoFiltered")+`",
		"infoPostFix":    "`+tr.S("bud.datatables.infoPostFix")+`",
		"thousands":      "`+tr.S("bud.datatables.thousands")+`",
		"lengthMenu":     "`+tr.S("bud.datatables.lengthMenu")+`",
		"loadingRecords": "`+tr.S("bud.datatables.loadingRecords")+`",
		"processing":     "`+tr.S("bud.datatables.processing")+`",
		"search":         "`+tr.S("bud.datatables.search")+`",
		"zeroRecords":    "`+tr.S("bud.datatables.zeroRecords")+`",
		"paginate": {
			"first":      "`+tr.S("bud.datatables.Pagination.first")+`",
			"last":       "`+tr.S("bud.datatables.Pagination.last")+`",
			"next":       "`+tr.S("bud.datatables.Pagination.next")+`",
			"previous":   "`+tr.S("bud.datatables.Pagination.previous")+`"
		},
		"aria": {
			"sortAscending":  "`+tr.S("bud.datatables.sortAscending")+`",
			"sortDescending": "`+tr.S("bud.datatables.sortDescending")+`"
		}
	}
} );`)

	s.LoadJavaScript(settings.UrlBulldozerResources+"libs/kepler/js/dataTables.kepler.min.js", "")
}
