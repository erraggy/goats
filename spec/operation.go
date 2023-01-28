package spec

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/valyala/fastjson"
)

// Operation defines a swagger operation object
// https://swagger.io/specification/v2/#operation-object
type Operation struct {
	Extensions
	ID                    string
	Summary               string
	Description           string
	Deprecated            bool
	Tags                  []string
	Consumes              []string
	Produces              []string
	Schemes               []string
	Parameters            []Parameter
	Responses             Responses
	Security              []SecurityRequirements
	ExternalDocumentation *ExternalDocumentation
	Key                   OperationKey
	docLoc                string
}

// NewOperation returns a new Operation object
func NewOperation(path string, method string) *Operation {
	return &Operation{
		Extensions: make(Extensions),
		Key: OperationKey{
			Path:   path,
			Method: method,
		},
	}
}

// DocumentLocation returns this object's JSON path location
func (o *Operation) DocumentLocation() string {
	return o.docLoc
}

// GatherRefs will add any definition reference keys to the specified refs
func (o *Operation) GatherRefs(refs map[string]struct{}) {
	if o == nil {
		return
	}
	for _, itm := range o.Parameters {
		itm.GatherRefs(refs)
	}
	o.Responses.GatherRefs(refs)
}

// ReferencedDefinitions will return all definition names from all the Reference values within this
func (o *Operation) ReferencedDefinitions() *UniqueDefinitionRefs {
	if o == nil {
		return nil
	}

	var result *UniqueDefinitionRefs
	for _, param := range o.Parameters {
		if moreRefs := param.ReferencedDefinitions(); moreRefs.Len() > 0 {
			result = result.Merge(moreRefs)
		}
	}
	if moreRefs := o.Responses.ReferencedDefinitions(); moreRefs.Len() > 0 {
		result = result.Merge(moreRefs)
	}

	return result
}

// OperationKey defines the natural key for any swagger Operation
type OperationKey struct {
	Path   string
	Method string
}

// Canonicalize will make sure that the method is in all upper-case
func (k OperationKey) Canonicalize() OperationKey {
	return OperationKey{
		Path:   k.Path,
		Method: strings.ToUpper(k.Method),
	}
}

// Operations defines a slice of Operation objects
type Operations []*Operation

// Sorted returns a sorted slice of Operation objects
func (ops Operations) Sorted() Operations {
	if len(ops) == 0 {
		return ops
	}
	sort.Slice(ops, func(i, j int) bool {
		oi, oj := ops[i], ops[j]
		if oi.Key.Path < oj.Key.Path {
			return true
		}
		if oi.Key.Path > oj.Key.Path {
			return false
		}
		if oi.Key.Method < oj.Key.Method {
			return true
		}
		return oi.Key.Method > oj.Key.Method
	})
	return ops
}

// OperationMap defines a mapping of Operation objects by each natural OperationKey
type OperationMap map[OperationKey]*Operation

// Sorted returns a sorted slice of Operation objects
func (om OperationMap) Sorted() Operations {
	var (
		result = make(Operations, len(om))
		i      = 0
	)
	for k := range om {
		result[i] = om[k]
		i++
	}
	return result.Sorted()
}

func (om OperationMap) Contains(opKey OperationKey) bool {
	if len(om) == 0 {
		return false
	}
	_, found := om[opKey]
	return found
}

// Union returns a new map with all items in both maps
func (om OperationMap) Union(other OperationMap) OperationMap {
	n1, n2, max := len(om), len(other), 0
	if n1 > max {
		max = n1
	}
	if n2 > max {
		max = n2
	}
	result := make(OperationMap, max)
	if max == 0 {
		return result
	}
	for key := range om {
		result[key] = om[key]
	}
	for key := range other {
		result[key] = other[key]
	}
	return result
}

// Difference returns a new map with items in the current map but not in the other
func (om OperationMap) Difference(other OperationMap) OperationMap {
	n := len(om)
	result := make(OperationMap, n)
	if n > 0 {
		otherN := len(other)
		for key := range om {
			if otherN == 0 || !other.Contains(key) {
				result[key] = om[key]
			}
		}
	}
	return result
}

// Intersect returns a new map with items that exist only in both maps
func (om OperationMap) Intersect(other OperationMap) OperationMap {
	result := make(OperationMap)
	// outer loop the smaller
	if nLeft, nRight := len(om), len(other); nLeft < nRight {
		for key := range om {
			if op, exists := other[key]; exists {
				result[key] = op
			}
		}
	} else {
		for key := range other {
			if op, exists := om[key]; exists {
				result[key] = op
			}
		}
	}
	return result
}

