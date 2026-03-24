<p align="center">
  <img src="assets/psl.svg" width="80" alt="PSL">
</p>

<h1 align="center">PSL — 协议规格语言</h1>

<p align="center">
  用文本定义任意二进制协议，自动完成编解码。
</p>

<p align="center">
  <a href="README.md">English</a> | <a href="README_ZH.md">中文</a>
</p>

---

PSL 是一个基于 Go 的通用二进制协议编解码引擎。你只需在 `.psl` 文本文件中描述协议结构 — 字段名、位宽、字节序、校验算法、条件字段 — 引擎自动完成编解码，无需为每个协议编写代码。

内置协议：IPv4、TCP、UDP、ICMP、ARP、DNS、HTTP/1.1、WebSocket。你可以用同样的方式定义自己的协议。

```bash
go build -o psl .
```

## 开源协议

[GPL-3.0](LICENSE)
