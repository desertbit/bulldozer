/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package protect

const templateText = `{{js load}}
{{emit GetContent()}}
{{event setContent(data)}}
	$('#{{$.Context.DomID}}').html(data);
{{end event}}
{{end js}}`
