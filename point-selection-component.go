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

type PointSelectionComponent struct {
	Site *Site

	BindValue *string

	AttrMap vugu.AttrMap

	options PointSelectionComponentOptions
}

func (c *PointSelectionComponent) Init(ctx vugu.InitCtx) {
	// Only load the list at creation of the component.
	// Updating it would cause the dropdown to show the wrong option.

	options := PointSelectionComponentOptions{
		keys:    []string{""},
		mapping: map[string]string{"": "-"},
	}

	// Generate options.
	for _, point := range c.Site.PointsSorted() {
		key := point.Key()
		options.keys = append(options.keys, key)
		options.mapping[key] = point.Name
	}

	c.options = options
}

// PointSelectionComponentOptions contains a list of sorted options.
type PointSelectionComponentOptions struct {
	keys    []string
	mapping map[string]string
}

// KeyList implements vgform.KeyLister.
func (o PointSelectionComponentOptions) KeyList() []string { return o.keys }

// TextMap implements vgform.TextMapper.
func (o PointSelectionComponentOptions) TextMap(key string) string { return o.mapping[key] }
