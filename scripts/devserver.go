// Copyright (C) 2021 David Vogel
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

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
