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
}

func NewUniqueDefinitionRefs(cap int) *UniqueDefinitionRefs {
	return &UniqueDefinitionRefs{
		unique: make(map[string]struct{}, cap),
	}
}

func (u *UniqueDefinitionRefs) Len() int {
	return len(u.unique)
}

func (u *UniqueDefinitionRefs) Contains(defRef string) bool {
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

func (u *UniqueDefinitionRefs) addRefStrings(refs map[string]struct{}) {
	if u == nil {
		return
	}
	for ref := range refs {
		if _, exists := u.unique[ref]; !exists {
			u.unique[ref] = struct{}{}
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
