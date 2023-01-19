package spec

import (
	"fmt"

	"github.com/valyala/fastjson"
)

// Schema represents the subset of JSONSchema used by swagger
// https://swagger.io/specification/v2/#schema-object
type Schema struct {
	Extensions
	Ref                   *Reference
	Discriminator         string
	IsReadOnly            bool
	XML                   *XML
	Example               any
	Format                string
	Title                 string
	Description           string
	MultipleOf            int
	Maximum               int
	ExclusiveMaximum      bool
	Minimum               int
	ExclusiveMinimum      bool
	MaxLength             int
	MinLength             int
	Pattern               string
	MaxItems              int
	MinItems              int
	UniqueItems           bool
	MaxProperties         int
	MinProperties         int
	Required              []string
	Enum                  []any
	Type                  *StringOrStrings
	Items                 *SchemaOrSchemas
	AdditionalItems       *SchemaOrBool
	AllOf                 []Schema
	Properties            map[string]Schema
	AdditionalProperties  *SchemaOrBool
	ExternalDocumentation *ExternalDocumentation
	Default               any
}

// NewSchema returns a new Schema
func NewSchema() *Schema {
	return &Schema{
		Extensions: make(Extensions),
	}
}

func (s *Schema) ReferencedDefinitions() *UniqueDefinitionRefs {
	if refs := s.allRefs(); len(refs) > 0 {
		result := NewUniqueDefinitionRefs(len(refs))
		result.AddRefs(refs...)
		return result
	}
	return nil
}

// allRefs will gather all Reference pointers from within
func (s *Schema) allRefs() []*Reference {
	if s == nil {
		return nil
	}
	var results []*Reference
	if r := s.Ref; r != nil {
		results = append(results, r)
	}
	if s.Items != nil {
		if sch := s.Items.value; sch != nil {
			for _, r := range sch.allRefs() {
				results = append(results, r)
			}
		} else {
			for _, itm := range s.Items.items {
				for _, r := range itm.allRefs() {
					results = append(results, r)
				}
			}
		}
	}
	if sch, ok := s.AdditionalItems.AsSchema(); ok {
		for _, r := range sch.allRefs() {
			results = append(results, r)
		}
	}
	for _, sch := range s.AllOf {
		for _, r := range sch.allRefs() {
			results = append(results, r)
		}
	}
	for _, sch := range s.Properties {
		for _, r := range sch.allRefs() {
			results = append(results, r)
		}
	}
	if sch, ok := s.AdditionalProperties.AsSchema(); ok {
		for _, r := range sch.allRefs() {
			results = append(results, r)
		}
	}
	return results
}

// StringOrStrings is either a single string or a slice of them
type StringOrStrings struct {
	value *string
	items []string
}

// NewStringOrStrings returns a combo type for either a single string or many otherwise nil
func NewStringOrStrings(s ...string) *StringOrStrings {
	switch len(s) {
	case 0:
		return nil
	case 1:
		v := s[0]
		return &StringOrStrings{
			value: &v,
		}
	default:
		return &StringOrStrings{
			items: s,
		}
	}
}

// Values will return itself as a slice of either the single value or the many
func (s *StringOrStrings) Values() []string {
	if s == nil {
		return nil
	}
	if s.value != nil {
		return []string{*s.value}
	}
	return s.items
}

// SchemaOrSchemas intended for Schema.Items as it may be either one Schema or many
type SchemaOrSchemas struct {
	value *Schema
	items []Schema
}

// NewSchemaOrSchemas returns a combo type for either one or many Schema otherwise nil
func NewSchemaOrSchemas(ss ...Schema) *SchemaOrSchemas {
	switch len(ss) {
	case 0:
		return nil
	case 1:
		s := ss[0]
		return &SchemaOrSchemas{
			value: &s,
		}
	default:
		return &SchemaOrSchemas{
			items: ss,
		}
	}
}

// SchemaOrBool intended for Schema.AdditionalItems as it may be either a Schema or a bool
type SchemaOrBool struct {
	object *Schema
	value  bool
}

// NewSchemaOrBoolValue will return a combo SchemaOrBool with the specified bool value
func NewSchemaOrBoolValue(value bool) *SchemaOrBool {
	return &SchemaOrBool{
		value: value,
	}
}

// NewSchemaOrBoolObject will return a combo SchemaOrBool with the specified Schema object
func NewSchemaOrBoolObject(obj Schema) *SchemaOrBool {
	return &SchemaOrBool{
		object: &obj,
	}
}

// AsBool returns this as a bool value and if it is a bool
func (sb *SchemaOrBool) AsBool() (value bool, isBool bool) {
	if sb == nil {
		return false, false
	}
	if sb.object != nil {
		return false, false
	}
	return sb.value, true
}

// AsSchema returns this as a Schema and if it is a Schema
func (sb *SchemaOrBool) AsSchema() (*Schema, bool) {
	if sb == nil {
		return nil, false
	}
	if sb.object == nil {
		return nil, false
	}
	return sb.object, true
}

