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

package main

import (
	"github.com/vugu/vugu"
)

type OptimizerComponent struct {
	AttrMap vugu.AttrMap

	OptimizerState   *OptimizerState
	fontAwesomeClass string
}

func (c *OptimizerComponent) Compute(ctx vugu.ComputeCtx) {
	if c.OptimizerState.Running() {
		// Step through hourglass icons to show optimizer is running.
		switch c.fontAwesomeClass {
		case "fas fa-hourglass-start":
			c.fontAwesomeClass = "fas fa-hourglass-half"
		case "fas fa-hourglass-half":
			c.fontAwesomeClass = "fas fa-hourglass-end"
		default:
			c.fontAwesomeClass = "fas fa-hourglass-start"
		}
	} else {
		c.fontAwesomeClass = "fas fa-sync"
	}
}

func (c *OptimizerComponent) handleClick(event vugu.DOMEvent) {
	if c.OptimizerState.Running() { // Side not: This is not perfectly race condition free, but it doesn't matter here.
		c.OptimizerState.Stop()
	} else {
		c.OptimizerState.Start(event)
	}
}
