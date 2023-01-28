package spec

import (
	"fmt"
	"net/http"

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
	docLoc     string
}

// NewPathItem returns a new PathItem
func NewPathItem() *PathItem {
	return &PathItem{
		Extensions: make(Extensions),
	}
}

// DocumentLocation returns this object's JSON path location
func (pi *PathItem) DocumentLocation() string {
	return pi.docLoc
}

// GatherRefs will add any definition reference keys to the specified refs
func (pi *PathItem) GatherRefs(refs map[string]struct{}) {
	if pi == nil {
		return
	}
	pi.Ref.GatherRefs(refs)
	for _, itm := range pi.Parameters {
		itm.GatherRefs(refs)
	}
	pi.Head.GatherRefs(refs)
	pi.Get.GatherRefs(refs)
	pi.Put.GatherRefs(refs)
	pi.Post.GatherRefs(refs)
	pi.Patch.GatherRefs(refs)
	pi.Options.GatherRefs(refs)
	pi.Delete.GatherRefs(refs)
}

// ReferencedDefinitions will return all definition names from all the Reference values within this
func (pi *PathItem) ReferencedDefinitions() *UniqueDefinitionRefs {
	if pi == nil {
		return nil
	}
	result := NewUniqueDefinitionRefs(len(pi.Parameters))
	result.AddRefs(pi.Ref)
	for _, itm := range pi.Parameters {
		if moreRefs := itm.ReferencedDefinitions(); moreRefs.Len() > 0 {
			result = result.Merge(moreRefs)
		}
	}
	if moreRefs := pi.Put.ReferencedDefinitions(); moreRefs.Len() > 0 {
		result = result.Merge(moreRefs)
	}
	if moreRefs := pi.Post.ReferencedDefinitions(); moreRefs.Len() > 0 {
		result = result.Merge(moreRefs)
	}
	if moreRefs := pi.Get.ReferencedDefinitions(); moreRefs.Len() > 0 {
		result = result.Merge(moreRefs)
	}
	if moreRefs := pi.Patch.ReferencedDefinitions(); moreRefs.Len() > 0 {
		result = result.Merge(moreRefs)
	}
	if moreRefs := pi.Delete.ReferencedDefinitions(); moreRefs.Len() > 0 {
		result = result.Merge(moreRefs)
	}
	if moreRefs := pi.Head.ReferencedDefinitions(); moreRefs.Len() > 0 {
		result = result.Merge(moreRefs)
	}
	if moreRefs := pi.Options.ReferencedDefinitions(); moreRefs.Len() > 0 {
		result = result.Merge(moreRefs)
	}
	return result
}

// Paths defines the Paths swagger object
// https://swagger.io/specification/v2/#paths-object
type Paths struct {
	Extensions
	Items  map[string]*PathItem
	docLoc string
}

// NewPaths returns a new Paths object
func NewPaths() *Paths {
	return &Paths{
		Extensions: make(Extensions),
		Items:      make(map[string]*PathItem),
	}
}

// DocumentLocation returns this object's JSON path location
func (p *Paths) DocumentLocation() string {
	return p.docLoc
}

// GatherRefs will add any definition reference keys to the specified refs
func (p *Paths) GatherRefs(refs map[string]struct{}) {
	if p == nil {
		return
	}
	for _, itm := range p.Items {
		itm.GatherRefs(refs)
	}
}

// ReferencedDefinitions will return all definition names from all the Reference values within this
func (p *Paths) ReferencedDefinitions() *UniqueDefinitionRefs {
	if p == nil {
		return nil
	}
	var result *UniqueDefinitionRefs
	for _, itm := range p.Items {
		if moreRefs := itm.ReferencedDefinitions(); moreRefs.Len() > 0 {
			result = result.Merge(moreRefs)
		}
	}
	return result
}

func parsePathItem(val *fastjson.Value, parser *Parser, path string) *PathItem {
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
	result.docLoc = parser.currentLoc
	obj.Visit(func(key []byte, v *fastjson.Value) {
		parser.currentLoc = fmt.Sprintf("%s.%s", fromLoc, key)
		switch {
		case matchString(key, "get"):
			result.Get = parseOperation(v, parser, path, http.MethodGet)
		case matchString(key, "put"):
			result.Put = parseOperation(v, parser, path, http.MethodPut)
		case matchString(key, "post"):
			result.Post = parseOperation(v, parser, path, http.MethodPost)
		case matchString(key, "delete"):
			result.Delete = parseOperation(v, parser, path, http.MethodDelete)
		case matchString(key, "options"):
			result.Options = parseOperation(v, parser, path, http.MethodOptions)
		case matchString(key, "head"):
			result.Head = parseOperation(v, parser, path, http.MethodHead)
		case matchString(key, "patch"):
			result.Patch = parseOperation(v, parser, path, http.MethodPatch)
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
	result.docLoc = parser.currentLoc
	obj.Visit(func(key []byte, v *fastjson.Value) {
		parser.currentLoc = fmt.Sprintf("%s.%s", fromLoc, key)
		keyStr := string(key)
		switch {
		case matchPath(key):
			if pi := parsePathItem(v, parser, keyStr); pi != nil {
				result.Items[keyStr] = pi
			}
		case matchExtension(key):
			result.Extensions[keyStr] = v
		default:
			parser.appendError(fmt.Errorf("invalid field name: '%s'", key))
		}
	})
	return result
}
