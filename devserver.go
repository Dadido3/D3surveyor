// +build !wasm

package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/vugu/vugu/simplehttp"
)

func main() {

	// A hackish way to insert metadata into the head tag.
	simplehttp.DefaultStaticData["MetaTags"] = map[string]string{"viewport": "width=device-width, initial-scale=1"}

	wd, _ := os.Getwd()
	uiDir := filepath.Join(wd)
	l := ":8875"
	log.Printf("Starting HTTP Server at %q", l)
	h := simplehttp.New(uiDir, true)
	http.Handle("/", h)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(filepath.Join(uiDir, "static")))))

	log.Fatal(http.ListenAndServe(l, nil))
}
