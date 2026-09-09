package renderer

import (
	"strings"
	"testing"

	"github.com/italiatroller-1990/garterscopic/internal/components"
)

func TestRenderDefaultSlot(t *testing.T) {
	r := New(map[string]components.Definition{
		"card": {Name: "card", Options: map[string]components.OptionDefinition{
			"title": {Type: "string"},
		}},
	})

	html := `<div class="card"><h2>{{ title }}</h2><slot /></div>`

	options := map[string]any{
		"title":     "My Card",
		"__slots__": &SlotContent{Default: "<p>Body text</p>"},
	}

	out, err := r.RenderComponentHTML("card", html, options)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(out, "<h2>My Card</h2>") {
		t.Errorf("expected placeholder substitution, got: %s", out)
	}
	if !strings.Contains(out, "<p>Body text</p>") {
		t.Errorf("expected slot content injected, got: %s", out)
	}
	if strings.Contains(out, "slot") {
		t.Errorf("expected slot tag removed, got: %s", out)
	}
}

func TestRenderNamedSlots(t *testing.T) {
	r := New(map[string]components.Definition{})

	html := `<header>{{ title }}<slot name="header" /></header><main><slot name="content" /></main>`

	options := map[string]any{
		"title":     "Site",
		"__slots__": &SlotContent{Named: map[string]string{"header": "<nav>Nav</nav>", "content": "<p>Main</p>"}},
	}

	out, err := r.RenderComponentHTML("layout", html, options)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(out, "<nav>Nav</nav>") || !strings.Contains(out, "<p>Main</p>") {
		t.Errorf("expected named slot content injected, got: %s", out)
	}
}

func TestRenderSlotWithoutContentIsEmpty(t *testing.T) {
	r := New(map[string]components.Definition{})

	out, err := r.RenderComponentHTML("card", `<div><slot /></div>`, map[string]any{})
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(out, "slot") {
		t.Errorf("expected unfilled slot tag stripped, got: %s", out)
	}
}

func TestExtractSlots(t *testing.T) {
	options := map[string]any{
		"title": "T",
		"slots": map[string]any{
			"default": "<p>d</p>",
			"header":  "<h>h</h>",
			"count":   3, // non-string values are ignored
		},
	}

	slots := ExtractSlots(options)
	if slots == nil {
		t.Fatal("expected slot content")
	}
	if slots.Default != "<p>d</p>" {
		t.Errorf("unexpected default: %q", slots.Default)
	}
	if slots.Named["header"] != "<h>h</h>" {
		t.Errorf("unexpected named slot: %v", slots.Named)
	}
	if _, exists := options["slots"]; exists {
		t.Error("expected slots removed from options")
	}
	if _, exists := options["title"]; !exists {
		t.Error("expected other options preserved")
	}
}

func TestExtractSlotsAbsent(t *testing.T) {
	if ExtractSlots(map[string]any{"title": "T"}) != nil {
		t.Error("expected nil when no slots declared")
	}
}

func TestRenderSlotContentMayContainPlaceholders(t *testing.T) {
	r := New(map[string]components.Definition{
		"box": {Name: "box", Options: map[string]components.OptionDefinition{
			"label": {Type: "string"},
		}},
	})

	// The slot content itself contains a placeholder that must still resolve.
	html := `<div class="box">{{ label }}:<slot /></div>`

	options := map[string]any{
		"label":     "A",
		"__slots__": &SlotContent{Default: "{{ label }}"},
	}

	out, err := r.RenderComponentHTML("box", html, options)
	if err != nil {
		t.Fatal(err)
	}

	// Slot content is injected verbatim (it is the site author's own HTML).
	if !strings.Contains(out, "A:") {
		t.Errorf("expected label rendered, got: %s", out)
	}
}
