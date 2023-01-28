package spec

import (
	"fmt"

	"github.com/valyala/fastjson"
)

// Swagger defines the root swagger object
// https://swagger.io/specification/v2/#swagger-object
type Swagger struct {
	Extensions
	Swagger               string
	Info                  Info
	Host                  string
	BasePath              string
	Schemes               []string
	Consumes              []string
	Produces              []string
	Paths                 Paths
	Definitions           map[string]Schema
	Parameters            map[string]Parameter
	Responses             map[string]Response
	SecurityDefinitions   map[string]SecurityScheme
	Security              []SecurityRequirements
	Tags                  []Tag
	ExternalDocumentation *ExternalDocumentation
	operationMap          OperationMap
	docLoc                string
}

// DocumentLocation returns this object's JSON path location
func (s *Swagger) DocumentLocation() string {
	return s.docLoc
}

// OperationCount returns the count of total operations contained within this spec
func (s *Swagger) OperationCount() int {
	if s == nil {
		return 0
	}
	return len(s.operationMap)
}

// OperationMap returns a copy of the OperationMap within this Swagger
func (s *Swagger) OperationMap() OperationMap {
	if s == nil {
		return nil
	}
	results := make(OperationMap, s.OperationCount())
	for key := range s.operationMap {
		results[key] = s.operationMap[key]
	}
	return results
}

// Operations returns a sorted slice of all of the Operation objects contained within this spec
func (s *Swagger) Operations() Operations {
	if s == nil {
		return nil
	}
	return s.operationMap.Sorted()
}

// GatherRefs will add any definition reference keys to the specified refs
func (s *Swagger) GatherRefs(refs map[string]struct{}) {
	if s == nil {
		return
	}
	s.Paths.GatherRefs(refs)
	for _, itm := range s.Definitions {
		itm.GatherRefs(refs)
	}
	for _, itm := range s.Parameters {
		itm.GatherRefs(refs)
	}
	for _, itm := range s.Responses {
		itm.GatherRefs(refs)
	}
}

// ReferencedDefinitions will return all definition names from all the Reference values within this
func (s *Swagger) ReferencedDefinitions() *UniqueDefinitionRefs {
	if s == nil {
		return nil
	}

	var result *UniqueDefinitionRefs
	if moreRefs := s.Paths.ReferencedDefinitions(); moreRefs.Len() > 0 {
		result = result.Merge(moreRefs)
	}
	for _, itm := range s.Definitions {
		if moreRefs := itm.ReferencedDefinitions(); moreRefs.Len() > 0 {
			result = result.Merge(moreRefs)
		}
	}
	for _, itm := range s.Parameters {
		if moreRefs := itm.ReferencedDefinitions(); moreRefs.Len() > 0 {
			result = result.Merge(moreRefs)
		}
	}
	for _, itm := range s.Responses {
		if moreRefs := itm.ReferencedDefinitions(); moreRefs.Len() > 0 {
			result = result.Merge(moreRefs)
		}
	}
	return result
}

// addOperation will add the specified Operation to metadata and return true only if it as added and not preexisting.
func (s *Swagger) addOperation(op *Operation) bool {
	if s == nil || op == nil {
		return false
	}
	op.Key = op.Key.Canonicalize()
	if _, preexisting := s.operationMap[op.Key]; preexisting {
		return false
	}
	s.operationMap[op.Key] = op
	return true
}

// NewSwagger returns a new Swagger
func NewSwagger() *Swagger {
	return &Swagger{
		Extensions:   make(Extensions),
		operationMap: make(map[OperationKey]*Operation),
	}
}

