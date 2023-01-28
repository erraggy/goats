package spec

import (
	"fmt"

	"github.com/valyala/fastjson"
)

// SecurityRequirements defines https://swagger.io/specification/v2/#security-requirement-object
type SecurityRequirements map[string][]string

// SecurityScheme defines https://swagger.io/specification/v2/#security-scheme-object
type SecurityScheme struct {
	Extensions
	Type             string
	Description      string
	Name             string
	In               string
	Flow             string
	AuthorizationURL string
	TokenURL         string
	Scopes           Scopes
	docLoc           string
}

// NewSecurityScheme returns a new SecurityScheme object
func NewSecurityScheme() *SecurityScheme {
	return &SecurityScheme{
		Extensions: make(Extensions),
	}
}

// DocumentLocation returns this object's JSON path location
func (ss *SecurityScheme) DocumentLocation() string {
	return ss.docLoc
}

// Scopes defines https://swagger.io/specification/v2/#scopes-object
type Scopes struct {
	Extensions
	Values map[string]string
	docLoc string
}

// NewScopes returns a new Scopes object
func NewScopes() *Scopes {
	return &Scopes{
		Extensions: make(Extensions),
	}
}

// DocumentLocation returns this object's JSON path location
func (s *Scopes) DocumentLocation() string {
	return s.docLoc
}

func parseSecurityDefinitions(val *fastjson.Value, parser *Parser) map[string]SecurityScheme {
	// first be sure to capture and reset our parser's location
	fromLoc := parser.currentLoc
	defer func() {
		parser.currentLoc = fromLoc
	}()
	obj, err := val.Object()
	if err != nil {
		parser.appendError(fmt.Errorf("invalid security definitions value: %w", err))
		return nil
	}
	result := make(map[string]SecurityScheme, obj.Len())
	obj.Visit(func(key []byte, v *fastjson.Value) {
		parser.currentLoc = fmt.Sprintf("%s.%s", fromLoc, key)
		if ss := parseSecurityScheme(v, parser); ss != nil {
			result[string(key)] = *ss
		}
	})
	return result
}

func parseSecurityScheme(val *fastjson.Value, parser *Parser) *SecurityScheme {
	// first be sure to capture and reset our parser's location
	fromLoc := parser.currentLoc
	defer func() {
		parser.currentLoc = fromLoc
	}()
	obj, err := val.Object()
	if err != nil {
		parser.appendError(fmt.Errorf("invalid security scheme value: %w", err))
		return nil
	}
	result := NewSecurityScheme()
	result.docLoc = parser.currentLoc
	obj.Visit(func(key []byte, v *fastjson.Value) {
		parser.currentLoc = fmt.Sprintf("%s.%s", fromLoc, key)
		switch {
		case matchString(key, "type"):
			parser.parseString(v, "type", false, func(s string) {
				result.Type = s
			})
		case matchString(key, "description"):
			parser.parseString(v, "description", true, func(s string) {
				result.Description = s
			})
		case matchString(key, "name"):
			parser.parseString(v, "name", false, func(s string) {
				result.Name = s
			})
		case matchString(key, "in"):
			parser.parseString(v, "in", false, func(s string) {
				result.In = s
			})
		case matchString(key, "flow"):
			parser.parseString(v, "flow", false, func(s string) {
				result.Flow = s
			})
		case matchString(key, "authorizationUrl"):
			parser.parseString(v, "authorizationUrl", false, func(s string) {
				result.AuthorizationURL = s
			})
		case matchString(key, "tokenUrl"):
			parser.parseString(v, "tokenUrl", false, func(s string) {
				result.TokenURL = s
			})
		case matchString(key, "scopes"):
			if scopes := parseScopes(v, parser); scopes != nil {
				result.Scopes = *scopes
			}
		case matchExtension(key):
			result.Extensions[string(key)] = v
		default:
			parser.appendError(fmt.Errorf("invalid field name: '%s'", key))
		}
	})
	return result
}

func parseScopes(val *fastjson.Value, parser *Parser) *Scopes {
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
	result := NewScopes()
	result.docLoc = parser.currentLoc
	result.Values = make(map[string]string, obj.Len())
	obj.Visit(func(key []byte, v *fastjson.Value) {
		parser.currentLoc = fmt.Sprintf("%s.%s", fromLoc, key)
		if matchExtension(key) {
			result.Extensions[string(key)] = v
		} else {
			parser.parseString(v, fmt.Sprintf("scopes[%s]", key), true, func(s string) {
				result.Values[string(key)] = s
			})
		}
	})
	return result
}

func parseSecurityRequirements(val *fastjson.Value, parser *Parser) SecurityRequirements {
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
	sec := make(SecurityRequirements, obj.Len())
	obj.Visit(func(key []byte, v *fastjson.Value) {
		parser.currentLoc = fmt.Sprintf("%s.%s", fromLoc, key)
		if secVals, e := v.Array(); e != nil {
			parser.appendError(fmt.Errorf("invalid value: %w", e))
		} else {
			secLoc := parser.currentLoc
			for i, secVal := range secVals {
				parser.currentLoc = fmt.Sprintf("%s[%d]", secLoc, i)
				parser.parseString(secVal, "security scheme", true, func(s string) {
					keyStr := string(key)
					sec[keyStr] = append(sec[keyStr], s)
				})
			}
		}
	})
	return sec
}
