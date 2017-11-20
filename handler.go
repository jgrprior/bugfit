package bugfit

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/memcache"
)

// NewRefreshHandler creates a new HTTP handler for handling data refresh requests.
func NewRefreshHandler(cnf *config) http.Handler {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := appengine.NewContext(r)
		_, err := makeGeoJSON(c, cnf)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
	})
	return handler
}

// featureData wraps our JSON blob when we put it in the datastore.
type featureData struct {
	Bytes []byte `datastore:",noindex"`
}

// NewCachedFeatureHandler responds with features from the cache if it exists, otherwise falls back
// to the datastore
func NewCachedFeatureHandler(cnf *config, fallback http.Handler) http.Handler {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := appengine.NewContext(r)
		itm, err := memcache.Get(c, cnf.bugfitCacheKey)

		switch err {
		case nil:
			w.Header().Set("Content-Type", "aplication/json")
			w.Write(itm.Value)
		case memcache.ErrCacheMiss:
			log.Infof(c, "cache miss for BuggyFit geo data, falling back to datastore: %s", err)
			fallback.ServeHTTP(w, r)
		default:
			log.Errorf(c, "unexpected memcache error, falling back to datastore: %s", err)
			fallback.ServeHTTP(w, r)
		}
	})
	return handler
}

// WithDatastoreFallback caches features from the datastore and responds with those features.
func WithDatastoreFallback(cnf *config, fallback http.Handler) http.Handler {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := appengine.NewContext(r)
		k := datastore.NewKey(c, cnf.bugfitKeyKind, cnf.bugfitKeyStringID, 0, nil)

		data := featureData{}
		err := datastore.Get(c, k, &data)

		switch err {
		case nil:
			// Write data to memcache
			itm := &memcache.Item{
				Key:        cnf.bugfitCacheKey,
				Value:      data.Bytes,
				Expiration: time.Duration(time.Hour),
			}
			if err := memcache.Set(c, itm); err != nil {
				log.Errorf(c, "failed to write to cache: %s", err)
			}

			w.Header().Set("Content-Type", "aplication/json")
			w.Write(data.Bytes)
		case datastore.ErrNoSuchEntity:
			log.Errorf(c, "no data in store, falling back to live fetch: %s", err)
			fallback.ServeHTTP(w, r)
		default:
			log.Errorf(c, "unknown error reading from store, falling back to live fetch: %s", err)
			fallback.ServeHTTP(w, r)
		}
	})
	return handler
}

// WithFetchFallback fetches regenerates GeoJSON data and saves it to the datastore.
func WithFetchFallback(cnf *config) http.Handler {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := appengine.NewContext(r)
		data, err := makeGeoJSON(c, cnf)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.Header().Set("Content-Type", "aplication/json")
		w.Write(data)
	})
	return handler
}

func makeGeoJSON(c context.Context, cnf *config) ([]byte, error) {
	body, err := cnf.fetcher(c, cnf)
	if err != nil {
		return nil, fmt.Errorf("location data refresh error: %s", err)
	}

	defer body.Close()
	script, err := parseHTML(body, cnf.bugfitReScript)
	if err != nil {
		return nil, fmt.Errorf("location data refresh error: %s", err)
	}

	objs, err := unmarshalScript(script)
	if err != nil {
		return nil, fmt.Errorf("location data refresh error: %s", err)
	}

	features := newGeoFeatures(objs)
	b, err := json.Marshal(features)
	if err != nil {
		return nil, fmt.Errorf("location data refresh error: %s", err)
	}

	k := datastore.NewKey(c, cnf.bugfitKeyKind, cnf.bugfitKeyStringID, 0, nil)
	if _, err := datastore.Put(c, k, &featureData{b}); err != nil {
		return nil, fmt.Errorf("location data refresh error: %s", err)
	}

	return b, nil
}
