package spec

import (
	"fmt"

	"github.com/valyala/fastjson"
)

// Items defines the items swagger object
// https://swagger.io/specification/v2/#items-object
type Items struct {
	Extensions
	Type             string
	Format           string
	Items            *Items
	CollectionFormat string
	Default          any
	MultipleOf       int
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
	docLoc           string
}

// NewItems returns a new Items
func NewItems() *Items {
	return &Items{
		Extensions: make(Extensions),
	}
}

// DocumentLocation returns this object's JSON path location
func (i *Items) DocumentLocation() string {
	return i.docLoc
}

//nolint:funlen // it just doesn't get shorter than this
func parseItems(val *fastjson.Value, parser *Parser) *Items {
	// first be sure to capture and reset our parser's location
	fromLoc := parser.currentLoc
	defer func() {
		parser.currentLoc = fromLoc
	}()
	obj, err := val.Object()
	if err != nil {
		parser.appendError(fmt.Errorf("invalid result value: %w", err))
		return nil
	}
	result := NewItems()
	result.docLoc = parser.currentLoc
	obj.Visit(func(key []byte, v *fastjson.Value) {
		parser.currentLoc = fmt.Sprintf("%s.%s", fromLoc, key)
		switch {
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
		case matchString(key, "multipleOf"):
			parser.parseInt(v, "multipleOf", func(i int) {
				result.MultipleOf = i
			})
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
		case matchString(key, "maxProperties"):
			parser.parseInt(v, "maxProperties", func(i int) {
				result.MaxProperties = i
			})
		case matchString(key, "minProperties"):
			parser.parseInt(v, "minProperties", func(i int) {
				result.MinProperties = i
			})
		case matchString(key, "required"):
			parser.parseBool(v, "required", func(b bool) {
				result.Required = b
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
		case matchExtension(key):
			result.Extensions[string(key)] = v
		default:
			parser.appendError(fmt.Errorf("invalid field name: '%s'", key))
		}
	})
	return result
}
