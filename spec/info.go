package spec

import (
	"bytes"
	"fmt"

	"github.com/valyala/fastjson"
)

// Info represents the swagger .info object
// https://swagger.io/specification/v2/#info-object
type Info struct {
	Extensions
	Title          string
	Description    string
	TermsOfService string
	Version        string
	Contact        *Contact
	License        *License
}

// NewInfo returns a new Info
func NewInfo() *Info {
	return &Info{
		Extensions: make(Extensions),
	}
}

// Contact represents the swagger .info.contact object
// https://swagger.io/specification/v2/#contact-object
type Contact struct {
	Extensions
	Name  string
	URL   string
	Email string
}

// NewContact returns a new Contact
func NewContact() *Contact {
	return &Contact{
		Extensions: make(Extensions),
	}
}

// License represents the swagger .info.license object
// https://swagger.io/specification/v2/#license-object
type License struct {
	Extensions
	Name string
	URL  string
}

// NewLicense returns a new License
func NewLicense() *License {
	return &License{
		Extensions: make(Extensions),
	}
}

// parseInfo will attempt to parse an Info from the source swagger .info JSON value
func parseInfo(infoVal *fastjson.Value, parser *Parser) *Info {
	// first be sure to capture and reset our parser's location
	fromLoc := parser.currentLoc
	defer func() {
		parser.currentLoc = fromLoc
	}()
	infoObj, err := infoVal.Object()
	if err != nil {
		parser.appendError(fmt.Errorf("invalid result value: %w", err))
		return nil
	}
	result := NewInfo()
	infoObj.Visit(func(key []byte, v *fastjson.Value) {
		parser.currentLoc = fmt.Sprintf("%s.%s", fromLoc, key)
		switch {
		case matchString(key, "title"):
			parser.parseString(v, "title", false, func(s string) {
				result.Title = s
			})
		case matchString(key, "version"):
			parser.parseString(v, "version", false, func(s string) {
				result.Version = s
			})
		case matchString(key, "description"):
			parser.parseString(v, "description", true, func(s string) {
				result.Description = s
			})
		case matchString(key, "termsOfService"):
			parser.parseString(v, "termsOfService", true, func(s string) {
				result.TermsOfService = s
			})
		case bytes.Equal(key, []byte("contact")):
			result.Contact = parseContact(v, parser)
		case bytes.Equal(key, []byte("license")):
			result.License = parseLicense(v, parser)
		case matchExtension(key):
			result.Extensions[string(key)] = v
		default:
			parser.appendError(fmt.Errorf("invalid field name: '%s'", key))
		}
	})
	return result
}

// parseContact will attempt to parse a Contact from the source swagger .info.contact JSON value
func parseContact(contactVal *fastjson.Value, parser *Parser) *Contact {
	// first be sure to capture and reset our parser's location
	fromLoc := parser.currentLoc
	defer func() {
		parser.currentLoc = fromLoc
	}()
	contactObj, err := contactVal.Object()
	if err != nil {
		parser.appendError(fmt.Errorf("invalid result value: %w", err))
		return nil
	}
	result := NewContact()
	contactObj.Visit(func(key []byte, v *fastjson.Value) {
		parser.currentLoc = fmt.Sprintf("%s.%s", fromLoc, key)
		switch {
		case matchString(key, "name"):
			parser.parseString(v, "name", true, func(s string) {
				result.Name = s
			})
		case matchString(key, "email"):
			parser.parseString(v, "email", true, func(s string) {
				result.Email = s
			})
		case matchString(key, "url"):
			parser.parseString(v, "url", true, func(s string) {
				result.URL = s
			})
		case matchExtension(key):
			result.Extensions[string(key)] = v
		default:
			parser.appendError(fmt.Errorf("invalid field name: '%s'", key))
		}
	})
	return result
}

// parseLicense will attempt to parse a License from the source swagger .info.license JSON value
func parseLicense(licenseVal *fastjson.Value, parser *Parser) *License {
	// first be sure to capture and reset our parser's location
	fromLoc := parser.currentLoc
	defer func() {
		parser.currentLoc = fromLoc
	}()
	licenseObj, err := licenseVal.Object()
	if err != nil {
		parser.appendError(fmt.Errorf("invalid result value: %w", err))
		return nil
	}
	result := NewLicense()
	licenseObj.Visit(func(key []byte, v *fastjson.Value) {
		parser.currentLoc = fmt.Sprintf("%s.%s", fromLoc, key)
		switch {
		case matchString(key, "name"):
			parser.parseString(v, "name", false, func(s string) {
				result.Name = s
			})
		case matchString(key, "url"):
			parser.parseString(v, "url", true, func(s string) {
				result.URL = s
			})
		case matchExtension(key):
			result.Extensions[string(key)] = v
		default:
			parser.appendError(fmt.Errorf("invalid field name: '%s'", key))
		}
	})
	return result
}
