package spec

import (
	"strings"

	"github.com/valyala/fastjson"
)

// Extensions defines a map of keys prefixed with 'x-' and any type of value
type Extensions map[string]*fastjson.Value

func (exts Extensions) marshalExtensions(val *fastjson.Value) {
	for k, v := range exts {
		if strings.HasPrefix(k, "x-") {
			val.Set(k, v)
		}
	}
}
