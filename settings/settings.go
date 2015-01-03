/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package settings

import (
	"code.desertbit.com/bulldozer/bulldozer/utils"
	"fmt"
	"github.com/golang/glog"
	"os"
	"strings"
)

const (
	/*
	 *  Public
	 */

	TemplateSuffix = ".tmpl"

	// The socket types
	TypeTcpSocket  SocketType = 1 << iota
	TypeUnixSocket SocketType = 1 << iota

	// Static URL paths
	UrlPublic             = "/public/"
	UrlBulldozerResources = "/bulldozer/res/"

	/*
	 *  Private
	 */

	bulldozerGoPath      = "src/code.desertbit.com/bulldozer/bulldozer/"
	tmpDirName           = "bulldozer"
	sessionsDatabaseName = "sessions.db"

	// Default cookie keys
	defaultCookieHashKey  = "R7DqYdgWlztQ06diRM4z7ByuDwfiAvehLxTwAEDHFvgjkA4CcPrWBhZk6FJIBuDs"
	defaultCookieBlockKey = "2Mox41MlNDHOzShGfiO6AMq3isx5hz9r"
)

var (
	Settings settings
)

type SocketType int

func init() {
	// Set the default values
	Settings = settings{
		AutoSetGOMAXPROCS:   true,
		AutoParseFlags:      true,
		AutoCatchInterrupts: true,

		SiteUrl:           "http://127.0.0.1:9000",
		SecureHttpsAccess: false,
		SocketType:        TypeTcpSocket,
		ListenAddress:     ":9000",
		ServeFiles:        true,

		CookieHashKey:  []byte(defaultCookieHashKey),
		CookieBlockKey: []byte(defaultCookieBlockKey),
		SessionMaxAge:  60 * 60 * 24 * 14, // 14 Days

		FirewallMaxRequestsPerMinute: 100,
		FirewallReleaseBlockAfter:    60 * 5, // 5 minutes

		// Set the required stylesheets
		StaticStyleSheets: []string{
			UrlBulldozerResources + "css/bulldozer.css",
		},

		// Set the required scripts
		StaticJavaScripts: []string{
			UrlBulldozerResources + "js/jquery.min.js",
			UrlBulldozerResources + "js/jquery.history.js",
			UrlBulldozerResources + "libs/kepler/js/vendors/fastclick/fastclick.min.js",
			UrlBulldozerResources + "libs/kepler/js/kepler.min.js",
			UrlBulldozerResources + "js/sha256.js",
			UrlBulldozerResources + "js/bulldozer.min.js",
		},
	}

	// Set the temporary directory path
	Settings.TmpPath = utils.AddTrailingSlashToPath(utils.AddTrailingSlashToPath(os.TempDir()) + tmpDirName)

	// Get the current working directory path
	var err error
	Settings.WorkingPath, err = os.Getwd()
	if err != nil {
		glog.Fatalf("failed to obtain current working directory path: %v", err)
	}

	// Set the GOPATH
	Settings.GoPath = os.Getenv("GOPATH")
	if len(Settings.GoPath) == 0 {
		glog.Fatalf("GOPATH is not set!")
	}

	// Append a trailing slash if not already present
	Settings.GoPath = utils.AddTrailingSlashToPath(Settings.GoPath)
	Settings.WorkingPath = utils.AddTrailingSlashToPath(Settings.WorkingPath)

	// Set the paths
	Settings.SessionsDatabasePath = Settings.TmpPath + sessionsDatabaseName

	Settings.PublicPath = Settings.WorkingPath + "public"
	Settings.PagesPath = Settings.WorkingPath + "pages"
	Settings.TemplatesPath = Settings.WorkingPath + "templates"
	Settings.CoreTemplatesPath = Settings.TemplatesPath + "/core"
	Settings.TranslationPath = Settings.WorkingPath + "translations"

	Settings.BulldozerSourcePath = Settings.GoPath + bulldozerGoPath
	Settings.BulldozerCoreTemplatesPath = Settings.BulldozerSourcePath + "/data/templates"
	Settings.BulldozerResourcesPath = Settings.BulldozerSourcePath + "/data/resources"
	Settings.BulldozerTranslationPath = Settings.BulldozerSourcePath + "/data/translations"

}

