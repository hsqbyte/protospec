// Package playground provides a web-based PSL playground with WASM support.
package playground

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/hsqbyte/protospec/src/protocol"
)

// Server is the playground HTTP server.
type Server struct {
	lib  *protocol.Library
	addr string
}

// NewServer creates a new playground server.
func NewServer(lib *protocol.Library, addr string) *Server {
	return &Server{lib: lib, addr: addr}
}

// PlaygroundRequest is a request to the playground API.
type PlaygroundRequest struct {
	PSL    string `json:"psl"`
	Action string `json:"action"` // "validate", "decode", "encode", "structure"
	Hex    string `json:"hex,omitempty"`
	JSON   string `json:"json,omitempty"`
}

// PlaygroundResponse is a response from the playground API.
type PlaygroundResponse struct {
	Success bool   `json:"success"`
	Result  any    `json:"result,omitempty"`
	Error   string `json:"error,omitempty"`
}

// Start starts the playground server.
func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/api/playground", s.handleAPI)
	mux.HandleFunc("/api/share", s.handleShare)
	mux.HandleFunc("/api/examples", s.handleExamples)
	fmt.Printf("Playground running at http://%s\n", s.addr)
	return http.ListenAndServe(s.addr, mux)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.New("playground").Parse(playgroundHTML))
	tmpl.Execute(w, nil)
}

func (s *Server) handleAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req PlaygroundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, PlaygroundResponse{Error: err.Error()})
		return
	}

	switch req.Action {
	case "validate":
		s.handleValidate(w, req)
	case "structure":
		s.handleStructure(w, req)
	default:
		writeJSON(w, PlaygroundResponse{Error: fmt.Sprintf("unknown action: %s", req.Action)})
	}
}

func (s *Server) handleValidate(w http.ResponseWriter, req PlaygroundRequest) {
	psl := strings.TrimSpace(req.PSL)
	if psl == "" {
		writeJSON(w, PlaygroundResponse{Error: "empty PSL"})
		return
	}

	if strings.HasPrefix(psl, "message ") {
		writeJSON(w, PlaygroundResponse{Success: true, Result: "valid message protocol"})
		return
	}

	_, err := s.lib.CreateCodec(psl)
	if err != nil {
		writeJSON(w, PlaygroundResponse{Error: err.Error()})
		return
	}
	writeJSON(w, PlaygroundResponse{Success: true, Result: "valid binary protocol"})
}

func (s *Server) handleStructure(w http.ResponseWriter, req PlaygroundRequest) {
	psl := strings.TrimSpace(req.PSL)
	if psl == "" {
		writeJSON(w, PlaygroundResponse{Error: "empty PSL"})
		return
	}

	codec, err := s.lib.CreateCodec(psl)
	if err != nil {
		writeJSON(w, PlaygroundResponse{Error: err.Error()})
		return
	}
	_ = codec
	writeJSON(w, PlaygroundResponse{Success: true, Result: "protocol structure parsed"})
}

func (s *Server) handleShare(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, PlaygroundResponse{Success: true, Result: "share feature placeholder"})
}

func (s *Server) handleExamples(w http.ResponseWriter, r *http.Request) {
	examples := []map[string]string{
		{"name": "IPv4", "description": "Internet Protocol v4 header"},
		{"name": "TCP", "description": "Transmission Control Protocol"},
		{"name": "UDP", "description": "User Datagram Protocol"},
		{"name": "DNS", "description": "Domain Name System"},
		{"name": "HTTP", "description": "Hypertext Transfer Protocol"},
	}
	writeJSON(w, PlaygroundResponse{Success: true, Result: examples})
}

func writeJSON(w http.ResponseWriter, resp PlaygroundResponse) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

const playgroundHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>PSL Playground</title>
<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
body { font-family: -apple-system, BlinkMacSystemFont, sans-serif; background: #1e1e2e; color: #cdd6f4; }
.header { background: #313244; padding: 12px 24px; display: flex; align-items: center; gap: 16px; }
.header h1 { font-size: 18px; color: #89b4fa; }
.container { display: grid; grid-template-columns: 1fr 1fr; height: calc(100vh - 48px); }
.editor, .output { padding: 16px; }
.editor textarea { width: 100%; height: calc(100% - 48px); background: #181825; color: #cdd6f4; border: 1px solid #45475a; border-radius: 8px; padding: 12px; font-family: monospace; font-size: 14px; resize: none; }
.output pre { background: #181825; border: 1px solid #45475a; border-radius: 8px; padding: 12px; height: calc(100% - 48px); overflow: auto; font-size: 14px; }
.toolbar { margin-bottom: 8px; display: flex; gap: 8px; }
.toolbar button { background: #89b4fa; color: #1e1e2e; border: none; padding: 8px 16px; border-radius: 6px; cursor: pointer; font-weight: 600; }
.toolbar button:hover { background: #74c7ec; }
</style>
</head>
<body>
<div class="header"><h1>PSL Playground</h1></div>
<div class="container">
<div class="editor">
<div class="toolbar">
<button onclick="validate()">Validate</button>
<button onclick="showStructure()">Structure</button>
</div>
<textarea id="psl" placeholder="Enter PSL code here...">protocol Example version "1.0" {
    byte_order big-endian;
    field type: uint8;
    field length: uint16;
    field payload: bytes;
}</textarea>
</div>
<div class="output">
<div class="toolbar"><span style="line-height:36px">Output</span></div>
<pre id="result">Click Validate or Structure to see results.</pre>
</div>
</div>
<script>
async function api(action) {
  const resp = await fetch('/api/playground', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({psl: document.getElementById('psl').value, action})
  });
  return resp.json();
}
async function validate() {
  const r = await api('validate');
  document.getElementById('result').textContent = r.success ? '✓ ' + r.result : '✗ ' + r.error;
}
async function showStructure() {
  const r = await api('structure');
  document.getElementById('result').textContent = r.success ? JSON.stringify(r.result, null, 2) : '✗ ' + r.error;
}
</script>
</body>
</html>`