func parseDefinitions(val *fastjson.Value, parser *Parser) map[string]Schema {
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
	result := make(map[string]Schema, obj.Len())
	obj.Visit(func(key []byte, v *fastjson.Value) {
		parser.currentLoc = fmt.Sprintf("%s.%s", fromLoc, key)
		if s := parseSchema(v, parser); s != nil {
			result[string(key)] = *s
		}
	})
	return result
}

func parseSchema(val *fastjson.Value, parser *Parser) *Schema {
	// first be sure to capture and reset our parser's location
	fromLoc := parser.currentLoc
	defer func() {
		parser.currentLoc = fromLoc
	}()
	obj, err := val.Object()
	if err != nil {
		parser.appendError(fmt.Errorf("invalid schema value: %w", err))
		return nil
	}
	result := NewSchema()
	obj.Visit(func(key []byte, v *fastjson.Value) {
		parser.currentLoc = fmt.Sprintf("%s.%s", fromLoc, key)
		switch {
		case matchString(key, "$ref"):
			parser.parseString(v, "$ref", false, func(s string) {
				result.Ref = NewRef(s)
			})
		case matchString(key, "format"):
			parser.parseString(v, "format", true, func(s string) {
				result.Format = s
			})
		case matchString(key, "title"):
			parser.parseString(v, "title", true, func(s string) {
				result.Title = s
			})
		case matchString(key, "description"):
			parser.parseString(v, "description", true, func(s string) {
				result.Description = s
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
			// should be an array of strings representing the property names that are required
			if vals, e := v.Array(); e != nil {
				parser.appendError(fmt.Errorf("invalid 'required' value: %w", e))
			} else {
				reqLoc := parser.currentLoc
				for i, reqVal := range vals {
					parser.currentLoc = fmt.Sprintf("%s[%d]", reqLoc, i)
					parser.parseString(reqVal, fmt.Sprintf("required[%d]", i), false, func(s string) {
						result.Required = append(result.Required, s)
					})
				}
			}
		case matchString(key, "enum"):
			if vals, e := v.Array(); e != nil {
				parser.appendError(fmt.Errorf("invalid enum value: %w", e))
			} else {
				result.Enum = make([]any, len(vals))
				for i := range vals {
					result.Enum[i] = vals[i]
				}
			}
		case matchString(key, "type"):
			parser.parseString(v, "type", true, func(s string) {
				result.Type = NewStringOrStrings(s)
			})
		case matchString(key, "items"):
			if v.Type() == fastjson.TypeArray {
				vals := v.GetArray()
				n := len(vals)
				if n > 0 {
					schemas := make([]Schema, 0, n)
					itemsLoc := parser.currentLoc
					for i, sVal := range vals {
						parser.currentLoc = fmt.Sprintf("%s[%d]", itemsLoc, i)
						if schema := parseSchema(sVal, parser); schema != nil {
							schemas = append(schemas, *schema)
						}
					}
					result.Items = NewSchemaOrSchemas(schemas...)
				}
			} else {
				if schema := parseSchema(v, parser); schema != nil {
					result.Items = NewSchemaOrSchemas(*schema)
				}
			}
		case matchString(key, "properties"):
			if props := parseProperties(v, parser); len(props) > 0 {
				result.Properties = props
			}
		case matchString(key, "additionalProperties"):
			if v.Type() == fastjson.TypeObject {
				if schema := parseSchema(v, parser); schema != nil {
					result.AdditionalProperties = NewSchemaOrBoolObject(*schema)
				}
			} else {
				parser.parseBool(v, "additionalProperties", func(b bool) {
					result.AdditionalProperties = NewSchemaOrBoolValue(b)
				})
			}
		case matchString(key, "discriminator"):
			parser.parseString(v, "discriminator", true, func(s string) {
				result.Discriminator = s
			})
		case matchString(key, "readOnly"):
			parser.parseBool(v, "readOnly", func(b bool) {
				result.IsReadOnly = b
			})
		case matchString(key, "xml"):
			if x := parseXML(v, parser); x != nil {
				result.XML = x
			}
		case matchString(key, "externalDocs"):
			result.ExternalDocumentation = parseExternalDocumentation(v, parser)
		case matchString(key, "example"):
			result.Example = v
		case matchExtension(key):
			result.Extensions[string(key)] = v
		default:
			parser.appendError(fmt.Errorf("invalid field name: '%s'", key))
		}
	})
	return result
}

func parseProperties(val *fastjson.Value, parser *Parser) map[string]Schema {
	// first be sure to capture and reset our parser's location
	fromLoc := parser.currentLoc
	defer func() {
		parser.currentLoc = fromLoc
	}()
	obj, err := val.Object()
	if err != nil {
		parser.appendError(fmt.Errorf("invalid properties value: %w", err))
		return nil
	}
	result := make(map[string]Schema, obj.Len())
	obj.Visit(func(key []byte, v *fastjson.Value) {
		parser.currentLoc = fmt.Sprintf("%s.%s", fromLoc, key)
		if schema := parseSchema(v, parser); schema != nil {
			result[string(key)] = *schema
		}
	})
	return result
}
