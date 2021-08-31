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

//go:build wasm

package main

import (
	"github.com/vugu/vgrouter"
	"github.com/vugu/vugu"
)

// urlPathPrefix contains the path prefix for the router.
//
// If the application's URL is `example.com/D3surveyor/`, use `go build -ldflags="-X 'main.urlPathPrefix=/D3surveyor'"`.
// This is already done by `dist.go`, and it's not needed when this app is run via the dev server.
var urlPathPrefix = ""

// vuguSetup sets up overall wiring and routing.
func vuguSetup(buildEnv *vugu.BuildEnv, eventEnv vugu.EventEnv) vugu.Builder {

	// Create new router instance.
	router := vgrouter.New(eventEnv)

	// Set prefix if available.
	router.SetPathPrefix(urlPathPrefix)

	// Create root object.
	root := &Root{}

	buildEnv.SetWireFunc(func(b vugu.Builder) {
		// MAKE OUR WIRE FUNCTION POPULATE ANYTHING THAT WANTS A "NAVIGATOR".
		if c, ok := b.(vgrouter.NavigatorSetter); ok {
			c.NavigatorSet(router)
		}
		if c, ok := b.(*TitleBar); ok {
			c.root = root
		}
	})

	// Wire the root component. (Not sure if that is really needed)
	buildEnv.WireComponent(root)

	// Add routes.
	router.MustAddRouteExact("/",
		vgrouter.RouteHandlerFunc(func(rm *vgrouter.RouteMatch) {
			root.Body = globalSite
			root.sidebarDisplay = "none"
		}))

	router.MustAddRouteExact("/points",
		vgrouter.RouteHandlerFunc(func(rm *vgrouter.RouteMatch) {
			root.Body = &PagePoints{Site: globalSite}
			root.sidebarDisplay = "none"
		}))

	router.MustAddRoute("/point/:key",
		vgrouter.RouteHandlerFunc(func(rm *vgrouter.RouteMatch) {
			keyParams := rm.Params["key"]
			if len(keyParams) < 1 {
				root.Body = &PageNotFound{}
				return
			}
			key := keyParams[0]
			if point, ok := globalSite.Points[key]; ok {
				root.Body = point
			} else {
				root.Body = &PageNonExistant{}
			}
			root.sidebarDisplay = "none"
		}))

	router.MustAddRouteExact("/cameras",
		vgrouter.RouteHandlerFunc(func(rm *vgrouter.RouteMatch) {
			root.Body = &PageCameras{Site: globalSite}
			root.sidebarDisplay = "none"
		}))

	router.MustAddRoute("/camera/:key",
		vgrouter.RouteHandlerFunc(func(rm *vgrouter.RouteMatch) {
			keyParams := rm.Params["key"]
			if len(keyParams) < 1 {
				root.Body = &PageNotFound{}
				return
			}
			key := keyParams[0]
			if camera, ok := globalSite.Cameras[key]; ok {
				root.Body = camera
			} else {
				root.Body = &PageNonExistant{}
			}
			root.sidebarDisplay = "none"
		}))

	router.MustAddRoute("/camera/:key1/photo/:key2",
		vgrouter.RouteHandlerFunc(func(rm *vgrouter.RouteMatch) {

			key1Params, key2Params := rm.Params["key1"], rm.Params["key2"]
			if len(key1Params) < 1 || len(key2Params) < 1 {
				root.Body = &PageNotFound{}
				return
			}
			key1, key2 := key1Params[0], key2Params[0]
			if camera, ok := globalSite.Cameras[key1]; ok {
				if photo, ok := camera.Photos[key2]; ok {
					root.Body = photo
				} else {
					root.Body = &PageNonExistant{}
				}
			} else {
				root.Body = &PageNonExistant{}
			}
			root.sidebarDisplay = "none"
		}))

	router.MustAddRouteExact("/rangefinders",
		vgrouter.RouteHandlerFunc(func(rm *vgrouter.RouteMatch) {
			root.Body = &PageRangefinders{Site: globalSite}
			root.sidebarDisplay = "none"
		}))

	router.MustAddRoute("/rangefinder/:key",
		vgrouter.RouteHandlerFunc(func(rm *vgrouter.RouteMatch) {
			keyParams := rm.Params["key"]
			if len(keyParams) < 1 {
				root.Body = &PageNotFound{}
				return
			}
			key := keyParams[0]
			if rangefinder, ok := globalSite.Rangefinders[key]; ok {
				root.Body = rangefinder
			} else {
				root.Body = &PageNonExistant{}
			}
			root.sidebarDisplay = "none"
		}))

	router.MustAddRoute("/rangefinder/:key1/measurement/:key2",
		vgrouter.RouteHandlerFunc(func(rm *vgrouter.RouteMatch) {

			key1Params, key2Params := rm.Params["key1"], rm.Params["key2"]
			if len(key1Params) < 1 || len(key2Params) < 1 {
				root.Body = &PageNotFound{}
				return
			}
			key1, key2 := key1Params[0], key2Params[0]
			if rangefinder, ok := globalSite.Rangefinders[key1]; ok {
				if rangefinderMeasurement, ok := rangefinder.Measurements[key2]; ok {
					root.Body = rangefinderMeasurement
				} else {
					root.Body = &PageNonExistant{}
				}
			} else {
				root.Body = &PageNonExistant{}
			}
			root.sidebarDisplay = "none"
		}))

	router.MustAddRouteExact("/tripods",
		vgrouter.RouteHandlerFunc(func(rm *vgrouter.RouteMatch) {
			root.Body = &PageTripods{Site: globalSite}
			root.sidebarDisplay = "none"
		}))

	router.MustAddRoute("/tripod/:key",
		vgrouter.RouteHandlerFunc(func(rm *vgrouter.RouteMatch) {
			keyParams := rm.Params["key"]
			if len(keyParams) < 1 {
				root.Body = &PageNotFound{}
				return
			}
			key := keyParams[0]
			if tripod, ok := globalSite.Tripods[key]; ok {
				root.Body = tripod
			} else {
				root.Body = &PageNonExistant{}
			}
			root.sidebarDisplay = "none"
		}))

	router.MustAddRoute("/tripod/:key1/measurement/:key2",
		vgrouter.RouteHandlerFunc(func(rm *vgrouter.RouteMatch) {

			key1Params, key2Params := rm.Params["key1"], rm.Params["key2"]
			if len(key1Params) < 1 || len(key2Params) < 1 {
				root.Body = &PageNotFound{}
				return
			}
			key1, key2 := key1Params[0], key2Params[0]
			if tripod, ok := globalSite.Tripods[key1]; ok {
				if tripodMeasurement, ok := tripod.Measurements[key2]; ok {
					root.Body = tripodMeasurement
				} else {
					root.Body = &PageNonExistant{}
				}
			} else {
				root.Body = &PageNonExistant{}
			}
			root.sidebarDisplay = "none"
		}))

	router.SetNotFound(vgrouter.RouteHandlerFunc(
		func(rm *vgrouter.RouteMatch) {
			root.Body = &PageNotFound{}
			root.sidebarDisplay = "none"
		}))

	// Tell the router to listen to the browser changing URLs.
	err := router.ListenForPopState()
	if err != nil {
		panic(err)
	}

	// Grab the current browser URL and process it as a route.
	err = router.Pull()
	if err != nil {
		panic(err)
	}

	return root
}
