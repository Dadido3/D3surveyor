package main

import (
	"github.com/vugu/vgrouter"
	"github.com/vugu/vugu"
)

// OVERALL APPLICATION WIRING IN vuguSetup
func vuguSetup(buildEnv *vugu.BuildEnv, eventEnv vugu.EventEnv) vugu.Builder {

	// CREATE A NEW ROUTER INSTANCE
	router := vgrouter.New(eventEnv)

	// MAKE OUR WIRE FUNCTION POPULATE ANYTHING THAT WANTS A "NAVIGATOR".
	buildEnv.SetWireFunc(func(b vugu.Builder) {
		if c, ok := b.(vgrouter.NavigatorSetter); ok {
			c.NavigatorSet(router)
		}
	})

	// CREATE THE ROOT COMPONENT
	root := &Root{}
	buildEnv.WireComponent(root) // WIRE IT

	// Add routes.
	router.MustAddRouteExact("/",
		vgrouter.RouteHandlerFunc(func(rm *vgrouter.RouteMatch) {
			root.Body = globalSite
		}))

	router.MustAddRouteExact("/rangefinders",
		vgrouter.RouteHandlerFunc(func(rm *vgrouter.RouteMatch) {
			root.Body = &PageRangefinders{Site: globalSite}
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
		}))

	router.MustAddRouteExact("/points",
		vgrouter.RouteHandlerFunc(func(rm *vgrouter.RouteMatch) {
			root.Body = &PagePoints{Site: globalSite}
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
		}))

	router.SetNotFound(vgrouter.RouteHandlerFunc(
		func(rm *vgrouter.RouteMatch) {
			root.Body = &PageNotFound{}
		}))

	// TELL THE ROUTER TO LISTEN FOR THE BROWSER CHANGING URLS
	err := router.ListenForPopState()
	if err != nil {
		panic(err)
	}

	// GRAB THE CURRENT BROWSER URL AND PROCESS IT AS A ROUTE
	err = router.Pull()
	if err != nil {
		panic(err)
	}

	return root
}
