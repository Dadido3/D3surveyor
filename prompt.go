// Copyright (C) 2025 David Vogel
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

import js "github.com/vugu/vugu/js"

// Prompt opens a js input prompt and modifies bindValue with the given input.
func Prompt(promptText string, bindValue GeneralInputValuer) {
	if bindValue == nil {
		return
	}

	result := js.Global().Call("prompt", promptText, bindValue.InputValue())
	if result.Type() == js.TypeString {
		bindValue.SetInputValue(result.String())
	}
}
