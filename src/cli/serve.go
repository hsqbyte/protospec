package cli

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func runServe(ctx *Context, args []string) error {
	port := "8080"
	for i := 0; i < len(args); i++ {
		if args[i] == "-p" && i+1 < len(args) {
			port = args[i+1]
			i++
		}
	}

	mux := http.NewServeMux()

	// API endpoints
	mux.HandleFunc("/api/protocols", func(w http.ResponseWriter, r *http.Request) {
		names := ctx.Lib.AllNames()
		type protoInfo struct {
			Name  string `json:"name"`
			Type  string `json:"type"`
			Layer string `json:"layer,omitempty"`
			Title string `json:"title,omitempty"`
		}
		var list []protoInfo
		for _, n := range names {
			info := protoInfo{Name: n, Type: "binary"}
			if ctx.Lib.Message(n) != nil {
				info.Type = "message"
			}
			if meta := ctx.Lib.Meta(n); meta != nil {
				info.Layer = meta.Layer
				if t, ok := meta.Title["en"]; ok {
					info.Title = t
				}
			}
			list = append(list, info)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(list)
	})

	mux.HandleFunc("/api/protocol/", func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/api/protocol/")
		w.Header().Set("Content-Type", "application/json")

		if ms := ctx.Lib.Message(name); ms != nil {
			json.NewEncoder(w).Encode(ms)
			return
		}
		s, err := ctx.Lib.Registry().GetSchema(name)
		if err != nil {
			http.Error(w, err.Error(), 404)
			return
		}
		json.NewEncoder(w).Encode(s)
	})

	mux.HandleFunc("/api/decode", func(w http.ResponseWriter, r *http.Request) {
		proto := r.URL.Query().Get("protocol")
		hexData := r.URL.Query().Get("hex")
		if proto == "" || hexData == "" {
			http.Error(w, "need protocol and hex params", 400)
			return
		}
		data, err := hex.DecodeString(hexData)
		if err != nil {
			http.Error(w, "invalid hex", 400)
			return
		}
		result, err := ctx.Lib.Decode(proto, data)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result.Packet)
	})

	// Serve embedded HTML UI
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, webUI)
	})

	fmt.Printf("PSL Web UI: http://localhost:%s\n", port)
	return http.ListenAndServe(":"+port, mux)
}

const webUI = `<!DOCTYPE html>
<html><head>
<title>PSL Protocol Browser</title>
<style>
* { box-sizing: border-box; margin: 0; padding: 0; }
body { font-family: -apple-system, sans-serif; background: #f5f5f5; }
.header { background: #1a73e8; color: white; padding: 16px 24px; }
.header h1 { font-size: 20px; }
.container { max-width: 1200px; margin: 0 auto; padding: 20px; }
.search { width: 100%; padding: 10px; border: 1px solid #ddd; border-radius: 6px; font-size: 16px; margin-bottom: 20px; }
.grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 16px; }
.card { background: white; border-radius: 8px; padding: 16px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); cursor: pointer; transition: box-shadow 0.2s; }
.card:hover { box-shadow: 0 4px 12px rgba(0,0,0,0.15); }
.card h3 { color: #1a73e8; margin-bottom: 4px; }
.card .type { font-size: 12px; color: #666; background: #f0f0f0; padding: 2px 8px; border-radius: 10px; display: inline-block; }
.card .layer { font-size: 12px; color: #999; margin-top: 4px; }
.decode-box { background: white; border-radius: 8px; padding: 20px; margin-top: 20px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
.decode-box input, .decode-box select { padding: 8px; margin: 4px; border: 1px solid #ddd; border-radius: 4px; }
.decode-box button { padding: 8px 16px; background: #1a73e8; color: white; border: none; border-radius: 4px; cursor: pointer; }
pre { background: #f8f8f8; padding: 12px; border-radius: 4px; overflow-x: auto; margin-top: 10px; }
</style>
</head><body>
<div class="header"><h1>PSL Protocol Browser</h1></div>
<div class="container">
<input class="search" placeholder="Search protocols..." oninput="filterProtocols(this.value)">
<div class="grid" id="grid"></div>
<div class="decode-box">
<h3>Decode</h3>
<select id="proto"></select>
<input id="hexInput" placeholder="Hex bytes..." size="40">
<button onclick="decode()">Decode</button>
<pre id="result"></pre>
</div>
</div>
<script>
let protocols = [];
fetch('/api/protocols').then(r=>r.json()).then(data=>{
  protocols=data;
  renderGrid(data);
  const sel=document.getElementById('proto');
  data.forEach(p=>{const o=document.createElement('option');o.value=p.name;o.text=p.name;sel.add(o);});
});
function renderGrid(list){
  document.getElementById('grid').innerHTML=list.map(p=>
    '<div class="card"><h3>'+p.name+'</h3><span class="type">'+p.type+'</span>'
    +(p.title?'<div class="layer">'+p.title+'</div>':'')
    +(p.layer?'<div class="layer">Layer: '+p.layer+'</div>':'')+'</div>'
  ).join('');
}
function filterProtocols(q){
  const f=protocols.filter(p=>p.name.toLowerCase().includes(q.toLowerCase())||
    (p.title&&p.title.toLowerCase().includes(q.toLowerCase())));
  renderGrid(f);
}
function decode(){
  const p=document.getElementById('proto').value;
  const h=document.getElementById('hexInput').value.replace(/\s/g,'');
  fetch('/api/decode?protocol='+p+'&hex='+h).then(r=>r.json()).then(d=>{
    document.getElementById('result').textContent=JSON.stringify(d,null,2);
  }).catch(e=>document.getElementById('result').textContent='Error: '+e);
}
</script>
</body></html>`
