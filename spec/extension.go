package spec

import "github.com/valyala/fastjson"

// Extensions defines a map of keys prefixed with 'x-' and any type of value
type Extensions map[string]*fastjson.Value
