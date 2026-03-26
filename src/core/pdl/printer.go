package pdl

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hsqbyte/protospec/src/core/schema"
)

// PDLPrinter formats a ProtocolSchema back into valid PDL text.
type PDLPrinter struct{}

// Print formats a ProtocolSchema as PDL text that can be re-parsed by PDLParser.
func (p *PDLPrinter) Print(s *schema.ProtocolSchema) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("protocol %s version %q {\n", s.Name, s.Version))
	b.WriteString(fmt.Sprintf("  byte_order %s;\n", s.DefaultByteOrder.String()))

	// Print constants
	if len(s.Constants) > 0 {
		b.WriteByte('\n')
		// Sort constant names for deterministic output
		constNames := make([]string, 0, len(s.Constants))
		for name := range s.Constants {
			constNames = append(constNames, name)
		}
		sort.Strings(constNames)
		for _, name := range constNames {
			b.WriteString(fmt.Sprintf("  const %s = %d;\n", name, s.Constants[name]))
		}
	}

	for i, f := range s.Fields {
		if i == 0 {
			b.WriteByte('\n')
		}
		if f.IsBitfieldGroup {
			p.printBitfield(&b, &f)
		} else {
			p.printField(&b, &f, "  ")
		}
	}

	b.WriteString("}\n")
	return b.String()
}

// printField writes a single field definition at the given indent level.
func (p *PDLPrinter) printField(b *strings.Builder, f *schema.FieldDef, indent string) {
	b.WriteString(fmt.Sprintf("%sfield %s: %s", indent, f.Name, formatType(f)))

	// Collect modifiers
	mods := p.collectModifiers(f)

	if len(mods) == 0 {
		b.WriteString(";\n")
		return
	}

	// Write modifiers on continuation lines
	for _, mod := range mods {
		b.WriteString("\n")
		b.WriteString(indent)
		b.WriteString("  ")
		b.WriteString(mod)
	}
	b.WriteString(";\n")
}

// collectModifiers returns the modifier strings for a field.
func (p *PDLPrinter) collectModifiers(f *schema.FieldDef) []string {
	var mods []string

	if f.Checksum != nil {
		mods = append(mods, formatChecksum(f.Checksum))
	}
	if f.LengthRef != nil {
		mods = append(mods, formatLengthRef(f.LengthRef))
	}
	if len(f.EnumMap) > 0 {
		mods = append(mods, formatEnum(f.EnumMap))
	}
	if f.Condition != nil {
		mods = append(mods, formatWhen(f.Condition))
	}
	if f.DisplayFormat != "" {
		mods = append(mods, fmt.Sprintf("display %s", f.DisplayFormat))
	}
	if f.DefaultValue != nil {
		mods = append(mods, fmt.Sprintf("= %d", f.DefaultValue.(int64)))
	}
	if f.RangeMin != nil && f.RangeMax != nil {
		mods = append(mods, fmt.Sprintf("range [%d..%d]", *f.RangeMin, *f.RangeMax))
	}

	return mods
}

// printBitfield writes a bitfield group.
func (p *PDLPrinter) printBitfield(b *strings.Builder, f *schema.FieldDef) {
	b.WriteString("  bitfield {\n")
	for _, sub := range f.BitfieldFields {
		p.printField(b, &sub, "    ")
	}
	b.WriteString("  }\n")
}

// formatType returns the PDL type string for a field (e.g. "uint16", "bytes", "bool").
func formatType(f *schema.FieldDef) string {
	switch f.Type {
	case schema.Uint:
		return fmt.Sprintf("uint%d", f.BitWidth)
	case schema.Int:
		return fmt.Sprintf("int%d", f.BitWidth)
	case schema.Bytes:
		if f.FixedLength > 0 {
			return fmt.Sprintf("bytes[%d]", f.FixedLength)
		}
		return "bytes"
	case schema.String:
		return "string"
	case schema.Bool:
		return "bool"
	default:
		return "unknown"
	}
}

// formatChecksum returns the PDL checksum modifier string.
func formatChecksum(c *schema.ChecksumConfig) string {
	var fields []string
	for _, cf := range c.CoverFields {
		// Range notation "field1..field2" is stored as-is
		fields = append(fields, cf)
	}
	return fmt.Sprintf("checksum %s covers [%s]", c.Algorithm, strings.Join(fields, ", "))
}

// formatLengthRef returns the PDL length_ref modifier string.
func formatLengthRef(lr *schema.LengthRef) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("length_ref %s", lr.FieldName))
	if lr.Scale != 1 {
		parts = append(parts, fmt.Sprintf("scale %d", lr.Scale))
	}
	if lr.Offset != 0 {
		parts = append(parts, fmt.Sprintf("offset %d", lr.Offset))
	}
	return strings.Join(parts, " ")
}

