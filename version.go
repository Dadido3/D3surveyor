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

// File for handling versioning.
// This relies on naming git tags by the semantic version scheme, and on the correct forwarding of those tag names into the build process.
// If an invalid version string is supplied, the software will create a runtime error on startup.

package main

import (
	"strings"

	"github.com/coreos/go-semver/semver"
)

// versionString contains the semantic version of the software as a string.
//
// This variable is only used to transfer the correct version information into the build.
// Don't use this variable in the software, use `version` instead.
//
// When building the software, the default `v0.0.0-development` is used.
// To compile the program with the correct version information, compile the following way:
//
// `go build -ldflags="-X 'main.versionString=x.y.z'"`, where `x.y.z` is the correct and valid version from the git tag.
// This variable may or may not contain the prefix v.
var versionString = "0.0.0-development"

// version of the program.
//
// When converted into a string, this will not contain the v prefix.
var version = semver.Must(semver.NewVersion(strings.TrimPrefix(versionString, "v")))
