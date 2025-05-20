// Copyright (C) 2021-2025 David Vogel
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
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/vugu/vugu/distutil"
)

func main() {
	clean := flag.Bool("clean", true, "Remove dist dir before starting")
	dist := flag.String("dist", "dist", "Directory to put distribution files in")
	tagName := flag.String("tagname", "0.0.0-undefined", "Tag name that contains the semantic version")
	urlPathPrefix := flag.String("urlpathprefix", "", "Path prefix for the router. If you serve the app from an URL like example.com/foo, use \"/foo\" as prefix")
	flag.Parse()

	start := time.Now()

	if *clean {
		os.RemoveAll(*dist)
	}

	// Create dist directory tree.
	os.MkdirAll(filepath.Join(*dist, "static"), 0755)

	// Copy static files.
	distutil.MustCopyDirFiltered(filepath.Join(".", "static"), filepath.Join(*dist, "static"), nil)

	// Find and copy wasm_exec.js.
	distutil.MustCopyFile(MustWasmExecJsPath(), filepath.Join(*dist, "wasm_exec.js"))

	// Check for vugugen and go get if not there.
	if _, err := exec.LookPath("vugugen"); err != nil {
		fmt.Print(distutil.MustExec("go", "install", "github.com/vugu/vugu/cmd/vugugen"))
	}

	// Run go generate.
	fmt.Print(distutil.MustExec("go", "generate", "."))

	// Prepare ldflags with version information.
	ldFlags := fmt.Sprintf("-ldflags= -X 'main.versionString=%s' -X 'main.urlPathPrefix=%s'", *tagName, *urlPathPrefix)

	// Run go build for wasm binary and store result in dist directory.
	fmt.Print(distutil.MustEnvExec([]string{"GOOS=js", "GOARCH=wasm"}, "go", "build", ldFlags, "-o", filepath.Join(*dist, "main.wasm"), "."))

	// STATIC INDEX FILE:
	// If you are hosting with a static file server or CDN, you can write out the default index.html from simplehttp.
	req, _ := http.NewRequest("GET", "/index.html", nil)
	indexFile, err := os.OpenFile(filepath.Join(*dist, "index.html"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	distutil.Must(err)
	defer indexFile.Close()
	template.Must(template.New("_page_").Parse(pageTemplateSource)).Execute(indexFile, map[string]interface{}{
		"Request":    req,
		"Title":      "D3surveyor",
		"CSSFiles":   []string{*urlPathPrefix + "/static/css/w3.css", *urlPathPrefix + "/static/font-awesome/css/all.min.css", *urlPathPrefix + "/static/css/styles.css"},
		"MetaTags":   map[string]string{"viewport": "width=device-width, initial-scale=1"},
		"PathPrefix": *urlPathPrefix,
	})

	// Generate 404 redirect page.
	redirectFile, err := os.OpenFile(filepath.Join(*dist, "404.html"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	distutil.Must(err)
	defer redirectFile.Close()
	template.Must(template.New("_page_").Parse(page404TemplateSource)).Execute(redirectFile, map[string]interface{}{
		"PathPrefix": *urlPathPrefix,
	})

	log.Printf("dist.go complete in %v", time.Since(start))
}

var pageTemplateSource = `
<!doctype html>
<html>
	<head>
		<title>{{.Title}}</title>
		<meta charset="utf-8"/>
		{{if .MetaTags}}{{range $k, $v := .MetaTags}}
			<meta name="{{$k}}" content="{{$v}}"/>
		{{end}}{{end}}
		{{if .CSSFiles}}{{range $f := .CSSFiles}}
			<link rel="stylesheet" href="{{$f}}" />
		{{end}}{{end}}
		<script src="https://cdn.jsdelivr.net/npm/text-encoding@0.7.0/lib/encoding.min.js"></script> <!-- MS Edge polyfill -->
		<script src="{{.PathPrefix}}/wasm_exec.js"></script>
	</head>
	<body>
		<div id="vugu_mount_point">
			{{if .ServerRenderedOutput}}{{.ServerRenderedOutput}}{{else}}
				<img style="position: absolute; top: 50%; left: 50%;" src="https://cdnjs.cloudflare.com/ajax/libs/galleriffic/2.0.1/css/loader.gif">
			{{end}}
		</div>
		<script>
			var wasmSupported = (typeof WebAssembly === "object");
			if (wasmSupported) {
				if (!WebAssembly.instantiateStreaming) { // polyfill
					WebAssembly.instantiateStreaming = async (resp, importObject) => {
						const source = await (await resp).arrayBuffer();
						return await WebAssembly.instantiate(source, importObject);
					};
				}
				const go = new Go();
				WebAssembly.instantiateStreaming(fetch("{{.PathPrefix}}/main.wasm"), go.importObject).then((result) => {
					go.run(result.instance);
				});
			} else {
				document.getElementById("vugu_mount_point").innerHTML = 'This application requires WebAssembly support.  Please upgrade your browser.';
			}
		</script>
	</body>
</html>
`

var page404TemplateSource = `
<!doctype HTML>
<html>
	<head>
		<meta charset="utf-8"/>
		<meta http-equiv="refresh" content="0; url={{.PathPrefix}}/">
		<script type="text/javascript">
			window.location.href = "{{.PathPrefix}}/"
		</script>
		<title>Page Redirection</title>
	</head>
	<body>
		If you are not redirected automatically, follow this <a href='{{.PathPrefix}}/'>link back to the index page</a>.
	</body>
</html>
`
