// Package lsp implements a Language Server Protocol server for PSL files.
package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// Server is the PSL language server.
type Server struct {
	reader *bufio.Reader
	writer io.Writer
	docs   map[string]string // uri -> content
}

// NewServer creates a new LSP server reading from stdin and writing to stdout.
func NewServer() *Server {
	return &Server{
		reader: bufio.NewReader(os.Stdin),
		writer: os.Stdout,
		docs:   make(map[string]string),
	}
}

// Run starts the LSP server main loop.
func (s *Server) Run() error {
	for {
		msg, err := s.readMessage()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		s.handleMessage(msg)
	}
}

type jsonrpcMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

func (s *Server) readMessage() (*jsonrpcMessage, error) {
	// Read headers
	contentLength := 0
	for {
		line, err := s.reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		if strings.HasPrefix(line, "Content-Length:") {
			val := strings.TrimSpace(line[len("Content-Length:"):])
			contentLength, _ = strconv.Atoi(val)
		}
	}

	if contentLength == 0 {
		return nil, fmt.Errorf("missing Content-Length")
	}

	body := make([]byte, contentLength)
	if _, err := io.ReadFull(s.reader, body); err != nil {
		return nil, err
	}

	var msg jsonrpcMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

func (s *Server) sendResponse(id any, result any) {
	resp := map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"result":  result,
	}
	data, _ := json.Marshal(resp)
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(data))
	s.writer.Write([]byte(header))
	s.writer.Write(data)
}

func (s *Server) sendNotification(method string, params any) {
	msg := map[string]any{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
	}
	data, _ := json.Marshal(msg)
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(data))
	s.writer.Write([]byte(header))
	s.writer.Write(data)
}

func (s *Server) handleMessage(msg *jsonrpcMessage) {
	switch msg.Method {
	case "initialize":
		s.handleInitialize(msg)
	case "initialized":
		// No-op
	case "shutdown":
		s.sendResponse(msg.ID, nil)
	case "exit":
		os.Exit(0)
	case "textDocument/didOpen":
		s.handleDidOpen(msg)
	case "textDocument/didChange":
		s.handleDidChange(msg)
	case "textDocument/didClose":
		s.handleDidClose(msg)
	case "textDocument/completion":
		s.handleCompletion(msg)
	case "textDocument/hover":
		s.handleHover(msg)
	case "textDocument/formatting":
		s.handleFormatting(msg)
	case "textDocument/definition":
		s.handleDefinition(msg)
	}
}

func (s *Server) handleInitialize(msg *jsonrpcMessage) {
	result := map[string]any{
		"capabilities": map[string]any{
			"textDocumentSync": 1, // Full sync
			"completionProvider": map[string]any{
				"triggerCharacters": []string{":", " "},
			},
			"hoverProvider":              true,
			"documentFormattingProvider": true,
			"definitionProvider":         true,
		},
		"serverInfo": map[string]any{
			"name":    "psl-lsp",
			"version": "0.9.0",
		},
	}
	s.sendResponse(msg.ID, result)
}
