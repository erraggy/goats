package spec

import (
	"fmt"

	"github.com/valyala/fastjson"
)

// Tag defines the swagger tag object
// https://swagger.io/specification/v2/#tag-object
type Tag struct {
	Extensions
	Name                  string
	Description           string
	ExternalDocumentation *ExternalDocumentation
}

// NewTag returns a new Tag
func NewTag() *Tag {
	return &Tag{
		Extensions: make(Extensions),
	}
}

// parseTag will attempt to parse a Tag from the source swagger .tags JSON array values
func parseTag(tagVal *fastjson.Value, parser *Parser) *Tag {
	// first be sure to capture and reset our parser's location
	fromLoc := parser.currentLoc
	defer func() {
		parser.currentLoc = fromLoc
	}()
	tagObj, err := tagVal.Object()
	if err != nil {
		parser.appendError(fmt.Errorf("invalid result value: %w", err))
	}
	result := NewTag()
	tagObj.Visit(func(key []byte, v *fastjson.Value) {
		parser.currentLoc = fmt.Sprintf("%s.%s", fromLoc, key)
		switch {
		case matchString(key, "name"):
			parser.parseString(v, "name", true, func(s string) {
				result.Name = s
			})
		case matchString(key, "description"):
			parser.parseString(v, "description", true, func(s string) {
				result.Description = s
			})
		case matchString(key, "externalDocs"):
			result.ExternalDocumentation = parseExternalDocumentation(v, parser)
		case matchExtension(key):
			result.Extensions[string(key)] = v
		default:
			parser.appendError(fmt.Errorf("invalid field name: '%s'", key))
		}
	})
	return result
}