//##############//
//### Public ###//
//##############//

// Check checks if the settings are correct and valid
func Check() error {
	// Check if the Site url is valid
	if !strings.HasPrefix(Settings.SiteUrl, "http://") &&
		!strings.HasPrefix(Settings.SiteUrl, "https://") {
		return fmt.Errorf("settings: site url is invalid: missing 'http://' or 'https://': '%s'", Settings.SiteUrl)
	}

	// Check if the length of the cookie keys are valid
	l := len(Settings.CookieHashKey)
	if l != 32 && l != 64 {
		return fmt.Errorf("settings: the cookie hash key has an invalid length! Valid lengths are 32 or 64 bytes...")
	}
	l = len(Settings.CookieBlockKey)
	if l != 16 && l != 24 && l != 32 {
		return fmt.Errorf("settings: the cookie block key has an invalid length! For AES, used by default, valid lengths are 16, 24, or 32 bytes to select AES-128, AES-192, or AES-256.")
	}

	// Print a warning if the default cookie keys are set
	if string(Settings.CookieHashKey) == defaultCookieHashKey {
		glog.Warningf("[WARNING] settings: the default cookie hash key is set! You should replace this with a secret key!")
	}
	if string(Settings.CookieBlockKey) == defaultCookieBlockKey {
		glog.Warningf("[WARNING] settings: the default cookie block key is set! You should replace this with a secret key!")
	}

	// Print a warning if the SecureHttpsAccess flag is false
	if !Settings.SecureHttpsAccess {
		glog.Warningf("[WARNING] settings: the secure https access flag is false! You should provide a secure https access!")
	}

	// Remove leading / from the site url
	Settings.SiteUrl = strings.TrimSuffix(Settings.SiteUrl, "/")

	return nil
}

//###############//
//### Private ###//
//###############//

type settings struct {
	// If some jobs should be done automatically by the Bulldoze() function
	AutoSetGOMAXPROCS   bool
	AutoParseFlags      bool
	AutoCatchInterrupts bool

	// This is the address to access this goji application. It should include the http:// part too.
	SiteUrl string

	// Whenever this application is accessible through a secure HTTPs connection.
	// This flag affects some important security mechanisms, as settings the secure flag on cookies.
	SecureHttpsAccess bool

	SocketType    SocketType
	ListenAddress string
	ServeFiles    bool

	GoPath               string
	WorkingPath          string
	TmpPath              string
	SessionsDatabasePath string

	PublicPath        string
	PagesPath         string
	TemplatesPath     string
	CoreTemplatesPath string
	TranslationPath   string

	BulldozerSourcePath        string
	BulldozerCoreTemplatesPath string
	BulldozerResourcesPath     string
	BulldozerTranslationPath   string

	// The CookieHashKey is required, used to authenticate the cookie value using HMAC.
	// It is recommended to use a key with 32 or 64 bytes.
	CookieHashKey []byte
	// The CookieBlockKey is used to encrypt the cookie value.
	// The length must correspond to the block size of the encryption algorithm.
	// For AES, used by default, valid lengths are 16, 24, or 32 bytes to select AES-128, AES-192, or AES-256.
	CookieBlockKey []byte

	// The maximum session age in seconds
	SessionMaxAge int

	// The maximum allowed requests per minute before the IP is blocked
	FirewallMaxRequestsPerMinute int
	// Release the blocked remote address after x seconds
	FirewallReleaseBlockAfter int

	// This are the static stylesheets and javascripts which
	// will be always loaded.
	// Don't manipulate this slices after Bulldozer initialization!
	StaticJavaScripts []string
	StaticStyleSheets []string
}
