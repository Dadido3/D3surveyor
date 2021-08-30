//go:build ignore

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
	"github.com/vugu/vugu/simplehttp"
)

func main() {
	clean := flag.Bool("clean", true, "Remove dist dir before starting")
	dist := flag.String("dist", "dist", "Directory to put distribution files in")
	tagName := flag.String("tagname", "0.0.0-undefined", "Tag name that contains the semantic version")
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
	distutil.MustCopyFile(distutil.MustWasmExecJsPath(), filepath.Join(*dist, "wasm_exec.js"))

	// Check for vugugen and go get if not there.
	if _, err := exec.LookPath("vugugen"); err != nil {
		fmt.Print(distutil.MustExec("go", "install", "github.com/vugu/vugu/cmd/vugugen"))
	}

	// Run go generate.
	fmt.Print(distutil.MustExec("go", "generate", "."))

	// Prepare ldflags with version information.
	ldFlags := fmt.Sprintf("-ldflags=\"-X 'main.versionString=%s'\"", *tagName)

	// Run go build for wasm binary and store result in dist directory.
	fmt.Print(distutil.MustEnvExec([]string{"GOOS=js", "GOARCH=wasm"}, "go", "build", ldFlags, "-o", filepath.Join(*dist, "main.wasm"), "."))

	// STATIC INDEX FILE:
	// If you are hosting with a static file server or CDN, you can write out the default index.html from simplehttp.
	req, _ := http.NewRequest("GET", "/index.html", nil)
	indexFile, err := os.OpenFile(filepath.Join(*dist, "index.html"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	distutil.Must(err)
	defer indexFile.Close()
	template.Must(template.New("_page_").Parse(simplehttp.DefaultPageTemplateSource)).Execute(indexFile, map[string]interface{}{
		"Request":  req,
		"Title":    "D3surveyor",
		"MetaTags": map[string]string{"viewport": "width=device-width, initial-scale=1"},
	})

	log.Printf("dist.go complete in %v", time.Since(start))
}
