/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package text

const templateText = `{{if #.EditModeActive}}
{{if %.store.IsBlocked}}
	<div class="bulldozer_blocked_border">{{#.Text}}</div>
{{else}}
	<div id="{{id "text"}}" class="bulldozer_text_plugin_edit">{{#.Text}}</div>
	{{js load}}
		$("#{{id "text"}}").attr("contenteditable", "false")
			.addClass('bulldozer_click_to_edit_hover_border');

		$("#{{$.Context.DomID}}").click(function(e) {
		    e.stopPropagation();
		    return false;
		});

		$("#{{id "text"}}").click(function() {
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

				var updateText = function() {
					if (editor.checkDirty()) {
						{{emit SetText(editor.getData())}}
						editor.resetDirty();
					}
				};

				editor.on("change", Kepler.utils.throttle(function() {
					if (el.data("isActive")) {
						updateText();
					}
				}, 5000));

				editor.on("focus", function() {
					{{emit Lock()}}
				});

				editor.on("blur", function() {
					el.data("isActive", false);
					$("#{{id "text"}}").attr("contenteditable", "false")
						.addClass("bulldozer_click_to_edit_hover_border")
						.removeClass("bulldozer_text_plugin_edit_border");
					if (el.data("lockFailed")) {return;}
					updateText();
					{{emit Unlock()}}
				});
			}
			else {
				editor.focus();
				editor;
			}

			el.data("isActive", true);
		});

		{{event lockFailed()}}
			var el = $("#{{id "text"}}");
			el.data("lockFailed", true);
			setTimeout(function() {
				el.data("lockFailed", false);
			}, 1000);
			var editor = $("#{{id "text"}}").data("ckeditor");
			if (!editor) {return;}
			editor.focusManager.blur();
		{{end event}}
	{{end js}}
{{end}}
{{else}}
	{{if #.Protect}}
	{{js load}}
		{{emit GetProtectedData()}}
		{{event setProtectedData(data)}}
			$('#{{$.Context.DomID}}').html(data);
		{{end event}}
	{{end js}}
	{{else}}{{#.Text}}{{end}}
{{end}}`
