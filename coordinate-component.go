package main

import "github.com/vugu/vugu"

type CoordinateComponent struct {
	AttrMap vugu.AttrMap

	BindValue *Coordinate

	Editable bool
}
