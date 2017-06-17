package bugfit

import (
	"log"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

type geoFeatures struct {
	Type     string     `json:"type"`
	Features []*feature `json:"features"`
}

func newGeoFeatures(objs *BFObjects) *geoFeatures {
	geo := &geoFeatures{
		Type:     "FeatureCollection",
		Features: make([]*feature, 0),
	}

	for _, l := range objs.Objs[0].Locations {
		geo.Features = append(geo.Features, newFeature(l))
	}

	return geo
}

// feature represents a GeoJSON feature.
type feature struct {
	Geometry   *geometry         `json:"geometry"`
	Type       string            `json:"type"`
	Properties map[string]string `json:"properties"`
}

// addProperty cleans the property value and adds it to the feature's Properties map.
func (f *feature) addProperty(k, v string) {
	f.Properties[k] = htmlText(v)
}

// Geometry represents a GeoJSON feature geometry.
type geometry struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

// NewFeature creates a new Feature from a BuggyFit location.
func newFeature(l *BFLocation) *feature {
	f := &feature{
		Geometry:   &geometry{"Point", make([]float64, 0)},
		Type:       "Feature",
		Properties: make(map[string]string),
	}

	f.addProperty("locationUrl", l.URL)
	f.addProperty("title", l.Title)
	f.addProperty("address", l.Address)

	long, err := strconv.ParseFloat(l.Longitude, 64)
	if err != nil {
		log.Fatal(err)
	}

	lat, err := strconv.ParseFloat(l.Latitude, 64)
	if err != nil {
		log.Fatal(err)
	}

	f.Geometry.Coordinates = append(f.Geometry.Coordinates, long, lat)

	return f
}

// htmlText strips HTML markup from a string and returns plain text.
func htmlText(s string) string {
	text := make([]string, 0)
	to := html.NewTokenizer(strings.NewReader(s))

	for {
		nt := to.Next()

		switch nt {
		case html.ErrorToken:
			return strings.Join(text, " ")
		case html.TextToken:
			t := to.Token()
			text = append(text, t.Data)
		}
	}
}
