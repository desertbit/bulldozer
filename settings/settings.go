/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package settings

import (
	"code.desertbit.com/bulldozer/bulldozer/log"
	"code.desertbit.com/bulldozer/bulldozer/utils"
	"fmt"
	"github.com/BurntSushi/toml"
	"os"
	"reflect"
	"strings"
	"time"
)

const (
	/*
	 *  Public
	 */

	TemplateExtension = ".bt"
	ScssSuffix        = ".scss"

	DefaultSettingsFileName = "settings.toml"

	// If this prefix is set to string values, then
	// the value is obtained from the environment variables.
	ParseEnvVarPrefix = "ENV:"

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

	// Default password key
	defaultPasswordEncryptionKey = "gNlmWx0jurl8ohIVZMi8k9eRZxP25kEeQq68TeTVLD9omFZmP7sSqLK"
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

		DatabaseAddr:    "localhost",
		DatabasePort:    "28015",
		DatabaseName:    "test",
		DatabaseMaxIdle: 50,
		DatabaseMaxOpen: 50,
		DatabaseTimeout: time.Minute,

		CookieHashKey:  defaultCookieHashKey,
		CookieBlockKey: defaultCookieBlockKey,
		SessionMaxAge:  60 * 60 * 24 * 14, // 14 Days

		FirewallMaxRequestsPerMinute: 100,
		FirewallReleaseBlockAfter:    60 * 5, // 5 minutes

		ScssCmd: "scss",

		RegistrationDisabled:           true,
		PasswordEncryptionKey:          defaultPasswordEncryptionKey,
		RemoveNotConfirmedUsersTimeout: 60 * 60 * 24 * 14, // 14 Days
	}

	// Set the temporary directory path
	Settings.TmpPath = utils.AddTrailingSlashToPath(utils.AddTrailingSlashToPath(os.TempDir()) + tmpDirName)

	// Get the current working directory path
	var err error
	Settings.WorkingPath, err = os.Getwd()
	if err != nil {
		log.L.Fatalf("failed to obtain current work directory path: %v", err)
	}

	// Set the GOPATH
	Settings.GoPath = os.Getenv("GOPATH")
	if len(Settings.GoPath) == 0 {
		log.L.Fatalf("GOPATH is not set!")
	}

	// Append a trailing slash if not already present
	Settings.GoPath = utils.AddTrailingSlashToPath(Settings.GoPath)
	Settings.WorkingPath = utils.AddTrailingSlashToPath(Settings.WorkingPath)

	// Set the paths
	Settings.SessionsDatabasePath = Settings.TmpPath + sessionsDatabaseName

	Settings.PublicPath = Settings.WorkingPath + "public"
	Settings.TemplatesPath = Settings.WorkingPath + "templates"
	Settings.PagesPath = Settings.TemplatesPath + "/pages"
	Settings.TranslationPath = Settings.WorkingPath + "translations"
	Settings.DataPath = Settings.WorkingPath + "data"
	Settings.ScssPath = Settings.DataPath + "/scss"
	Settings.CssPath = Settings.PublicPath + "/css"

	Settings.ScssArgs = []string{
		"--unix-newlines",
		"--no-cache",
		"--sourcemap=none",
		"-t",
		"compressed",
		"--update",
		Settings.ScssPath + ":" + Settings.CssPath,
	}

	Settings.BulldozerSourcePath = Settings.GoPath + bulldozerGoPath
	Settings.BulldozerTemplatesPath = Settings.BulldozerSourcePath + "/data/templates"
	Settings.BulldozerCoreTemplatesPath = Settings.BulldozerTemplatesPath + "/core"
	Settings.BulldozerResourcesPath = Settings.BulldozerSourcePath + "/data/resources"
	Settings.BulldozerTranslationPath = Settings.BulldozerSourcePath + "/data/translations"
	Settings.BulldozerPrototypesPath = Settings.BulldozerSourcePath + "/data/prototypes"

}

//##############//
//### Public ###//
//##############//

