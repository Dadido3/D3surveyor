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

	router.MustAddRouteExact("/rooms",
		vgrouter.RouteHandlerFunc(func(rm *vgrouter.RouteMatch) {
			root.Body = &PageRooms{Site: globalSite}
		}))

	router.MustAddRoute("/room/:key",
		vgrouter.RouteHandlerFunc(func(rm *vgrouter.RouteMatch) {
			keyParams := rm.Params["key"]
			if len(keyParams) < 1 {
				root.Body = &PageNotFound{}
				return
			}
			key := keyParams[0]
			if room, ok := globalSite.Rooms[key]; ok {
				root.Body = room
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
