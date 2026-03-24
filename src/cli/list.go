package cli

import (
	"fmt"
	"strings"
)

// layerOrder defines the display order for protocol layers.
var layerOrder = []string{"link", "network", "transport", "application"}

var layerLabels = map[string]map[string]string{
	"link":        {"en": "Link Layer", "zh": "链路层"},
	"network":     {"en": "Network Layer", "zh": "网络层"},
	"transport":   {"en": "Transport Layer", "zh": "传输层"},
	"application": {"en": "Application Layer", "zh": "应用层"},
	"":            {"en": "Other", "zh": "其他"},
}

func runList(ctx *Context) error {
	names := ctx.Lib.AllNames()

	if len(names) == 0 {
		fmt.Println(msg(ctx, "list.empty"))
		return nil
	}

	fmt.Printf(msg(ctx, "list.header")+"\n\n", len(names))

	// Group by layer
	groups := make(map[string][]string)
	for _, name := range names {
		layer := ""
		if meta := ctx.Lib.Meta(name); meta != nil {
			layer = meta.Layer
		}
		groups[layer] = append(groups[layer], name)
	}

	tw := termWidth()

	// Column widths
	const colName = 14
	const colRFC = 10

	hName := msg(ctx, "list.col.name")
	hRFC := msg(ctx, "list.col.rfc")
	hDesc := msg(ctx, "list.col.desc")

	for _, layer := range layerOrder {
		protos, ok := groups[layer]
		if !ok || len(protos) == 0 {
			continue
		}

		// Layer header
		label := metaLang(layerLabels[layer], ctx.Lang)
		fmt.Printf("  %s%s%s\n", cBold+cCyan, label, cReset)

		fmt.Printf("  %s%s  %s  %s%s\n",
			cBold,
			padRight(hName, colName), padRight(hRFC, colRFC), hDesc,
			cReset)

		lineWidth := min(tw, 80)
		fmt.Printf("  %s%s%s\n", cDim, repeat("─", lineWidth-2), cReset)

		for _, name := range protos {
			printListRow(ctx, name, colName, colRFC, tw)
		}
		fmt.Println()
	}

	// Print any protocols without a layer
	if protos, ok := groups[""]; ok && len(protos) > 0 {
		label := metaLang(layerLabels[""], ctx.Lang)
		fmt.Printf("  %s%s%s\n", cBold+cCyan, label, cReset)
		for _, name := range protos {
			printListRow(ctx, name, colName, colRFC, tw)
		}
		fmt.Println()
	}

	return nil
}

func printListRow(ctx *Context, name string, colName, colRFC, tw int) {
	meta := ctx.Lib.Meta(name)
	rfc := ""
	desc := ""
	deps := ""
	if meta != nil {
		if meta.RFC != "" {
			rfc = "RFC " + meta.RFC
		}
		desc = metaLang(meta.Title, ctx.Lang)
		if len(meta.DependsOn) > 0 {
			deps = " " + cDim + "← " + strings.Join(meta.DependsOn, ", ") + cReset
		}
	}

	descWidth := max(tw-2-colName-2-colRFC-2, 10)
	desc = truncateToWidth(desc, descWidth)

	fmt.Printf("  %s%s%s  %s%s%s  %s%s\n",
		cGreen, padRight(name, colName), cReset,
		cYellow, padRight(rfc, colRFC), cReset,
		desc, deps)
}

func repeat(s string, n int) string {
	if n <= 0 {
		return ""
	}
	result := ""
	for range n {
		result += s
	}
	return result
}
