package graph

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/italiatroller-1990/garterscopic/internal/components"
	"github.com/italiatroller-1990/garterscopic/internal/layouts"
)

// Node describes the resolved structure of a single page.
type Node struct {
	Route    string   `json:"route"`
	Source   string   `json:"source,omitempty"`
	Layout   string   `json:"layout,omitempty"`
	Children []Branch `json:"children,omitempty"`
}

// Branch is one entry under a page node (layout component or stylesheet).
type Branch struct {
	Label    string   `json:"label"`
	Children []Branch `json:"children,omitempty"`
}

// SiteGraph is the resolved dependency/component structure of the whole site.
type SiteGraph struct {
	Nodes []Node `json:"pages"`
}

// Build assembles the site graph from loaded layouts and component
// definitions. Page entries come as (route, sourcePath, layoutName) tuples
// so the graph package stays decoupled from the builder.
func Build(entries []Entry, comps map[string]components.Definition, lookupLayout func(string) (layouts.Layout, bool)) *SiteGraph {
	g := &SiteGraph{Nodes: make([]Node, 0, len(entries))}

	for _, e := range entries {
		node := Node{
			Route:  e.Route,
			Source: e.Source,
			Layout: e.Layout,
		}

		if e.Layout != "" && lookupLayout != nil {
			layoutBranch := Branch{Label: "layout: " + e.Layout}

			if layo, ok := lookupLayout(e.Layout); ok {
				for _, inst := range layo.Components {
					compBranch := Branch{Label: inst.Name}

					if def, ok := comps[inst.Name]; ok && def.Style != "" {
						compBranch.Children = append(compBranch.Children, Branch{Label: "style: " + def.Style})
					}

					layoutBranch.Children = append(layoutBranch.Children, compBranch)
				}
			}

			node.Children = append(node.Children, layoutBranch)
		}

		g.Nodes = append(g.Nodes, node)
	}

	return g
}

// Entry is one page to include in the graph.
type Entry struct {
	Route  string
	Source string
	Layout string
}

// RenderText produces a readable tree for terminals.
func (g *SiteGraph) RenderText() string {
	var sb strings.Builder

	for i, node := range g.Nodes {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(node.Route)
		if node.Source != "" {
			fmt.Fprintf(&sb, "    (%s)", node.Source)
		}
		sb.WriteString("\n")

		renderBranch(&sb, node.Children, "")
	}

	return sb.String()
}

func renderBranch(sb *strings.Builder, branches []Branch, prefix string) {
	for i, b := range branches {
		connector := "├── "
		childPrefix := prefix + "│   "
		if i == len(branches)-1 {
			connector = "└── "
			childPrefix = prefix + "    "
		}

		sb.WriteString(prefix + connector + b.Label + "\n")
		if len(b.Children) > 0 {
			renderBranch(sb, b.Children, childPrefix)
		}
	}
}

// RenderJSON produces machine-readable output for --json.
func (g *SiteGraph) RenderJSON() (string, error) {
	data, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// SortNodes orders nodes by route for deterministic output.
func (g *SiteGraph) SortNodes() {
	sort.Slice(g.Nodes, func(i, j int) bool {
		return g.Nodes[i].Route < g.Nodes[j].Route
	})
}
