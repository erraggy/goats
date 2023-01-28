package spec

import (
	"bytes"
	"fmt"

	"github.com/valyala/fastjson"
)

// Header defines https://swagger.io/specification/v2/#header-object
type Header struct {
	Extensions
	Description      string
	Type             string
	Format           string
	Items            *Items
	CollectionFormat string
	Default          any
	Maximum          int
	ExclusiveMaximum bool
	Minimum          int
	ExclusiveMinimum bool
	MaxLength        int
	MinLength        int
	Pattern          string
	MaxItems         int
	MinItems         int
	UniqueItems      bool
	MaxProperties    int
	MinProperties    int
	Required         bool
	Enum             []any
	MultipleOf       int
	docLoc           string
}

// NewHeader returns a new Header object
func NewHeader() *Header {
	return &Header{
		Extensions: make(Extensions),
	}
}

// DocumentLocation returns this object's JSON path location
func (h *Header) DocumentLocation() string {
	return h.docLoc
}

func parseHeader(val *fastjson.Value, parser *Parser) *Header {
	// first be sure to capture and reset our parser's location
	fromLoc := parser.currentLoc
	defer func() {
		parser.currentLoc = fromLoc
	}()
	obj, err := val.Object()
	if err != nil {
		parser.appendError(fmt.Errorf("invalid header value: %w", err))
		return nil
	}
	result := NewHeader()
	result.docLoc = parser.currentLoc
	obj.Visit(func(key []byte, v *fastjson.Value) {
		parser.currentLoc = fmt.Sprintf("%s.%s", fromLoc, key)
		switch {
		case matchString(key, "description"):
			parser.parseString(v, "description", true, func(s string) {
				result.Description = s
			})
		case matchString(key, "type"):
			parser.parseString(v, "type", false, func(s string) {
				result.Type = s
			})
		case matchString(key, "format"):
			parser.parseString(v, "format", true, func(s string) {
				result.Format = s
			})
		case matchString(key, "items"):
			result.Items = parseItems(v, parser)
		case matchString(key, "collectionFormat"):
			parser.parseString(v, "collectionFormat", true, func(s string) {
				result.CollectionFormat = s
			})
		case matchString(key, "default"):
			result.Default = v
		case matchString(key, "maximum"):
			parser.parseInt(v, "maximum", func(i int) {
				result.Maximum = i
			})
		case matchString(key, "exclusiveMaximum"):
			parser.parseBool(v, "exclusiveMaximum", func(b bool) {
				result.ExclusiveMaximum = b
			})
		case matchString(key, "minimum"):
			parser.parseInt(v, "minimum", func(i int) {
				result.Minimum = i
			})
		case matchString(key, "exclusiveMinimum"):
			parser.parseBool(v, "exclusiveMinimum", func(b bool) {
				result.ExclusiveMinimum = b
			})
		case matchString(key, "maxLength"):
			parser.parseInt(v, "maxLength", func(i int) {
				result.MaxLength = i
			})
		case matchString(key, "minLength"):
			parser.parseInt(v, "minLength", func(i int) {
				result.MinLength = i
			})
		case matchString(key, "pattern"):
			parser.parseString(v, "pattern", true, func(s string) {
				result.Pattern = s
			})
		case matchString(key, "maxItems"):
			parser.parseInt(v, "maxItems", func(i int) {
				result.MaxItems = i
			})
		case matchString(key, "minItems"):
			parser.parseInt(v, "minItems", func(i int) {
				result.MinItems = i
			})
		case matchString(key, "uniqueItems"):
			parser.parseBool(v, "uniqueItems", func(b bool) {
				result.UniqueItems = b
			})
		case matchString(key, "enum"):
			if vals, e := v.Array(); e != nil {
				parser.appendError(fmt.Errorf("invalid enum value: %w", e))
			} else {
				result.Enum = make([]any, len(vals))
				for i := range vals {
					result.Enum[i] = vals[i]
				}
			}
		case matchString(key, "multipleOf"):
			parser.parseInt(v, "multipleOf", func(i int) {
				result.MultipleOf = i
			})
		case bytes.HasPrefix(key, []byte("x-")):
			result.Extensions[string(key)] = v
		default:
			parser.appendError(fmt.Errorf("invalid field name '%s'", key))
		}
	})
	return result
}
