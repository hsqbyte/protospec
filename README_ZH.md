<p align="center">
  <img src="assets/psl.svg" width="80" alt="PSL">
</p>

<h1 align="center">PSL — 协议规范语言</h1>

<p align="center">
  用文本定义任意二进制协议，自动完成编解码。
</p>

<p align="center">
  <a href="README.md">English</a> | <a href="README_ZH.md">中文</a>
</p>

---

PSL 是一个基于 Go 的通用二进制协议编解码引擎。你只需在 `.psl` 文本文件中描述协议结构——字段名、位宽、字节序、校验算法、条件字段——引擎自动完成编解码，无需为每个协议编写代码。

内置协议：IPv4、TCP、UDP、ICMP、ARP、DNS、HTTP/1.1、WebSocket、MQTT、TLS 等 40+ 协议。

```bash
go build -o psl .
```

## Go 集成

通过 [pkg.go.dev](https://pkg.go.dev/github.com/hsqbyte/protospec) 安装：

```bash
go get github.com/hsqbyte/protospec@v1.0.0
```

```go
package main

import (
    "fmt"
    "github.com/hsqbyte/protospec/src/protocol"
)

func main() {
    lib, _ := protocol.NewLibrary()

    // 列出所有协议
    for _, name := range lib.AllNames() {
        fmt.Println(name)
    }

    // 编码
    data, _ := lib.Encode("UDP", map[string]any{
        "source_port": 12345, "destination_port": 53,
        "length": 20, "checksum": 0,
    })
    fmt.Printf("hex: %x\n", data)

    // 解码
    result, _ := lib.Decode("UDP", data)
    fmt.Println(result.Fields)
}
```

## Web / 浏览器集成（WASM）

PSL 也提供 WebAssembly 模块，发布在 [npm](https://www.npmjs.com/package/protospec-wasm)，可以直接在浏览器中运行完整的协议引擎——无需后端。

```bash
npm install protospec-wasm
```

将 `node_modules/protospec-wasm/` 中的 `protospec.wasm` 和 `wasm_exec.js` 复制到项目的 public 目录，然后：

```js
// 1. 加载 WASM 引擎
const go = new Go();  // 来自 wasm_exec.js
const resp = await fetch("/protospec.wasm");
const result = await WebAssembly.instantiateStreaming(resp, go.importObject);
go.run(result.instance);

// 2. 列出所有协议（返回 JSON 字符串）
const protocols = JSON.parse(psl_listProtocols());

// 3. 编码
const encoded = JSON.parse(psl_encode("UDP", JSON.stringify({
  source_port: 12345, destination_port: 53, length: 20, checksum: 0
})));
console.log(encoded.hex);

// 4. 解码
const decoded = JSON.parse(psl_decode("UDP", encoded.hex));
console.log(decoded);
```

### WASM 可用函数

| 函数 | 参数 | 返回值 |
|---|---|---|
| `psl_listProtocols()` | — | 所有协议及元数据的 JSON 数组 |
| `psl_getProtocol(name)` | 协议名 | 协议 schema JSON |
| `psl_getMeta(name)` | 协议名 | 元数据 JSON（RFC、描述、层级等） |
| `psl_encode(name, json)` | 协议名, JSON 字段 | `{"hex": "..."}` |
| `psl_decode(name, hex)` | 协议名, 十六进制字符串 | 解码后的字段 JSON |

### 从源码构建 WASM

```bash
task wasm          # 编译 → npm/protospec.wasm
task npm:publish   # 编译 + 发布到 npm
```

## Web UI

[PSL UI](https://github.com/hsqbyte/psl_ui) 提供了基于 Web 的协议浏览、可视化和交互界面——由上述 WASM 引擎驱动。

<p align="center">
  <img src="assets/psl_ui_screenshot.png" alt="PSL UI">
</p>

## 项目结构

```
cmd/wasm/          WASM 入口
npm/               npm 包（protospec-wasm）
psl/               内置协议定义（.psl + meta.json）
src/
├── core/          编解码引擎、解析器、schema、校验和、格式化
├── cli/           命令行
├── protocol/      协议库 API
├── i18n/          国际化
├── codegen/       多语言代码生成
├── lsp/           语言服务器
├── tools/         Lint、格式化、覆盖率、测试生成、调试等
├── integrations/  eBPF、DPDK、网关、加密、合规等
├── platform/      云平台、插件、SDK、CI、迁移等
└── docs/          文档生成
```

## 开源协议

[GPL-3.0](LICENSE)
