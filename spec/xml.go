package spec

import (
	"fmt"

	"github.com/valyala/fastjson"
)

// XML defines a swagger XML object
// https://swagger.io/specification/v2/#xml-object
type XML struct {
	Extensions
	Name        string
	Namespace   string
	Prefix      string
	IsAttribute bool
	IsWrapped   bool
}

// NewXML returns a new XML object
func NewXML() *XML {
	return &XML{
		Extensions: make(Extensions),
	}
}

func parseXML(val *fastjson.Value, parser *Parser) *XML {
	// first be sure to capture and reset our parser's location
	fromLoc := parser.currentLoc
	defer func() {
		parser.currentLoc = fromLoc
	}()
	obj, err := val.Object()
	if err != nil {
		parser.appendError(fmt.Errorf("invalid security value: %w", err))
		return nil
	}
	result := NewXML()
	obj.Visit(func(key []byte, v *fastjson.Value) {
		parser.currentLoc = fmt.Sprintf("%s.%s", fromLoc, key)
		switch {
		case matchString(key, "name"):
			parser.parseString(v, "name", true, func(s string) {
				result.Name = s
			})
		case matchString(key, "namespace"):
			parser.parseString(v, "namespace", true, func(s string) {
				result.Namespace = s
			})
		case matchString(key, "prefix"):
			parser.parseString(v, "prefix", true, func(s string) {
				result.Prefix = s
			})
		case matchString(key, "attribute"):
			parser.parseBool(v, "attribute", func(b bool) {
				result.IsAttribute = b
			})
		case matchString(key, "wrapped"):
			parser.parseBool(v, "wrapped", func(b bool) {
				result.IsWrapped = b
			})
		case matchExtension(key):
			result.Extensions[string(key)] = v
		default:
			parser.appendError(fmt.Errorf("invalid field name: '%s'", key))
		}
	})
	return result
}
