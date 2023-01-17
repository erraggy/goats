package spec

import (
	"fmt"

	"github.com/valyala/fastjson"
)

// PathItem defines the PathItem swagger object
// https://swagger.io/specification/v2/#path-item-object
type PathItem struct {
	Extensions
	Ref        *Reference
	Get        *Operation
	Put        *Operation
	Post       *Operation
	Delete     *Operation
	Options    *Operation
	Head       *Operation
	Patch      *Operation
	Parameters []Parameter
}

// NewPathItem returns a new PathItem
func NewPathItem() *PathItem {
	return &PathItem{
		Extensions: make(Extensions),
	}
}

// Paths defines the Paths swagger object
// https://swagger.io/specification/v2/#paths-object
type Paths struct {
	Extensions
	Items map[string]*PathItem
}

// NewPaths returns a new Paths object
func NewPaths() *Paths {
	return &Paths{
		Extensions: make(Extensions),
		Items:      make(map[string]*PathItem),
	}
}

func parsePathItem(val *fastjson.Value, parser *Parser) *PathItem {
	fromLoc := parser.currentLoc
	defer func() {
		parser.currentLoc = fromLoc
	}()
	obj, err := val.Object()
	if err != nil {
		parser.appendError(fmt.Errorf("invalid path item value: %w", err))
		return nil
	}
	result := NewPathItem()
	obj.Visit(func(key []byte, v *fastjson.Value) {
		parser.currentLoc = fmt.Sprintf("%s.%s", fromLoc, key)
		switch {
		case matchString(key, "get"):
			result.Get = parseOperation(v, parser)
		case matchString(key, "put"):
			result.Put = parseOperation(v, parser)
		case matchString(key, "post"):
			result.Post = parseOperation(v, parser)
		case matchString(key, "delete"):
			result.Delete = parseOperation(v, parser)
		case matchString(key, "options"):
			result.Options = parseOperation(v, parser)
		case matchString(key, "head"):
			result.Head = parseOperation(v, parser)
		case matchString(key, "patch"):
			result.Patch = parseOperation(v, parser)
		case matchString(key, "parameters"):
			if vals, e := v.Array(); e != nil {
				parser.appendError(fmt.Errorf("invalid parameters value: %w", e))
			} else {
				paramsLoc := parser.currentLoc
				for i, paramVal := range vals {
					parser.currentLoc = fmt.Sprintf("%s[%d]", paramsLoc, i)
					if p := parseParameter(paramVal, parser); p != nil {
						result.Parameters = append(result.Parameters, *p)
					}
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

func parsePaths(val *fastjson.Value, parser *Parser) *Paths {
	fromLoc := parser.currentLoc
	defer func() {
		parser.currentLoc = fromLoc
	}()
	obj, err := val.Object()
	if err != nil {
		parser.appendError(fmt.Errorf("invalid result value: %w", err))
		return nil
	}
	result := NewPaths()
	obj.Visit(func(key []byte, v *fastjson.Value) {
		parser.currentLoc = fmt.Sprintf("%s.%s", fromLoc, key)
		switch {
		case matchPath(key):
			if pi := parsePathItem(v, parser); pi != nil {
				result.Items[string(key)] = pi
			}
		case matchExtension(key):
			result.Extensions[string(key)] = v
		default:
			parser.appendError(fmt.Errorf("invalid field name: '%s'", key))
		}
	})
	return result
}