// formatEnum returns the PDL enum modifier string with sorted keys for deterministic output.
func formatEnum(em map[int]string) string {
	keys := make([]int, 0, len(em))
	for k := range em {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	var entries []string
	for _, k := range keys {
		entries = append(entries, fmt.Sprintf("%d = %q", k, em[k]))
	}
	return fmt.Sprintf("enum { %s }", strings.Join(entries, ", "))
}

// formatWhen returns the PDL when modifier string.
func formatWhen(c *schema.ConditionExpr) string {
	return fmt.Sprintf("when %s %s %v", c.FieldName, c.Operator, c.Value)
}

// PrintMessage formats a MessageSchema back into valid PSL text.
func (p *PDLPrinter) PrintMessage(ms *schema.MessageSchema) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("message %s version %q {\n", ms.Name, ms.Version))

	// If this schema has a transport reference, print it
	if ms.Transport != "" {
		b.WriteString(fmt.Sprintf("  transport %s;\n", ms.Transport))
	}

	// Determine if we have a transport definition for field-level context
	var transportDef *schema.TransportDef
	if ms.TransportDef != nil {
		transportDef = ms.TransportDef
	}

	for _, msg := range ms.Messages {
		b.WriteByte('\n')
		kind := msg.Kind

		if transportDef != nil && ms.Transport == "" {
			// Transport PSL mode: print transport-level message type with field defaults
			p.printTransportMessageType(&b, &msg, transportDef, "  ")
		} else if transportDef != nil && ms.Transport != "" {
			// Upper-layer dynamic mode: print using transport field keywords
			p.printDynamicMessage(&b, &msg, transportDef, "  ")
		} else {
			// Legacy mode: print as before
			if len(msg.Fields) == 0 && msg.Response == nil {
				b.WriteString(fmt.Sprintf("  %s %s;\n", kind, msg.Name))
			} else {
				b.WriteString(fmt.Sprintf("  %s %s {\n", kind, msg.Name))
				for _, f := range msg.Fields {
					p.printMessageField(&b, &f, "    ")
				}
				if msg.Response != nil {
					p.printLegacyResponse(&b, msg.Response, "    ")
				}
				b.WriteString("  }\n")
			}
		}
	}

	b.WriteString("}\n")
	return b.String()
}

// printTransportMessageType prints a message type block in transport PSL format,
// including field default annotations and nested response blocks.
func (p *PDLPrinter) printTransportMessageType(b *strings.Builder, msg *schema.MessageDef, td *schema.TransportDef, indent string) {
	// Find the matching MessageTypeDef from the transport
	var typeDef *schema.MessageTypeDef
	for i := range td.MessageTypes {
		if td.MessageTypes[i].Name == msg.Kind {
			typeDef = &td.MessageTypes[i]
			break
		}
	}

	b.WriteString(fmt.Sprintf("%s%s {\n", indent, msg.Kind))
	innerIndent := indent + "  "

	if typeDef != nil {
		// Print transport fields with default annotations
		for _, tf := range typeDef.Fields {
			p.printTransportField(b, &tf, innerIndent)
		}
	} else {
		// Fallback: print regular message fields
		for _, f := range msg.Fields {
			p.printMessageField(b, &f, innerIndent)
		}
	}

	// Print nested response block if present
	if msg.Response != nil && typeDef != nil && typeDef.ResponseDef != nil {
		b.WriteByte('\n')
		b.WriteString(fmt.Sprintf("%sresponse {\n", innerIndent))
		for _, tf := range typeDef.ResponseDef.Fields {
			p.printTransportField(b, &tf, innerIndent+"  ")
		}
		b.WriteString(fmt.Sprintf("%s}\n", innerIndent))
	}

	b.WriteString(fmt.Sprintf("%s}\n", indent))
}

// printTransportField prints a single transport field definition with default annotations.
func (p *PDLPrinter) printTransportField(b *strings.Builder, tf *schema.TransportFieldDef, indent string) {
	b.WriteString(fmt.Sprintf("%sfield %s: %s", indent, tf.Name, tf.Type.String()))

	// Print nested object fields
	if tf.Type == schema.MsgObject && len(tf.Fields) > 0 {
		b.WriteString(" {\n")
		for _, sub := range tf.Fields {
			p.printTransportField(b, &sub, indent+"  ")
		}
		b.WriteString(indent + "}")
	}

	if tf.Optional {
		b.WriteString(" optional")
	}

	// Print default annotation
	if tf.AutoValue {
		b.WriteString(" default auto")
	} else if tf.DefaultValue != nil {
		switch v := tf.DefaultValue.(type) {
		case string:
			b.WriteString(fmt.Sprintf(" default %q", v))
		default:
			b.WriteString(fmt.Sprintf(" default %v", v))
		}
	}

	b.WriteString(";\n")
}

