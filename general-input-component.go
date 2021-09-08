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

type GeneralInputValuer interface {
	InputParse(string)
	InputString() string
}

// GeneralInputComponent is a generalized input component that takes any value that implement the GeneralInputValuer interface.
// TODO: Replace most input components with the general input component
type GeneralInputComponent struct {
	BindValue  GeneralInputValuer
	BindLocked *bool
	LabelText  string
	InputType  string // Input type of the HTML input element. Defaults to "text"

	AttrMap vugu.AttrMap
}

func (c *GeneralInputComponent) Init(ctx vugu.InitCtx) {
	if c.InputType == "" {
		c.InputType = "text"
	}
}

func (c *GeneralInputComponent) handleValueChange(event vugu.DOMEvent) {
	strVal := event.PropString("target", "value")

	c.BindValue.InputParse(strVal)
}

func (c *GeneralInputComponent) inputContent() string {
	if c.BindValue != nil {
		return c.BindValue.InputString()
	}

	return ""
}

func (c *GeneralInputComponent) handleLockedChange(event vugu.DOMEvent) {
	*c.BindLocked = event.PropBool("target", "checked")
}
