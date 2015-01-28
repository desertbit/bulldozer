/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package text

const templateText = `{{if #.EditModeActive}}
{{if %.store.IsBlocked}}
	<div class="bulldozer_blocked_border">{{#.Text}}</div>
{{else}}
	<div id="{{id "text"}}" data-kepler-popover="#{{id "pop"}}" data-kepler-popover-options="placement:auto top;">{{#.Text}}</div>
	<div id="{{id "pop"}}" class="kepler popover radius shadow bulldozer_popover">
		<a id="{{id "edit"}}">
			<i class="fa fa-pencil"></i>
			<span>{{tr "blz.plugin.text.edit"}}</span>
		</a>
	</div>
	{{js load}}
		$("#{{id "text"}}").attr("contenteditable", "false")
			.addClass('bulldozer_click_to_edit_hover_border');

		Kepler.popover.openRequested("#{{id "pop"}}", function() {
			return !($("#{{id "text"}}").hasClass("bulldozer_text_plugin_edit_border"));
		});

		Kepler.popover.opened("#{{id "pop"}}", function() {
			$("#{{id "text"}}").addClass("bulldozer_click_to_edit_border");
		});

		Kepler.popover.closed("#{{id "pop"}}", function() {
			$("#{{id "text"}}").removeClass("bulldozer_click_to_edit_border");
		});

		$("#{{id "edit"}}").click(function() {
			{{emit Lock()}}
		});

		{{event edit()}}
			var el = $("#{{id "text"}}");

			el.attr("contenteditable", "true")
				.removeClass("bulldozer_click_to_edit_hover_border")
				.addClass("bulldozer_text_plugin_edit_border");

			var editor = el.data("ckeditor");

			if(!editor){
				editor = CKEDITOR.inline("{{id "text"}}",{
					startupFocus: true,
					{{/*{{if $.IsModeFull}}*/}}
						toolbarGroups: [
							{ name: 'clipboard',   groups: [ 'clipboard', 'undo' ] },
							{ name: 'editing',     groups: [ 'find', 'selection', 'spellchecker' ] },
							{ name: 'links' },
							{ name: 'insert' },
							{ name: 'forms' },
							{ name: 'tools' },
							{ name: 'document',	   groups: [ 'mode', 'document', 'doctools' ] },
							{ name: 'others' },
							'/',
							{ name: 'basicstyles', groups: [ 'basicstyles', 'cleanup' ] },
							{ name: 'paragraph',   groups: [ 'list', 'indent', 'blocks', 'align', 'bidi' ] },
							{ name: 'styles' },
							{ name: 'colors' },
							{ name: 'about' }
						],
						removeButtons: 'Underline,Subscript,Superscript'
					{{/*{{else if $.IsModeMinimal}}
						toolbar: [
						    [ 'Bold', 'Italic', 'Underline', 'Strike', '-', 'Link', 'Unlink' ],
						    [ 'TextColor', 'BGColor' ],
						    [ 'Format' ],
						    [ 'Undo', 'Redo' ],
						    [ 'Cut', 'Copy', 'Paste' ]
						]
					{{end}}*/}}
				});

				el.data("ckeditor", editor);
				
				editor.on("blur", function(e) {
					$("#{{id "text"}}").attr("contenteditable", "false")
						.addClass("bulldozer_click_to_edit_hover_border")
						.removeClass("bulldozer_text_plugin_edit_border");
					{{emit SetText(e.editor.getData())}}
					$(e.editor).data("focusManagerLocked", true);
					e.editor.focusManager.lock();
					{{emit Unlock()}}
				});
			}
			else {
				editor.focusManager.unlock();
				editor.focus();
			}

			Kepler.popover.close("#{{id "pop"}}");
		{{end event}}
	{{end js}}
{{end}}
{{else}}{{#.Text}}{{end}}`