// printDynamicMessage prints an upper-layer message block using transport field keywords.
func (p *PDLPrinter) printDynamicMessage(b *strings.Builder, msg *schema.MessageDef, td *schema.TransportDef, indent string) {
	kind := msg.Kind

	// Find the matching MessageTypeDef from the transport
	var typeDef *schema.MessageTypeDef
	for i := range td.MessageTypes {
		if td.MessageTypes[i].Name == kind {
			typeDef = &td.MessageTypes[i]
			break
		}
	}

	if len(msg.Fields) == 0 && msg.Response == nil {
		b.WriteString(fmt.Sprintf("%s%s {\n", indent, kind))
		b.WriteString(fmt.Sprintf("%s}\n", indent))
		return
	}

	b.WriteString(fmt.Sprintf("%s%s {\n", indent, kind))
	innerIndent := indent + "  "

	// Print fields using transport field keywords
	for _, f := range msg.Fields {
		if typeDef != nil {
			p.printDynamicField(b, &f, typeDef, innerIndent)
		} else {
			// Fallback to regular field printing
			p.printMessageField(b, &f, innerIndent)
		}
	}

	// Print nested response block
	if msg.Response != nil {
		p.printDynamicResponseBlock(b, msg.Response, typeDef, innerIndent)
	}

	b.WriteString(fmt.Sprintf("%s}\n", indent))
}

// printDynamicField prints a field in upper-layer dynamic format using transport keywords.
func (p *PDLPrinter) printDynamicField(b *strings.Builder, f *schema.MessageFieldDef, typeDef *schema.MessageTypeDef, indent string) {
	// Find the matching transport field definition
	var tf *schema.TransportFieldDef
	for i := range typeDef.Fields {
		if typeDef.Fields[i].Name == f.Name {
			tf = &typeDef.Fields[i]
			break
		}
	}

	if tf != nil {
		// Print using transport field keyword format
		switch tf.Type {
		case schema.MsgString:
			// method "initialize";
			b.WriteString(fmt.Sprintf("%s%s %q;\n", indent, tf.Name, f.Name))
			return
		case schema.MsgNumber:
			// opcode 0x0006;
			b.WriteString(fmt.Sprintf("%s%s %s;\n", indent, tf.Name, f.Name))
			return
		case schema.MsgObject:
			// params { field ...; }
			b.WriteString(fmt.Sprintf("%s%s {\n", indent, tf.Name))
			for _, sub := range f.Fields {
				p.printMessageField(b, &sub, indent+"  ")
			}
			b.WriteString(fmt.Sprintf("%s}\n", indent))
			return
		}
	}

	// Fallback: print as regular message field
	p.printMessageField(b, f, indent)
}

// printDynamicResponseBlock prints a nested response block in upper-layer dynamic format.
func (p *PDLPrinter) printDynamicResponseBlock(b *strings.Builder, resp *schema.MessageDef, parentTypeDef *schema.MessageTypeDef, indent string) {
	b.WriteString(fmt.Sprintf("%sresponse {\n", indent))
	innerIndent := indent + "  "

	var responseDef *schema.MessageTypeDef
	if parentTypeDef != nil {
		responseDef = parentTypeDef.ResponseDef
	}

	for _, f := range resp.Fields {
		if responseDef != nil {
			p.printDynamicField(b, &f, responseDef, innerIndent)
		} else {
			p.printMessageField(b, &f, innerIndent)
		}
	}

	b.WriteString(fmt.Sprintf("%s}\n", indent))
}

// printLegacyResponse prints a nested response block in legacy format.
func (p *PDLPrinter) printLegacyResponse(b *strings.Builder, resp *schema.MessageDef, indent string) {
	b.WriteString(fmt.Sprintf("%sresponse {\n", indent))
	for _, f := range resp.Fields {
		p.printMessageField(b, &f, indent+"  ")
	}
	b.WriteString(fmt.Sprintf("%s}\n", indent))
}

func (p *PDLPrinter) printMessageField(b *strings.Builder, f *schema.MessageFieldDef, indent string) {
	b.WriteString(fmt.Sprintf("%sfield %s: %s", indent, f.Name, f.Type.String()))
	if f.Type == schema.MsgObject && len(f.Fields) > 0 {
		b.WriteString(" {\n")
		for _, sub := range f.Fields {
			p.printMessageField(b, &sub, indent+"  ")
		}
		b.WriteString(indent + "}")
	}
	if f.Optional {
		b.WriteString(" optional")
	}
	b.WriteString(";\n")
}

// PrintTransport formats a TransportDef back into valid transport PSL text.
func (p *PDLPrinter) PrintTransport(td *schema.TransportDef) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("message %s version %q {\n", td.Name, td.Version))

	for i, mt := range td.MessageTypes {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(fmt.Sprintf("  %s {\n", mt.Name))
		for _, tf := range mt.Fields {
			p.printTransportField(&b, &tf, "    ")
		}

		// Print nested response block
		if mt.ResponseDef != nil {
			b.WriteByte('\n')
			b.WriteString("    response {\n")
			for _, tf := range mt.ResponseDef.Fields {
				p.printTransportField(&b, &tf, "      ")
			}
			b.WriteString("    }\n")
		}

		b.WriteString("  }\n")
	}

	b.WriteString("}\n")
	return b.String()
}
