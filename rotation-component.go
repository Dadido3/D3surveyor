// Copyright (c) 2021 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import "github.com/vugu/vugu"

type RotationComponent struct {
	AttrMap vugu.AttrMap

	BindValue *Rotation

	Editable bool
}
