package feed

import (
	"strings"
	"testing"
)

func TestGenerateBasic(t *testing.T) {
	posts := []Post{
		{
			Title:       "Hello World",
			Route:       "/blog/hello/",
			Date:        "2026-03-01",
			Description: "First post",
			Author:      "Jane",
		},
	}

	xml := Generate(posts, "Test Site", "http://example.com")

	for _, want := range []string{
		"<title>Test Site</title>",
		"<link>http://example.com/</link>",
		"<item>",
		"<title>Hello World</title>",
		"<link>http://example.com/blog/hello/</link>",
		"<pubDate>",
		"<description>First post</description>",
		"<author>Jane</author>",
	} {
		if !strings.Contains(xml, want) {
			t.Errorf("expected feed to contain %s", want)
		}
	}
}

func TestGenerateSortsByDate(t *testing.T) {
	posts := []Post{
		{Title: "Old", Route: "/old/", Date: "2026-01-01"},
		{Title: "New", Route: "/new/", Date: "2026-06-01"},
	}

	xml := Generate(posts, "Test", "http://example.com")

	oldIdx := strings.Index(xml, "<title>Old</title>")
	newIdx := strings.Index(xml, "<title>New</title>")
	if oldIdx == -1 || newIdx == -1 {
		t.Fatal("expected both posts in feed")
	}
	if newIdx > oldIdx {
		t.Error("expected newest post first")
	}
}

func TestGenerateEscapes(t *testing.T) {
	posts := []Post{
		{Title: `A & B "quoted"`, Route: "/x/", Date: "2026-01-01"},
	}

	xml := Generate(posts, "Title & Stuff <tag>", "http://example.com")

	if !strings.Contains(xml, "<title>A &amp; B &#34;quoted&#34;</title>") {
		t.Error("expected post title to be escaped")
	}
	if !strings.Contains(xml, "<title>Title &amp; Stuff &lt;tag&gt;</title>") {
		t.Error("expected site title to be escaped")
	}
}

func TestGeneratePubDateFormat(t *testing.T) {
	posts := []Post{
		{Title: "Dated", Route: "/d/", Date: "2026-05-04"},
	}

	xml := Generate(posts, "Test", "http://example.com")

	if !strings.Contains(xml, "<pubDate>Mon, 04 May 2026 00:00:00 +0000</pubDate>") {
		t.Errorf("expected RFC1123Z pubDate, got: %s", xml)
	}
}

func TestGenerateSkipsPostsWithoutLink(t *testing.T) {
	posts := []Post{
		{Title: "No link", Date: "2026-01-01"},
	}

	xml := Generate(posts, "Test", "http://example.com")

	if strings.Contains(xml, "<item>") {
		t.Error("expected posts without a resolvable link to be skipped")
	}
}
