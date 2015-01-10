/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package text

const templateText = `{{#}}`

const templateTexta = `{{if $.EditableMode}}
	<div id="{{{id text}}}" data-kepler-popover="#{{{id pop}}}" data-kepler-popover-options="placement:auto top;">{{.Text}}</div>
	<div id="{{{id pop}}}" class="kepler popover radius shadow goji_popup_widget">
		<a id="{{{id edit}}}">
			<i class="fa fa-pencil"></i>
			<span>{{tr "textplugin.edit"}}</span>
		</a>
	</div>

	{{{js load}}}
		$("#{{{id text}}}").attr("contenteditable", "false")
			.addClass('goji_click_to_edit_hover_border');

		Kepler.popover.openRequested("#{{{id pop}}}", function() {
			return !($("#{{{id text}}}").hasClass("goji_text_plugin_edit_border"));
		});

		Kepler.popover.opened("#{{{id pop}}}", function() {
			$("#{{{id text}}}").addClass("goji_click_to_edit_border");
		});

		Kepler.popover.closed("#{{{id pop}}}", function() {
			$("#{{{id text}}}").removeClass("goji_click_to_edit_border");
		});

		$("#{{{id edit}}}").click(function() {
			var el = $("#{{{id text}}}");

			el.attr("contenteditable", "true")
				.removeClass("goji_click_to_edit_hover_border")
				.addClass("goji_text_plugin_edit_border");

			var editor = el.data("ckeditor");

			if(!editor){
				editor = CKEDITOR.inline("{{{id text}}}",{
					startupFocus: true,
					{{if $.IsModeFull}}
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
					{{else if $.IsModeMinimal}}
						toolbar: [
						    [ 'Bold', 'Italic', 'Underline', 'Strike', '-', 'Link', 'Unlink' ],
						    [ 'TextColor', 'BGColor' ],
						    [ 'Format' ],
						    [ 'Undo', 'Redo' ],
						    [ 'Cut', 'Copy', 'Paste' ]
						]
					{{end}}
				});

				el.data("ckeditor", editor);
				
				editor.on("blur", function(e) {
					$("#{{{id text}}}").attr("contenteditable", "false")
						.addClass("goji_click_to_edit_hover_border")
						.removeClass("goji_text_plugin_edit_border");
					{{{emit setText(e.editor.getData())}}}
					$(e.editor).data("focusManagerLocked", true);
					e.editor.focusManager.lock();
				});
			}
			else {
				editor.focusManager.unlock();
				editor.focus();
			}

			Kepler.popover.close("#{{{id pop}}}");
		});
	{{{end js}}}
{{else}}{{.Text}}{{end}}`
