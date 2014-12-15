/*
 *  Goji Framework
 *  Copyright (C) Roland Singer & Wlad Meixner
 */

package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"html/template"
	"math/big"
	"os"
	"strings"
)

func MkDirIfNotExist(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.Mkdir(path, 0700)
	}

	return nil
}

func AddTrailingSlashToPath(path string) string {
	return strings.TrimSuffix(path, "/") + "/"
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
	// Trim, to lower and replace all empty spaces
	path = strings.Replace(strings.ToLower(strings.TrimSpace(path)), " ", "-", -1)

	// Remove the following / if necessary
	if strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}

	// Append a leading / if necessary
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

// ErrorBox returns a styled div error box
func ErrorBox(err string) template.HTML {
	return template.HTML("<div class=\"kepler alert-box warning icon\">" + err + "</div>")
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
