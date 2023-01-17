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
}

// NewResponse returns a new Response object
func NewResponse() *Response {
	return &Response{
		Extensions: make(Extensions),
	}
}

// Responses defines the responses swagger object
// https://swagger.io/specification/v2/#responses-object
type Responses struct {
	Extensions
	Default      *Response
	ByStatusCode map[int]*Response
}

// NewResponses returns a new Responses object
func NewResponses() *Responses {
	return &Responses{
		Extensions:   make(Extensions),
		ByStatusCode: make(map[int]*Response),
	}
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