// SymmetricDifference returns a new map with items in the current map or the other map but not in both
func (om OperationMap) SymmetricDifference(other OperationMap) OperationMap {
	aDiff := om.Difference(other)
	bDiff := other.Difference(om)
	return aDiff.Union(bDiff)
}

func parseOperation(val *fastjson.Value, parser *Parser, path string, method string) *Operation {
	// first be sure to capture and reset our parser's location
	fromLoc := parser.currentLoc
	defer func() {
		parser.currentLoc = fromLoc
	}()
	obj, err := val.Object()
	if err != nil {
		parser.appendError(fmt.Errorf("invalid operation value: %w", err))
		return nil
	}
	result := NewOperation(path, method)
	result.docLoc = parser.currentLoc
	obj.Visit(func(key []byte, v *fastjson.Value) {
		parser.currentLoc = fmt.Sprintf("%s.%s", fromLoc, key)
		switch {
		case matchString(key, "operationId"):
			parser.parseAndValidateString(v, "operationId", func(id string) error {
				if id == "" {
					return errors.New("empty operationId")
				}
				if other, unique := parser.locationForOperation(id); !unique {
					return fmt.Errorf("duplicated operationID[%s]: also in: %s", id, other)
				}
				result.ID = id
				return nil
			})
		case matchString(key, "summary"):
			parser.parseString(v, "summary", true, func(s string) {
				result.Summary = s
			})
		case matchString(key, "description"):
			parser.parseString(v, "description", true, func(s string) {
				result.Description = s
			})
		case matchString(key, "deprecated"):
			parser.parseBool(v, "deprecated", func(b bool) {
				result.Deprecated = b
			})
		case matchString(key, "tags"):
			if tags, e := v.Array(); e != nil {
				parser.appendError(fmt.Errorf("invalid tags value: %w", e))
			} else {
				tagsLoc := parser.currentLoc
				for i, tVal := range tags {
					parser.currentLoc = fmt.Sprintf("%s[%d]", tagsLoc, i)
					parser.parseString(tVal, "tags item", true, func(s string) {
						result.Tags = append(result.Tags, s)
					})
				}
			}
		case matchString(key, "consumes"):
			if consumes, e := v.Array(); e != nil {
				parser.appendError(fmt.Errorf("invalid consumes value: %w", e))
			} else {
				consumesLoc := parser.currentLoc
				for i, cVal := range consumes {
					parser.currentLoc = fmt.Sprintf("%s[%d]", consumesLoc, i)
					parser.parseString(cVal, "consumes item", true, func(s string) {
						result.Consumes = append(result.Consumes, s)
					})
				}
			}
		case matchString(key, "produces"):
			if produces, e := v.Array(); e != nil {
				parser.appendError(fmt.Errorf("invalid produces value: %w", e))
			} else {
				producesLoc := parser.currentLoc
				for i, pVal := range produces {
					parser.currentLoc = fmt.Sprintf("%s[%d]", producesLoc, i)
					parser.parseString(pVal, "produces item", true, func(s string) {
						result.Produces = append(result.Produces, s)
					})
				}
			}
		case matchString(key, "schemes"):
			if schemes, e := v.Array(); e != nil {
				parser.appendError(fmt.Errorf("invalid schemes value: %w", e))
			} else {
				schemesLoc := parser.currentLoc
				for i, sVal := range schemes {
					parser.currentLoc = fmt.Sprintf("%s[%d]", schemesLoc, i)
					parser.parseString(sVal, "schemes item", true, func(s string) {
						result.Schemes = append(result.Schemes, s)
					})
				}
			}
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
		case matchString(key, "responses"):
			if rs := parseResponses(v, parser); rs != nil {
				result.Responses = *rs
			}
		case matchString(key, "security"):
			if vals, e := v.Array(); e != nil {
				parser.appendError(fmt.Errorf("invalid security value: %w", e))
			} else {
				secLoc := parser.currentLoc
				for i, secVal := range vals {
					parser.currentLoc = fmt.Sprintf("%s[%d]", secLoc, i)
					if sec := parseSecurityRequirements(secVal, parser); sec != nil {
						result.Security = append(result.Security, sec)
					}
				}
			}
		case matchString(key, "externalDocs"):
			result.ExternalDocumentation = parseExternalDocumentation(v, parser)
		case matchExtension(key):
			result.Extensions[string(key)] = v
		default:
			parser.appendError(fmt.Errorf("invalid field name: '%s'", key))
		}
	})
	// store this in our swagger's operations map
	parser.swagger.addOperation(result)

	return result
}