// parseSwagger will attempt to parse the root swagger object from the root JSON value
func parseSwagger(swagVal *fastjson.Value, parser *Parser) *Swagger {
	swagObj, err := swagVal.Object()
	if err != nil {
		err = fmt.Errorf("invalid swagger value: %w", err)
		parser.appendError(err)
		return nil
	}
	result := NewSwagger()
	result.docLoc = parser.currentLoc
	parser.swagger = result
	defer func() {
		// reset after
		parser.currentLoc = "."
	}()
	swagObj.Visit(func(key []byte, v *fastjson.Value) {
		parser.currentLoc = fmt.Sprintf(".%s", key)
		switch {
		case matchString(key, "swagger"):
			parser.parseAndValidateString(v, "swagger", func(s string) error {
				if s != "2.0" {
					return fmt.Errorf("swagger value should be '2.0' but got: '%s'", s)
				}
				result.Swagger = s
				return nil
			})
		case matchString(key, "host"):
			parser.parseString(v, "host", true, func(s string) {
				result.Host = s
			})
		case matchString(key, "basePath"):
			parser.parseString(v, "basePath", true, func(s string) {
				result.BasePath = s
			})
		case matchString(key, "schemes"):
			if schemes, e := v.Array(); e != nil {
				parser.appendError(fmt.Errorf("invalid schemes value: %w", e))
			} else {
				for i, sVal := range schemes {
					parser.currentLoc = fmt.Sprintf(".schemes[%d]", i)
					parser.parseString(sVal, "schemes item", true, func(s string) {
						result.Schemes = append(result.Schemes, s)
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
		case matchString(key, "info"):
			if info := parseInfo(v, parser); info != nil {
				result.Info = *info
			}
		case matchString(key, "definitions"):
			if defs := parseDefinitions(v, parser); len(defs) > 0 {
				result.Definitions = defs
			}
		case matchString(key, "paths"):
			if paths := parsePaths(v, parser); paths != nil {
				result.Paths = *paths
			}
		case matchString(key, "parameters"):
			if params := parseParameterDefinitions(v, parser); len(params) > 0 {
				result.Parameters = params
			}
		case matchString(key, "responses"):
			if responses := parseResponseDefinitions(v, parser); len(responses) > 0 {
				result.Responses = responses
			}
		case matchString(key, "securityDefinitions"):
			if secDefs := parseSecurityDefinitions(v, parser); len(secDefs) > 0 {
				result.SecurityDefinitions = secDefs
			}
		case matchString(key, "security"):
			// this is an array of security requirements, so parse the array then parse each
			if secReqs, e := v.Array(); e != nil {
				parser.appendError(fmt.Errorf("invalid 'security' value: %w", e))
			} else {
				secLoc := parser.currentLoc
				for i, secVal := range secReqs {
					parser.currentLoc = fmt.Sprintf("%s[%d]", secLoc, i)
					if sec := parseSecurityRequirements(secVal, parser); len(sec) > 0 {
						result.Security = append(result.Security, sec)
					}
				}
			}
		case matchString(key, "tags"):
			if tags, e := v.Array(); e != nil {
				parser.appendError(fmt.Errorf("invalid tags value: %w", e))
			} else {
				result.Tags = make([]Tag, 0, len(tags))
				tagsLoc := parser.currentLoc
				for i, tagVal := range tags {
					parser.currentLoc = fmt.Sprintf("%s[%d]", tagsLoc, i)
					if tag := parseTag(tagVal, parser); tag != nil {
						result.Tags = append(result.Tags, *tag)
					}
				}
			}
		case matchString(key, "externalDocs"):
			if ed := parseExternalDocumentation(v, parser); ed != nil {
				result.ExternalDocumentation = ed
			}
		case matchExtension(key):
			result.Extensions[string(key)] = v
		default:
			parser.appendError(fmt.Errorf("invalid field name: '%s'", key))
		}
	})
	parser.swagger = result
	return result
}

// ExternalDocumentation defines the external documentation swagger object
// https://swagger.io/specification/v2/#external-documentation-object
type ExternalDocumentation struct {
	Extensions
	Description string
	URL         string
	docLoc      string
}

// NewExternalDocumentation returns a new ExternalDocumentation
func NewExternalDocumentation() *ExternalDocumentation {
	return &ExternalDocumentation{
		Extensions: make(Extensions),
	}
}

// DocumentLocation returns this object's JSON path location
func (ed *ExternalDocumentation) DocumentLocation() string {
	return ed.docLoc
}

func (ed *ExternalDocumentation) marshal(a *fastjson.Arena) *fastjson.Value {
	val := a.NewObject()
	if ed.Description != "" {
		val.Set("description", a.NewString(ed.Description))
	}
	if ed.URL != "" {
		val.Set("url", a.NewString(ed.URL))
	}
	ed.marshalExtensions(val)
	return val
}

func (ed *ExternalDocumentation) String() string {
	if ed == nil {
		return ""
	}
	a := arenaPool.Get()
	defer func() {
		a.Reset()
		arenaPool.Put(a)
	}()
	val := ed.marshal(a)
	return string(val.MarshalTo(nil))
}

func (ed *ExternalDocumentation) description() string {
	if ed != nil {
		return ed.Description
	}
	return ""
}

func (ed *ExternalDocumentation) url() string {
	if ed != nil {
		return ed.URL
	}
	return ""
}

// parseExternalDocumentation will attempt to parse an ExternalDocumentation from the source swagger .externalDocumentation JSON values
func parseExternalDocumentation(edVal *fastjson.Value, parser *Parser) *ExternalDocumentation {
	// first be sure to capture and reset our parser's location
	fromLoc := parser.currentLoc
	defer func() {
		parser.currentLoc = fromLoc
	}()
	edObj, err := edVal.Object()
	if err != nil {
		parser.appendError(fmt.Errorf("invalid externalDocs value: %w", err))
		return nil
	}
	result := NewExternalDocumentation()
	result.docLoc = parser.currentLoc
	edObj.Visit(func(key []byte, v *fastjson.Value) {
		parser.currentLoc = fmt.Sprintf("%s.%s", fromLoc, key)
		switch {
		case matchString(key, "url"):
			parser.parseString(v, "url", true, func(s string) {
				result.URL = s
			})
		case matchString(key, "description"):
			parser.parseString(v, "description", true, func(s string) {
				result.Description = s
			})
		case matchExtension(key):
			result.Extensions[string(key)] = v
		default:
			parser.appendError(fmt.Errorf("invalid field name: '%s'", key))
		}
	})
	return result
}
