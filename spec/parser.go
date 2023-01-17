package spec

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/valyala/fastjson"
)

// Parser handles the parsing and validation of a swagger spec
type Parser struct {
	raw                []byte
	rootVal            *fastjson.Value
	swagger            *Swagger
	errorsByLocation   map[string][]error
	uniqueOperationIDs map[string]string
	currentLoc         string
}

// NewParser returns a new parser for the specified raw swagger JSON bytes
func NewParser(raw []byte) *Parser {
	return &Parser{
		raw:                raw,
		errorsByLocation:   make(map[string][]error),
		uniqueOperationIDs: make(map[string]string),
	}
}

func (p *Parser) HasError() bool {
	if p == nil {
		return false
	}
	return len(p.errorsByLocation) == 0
}

func (p *Parser) Parse() (*Swagger, error) {
	if p == nil {
		return nil, nil
	}
	if len(p.raw) == 0 {
		return nil, errors.New("cannot parse empty raw swagger JSON bytes")
	}
	var (
		jp  fastjson.Parser
		err error
	)
	p.currentLoc = "."
	if p.rootVal, err = jp.ParseBytes(p.raw); err != nil {
		err = fmt.Errorf("failed to parse raw swagger bytes as JSON: %w", err)
		p.appendError(err)
		return nil, err
	}

	parseSwagger(p.rootVal, p)
	return p.swagger, p.Err()
}

// Err returns an aggregated error or nil if none occurred
func (p *Parser) Err() error {
	if p == nil || len(p.errorsByLocation) == 0 {
		return nil
	}
	return &ParseError{ByLocation: p.errorsByLocation}
}

func (p *Parser) locationForOperation(id string) (string, bool) {
	loc, preExisting := p.uniqueOperationIDs[id]
	if preExisting {
		return loc, preExisting
	}
	p.uniqueOperationIDs[id] = p.currentLoc
	return p.currentLoc, true
}

func (p *Parser) parseString(v *fastjson.Value, fieldName string, allowEmpty bool, accept func(s string)) {
	var validator func(string) error
	if allowEmpty {
		validator = func(s string) error {
			accept(s)
			return nil
		}
	} else {
		validator = func(s string) error {
			if s == "" {
				return fmt.Errorf("empty '%s' value", fieldName)
			}
			accept(s)
			return nil
		}
	}
	p.parseAndValidateString(v, fieldName, validator)
}

func (p *Parser) parseAndValidateString(v *fastjson.Value, fieldName string, validate func(s string) error) {
	if s, e := v.StringBytes(); e != nil {
		p.appendError(fmt.Errorf("invalid '%s' value: %w", fieldName, e))
	} else if e = validate(string(s)); e != nil {
		p.appendError(e)
	}
}

func (p *Parser) parseInt(v *fastjson.Value, fieldName string, accept func(i int)) {
	if i, e := v.Int(); e != nil {
		p.appendError(fmt.Errorf("invalid '%s' value: %w", fieldName, e))
	} else if accept != nil {
		accept(i)
	}
}

func (p *Parser) parseBool(v *fastjson.Value, fieldName string, accept func(b bool)) {
	if b, e := v.Bool(); e != nil {
		p.appendError(fmt.Errorf("invalid '%s' value: %w", fieldName, e))
	} else if accept != nil {
		accept(b)
	}
}

func (p *Parser) appendError(err error) {
	if err != nil {
		p.errorsByLocation[p.currentLoc] = append(p.errorsByLocation[p.currentLoc], err)
	}
}

type ParseError struct {
	ByLocation map[string][]error
}

func (e *ParseError) Error() string {
	if e == nil || len(e.ByLocation) == 0 {
		return ""
	}
	var (
		b       strings.Builder
		numLocs = len(e.ByLocation)
		locs    = make([]string, numLocs)
		i       int
	)
	b.WriteString("invalid swagger: found validation errors from ")
	b.WriteString(strconv.Itoa(numLocs))
	b.WriteString(" locations: {")
	for loc := range e.ByLocation {
		locs[i] = loc
		i++
	}
	sort.Strings(locs)
	for y, loc := range locs {
		if y > 0 {
			b.WriteString(", ")
		}
		b.WriteRune('"')
		b.WriteString(loc)
		b.WriteString(`": [`)
		for z, err := range e.ByLocation[loc] {
			if z > 0 {
				b.WriteString(", ")
			}
			b.WriteRune('"')
			b.WriteString(err.Error())
			b.WriteRune('"')
		}
		b.WriteRune(']')
	}
	b.WriteRune('}')
	return b.String()
}

func matchString(key []byte, match string) bool {
	return bytes.Equal(key, []byte(match))
}

func matchExtension(key []byte) bool {
	return bytes.HasPrefix(key, []byte("x-"))
}

func matchPath(key []byte) bool {
	return bytes.HasPrefix(key, []byte("/"))
}

func matchHTTPStatusCode(key []byte) bool {
	if status := bytesToInt(key); 99 < status && status < 600 {
		return true
	}
	return false
}

func bytesToInt(b []byte) int {
	i, _ := strconv.Atoi(string(b))
	return i
}
