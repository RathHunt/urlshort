package urlshort

import (
	"encoding/json"
	"github.com/boltdb/bolt"
	"gopkg.in/yaml.v2"
	"net/http"
)

type shorts struct {
	Path string
	URL  string
}

// MapHandler will return an http.HandlerFunc (which also
// implements http.Handler) that will attempt to map any
// paths (keys in the map) to their corresponding URL (values
// that each key in the map points to, in string format).
// If the path is not provided in the map, then the fallback
// http.Handler will be called instead.
func MapHandler(pathsToUrls map[string]string, fallback http.Handler) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url := pathsToUrls[r.URL.Path]

		if url != "" {
			http.Redirect(w, r, url, http.StatusPermanentRedirect)
		} else {
			fallback.ServeHTTP(w, r)
		}
	})
}

func MakeMap(data []shorts) map[string]string {
	pathMap := make(map[string]string)

	for _, item := range data {
		pathMap[item.Path] = item.URL
	}
	return pathMap
}

// YAMLHandler will parse the provided YAML and then return
// an http.HandlerFunc (which also implements http.Handler)
// that will attempt to map any paths to their corresponding
// URL. If the path is not provided in the YAML, then the
// fallback http.Handler will be called instead.
//
// YAML is expected to be in the format:
//
//     - path: /some-path
//       url: https://www.some-url.com/demo
//
// The only errors that can be returned all related to having
// invalid YAML data.
//
// See MapHandler to create a similar http.HandlerFunc via
// a mapping of paths to urls.
func YAMLHandler(yml []byte, fallback http.Handler) (http.HandlerFunc, error) {
	var parsedYaml []shorts
	err := yaml.Unmarshal(yml, &parsedYaml)

	if err != nil {
		return nil, err
	}

	pathMap := MakeMap(parsedYaml)

	return MapHandler(pathMap, fallback), nil
}

func JSONHandler(jsondata []byte, fallback http.Handler) (http.HandlerFunc, error) {
	var parsedJSON []shorts

	err := json.Unmarshal(jsondata, &parsedJSON)
	if err != nil {
		return nil, err
	}

	pathMap := MakeMap(parsedJSON)

	return MapHandler(pathMap, fallback), nil
}

func DBHandler(db *bolt.DB, fallback http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var url string
		err := db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("shorts"))
			data := b.Get([]byte(r.URL.Path))
			if data != nil {
				url = string(data)
			}
			return nil
		})

		if err == nil && url != "" {
			http.Redirect(w, r, url, http.StatusTemporaryRedirect)
		} else {
			fallback.ServeHTTP(w, r)
		}
	})
}
