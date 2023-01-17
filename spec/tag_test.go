package spec

import (
	"testing"

	"github.com/valyala/fastjson"
)

func Test_parseTag(t *testing.T) {
	var arena fastjson.Arena
	defer arena.Reset()

	type testCase struct {
		location    string
		expectedTag *Tag
		expectedErr error
	}
	tests := map[string]testCase{
		"valid tag should return valid result without error": {
			location: ".tags[0]",
			expectedTag: &Tag{
				Name:        "TestTag",
				Description: "just a test tag",
			},
		},
		"valid tag with externalDocs should return valid result without error": {
			location: ".tags[0]",
			expectedTag: &Tag{
				Name:        "TestTag",
				Description: "just a test tag",
				ExternalDocumentation: func() *ExternalDocumentation {
					ed := NewExternalDocumentation()
					ed.URL = "https://example.com/docs"
					ed.Description = "example external docs"
					return ed
				}(),
				Extensions: func() Extensions {
					ext := make(Extensions, 1)
					ext["x-robbie"] = arena.NewString("poop!")
					return ext
				}(),
			},
		},
	}
	for should, tt := range tests {
		parser := NewParser(nil)
		parser.currentLoc = tt.location
		tagVal := newTagJSON(arena, tt.expectedTag)
		t.Run(should, func(t *testing.T) {
			got := parseTag(tagVal, parser)
			if tt.expectedErr != nil {
				if err := parser.Err(); err == nil {
					t.Errorf("error was nil but we expected: %s", tt.expectedErr)
				}
			} else if !tagsEqual(got, tt.expectedTag) {
				t.Errorf("parseTag() = %v, want %v", got, tt.expectedTag)
			}
			t.Log(got.String())
		})
	}
}

func tagsEqual(t1, t2 *Tag) bool {
	if t1.Name != t2.Name {
		return false
	}
	if t1.Description != t2.Description {
		return false
	}
	if t1.ExternalDocumentation.url() != t2.ExternalDocumentation.url() {
		return false
	}
	if t1.ExternalDocumentation.description() != t2.ExternalDocumentation.description() {
		return false
	}
	if len(t1.Extensions) != len(t2.Extensions) {
		return false
	}
	if len(t1.Extensions) > 0 {
		for k1, v1 := range t1.Extensions {
			if v1 != t2.Extensions[k1] {
				return false
			}
		}
	}
	return true
}

func newTagJSON(a fastjson.Arena, tag *Tag) *fastjson.Value {
	if tag == nil {
		return nil
	}
	tagVal := a.NewObject()
	if tag.Name != "" {
		tagVal.Set("name", a.NewString(tag.Name))
	}
	if tag.Description != "" {
		tagVal.Set("description", a.NewString(tag.Description))
	}
	tag.marshalExtensions(tagVal)
	if tag.ExternalDocumentation != nil {
		extDocs := a.NewObject()
		if tag.ExternalDocumentation.Description != "" {
			extDocs.Set("description", a.NewString(tag.ExternalDocumentation.Description))
		}
		if tag.ExternalDocumentation.URL != "" {
			extDocs.Set("url", a.NewString(tag.ExternalDocumentation.URL))
		}
		tagVal.Set("externalDocs", extDocs)
	}
	return tagVal
}
