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

	wd, _ := os.Getwd()
	uiDir := filepath.Join(wd)
	l := ":8875"
	log.Printf("Starting HTTP Server at %q", l)
	h := simplehttp.New(uiDir, true)
	log.Fatal(http.ListenAndServe(l, h))
}
