/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"html"
	"html/template"
	"math/big"
	"strings"
	"time"
)

func AddTrailingSlashToPath(path string) string {
	return strings.TrimSuffix(path, "/") + "/"
}

// Fields splits a string between spaces, but skips spaces in quotes.
// nil is returned, if the string is empty or if it only contains empty spaces.
func Fields(s string) ([]string, error) {
	// Trim all empty spaces
	s = strings.TrimSpace(s)

	// If s is empty, then return nil
	if len(s) == 0 {
		return nil, nil
	}

	var l []string
	var data []rune
	var dataStr string
	skip := false

	// Split the string
	for _, p := range s {
		if p == '"' {
			skip = !skip
			data = append(data, p)
		} else if !skip && p == ' ' {
			dataStr = strings.TrimSpace(string(data))
			if len(dataStr) > 0 {
				l = append(l, dataStr)
			}

			data = data[:0]
		} else {
			data = append(data, p)
		}
	}

	dataStr = strings.TrimSpace(string(data))
	if len(dataStr) > 0 {
		if skip {
			return nil, fmt.Errorf("utils fields: failed to split string. Missing end quote: '%s'", s)
		}

		l = append(l, dataStr)
	}

	return l, nil
}

// LimitLen will cut a string if it exceeds the max length.
func LimitLen(s string, max int) string {
	if len(s) <= max {
		return s
	}

	return s[:max]
}

// RandomString generates a random string with a length of n
func RandomString(n int) string {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	symbols := big.NewInt(int64(len(alphanum)))
	states := big.NewInt(0)
	states.Exp(symbols, big.NewInt(int64(n)), nil)
	r, err := rand.Int(rand.Reader, states)
	if err != nil {
		panic(err)
	}
	var bytes = make([]byte, n)
	r2 := big.NewInt(0)
	symbol := big.NewInt(0)
	for i := range bytes {
		r2.DivMod(r, symbols, symbol)
		r, r2 = r2, r
		bytes[i] = alphanum[symbol.Int64()]
	}
	return string(bytes)
}

// ToPath returns a valid path.
func ToPath(path string) string {
	// Trim, to lower and replace all empty spaces.
	path = strings.Replace(strings.ToLower(strings.TrimSpace(path)), " ", "-", -1)

	// Remove everything after #
	p := strings.Index(path, "#")
	if p >= 0 {
		path = path[:p]
	}

	// Remove the following / if necessary.
	if strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}

	// Append a leading / if necessary.
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	return path
}

// EscapeJS escapes a string to be send over the SendCommand method.
// This method escapes all backslaches, simple quotes and new lines.
func EscapeJS(data string) string {
	final := make([]rune, 0, len(data))

	// Replace all \ with \\, ' with \' and \n with \\n
	for _, p := range data {
		if p == '\n' {
			final = append(final, '\\', 'n')
		} else if p == '\\' || p == '\'' {
			final = append(final, '\\', p)
		} else {
			final = append(final, p)
		}
	}

	// Return the final escaped string
	return string(final)
}

// ErrorBox returns a styled div error box.
// One optional argument can be passed to show a detailed error code.
// This error code is only shown to the user if the sessions is authenticated
// as developer.
func ErrorBox(err string, vars ...interface{}) template.HTML {
	body := `<div class="kepler panel warning icon">` +
		`<h3 class="headline">` + html.EscapeString(err) + `</h3>`

	// TODO: Only show this if the user is authenticated as developer.
	if len(vars) >= 1 {
		body += "<code>" + html.EscapeString(fmt.Sprint(vars[0])) + "</code></div>"
	}

	return template.HTML(body)
}

func Sha256Sum(s string) string {
	// Generate the final SHA1 salt
	h := sha256.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// EncryptXorBase64 is a simple XOR encryption
func EncryptXorBase64(key string, s string) string {
	// Simply XOR the string with the key
	nk := len(key)
	n := len(s)
	b := make([]byte, n)
	for i := 0; i < n; i++ {
		b[i] = s[i] ^ key[i%nk]
	}

	// Encode the byte slice with Base64
	return base64.URLEncoding.EncodeToString(b)
}

// DecryptXorBase64 is a simple XOR decryption
func DecryptXorBase64(key string, s string) (string, error) {
	// Decode the Base64 string
	b, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}

	// Simply XOR the byte slice with the key
	nk := len(key)
	n := len(b)
	f := make([]byte, n)
	for i := 0; i < n; i++ {
		f[i] = b[i] ^ key[i%nk]
	}

	return string(f), nil
}

// EncryptDomId calculates a short unique DOM ID hash
func EncryptDomId(key string, s string) (hash string) {
	// Encrypt the string with the new key
	hash = EncryptXorBase64(key, s)

	// Remove the following Base64 equals if present
	hash = strings.TrimSuffix(strings.TrimSuffix(hash, "="), "=")
	return
}

// Debounce executes a function when it stops being invoked for the delay.
func Debounce(delay time.Duration, f func()) func() {
	// Create the timer.
	timer := time.AfterFunc(delay, f)
	timer.Stop()

	// Return the lazy function.
	return func() {
		// Reset the timer.
		timer.Reset(delay)
	}
}
