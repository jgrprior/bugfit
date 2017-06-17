// Package bugfit scrapes BuggyFit class location data from www.buggyfit.co.uk, converts it to
// GeoJSON format and exposes that GeoJSON data over HTTP.
package bugfit

import (
	"net/http"
	"regexp"

	"google.golang.org/appengine"
)

type config struct {
	bugfitURL         string         // BuggyFit "find a class" URL
	bugfitCacheKey    string         // Memcache key
	bugfitKeyKind     string         // Datastore key kind
	bugfitKeyStringID string         // Datastore key string ID
	bugfitReScript    *regexp.Regexp // Pattern for grabbing JavaScript object from fetched script
	fetcher           htmlFetcher
}

func init() {
	cnf := &config{
		bugfitURL:         "http://www.buggyfit.co.uk/find-a-class/",
		bugfitCacheKey:    "geo",
		bugfitKeyKind:     "geo",
		bugfitKeyStringID: "bugfit",
		bugfitReScript:    regexp.MustCompile(`var\s+maplistScriptParamsKo\s+=\s+({.+})`),
		fetcher:           urlFetchHTML,
	}

	if appengine.IsDevAppServer() {
		cnf.fetcher = cacheFetchHTML
	}

	http.Handle("/refresh", NewRefreshHandler(cnf))

	http.Handle("/features",
		NewCachedFeatureHandler(cnf,
			WithDatastoreFallback(cnf,
				WithFetchFallback(cnf))))

}
