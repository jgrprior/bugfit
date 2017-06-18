package bugfit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"

	"golang.org/x/net/context"
	"golang.org/x/net/html"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/urlfetch"
)

// BFObjects represents an array of map parameters from BF's map pluging script.
type BFObjects struct {
	Objs []*BFObject `json:"KOObject"`
}

// BFObject represents a single map's parameters from BF's map plugin script.
type BFObject struct {
	ID        int           `json:"id"`
	Locations []*BFLocation `json:"locations"`
}

// BFLocation represents a single map marker and some additional properties.
type BFLocation struct {
	URL       string `json:"locationUrl"`
	Title     string `json:"title"`
	Address   string `json:"address"`
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
}

type htmlFetcher func(context.Context, *config) (io.ReadCloser, error)

func urlFetchHTML(ctx context.Context, cnf *config) (io.ReadCloser, error) {
	cli := urlfetch.Client(ctx)
	resp, err := cli.Get(cnf.bugfitURL)
	if err != nil {
		return nil, fmt.Errorf("faild to fetch buggyfit class page: %s", err)
	}
	return resp.Body, nil
}

// cacheFetchHTML always reads from the datastore if bugfitCache exists. Otherwise it calls
// urlFetchHTML and writes the cache entity.
func cacheFetchHTML(ctx context.Context, cnf *config) (io.ReadCloser, error) {
	cache := cacheData{}
	k := datastore.NewKey(ctx, "html", "bugfitCache", 0, nil)

	err := datastore.Get(ctx, k, &cache)
	switch err {
	case datastore.ErrNoSuchEntity:
		body, err := urlFetchHTML(ctx, cnf)
		defer body.Close()

		if err != nil {
			return nil, err
		}

		var buf bytes.Buffer
		buf.ReadFrom(body)
		if _, err := datastore.Put(ctx, k, &cacheData{buf.Bytes()}); err != nil {
			return nil, err
		}
		return &mockBody{bytes.NewBuffer(buf.Bytes())}, nil
	case nil:
		return &mockBody{bytes.NewBuffer(cache.Bytes)}, nil
	default:
		return nil, fmt.Errorf("failed to read buggyfit HTML cache: %s", err)
	}
}

func parseHTML(body io.Reader, re *regexp.Regexp) (string, error) {
	doc, err := html.Parse(body)
	if err != nil {
		return "", fmt.Errorf("failed to parse buggyfit HTML: %s", err)
	}

	scripts := make([]string, 0)
	findScripts(doc, &scripts)

	for _, script := range scripts {
		match := re.FindStringSubmatch(script)
		if len(match) > 0 {
			return match[1], nil
		}
	}

	return "", fmt.Errorf("failed to find location markers from %d scripts", len(scripts))
}

func findScripts(n *html.Node, scripts *[]string) {
	if n.Type == html.ElementNode && n.Data == "script" {
		if n.FirstChild != nil {
			*scripts = append(*scripts, n.FirstChild.Data)
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		findScripts(c, scripts)
	}
}

func unmarshalScript(s string) (*BFObjects, error) {
	objs := BFObjects{}
	if err := json.Unmarshal([]byte(s), &objs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal script data: %s", err)
	}
	return &objs, nil
}

// mockBody makes data read from a file look like a io.ReadCloser.
type mockBody struct {
	io.Reader
}

func (mb *mockBody) Close() error { return nil }

// cacheData wraps our HTML blob when we put it in the datastore.
type cacheData struct {
	Bytes []byte `datastore:",noindex"`
}

// TODO: Add timestamp to datastore entity
