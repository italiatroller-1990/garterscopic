package markdown

import (
	"testing"
)

func TestRender(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "simple text",
			input: "Hello World",
			want:  "<p>Hello World</p>\n",
		},
		{
			name:  "heading",
			input: "# Hello",
			want:  `<h1 id="hello">Hello</h1>` + "\n",
		},
		{
			name:  "bold and italic",
			input: "**bold** and *italic*",
			want:  "<p><strong>bold</strong> and <em>italic</em></p>\n",
		},
		{
			name:  "list",
			input: "- item 1\n- item 2",
			want:  "<ul>\n<li>item 1</li>\n<li>item 2</li>\n</ul>\n",
		},
		{
			name:  "link",
			input: "[Google](https://google.com)",
			want:  "<p><a href=\"https://google.com\">Google</a></p>\n",
		},
		{
			name:  "GFM table",
			input: "| A | B |\n|---|---|\n| 1 | 2 |",
			want:  "<table>\n<thead>\n<tr>\n<th>A</th>\n<th>B</th>\n</tr>\n</thead>\n<tbody>\n<tr>\n<td>1</td>\n<td>2</td>\n</tr>\n</tbody>\n</table>\n",
		},
		{
			name:  "task list",
			input: "- [x] done\n- [ ] todo",
			want:  "<ul>\n<li><input checked=\"\" disabled=\"\" type=\"checkbox\" /> done</li>\n<li><input disabled=\"\" type=\"checkbox\" /> todo</li>\n</ul>\n",
		},
		{
			name:  "strikethrough",
			input: "~~strike~~",
			want:  "<p><del>strike</del></p>\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Render(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Render() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Render() got = %q, want %q", got, tt.want)
			}
		})
	}
}