// Prepare checks if the settings are correct and valid and initializes some values.
func Prepare() error {
	// Get environment variable values if the environment prefix is set on struct field strings.
	s := reflect.ValueOf(&Settings).Elem()
	for x := 0; x < s.NumField(); x++ {
		f := s.Field(x)

		if !f.CanSet() || f.Kind() != reflect.String {
			continue
		}

		// Get the struct field string value.
		v := f.String()

		// Skip if no environment prefix is set.
		if !strings.HasPrefix(v, ParseEnvVarPrefix) {
			continue
		}

		// Remove the prefix.
		v = strings.TrimPrefix(v, ParseEnvVarPrefix)

		// Get the value from the environment variable.
		envV := os.Getenv(v)
		if len(envV) == 0 {
			log.L.Warning("settings environment variable '%s' is not set!", v)
		}

		// Set the new value.
		f.SetString(envV)
	}

	// Check if the Site url is valid
	if !strings.HasPrefix(Settings.SiteUrl, "http://") &&
		!strings.HasPrefix(Settings.SiteUrl, "https://") {
		return fmt.Errorf("settings: site url is invalid: missing 'http://' or 'https://': '%s'", Settings.SiteUrl)
	}

	// Check if the length of the cookie keys are valid
	l := len(Settings.CookieHashKeyBytes())
	if l != 32 && l != 64 {
		return fmt.Errorf("settings: the cookie hash key has an invalid length of %v bytes! Valid lengths are 32 or 64 bytes...", l)
	}
	l = len(Settings.CookieBlockKeyBytes())
	if l != 16 && l != 24 && l != 32 {
		return fmt.Errorf("settings: the cookie block key has an invalid length of %v bytes! For AES, used by default, valid lengths are 16, 24, or 32 bytes to select AES-128, AES-192, or AES-256.", l)
	}

	// Print a warning if the default cookie keys are set
	if Settings.CookieHashKey == defaultCookieHashKey {
		log.L.Warning("[WARNING] settings: the default cookie hash key is set! You should replace this with a secret key!")
	}
	if Settings.CookieBlockKey == defaultCookieBlockKey {
		log.L.Warning("[WARNING] settings: the default cookie block key is set! You should replace this with a secret key!")
	}

	// Print a warning if the SecureHttpsAccess flag is false
	if !Settings.SecureHttpsAccess {
		log.L.Warning("[WARNING] settings: the secure https access flag is false! You should provide a secure https access!")
	}

	// Print a warning if the default password encryption key is used.
	if Settings.PasswordEncryptionKey == defaultPasswordEncryptionKey {
		log.L.Warning("[WARNING] settings: the default password encryption key is set! You should replace this with a secret key!")
	}

	// Warn the user about possible forgotten root slashes.
	for _, url := range Settings.StaticJavaScripts {
		if !strings.HasPrefix(url, "/") {
			log.L.Warning("static javascript url does not start with a slash: '%s'", url)
		}
	}
	for _, url := range Settings.StaticStyleSheets {
		if !strings.HasPrefix(url, "/") {
			log.L.Warning("static style sheet url does not start with a slash: '%s'", url)
		}
	}

	// Remove leading / from the site url
	Settings.SiteUrl = strings.TrimSuffix(Settings.SiteUrl, "/")

	return nil
}

// Load loads the settings file.
func Load(path string) error {
	log.L.Info("Loading settings from file: '%s'", path)

	exists, err := utils.Exists(path)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("settings file does not exists: '%s'", path)
	}

	// Parse the configuration
	if _, err = toml.DecodeFile(path, &Settings); err != nil {
		return fmt.Errorf("failed to parse settings file '%s': %v", path, err)
	}

	return nil
}

//#######################//
//### Settings struct ###//
//#######################//

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

	DatabaseAddr string
	DatabasePort string
	DatabaseName string

	// Maximum number of idle connections in the pool.
	DatabaseMaxIdle int

	// Maximum number of connections allocated by the pool at a given time.
	// When zero, there is no limit on the number of connections in the pool.
	DatabaseMaxOpen int

	// Close connections after remaining idle for this duration. If the value
	// is zero, then idle connections are not closed. Applications should set
	// the timeout to a value less than the server's timeout.
	DatabaseTimeout time.Duration

	GoPath               string
	WorkingPath          string
	TmpPath              string
	SessionsDatabasePath string

	PublicPath      string
	PagesPath       string
	TemplatesPath   string
	TranslationPath string
	DataPath        string

	ScssPath string
	CssPath  string

	ScssCmd  string
	ScssArgs []string

	BulldozerSourcePath        string
	BulldozerTemplatesPath     string
	BulldozerCoreTemplatesPath string
	BulldozerResourcesPath     string
	BulldozerTranslationPath   string
	BulldozerPrototypesPath    string

	// The CookieHashKey is required, used to authenticate the cookie value using HMAC.
	// It is recommended to use a key with 32 or 64 bytes.
	CookieHashKey string
	// The CookieBlockKey is used to encrypt the cookie value.
	// The length must correspond to the block size of the encryption algorithm.
	// For AES, used by default, valid lengths are 16, 24, or 32 bytes to select AES-128, AES-192, or AES-256.
	CookieBlockKey string

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

	DisallowedRobotsUrls []string

	// Authentication stuff
	RegistrationDisabled           bool
	PasswordEncryptionKey          string
	RemoveNotConfirmedUsersTimeout int
}

func (s *settings) CookieHashKeyBytes() []byte {
	return []byte(s.CookieHashKey)
}

func (s *settings) CookieBlockKeyBytes() []byte {
	return []byte(s.CookieBlockKey)
}

func (s *settings) AddDisallowedRobotsUrls(urls ...string) {
	s.DisallowedRobotsUrls = append(s.DisallowedRobotsUrls, urls...)
}
