/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package webcrawler

var (
	sitemapHooks []HookFunc
)

//#############//
//### Types ###//
//#############//

type HookFunc func() []string

//##############//
//### Public ###//
//##############//

// AddSitemapHook adds a hook function which is called
// during each sitemap render. Return a slice of additional
// sitemap urls. The path should not contain the site URL.
// This is prepended automatically.
func AddSitemapHook(f HookFunc) {
	sitemapHooks = append(sitemapHooks, f)
}

//###############//
//### Private ###//
//###############//

func callSitemapHooks() (result []string) {
	for _, f := range sitemapHooks {
		result = append(result, f()...)
	}
	return
}
