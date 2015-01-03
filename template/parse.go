/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package template

import (
	"fmt"
	"strings"
)

var (
	parseFuncs map[string]parseFunc = make(map[string]parseFunc)
)

//###############//
//### Private ###//
//###############//

type parseData struct {
	t *Template

	leftDelim  string
	rightDelim string

	src       *string
	final     *string
	lineCount *int
}

type parseFunc func(typeStr string, token string, d *parseData) error

// This has to be called during initialization. THis is not thread-safe.
func registerParseFunc(typeStr string, f parseFunc) {
	parseFuncs[typeStr] = f
}

func parse(t *Template, src string, linesCountOffset int) (final string, err error) {
	// Recover panics and return the error
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("panic: parse bulldozer code: %v", e)
		}
	}()

	// Lines count start with 1. Also add the lines count offset.
	lineCount := 1 + linesCountOffset

	// Create a new parse data value.
	parseData := &parseData{
		t:          t,
		leftDelim:  t.leftDelim,
		rightDelim: t.rightDelim,
		src:        &src,
		final:      &final,
		lineCount:  &lineCount,
	}

	leftDelimLen := len(parseData.leftDelim)

	// Find each left delimiter and process it.
	for pos := strings.Index(src, t.leftDelim); pos != -1; pos = strings.Index(src, t.leftDelim) {
		// Update the line count
		lineCount += countLinesUntilPos(src, pos)

		// Add text which is in front of the left delimiter
		final += src[0:pos]

		// Remove first part including the left delimiters
		src = src[pos+leftDelimLen:]

		// Get the template token with nested template code.
		token, err := getSectionBetweenDelim(parseData.leftDelim, parseData.rightDelim, &src)
		if err != nil {
			return "", fmt.Errorf("%d: '%s': %v", lineCount, parseData.leftDelim, err)
		}

		// Replace the custom tags '#.' and '%.'
		token = replaceCustomTags(token)

		// Save the template Code for error messages
		templateCode := parseData.leftDelim + token + parseData.rightDelim

		// Define the type string variable
		var typeStr string

		// Try to find the type string delimiter position
		typePos := strings.Index(token, " ")

		if typePos == -1 {
			// If not found, then this token has no arguments.
			// The complete token text is the type text.
			typeStr = token
			token = ""
		} else {
			// Extract the type string and remove it from the original token
			typeStr = token[0:typePos]
			token = token[typePos+1:]
		}

		// Try to obtain the parse function for the type if present
		f, ok := parseFuncs[typeStr]
		if ok {
			// Call the function
			err = f(typeStr, token, parseData)
			if err != nil {
				return "", fmt.Errorf("%d: '%s': %v", lineCount, templateCode, err)
			}
		} else {
			// Just add the template code to the final data.
			final += templateCode
		}
	}

	// Append the rest of the source data to the final string
	// and wrap the final source between a div tag with the template ID.
	// Also execute the js load event for the current template.
	final = `<div id="{{$.Context.DomID}}"{{with $.Context.StylesString}} class="{{.}}"{{end}}>` + final + src + `<script>Bulldozer.core.execJsLoad("{{$.Context.DomID}}");</script></div>`

	return final, nil
}

// Count all lines until the text index
func countLinesUntilPos(data string, pos int) (i int) {
	// Lines count
	i = 0

	// Create a substring to the text index
	if pos >= 0 {
		data = data[0:pos]
	}

	for p := strings.Index(data, "\n"); p != -1; p = strings.Index(data, "\n") {
		// Increase the new lines count
		i++

		// Remove all the text until the index of the new line,
		// to continue the search.
		data = data[p+1:]
	}

	return
}

func getSection(tag string, data *string, d *parseData) (string, error) {
	// Create the tags
	startTag := d.leftDelim + tag + " "
	endTag := d.leftDelim + "end " + tag + d.rightDelim

	// Get the section
	return getSectionBetweenDelim(startTag, endTag, data)
}

func getSectionBetweenDelim(startTag string, endTag string, data *string) (string, error) {
	// Remove all nested sections and update the final position
	pos, err := removeNestedSections(startTag, endTag, *data)
	if err != nil {
		return "", err
	}

	// Get the section data
	section := (*data)[0:pos]

	// Remove the section from the data string
	*data = (*data)[pos+len(endTag):]

	return section, nil
}

func removeNestedSections(startTag string, endTag string, data string) (int, error) {
	startTagLen := len(startTag)
	endTagLen := len(endTag)

	posFinal := 0

	// Find each start tag and skip nested tags with the same tag.
	// Find the right ending tag for this section...
	for {
		posStart := strings.Index(data, startTag)
		posEnd := strings.Index(data, endTag)

		// Check if the end tag exists
		if posEnd == -1 {
			return -1, fmt.Errorf("Invalid syntax! Missing end tag '%s'", endTag)
		}

		// Check if the found end tag belongs to a nested section with the same tag
		if posStart != -1 && posStart < posEnd {
			// Remove everything to the start tag
			data = data[posStart+startTagLen:]

			// Remove all nested sections and update the final position
			pos, err := removeNestedSections(startTag, endTag, data)
			if err != nil {
				return -1, err
			}

			// Remove the nested section from the temporary data string
			data = data[pos+endTagLen:]

			// Update the final position
			posFinal += posStart + startTagLen + pos + endTagLen

			// Continue the search
			continue
		}

		// Update the final position
		posFinal += posEnd

		// Break the loop
		break
	}

	return posFinal, nil
}

func replaceCustomTags(src string) string {
	var final []rune = make([]rune, 0)
	escaped, skip, injectContext := false, false, false
	l := len(src) - 1

	// Iterate through the string characters.
	for i, p := range src {
		if escaped {
			escaped = false
			final = append(final, p)
			continue
		}

		if p == '\\' {
			escaped = true
		} else if p == '"' {
			skip = !skip
		} else if injectContext && (p == ' ' || p == '}') {
			injectContext = false
			final = append(final, []rune(" $.Context")...)
		} else if !skip && p == '#' {
			final = append(final, []rune("$.Data")...)
			continue
		} else if !skip && p == '%' {
			final = append(final, []rune("$.Pkg")...)

			// Add the context as first function parameter if this is a package function call.
			if i < l && src[i+1] == '.' {
				injectContext = true
			}

			continue
		}

		final = append(final, p)
	}

	if injectContext {
		final = append(final, []rune(" $.Context")...)
	}

	return string(final)
}
