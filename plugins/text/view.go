/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package text

const templateText = `{{if #.EditModeActive}}
{{if %.store.IsBlocked}}
	<div class="bulldozer_blocked_border">{{#.Text}}</div>
{{else}}
	<div id="{{id "text"}}" class="bulldozer_text_plugin_edit" data-kepler-popover="#{{id "pop"}}" data-kepler-popover-options="placement:auto top;">{{#.Text}}</div>
	<div id="{{id "pop"}}" class="kepler popover radius shadow bulldozer_popover">
		<a id="{{id "edit"}}">
			<i class="fa fa-pencil"></i>
			<span>{{tr "blz.plugin.text.edit"}}</span>
		</a>
	</div>
	{{js load}}
		$("#{{id "text"}}").attr("contenteditable", "false")
			.addClass('bulldozer_click_to_edit_hover_border');

		$("#{{$.Context.DomID}}").click(function(e) {
		    e.stopPropagation();
		    return false;
		});

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
					{{if eq #.Mode "` + ModeFull + `"}}
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
							{ name: 'paragraph',   groups: [ 'list', 'indent', 'blocks' ] },
							{ name: 'paragraph',   groups: [ 'align', 'bidi' ] },
							{ name: 'colors' },
							'/',
							{ name: 'styles' }
						],
						removeButtons: 'Subscript,Superscript'
					{{else if eq #.Mode "` + ModePlain + `"}}
						toolbar: [ [ 'Undo', 'Redo' ] ]
					{{else if eq #.Mode "` + ModeMinimal + `"}}
						toolbarGroups: [
							{ name: 'undo' },
							{ name: 'basicstyles', groups: [ 'basicstyles', 'cleanup' ] },
							{ name: 'paragraph',   groups: [ 'align' ] },
							{ name: 'colors' }
						],
						removeButtons: 'Subscript,Superscript'
					{{else}}
						toolbarGroups: [
							{ name: 'undo' },
							{ name: 'links' },
							{ name: 'basicstyles', groups: [ 'basicstyles', 'cleanup' ] },
							{ name: 'paragraph',   groups: [ 'align' ] },
							'/',
							{ name: 'styles' },
							{ name: 'colors' },
							{ name: 'paragraph',   groups: [ 'list', 'blocks' ] }
						],
						removeButtons: 'Subscript,Superscript,Font,Styles'
					{{end}}
				});

				el.data("ckeditor", editor);

				editor.on("change", Kepler.utils.throttle(function() {
					if (el.data("isActive")) {
						{{emit SetText(editor.getData())}}
					}
				}, 5000));

				editor.on("blur", function() {
					el.data("isActive", false);
					$("#{{id "text"}}").attr("contenteditable", "false")
						.addClass("bulldozer_click_to_edit_hover_border")
						.removeClass("bulldozer_text_plugin_edit_border");
					{{emit SetText(editor.getData())}}
					editor.focusManager.lock();
					{{emit Unlock()}}
				});
			}
			else {
				editor.focusManager.unlock();
				editor.focus();
			}

			el.data("isActive", true);
			Kepler.popover.close("#{{id "pop"}}");
		{{end event}}
	{{end js}}
{{end}}
{{else}}{{#.Text}}{{end}}`
