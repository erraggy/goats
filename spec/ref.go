package spec

import "strings"

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

// definitionKey returns the definition name portion of the URI and if it is a definition key
func (r *Reference) definitionKey() (string, bool) {
	full := r.URI()
	if full == "" {
		return "", false
	}
	frag := strings.TrimPrefix(full, "#/definitions/")
	return frag, frag != full
}

type UniqueDefinitionRefs struct {
	unique map[string]struct{}
	refs   []string
}

func NewUniqueDefinitionRefs(cap int) *UniqueDefinitionRefs {
	return &UniqueDefinitionRefs{
		unique: make(map[string]struct{}, cap),
		refs:   make([]string, 0, cap),
	}
}

func (u *UniqueDefinitionRefs) Values() []string {
	if u == nil {
		return nil
	}
	results := make([]string, len(u.unique))
	for i := range u.refs {
		results[i] = u.refs[i]
	}
	return results
}

func (u *UniqueDefinitionRefs) AddRefs(refs ...*Reference) {
	if u == nil {
		return
	}
	for _, ref := range refs {
		if k, ok := ref.definitionKey(); ok {
			if _, exists := u.unique[k]; !exists {
				u.unique[k] = struct{}{}
				u.refs = append(u.refs, k)
			}
		}
	}
}

func (u *UniqueDefinitionRefs) addRefStrings(refs []string) {
	if u == nil {
		return
	}
	for _, ref := range refs {
		if _, exists := u.unique[ref]; !exists {
			u.unique[ref] = struct{}{}
			u.refs = append(u.refs, ref)
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
	result := NewUniqueDefinitionRefs(len(u.unique) + len(other.unique))
	result.addRefStrings(u.refs)
	result.addRefStrings(other.refs)
	return result
}
