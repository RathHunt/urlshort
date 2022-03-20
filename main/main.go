package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/boltdb/bolt"
	"github.com/gophercises/urlshort"
)

func main() {
	mux := defaultMux()

	// Build the MapHandler using the mux as the fallback
	pathsToUrls := map[string]string{
		"/urlshort-godoc": "https://godoc.org/github.com/gophercises/urlshort",
		"/yaml-godoc":     "https://godoc.org/gopkg.in/yaml.v2",
	}
	mapHandler := urlshort.MapHandler(pathsToUrls, mux)

	yamlfile := flag.String("yamlfile", "", "path to load yml file from")
	jsonfile := flag.String("jsonfile", "", "path to load json file from")
	dbfile := flag.String("dbfile", "", "path to load boltDB database file from")
	flag.Parse()

	// Build the YAMLHandler using the mapHandler as the
	// fallback

	var Handler http.Handler

	if *yamlfile != "" {
		yaml, err := os.ReadFile(*yamlfile)
		if err != nil {
			panic(err)
		}
		Handler, err = urlshort.YAMLHandler(yaml, mapHandler)
		if err != nil {
			panic(err)
		}
	}

	if *jsonfile != "" {
		json, err := os.ReadFile(*jsonfile)
		if err != nil {
			panic(err)
		}
		Handler, err = urlshort.JSONHandler(json, mapHandler)
		if err != nil {
			panic(err)
		}
	}

	if *dbfile != "" {
		db, err := bolt.Open(*dbfile, 0600, nil)
		if err != nil {
			panic(err)
		}

		err = db.Update(func(tx *bolt.Tx) error {
			b, err := tx.CreateBucket([]byte("shorts"))

			if err == bolt.ErrBucketExists {
				return nil
			}

			if err != nil {
				return err
			}

			err = b.Put([]byte("/urlshort"), []byte("https://github.com/gophercises/urlshort"))

			if err != nil {
				return err
			}

			err = b.Put([]byte("/urlshort-final"), []byte("https://github.com/gophercises/urlshort/tree/solution"))

			return err
		})

		if err != nil {
			panic(err)
		}

		Handler = urlshort.DBHandler(db, mux)
	}

	fmt.Println("Starting the server on :8080")
	http.ListenAndServe(":8080", Handler)
}

func defaultMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", hello)
	return mux
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, world!")
}
