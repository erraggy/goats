package spec

import (
	"fmt"

	"github.com/valyala/fastjson"
)

// Parameter defines a swagger parameter object
// https://swagger.io/specification/v2/#parameter-object
type Parameter struct {
	Extensions
	Name             string
	In               string
	Description      string
	Required         bool
	Schema           *Schema
	Type             string
	Format           string
	AllowEmptyValue  bool
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
	Enum             []any
	MultipleOf       int
}

// NewParameter returns a new Parameter object
func NewParameter() *Parameter {
	return &Parameter{
		Extensions: make(Extensions),
	}
}

func parseParameterDefinitions(val *fastjson.Value, parser *Parser) map[string]Parameter {
	fromLoc := parser.currentLoc
	defer func() {
		parser.currentLoc = fromLoc
	}()
	obj, err := val.Object()
	if err != nil {
		parser.appendError(fmt.Errorf("invalid parameters value: %w", err))
		return nil
	}
	result := make(map[string]Parameter, obj.Len())
	obj.Visit(func(key []byte, v *fastjson.Value) {
		parser.currentLoc = fmt.Sprintf("%s.%s", fromLoc, key)
		if param := parseParameter(v, parser); param != nil {
			result[string(key)] = *param
		}
	})
	return result
}

func parseParameter(val *fastjson.Value, parser *Parser) *Parameter {
	fromLoc := parser.currentLoc
	defer func() {
		parser.currentLoc = fromLoc
	}()
	obj, err := val.Object()
	if err != nil {
		parser.appendError(fmt.Errorf("invalid parameter value: %w", err))
		return nil
	}
	result := NewParameter()
	obj.Visit(func(key []byte, v *fastjson.Value) {
		parser.currentLoc = fmt.Sprintf("%s.%s", fromLoc, key)
		switch {
		case matchString(key, "name"):
			parser.parseString(v, "name", false, func(s string) {
				result.Name = s
			})
		case matchString(key, "in"):
			parser.parseString(v, "in", false, func(s string) {
				result.In = s
			})
		case matchString(key, "description"):
			parser.parseString(v, "description", true, func(s string) {
				result.Description = s
			})
		case matchString(key, "required"):
			parser.parseBool(v, "required", func(b bool) {
				result.Required = b
			})
		case matchString(key, "type"):
			parser.parseString(v, "type", false, func(s string) {
				result.Type = s
			})
		case matchString(key, "format"):
			parser.parseString(v, "format", true, func(s string) {
				result.Format = s
			})
		case matchString(key, "collectionFormat"):
			parser.parseString(v, "collectionFormat", true, func(s string) {
				result.CollectionFormat = s
			})
		case matchString(key, "allowEmptyValue"):
			parser.parseBool(v, "allowEmptyValue", func(b bool) {
				result.AllowEmptyValue = b
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
		case matchString(key, "multipleOf"):
			parser.parseInt(v, "multipleOf", func(i int) {
				result.MultipleOf = i
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
		case matchString(key, "items"):
			result.Items = parseItems(v, parser)
		case matchString(key, "default"):
			result.Default = v
		case matchString(key, "schema"):
			result.Schema = parseSchema(v, parser)
		case matchExtension(key):
			result.Extensions[string(key)] = v
		default:
			parser.appendError(fmt.Errorf("invalid field name: '%s'", key))
		}
	})
	return result
}
