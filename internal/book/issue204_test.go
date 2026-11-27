package book

import (
	"strings"
	"testing"
)

func TestIssue204ImagesAreNotLinks(t *testing.T) {
	tests := []struct {
		content string
		labels  []string
	}{
		{"# T\n\n![alt](#t) [real](#t)\n", []string{"real"}},
		{"![alt](#t)", nil},
		{"(![alt](#t)) ![alt](#t) a! [x](#t) word! [y](#t)", []string{"x", "y"}},
		{"`![alt](#t)` [real](#t)", []string{"real"}},
		{"| ![alt](#t) | [real](#t) |", []string{"real"}},
		{"![same](#t) [same](#t)", []string{"same"}},
	}
	for _, tt := range tests {
		t.Run(tt.content, func(t *testing.T) {
			extracted := ExtractLinks(tt.content)
			if len(extracted) != len(tt.labels) {
				t.Errorf("ExtractLinks = %+v, want labels %v", extracted, tt.labels)
			}
			for _, size := range [][2]int{{80, 20}, {10, 2}} {
				b, err := Read(strings.NewReader(tt.content), "x", size[0], size[1], false)
				if err != nil {
					t.Fatal(err)
				}
				var links []Link
				for _, p := range b.pages {
					links = append(links, p.Links...)
					if len(tt.labels) == 0 && len(p.LabelSpans) != 0 {
						t.Errorf("image has label spans: %+v", p.LabelSpans)
					}
				}
				if len(links) != len(tt.labels) {
					t.Errorf("size %v: links = %+v, want %v", size, links, tt.labels)
					continue
				}
				for i, label := range tt.labels {
					if links[i].Label != label || links[i].Target != "t" {
						t.Errorf("size %v: link = %+v", size, links[i])
					}
				}
				if size[0] == 80 {
					assertLinksLandWhereRendered(t, b, len(tt.labels))
				}
			}
		})
	}
}
