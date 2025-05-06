package main

import (
	"bytes"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type WASMExecHandler struct {
	wasmExecJsOnce    sync.Once
	wasmExecJsContent []byte
	wasmExecJsTs      time.Time
}

func (h *WASMExecHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b, err := exec.Command("go", "env", "GOROOT").CombinedOutput()
	if err != nil {
		http.Error(w, "failed to run `go env GOROOT`: "+err.Error(), 500)
		return
	}

	h.wasmExecJsOnce.Do(func() {
		h.wasmExecJsContent, err = os.ReadFile(filepath.Join(strings.TrimSpace(string(b)), "lib/wasm/wasm_exec.js"))
		if err != nil {
			http.Error(w, "failed to run `go env GOROOT`: "+err.Error(), 500)
			return
		}
		h.wasmExecJsTs = time.Now() // hack but whatever for now
	})

	if len(h.wasmExecJsContent) == 0 {
		http.Error(w, "failed to read wasm_exec.js from local Go environment", 500)
		return
	}

	w.Header().Set("Content-Type", "text/javascript")
	http.ServeContent(w, r, "/wasm_exec.js", h.wasmExecJsTs, bytes.NewReader(h.wasmExecJsContent))
}
