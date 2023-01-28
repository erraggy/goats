package spec

import "strings"

// Reference a JSON reference link
// https://swagger.io/specification/v2/#reference-object
type Reference struct {
	uri    string
	docLoc string
}

// NewRef will return a *Reference for the specified URI
func NewRef(uri string, loc string) *Reference {
	return &Reference{
		uri:    uri,
		docLoc: loc,
	}
}

// DocumentLocation returns this object's JSON path location
func (r *Reference) DocumentLocation() string {
	return r.docLoc
}

// URI is the link
func (r *Reference) URI() string {
	if r == nil {
		return ""
	}
	return r.uri
}

// GatherRefs will add any definition reference keys to the specified refs
func (r *Reference) GatherRefs(refs map[string]struct{}) {
	if r == nil {
		return
	}
	if ref, ok := r.definitionKey(); ok {
		refs[ref] = struct{}{}
	}
}

// definitionKey returns the definition name portion of the URI and if it is a definition key
func (r *Reference) definitionKey() (string, bool) {
	full := r.URI()
	if full == "" {
		return "", false
	}
	frag := strings.TrimPrefix(full, "#/definitions/")
	if frag == "" {
		return "", false
	}
	return frag, frag != full
}

type UniqueDefinitionRefs struct {
	unique map[string]struct{}
}

func NewUniqueDefinitionRefs(cap int) *UniqueDefinitionRefs {
	return &UniqueDefinitionRefs{
		unique: make(map[string]struct{}, cap),
	}
}

func (u *UniqueDefinitionRefs) Len() int {
	if u == nil {
		return 0
	}
	return len(u.unique)
}

func (u *UniqueDefinitionRefs) Contains(defRef string) bool {
	if u == nil {
		return false
	}
	_, found := u.unique[defRef]
	return found
}

func (u *UniqueDefinitionRefs) Values() []string {
	if u == nil {
		return nil
	}
	var (
		results = make([]string, len(u.unique))
		i       int
	)
	for k := range u.unique {
		results[i] = k
		i++
	}
	return results
}

func (u *UniqueDefinitionRefs) AddRefs(refs ...*Reference) {
	if u == nil {
		return
	}
	for _, ref := range refs {
		if k, ok := ref.definitionKey(); ok {
			u.unique[k] = struct{}{}
		}
	}
}

func (u *UniqueDefinitionRefs) Merge(other *UniqueDefinitionRefs) *UniqueDefinitionRefs {
	if u == nil {
		return other
	}
	if other == nil {
		return u
	}

	both := make(map[string]struct{}, len(u.unique))
	for s := range u.unique {
		both[s] = struct{}{}
	}
	for s := range other.unique {
		both[s] = struct{}{}
	}

	return &UniqueDefinitionRefs{unique: both}
}
