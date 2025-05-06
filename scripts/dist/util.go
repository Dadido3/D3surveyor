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
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// WasmExecJsPath find wasm_exec.js in the local Go distribution and return it's path.
// Return error if not found.
func WasmExecJsPath() (string, error) {

	b, err := exec.Command("go", "env", "GOROOT").CombinedOutput()
	if err != nil {
		return "", err
	}
	bstr := strings.TrimSpace(string(b))
	if bstr == "" {
		return "", fmt.Errorf("failed to find wasm_exec.js, empty path from `go env GOROOT`")
	}

	p := filepath.Join(bstr, "lib/wasm/wasm_exec.js")
	_, err = os.Stat(p)
	if err != nil {
		return "", err
	}

	return p, nil
}

// MustWasmExecJsPath find wasm_exec.js in the local Go distribution and return it's path.
// Panic if not found.
func MustWasmExecJsPath() string {
	s, err := WasmExecJsPath()
	if err != nil {
		panic(err)
	}
	return s
}
