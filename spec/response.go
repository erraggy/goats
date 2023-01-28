package spec

import (
	"fmt"

	"github.com/valyala/fastjson"
)

// Response defines the Response swagger object
// https://swagger.io/specification/v2/#response-object
type Response struct {
	Extensions
	Description string
	Schema      *Schema
	Headers     map[string]*Header
	docLoc      string
}

// NewResponse returns a new Response object
func NewResponse() *Response {
	return &Response{
		Extensions: make(Extensions),
	}
}

// DocumentLocation returns this object's JSON path location
func (r *Response) DocumentLocation() string {
	return r.docLoc
}

// GatherRefs will add any definition reference keys to the specified refs
func (r *Response) GatherRefs(refs map[string]struct{}) {
	if r == nil {
		return
	}
	r.Schema.GatherRefs(refs)
}

// ReferencedDefinitions will return all definition names from all the Reference values within this
func (r *Response) ReferencedDefinitions() *UniqueDefinitionRefs {
	if r == nil {
		return nil
	}

	return r.Schema.ReferencedDefinitions()
}

// Responses defines the responses swagger object
// https://swagger.io/specification/v2/#responses-object
type Responses struct {
	Extensions
	Default      *Response
	ByStatusCode map[int]*Response
	docLoc       string
}

// NewResponses returns a new Responses object
func NewResponses() *Responses {
	return &Responses{
		Extensions:   make(Extensions),
		ByStatusCode: make(map[int]*Response),
	}
}

// DocumentLocation returns this object's JSON path location
func (rr *Responses) DocumentLocation() string {
	return rr.docLoc
}

// GatherRefs will add any definition reference keys to the specified refs
func (rr *Responses) GatherRefs(refs map[string]struct{}) {
	if rr == nil {
		return
	}
	rr.Default.GatherRefs(refs)
	for _, resp := range rr.ByStatusCode {
		resp.GatherRefs(refs)
	}
}

// ReferencedDefinitions will return all definition names from all the Reference values within this
func (rr *Responses) ReferencedDefinitions() *UniqueDefinitionRefs {
	if rr == nil {
		return nil
	}
	var result *UniqueDefinitionRefs
	if moreRefs := rr.Default.ReferencedDefinitions(); moreRefs.Len() > 0 {
		result = result.Merge(moreRefs)
	}
	for _, itm := range rr.ByStatusCode {
		if moreRefs := itm.ReferencedDefinitions(); moreRefs.Len() > 0 {
			result = result.Merge(moreRefs)
		}
	}
	return result
}

func parseResponses(val *fastjson.Value, parser *Parser) *Responses {
	// first be sure to capture and reset our parser's location
	fromLoc := parser.currentLoc
	defer func() {
		parser.currentLoc = fromLoc
	}()
	obj, err := val.Object()
	if err != nil {
		parser.appendError(fmt.Errorf("invalid responses value: %w", err))
		return nil
	}
	result := NewResponses()
	result.docLoc = parser.currentLoc
	obj.Visit(func(key []byte, v *fastjson.Value) {
		parser.currentLoc = fmt.Sprintf("%s.%s", fromLoc, key)
		switch {
		case matchString(key, "default"):
			if r := parseResponse(v, parser); r != nil {
				result.Default = r
			}
		case matchHTTPStatusCode(key):
			if r := parseResponse(v, parser); r != nil {
				result.ByStatusCode[bytesToInt(key)] = r
			}
		case matchExtension(key):
			result.Extensions[string(key)] = v
		default:
			parser.appendError(fmt.Errorf("invalid field name: '%s'", key))
		}
	})
	return result
}

func parseResponseDefinitions(val *fastjson.Value, parser *Parser) map[string]Response {
	// first be sure to capture and reset our parser's location
	fromLoc := parser.currentLoc
	defer func() {
		parser.currentLoc = fromLoc
	}()
	obj, err := val.Object()
	if err != nil {
		parser.appendError(fmt.Errorf("invalid response definitions value: %w", err))
		return nil
	}
	result := make(map[string]Response, obj.Len())
	obj.Visit(func(key []byte, v *fastjson.Value) {
		parser.currentLoc = fmt.Sprintf("%s.%s", fromLoc, key)
		if resp := parseResponse(v, parser); resp != nil {
			result[string(key)] = *resp
		}
	})
	return result
}

func parseResponse(val *fastjson.Value, parser *Parser) *Response {
	// first be sure to capture and reset our parser's location
	fromLoc := parser.currentLoc
	defer func() {
		parser.currentLoc = fromLoc
	}()
	obj, err := val.Object()
	if err != nil {
		parser.appendError(fmt.Errorf("invalid response value: %w", err))
		return nil
	}
	result := NewResponse()
	result.docLoc = parser.currentLoc
	obj.Visit(func(key []byte, v *fastjson.Value) {
		parser.currentLoc = fmt.Sprintf("%s.%s", fromLoc, key)
		switch {
		case matchString(key, "description"):
			parser.parseString(v, "description", false, func(s string) {
				result.Description = s
			})
		case matchString(key, "schema"):
			result.Schema = parseSchema(v, parser)
		case matchString(key, "headers"):
			if hMap, e := v.Object(); e != nil {
				parser.appendError(fmt.Errorf("invalid headers type: %w", e))
			} else {
				result.Headers = make(map[string]*Header, hMap.Len())
				hdrLoc := parser.currentLoc
				hMap.Visit(func(hKey []byte, hVal *fastjson.Value) {
					parser.currentLoc = fmt.Sprintf("%s.%s", hdrLoc, hKey)
					if hdr := parseHeader(hVal, parser); hdr != nil {
						result.Headers[string(hKey)] = hdr
					}
				})
			}
		case matchExtension(key):
			result.Extensions[string(key)] = v
		default:
			parser.appendError(fmt.Errorf("invalid field name: '%s'", key))
		}
	})
	return result
}
