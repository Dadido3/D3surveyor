// Copyright (c) 2021 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//go:build ignore

package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/vugu/vugu/simplehttp"
)

func main() {

	// A hackish way to insert metadata and other stuff into the head tag.
	simplehttp.DefaultStaticData["MetaTags"] = map[string]string{"viewport": "width=device-width, initial-scale=1"}
	simplehttp.DefaultStaticData["Title"] = "D3surveyor dev"
	simplehttp.DefaultStaticData["CSSFiles"] = []string{"/static/css/w3.css", "/static/font-awesome/css/all.min.css"}

	wd, _ := os.Getwd()
	uiDir := filepath.Join(wd, "..")
	l := ":8875"
	log.Printf("Starting HTTP Server at %q", l)
	h := simplehttp.New(uiDir, true)
	http.Handle("/", h)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(filepath.Join(uiDir, "static")))))

	log.Fatal(http.ListenAndServe(l, nil))
}
