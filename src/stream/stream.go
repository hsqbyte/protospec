// Package stream provides streaming protocol decoding from io.Reader.
package stream

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/hsqbyte/protospec/src/protocol"
)

// FrameType defines how message boundaries are detected.
type FrameType int

const (
	LengthPrefixed FrameType = iota // Length field precedes payload
	Delimited                       // Delimiter separates messages
	FixedLength                     // All messages have the same length
)

// FrameConfig configures message framing.
type FrameConfig struct {
	Type         FrameType
	LengthBytes  int    // For LengthPrefixed: 1, 2, or 4
	LengthOffset int    // Offset to add to length value
	BigEndian    bool   // Byte order for length field
	Delimiter    []byte // For Delimited framing
	FixedSize    int    // For FixedLength framing
}

// StreamDecoder reads and decodes protocol messages from a stream.
type StreamDecoder struct {
	reader   io.Reader
	lib      *protocol.Library
	protocol string
	frame    FrameConfig
}

// NewStreamDecoder creates a new stream decoder.
func NewStreamDecoder(r io.Reader, lib *protocol.Library, proto string, frame FrameConfig) *StreamDecoder {
	return &StreamDecoder{
		reader:   r,
		lib:      lib,
		protocol: proto,
		frame:    frame,
	}
}

// Message holds a decoded message from the stream.
type Message struct {
	Raw    []byte
	Fields map[string]any
	Err    error
}

// Next reads and decodes the next message from the stream.
func (sd *StreamDecoder) Next() (*Message, error) {
	data, err := sd.readFrame()
	if err != nil {
		return nil, err
	}

	result, err := sd.lib.Decode(sd.protocol, data)
	if err != nil {
		return &Message{Raw: data, Err: err}, nil
	}

	return &Message{Raw: data, Fields: result.Packet}, nil
}

func (sd *StreamDecoder) readFrame() ([]byte, error) {
	switch sd.frame.Type {
	case LengthPrefixed:
		return sd.readLengthPrefixed()
	case Delimited:
		return sd.readDelimited()
	case FixedLength:
		buf := make([]byte, sd.frame.FixedSize)
		if _, err := io.ReadFull(sd.reader, buf); err != nil {
			return nil, err
		}
		return buf, nil
	default:
		return nil, fmt.Errorf("unknown frame type")
	}
}

func (sd *StreamDecoder) readLengthPrefixed() ([]byte, error) {
	lenBuf := make([]byte, sd.frame.LengthBytes)
	if _, err := io.ReadFull(sd.reader, lenBuf); err != nil {
		return nil, err
	}

	var length int
	switch sd.frame.LengthBytes {
	case 1:
		length = int(lenBuf[0])
	case 2:
		if sd.frame.BigEndian {
			length = int(binary.BigEndian.Uint16(lenBuf))
		} else {
			length = int(binary.LittleEndian.Uint16(lenBuf))
		}
	case 4:
		if sd.frame.BigEndian {
			length = int(binary.BigEndian.Uint32(lenBuf))
		} else {
			length = int(binary.LittleEndian.Uint32(lenBuf))
		}
	}

	length += sd.frame.LengthOffset
	if length <= 0 {
		return lenBuf, nil
	}

	payload := make([]byte, length)
	if _, err := io.ReadFull(sd.reader, payload); err != nil {
		return nil, err
	}

	// Return length prefix + payload
	return append(lenBuf, payload...), nil
}

func (sd *StreamDecoder) readDelimited() ([]byte, error) {
	var buf []byte
	single := make([]byte, 1)
	delimLen := len(sd.frame.Delimiter)

	for {
		if _, err := io.ReadFull(sd.reader, single); err != nil {
			if len(buf) > 0 {
				return buf, nil
			}
			return nil, err
		}
		buf = append(buf, single[0])

		if len(buf) >= delimLen {
			tail := buf[len(buf)-delimLen:]
			match := true
			for i := range tail {
				if tail[i] != sd.frame.Delimiter[i] {
					match = false
					break
				}
			}
			if match {
				return buf[:len(buf)-delimLen], nil
			}
		}
	}
}

// StateMachine tracks protocol state transitions.
type StateMachine struct {
	Current     string
	Transitions map[string]map[string]string // from -> trigger -> to
}

// NewStateMachine creates a state machine with an initial state.
func NewStateMachine(initial string) *StateMachine {
	return &StateMachine{
		Current:     initial,
		Transitions: make(map[string]map[string]string),
	}
}

// AddTransition adds a state transition rule.
func (sm *StateMachine) AddTransition(from, trigger, to string) {
	if sm.Transitions[from] == nil {
		sm.Transitions[from] = make(map[string]string)
	}
	sm.Transitions[from][trigger] = to
}

// Trigger attempts a state transition and returns the new state.
func (sm *StateMachine) Trigger(event string) (string, error) {
	trans, ok := sm.Transitions[sm.Current]
	if !ok {
		return sm.Current, fmt.Errorf("no transitions from state %q", sm.Current)
	}
	next, ok := trans[event]
	if !ok {
		return sm.Current, fmt.Errorf("invalid transition %q from state %q", event, sm.Current)
	}
	sm.Current = next
	return next, nil
}
