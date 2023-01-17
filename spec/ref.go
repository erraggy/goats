package spec

// Reference a JSON reference link
// https://swagger.io/specification/v2/#reference-object
type Reference struct {
	uri string
}

// NewRef will return a *Reference for the specified URI
func NewRef(uri string) *Reference {
	return &Reference{uri: uri}
}

// URI is the link
func (r *Reference) URI() string {
	if r == nil {
		return ""
	}
	return r.uri
}
