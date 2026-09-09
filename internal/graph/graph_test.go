package graph

import (
	"strings"
	"testing"

	"github.com/italiatroller-1990/garterscopic/internal/components"
	"github.com/italiatroller-1990/garterscopic/internal/layouts"
)

func testLookup(name string) (layouts.Layout, bool) {
	if name == "default" {
		return layouts.Layout{
			Name: "default",
			Components: []layouts.ComponentInstance{
				{Name: "navbar"},
				{Name: "content"},
				{Name: "footer"},
			},
		}, true
	}
	return layouts.Layout{}, false
}

func testComponents() map[string]components.Definition {
	return map[string]components.Definition{
		"navbar":  {Name: "navbar", Style: "styles/navbar.css"},
		"footer":  {Name: "footer", Style: "styles/footer.css"},
		"content": {Name: "content"},
	}
}

func TestBuildAndRenderText(t *testing.T) {
	entries := []Entry{
		{Route: "/", Source: "pages/index.md", Layout: "default"},
		{Route: "/about/", Source: "pages/about.md", Layout: "default"},
	}

	g := Build(entries, testComponents(), testLookup)
	g.SortNodes()

	text := g.RenderText()

	for _, want := range []string{
		"/",
		"/about/",
		"layout: default",
		"navbar",
		"content",
		"footer",
		"style: styles/navbar.css",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("expected graph text to contain %q, got:\n%s", want, text)
		}
	}

	// Tree connectors must be present.
	if !strings.Contains(text, "├──") || !strings.Contains(text, "└──") {
		t.Error("expected tree connectors in output")
	}
}

func TestRenderJSON(t *testing.T) {
	entries := []Entry{
		{Route: "/", Source: "pages/index.md", Layout: "default"},
	}

	g := Build(entries, testComponents(), testLookup)

	json, err := g.RenderJSON()
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{`"route": "/"`, `"layout": "default"`, `"pages"`} {
		if !strings.Contains(json, want) {
			t.Errorf("expected JSON to contain %s", want)
		}
	}
}

func TestBuildWithoutLayout(t *testing.T) {
	entries := []Entry{{Route: "/x/"}}

	g := Build(entries, testComponents(), testLookup)
	if len(g.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(g.Nodes))
	}
	if len(g.Nodes[0].Children) != 0 {
		t.Errorf("expected no children without a layout, got %d", len(g.Nodes[0].Children))
	}
}
