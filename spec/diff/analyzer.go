package diff

import (
	"errors"
	"fmt"

	"github.com/erraggy/goats/spec"
)

// Analyze will analyze the differences between 2 swagger specs in JSON format
func Analyze(fromSpecJSON, toSpecJSON []byte) (*Report, error) {
	var errs []error
	if len(fromSpecJSON) == 0 {
		errs = append(errs, errors.New("diff: fromSpecJSON must not be nil or empty"))
	}
	if len(toSpecJSON) == 0 {
		errs = append(errs, errors.New("diff: toSpecJSON must not be nil or empty"))
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	var (
		fromParser, toParser = spec.NewParser(fromSpecJSON), spec.NewParser(toSpecJSON)
		fromSwag, toSwag     *spec.Swagger
		err                  error
	)
	if fromSwag, err = fromParser.Parse(); err != nil {
		errs = append(errs, err)
	}
	if toSwag, err = toParser.Parse(); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	// no more errors possible

	report := NewReport()
	report.Changes[ClassRoot] = analyzeRoot(fromSwag, toSwag)
	report.Changes[ClassInfo] = analyzeInfo(&fromSwag.Info, &toSwag.Info)

	// TODO: redo paths to gather all changes for paths and path items
	report.Changes[ClassPaths] = analyzePaths(&fromSwag.Paths, &toSwag.Paths)

	// TODO: redo operations to gather all changes by spec.OperationKey
	report.Changes[ClassOperation] = analyzeOperations(fromSwag.Paths.Items, toSwag.Paths.Items)

	// TODO: still need to

	return report, nil
}

func analyzeOperations(fromPaths, toPaths map[string]*spec.PathItem) changesByLocation {
	changes := make(changesByLocation)
	set := make(map[spec.OperationKey][2]*spec.Operation, len(fromPaths))
	// initialize with our from values
	for path := range fromPaths {
		pi := fromPaths[path]
		if pi == nil {
			continue
		}
		if op := pi.Get; op != nil {
			set[op.Key] = [2]*spec.Operation{op, nil}
		}
		if op := pi.Put; op != nil {
			set[op.Key] = [2]*spec.Operation{op, nil}
		}
		if op := pi.Post; op != nil {
			set[op.Key] = [2]*spec.Operation{op, nil}
		}
		if op := pi.Delete; op != nil {
			set[op.Key] = [2]*spec.Operation{op, nil}
		}
		if op := pi.Options; op != nil {
			set[op.Key] = [2]*spec.Operation{op, nil}
		}
		if op := pi.Head; op != nil {
			set[op.Key] = [2]*spec.Operation{op, nil}
		}
		if op := pi.Patch; op != nil {
			set[op.Key] = [2]*spec.Operation{op, nil}
		}
		// take care of path item extensions in this loop
		if toPathItem, found := toPaths[path]; found {
			diffExtensions(fmt.Sprintf(".paths[%s]", path), changes, pi.Extensions, toPathItem.Extensions)
		}
	}
	// update with our to values
	for path := range toPaths {
		pi := toPaths[path]
		if pi == nil {
			continue
		}
		if op := pi.Get; op != nil {
			if tuple, exists := set[op.Key]; exists {
				tuple[1] = op
				set[op.Key] = tuple
			} else {
				set[op.Key] = [2]*spec.Operation{nil, op}
			}
		}
		if op := pi.Put; op != nil {
			if tuple, exists := set[op.Key]; exists {
				tuple[1] = op
				set[op.Key] = tuple
			} else {
				set[op.Key] = [2]*spec.Operation{nil, op}
			}
		}
		if op := pi.Post; op != nil {
			if tuple, exists := set[op.Key]; exists {
				tuple[1] = op
				set[op.Key] = tuple
			} else {
				set[op.Key] = [2]*spec.Operation{nil, op}
			}
		}
		if op := pi.Delete; op != nil {
			if tuple, exists := set[op.Key]; exists {
				tuple[1] = op
				set[op.Key] = tuple
			} else {
				set[op.Key] = [2]*spec.Operation{nil, op}
			}
		}
		if op := pi.Options; op != nil {
			if tuple, exists := set[op.Key]; exists {
				tuple[1] = op
				set[op.Key] = tuple
			} else {
				set[op.Key] = [2]*spec.Operation{nil, op}
			}
		}
		if op := pi.Head; op != nil {
			if tuple, exists := set[op.Key]; exists {
				tuple[1] = op
				set[op.Key] = tuple
			} else {
				set[op.Key] = [2]*spec.Operation{nil, op}
			}
		}
		if op := pi.Patch; op != nil {
			if tuple, exists := set[op.Key]; exists {
				tuple[1] = op
				set[op.Key] = tuple
			} else {
				set[op.Key] = [2]*spec.Operation{nil, op}
			}
		}
	}
	for opKey, fromAndTo := range set {
		fromOp, toOp := fromAndTo[0], fromAndTo[1]
		if toOp == nil {
			// Operation removed
			c := Change{
				FieldLocation: fromOp.DocumentLocation(),
				FieldName:     opKey.String(),
				OldValue:      "TODO: implement something to show here",
				Operation:     OpItemRemoved,
				Class:         ClassOperation,
			}
			changes.add(c)
			continue
		}
		if fromOp == nil {
			// Operation added
			c := Change{
				FieldLocation: toOp.DocumentLocation(),
				FieldName:     opKey.String(),
				NewValue:      "TODO: implement something to show here",
				Operation:     OpItemAdded,
				Class:         ClassOperation,
			}
			changes.add(c)
			continue
		}
		// TODO: Implement operation change reporting
		if fromOp != toOp {
			// Operation changed
			c := Change{
				FieldLocation: toOp.DocumentLocation(),
				FieldName:     opKey.String(),
				OldValue:      "TODO: implement something to show here",
				NewValue:      "TODO: implement something to show here",
				Operation:     OpUpdate,
				Class:         ClassOperation,
			}
			changes.add(c)
			continue
		}
	}
	return changes
}

func analyzePaths(fromPaths, toPaths *spec.Paths) changesByLocation {
	changes := make(changesByLocation)
	added, removed := diffStringMapKeys(fromPaths.Items, toPaths.Items)
	for _, path := range added {
		c := Change{
			FieldLocation: ".paths",
			FieldName:     "paths",
			NewValue:      path,
			Operation:     OpItemAdded,
			Class:         ClassPaths,
		}
		changes.add(c)
	}
	for _, path := range removed {
		c := Change{
			FieldLocation: ".paths",
			FieldName:     "paths",
			OldValue:      path,
			Operation:     OpItemRemoved,
			Class:         ClassPaths,
		}
		changes.add(c)
	}
	diffExtensions(".paths", changes, fromPaths.Extensions, toPaths.Extensions)
	return changes
}

type changesByLocation map[string][]Change

func (changes changesByLocation) add(c Change) {
	changes[c.FieldLocation] = append(changes[c.FieldLocation], c)
}

func analyzeRoot(fromSwag, toSwag *spec.Swagger) changesByLocation {
	changes := make(changesByLocation)
	if fromSwag.Host != toSwag.Host {
		c := Change{
			FieldLocation: ".host",
			FieldName:     "host",
			OldValue:      fromSwag.Host,
			NewValue:      toSwag.Host,
			Class:         ClassRoot,
		}
		switch {
		case fromSwag.Host == "":
			c.Operation = OpAdd
		case toSwag.Host == "":
			c.Operation = OpRemove
		default:
			c.Operation = OpUpdate
		}
		changes.add(c)
	}
	if fromSwag.BasePath != toSwag.BasePath {
		c := Change{
			FieldLocation: ".basePath",
			FieldName:     "basePath",
			OldValue:      fromSwag.BasePath,
			NewValue:      toSwag.BasePath,
			Class:         ClassRoot,
		}
		switch {
		case fromSwag.BasePath == "":
			c.Operation = OpAdd
		case toSwag.BasePath == "":
			c.Operation = OpRemove
		default:
			c.Operation = OpUpdate
		}
		changes.add(c)
	}
	if added, removed := diffStringSlice(fromSwag.Schemes, toSwag.Schemes); len(added) > 0 || len(removed) > 0 {
		mkchg := func(op Op, v string) Change {
			c := Change{
				FieldLocation: ".schemes",
				FieldName:     "schemes",
				Class:         ClassRoot,
				Operation:     op,
			}
			if op == OpItemAdded {
				c.NewValue = v
			} else if op == OpItemRemoved {
				c.OldValue = v
			}
			return c
		}
		for _, s := range added {
			c := mkchg(OpItemAdded, s)
			changes.add(c)
		}
		for _, s := range removed {
			c := mkchg(OpItemRemoved, s)
			changes.add(c)
		}
	}
	if added, removed := diffStringSlice(fromSwag.Consumes, toSwag.Consumes); len(added) > 0 || len(removed) > 0 {
		mkchg := func(op Op, v string) Change {
			c := Change{
				FieldLocation: ".consumes",
				FieldName:     "consumes",
				Class:         ClassRoot,
				Operation:     op,
			}
			if op == OpItemAdded {
				c.NewValue = v
			} else if op == OpItemRemoved {
				c.OldValue = v
			}
			return c
		}
		for _, s := range added {
			c := mkchg(OpItemAdded, s)
			changes.add(c)
		}
		for _, s := range removed {
			c := mkchg(OpItemRemoved, s)
			changes.add(c)
		}
	}
	if added, removed := diffStringSlice(fromSwag.Produces, toSwag.Produces); len(added) > 0 || len(removed) > 0 {
		mkchg := func(op Op, v string) Change {
			c := Change{
				FieldLocation: ".produces",
				FieldName:     "produces",
				Class:         ClassRoot,
				Operation:     op,
			}
			if op == OpItemAdded {
				c.NewValue = v
			} else if op == OpItemRemoved {
				c.OldValue = v
			}
			return c
		}
		for _, s := range added {
			c := mkchg(OpItemAdded, s)
			changes.add(c)
		}
		for _, s := range removed {
			c := mkchg(OpItemRemoved, s)
			changes.add(c)
		}
	}
	diffExtensions("", changes, fromSwag.Extensions, toSwag.Extensions)

	return changes
}

func analyzeInfo(fromInfo, toInfo *spec.Info) changesByLocation {
	changes := make(changesByLocation)
	if fromInfo.Title != toInfo.Title {
		c := Change{
			FieldLocation: ".info.title",
			FieldName:     "title",
			OldValue:      fromInfo.Title,
			NewValue:      toInfo.Title,
			Operation:     OpUpdate,
			Class:         ClassInfo,
		}
		changes.add(c)
	}
	if fromInfo.Description != toInfo.Description {
		c := Change{
			FieldLocation: ".info.description",
			FieldName:     "description",
			OldValue:      fromInfo.Description,
			NewValue:      toInfo.Description,
			Operation:     OpUpdate,
			Class:         ClassInfo,
		}
		changes.add(c)
	}
	if fromInfo.TermsOfService != toInfo.TermsOfService {
		c := Change{
			FieldLocation: ".info.termsOfService",
			FieldName:     "termsOfService",
			OldValue:      fromInfo.TermsOfService,
			NewValue:      toInfo.TermsOfService,
			Operation:     OpUpdate,
			Class:         ClassInfo,
		}
		changes.add(c)
	}
	if fromInfo.Version != toInfo.Version {
		c := Change{
			FieldLocation: ".info.version",
			FieldName:     "version",
			OldValue:      fromInfo.Version,
			NewValue:      toInfo.Version,
			Operation:     OpUpdate,
			Class:         ClassInfo,
		}
		changes.add(c)
	}
	if fromInfo.Contact != toInfo.Contact {
		if fromInfo.Contact == nil {
			// Contact added
			c := Change{
				FieldLocation: ".info.contact",
				FieldName:     "contact",
				NewValue:      toInfo.Contact.String(),
				Operation:     OpUpdate,
				Class:         ClassInfo,
			}
			changes.add(c)
		} else {
			if fromInfo.Contact.Name != toInfo.Contact.Name {
				c := Change{
					FieldLocation: ".info.contact.name",
					FieldName:     "name",
					OldValue:      fromInfo.Contact.Name,
					NewValue:      toInfo.Contact.Name,
					Operation:     OpUpdate,
					Class:         ClassInfo,
				}
				changes.add(c)
			}
			if fromInfo.Contact.Email != toInfo.Contact.Email {
				c := Change{
					FieldLocation: ".info.contact.email",
					FieldName:     "email",
					OldValue:      fromInfo.Contact.Email,
					NewValue:      toInfo.Contact.Email,
					Operation:     OpUpdate,
					Class:         ClassInfo,
				}
				changes.add(c)
			}
			if fromInfo.Contact.URL != toInfo.Contact.URL {
				c := Change{
					FieldLocation: ".info.contact.url",
					FieldName:     "url",
					OldValue:      fromInfo.Contact.URL,
					NewValue:      toInfo.Contact.URL,
					Operation:     OpUpdate,
					Class:         ClassInfo,
				}
				changes.add(c)
			}
			diffExtensions(".info.contact", changes, fromInfo.Contact.Extensions, toInfo.Contact.Extensions)
		}
	}
	if fromInfo.License != toInfo.License {
		if fromInfo.License == nil {
			// License added
			c := Change{
				FieldLocation: ".info.license",
				FieldName:     "license",
				NewValue:      toInfo.License.String(),
				Operation:     OpUpdate,
				Class:         ClassInfo,
			}
			changes.add(c)
		} else {
			if fromInfo.License.Name != toInfo.License.Name {
				c := Change{
					FieldLocation: ".info.license.name",
					FieldName:     "name",
					OldValue:      fromInfo.License.Name,
					NewValue:      toInfo.License.Name,
					Operation:     OpUpdate,
					Class:         ClassInfo,
				}
				changes.add(c)
			}
			if fromInfo.License.URL != toInfo.License.URL {
				c := Change{
					FieldLocation: ".info.license.url",
					FieldName:     "url",
					OldValue:      fromInfo.License.URL,
					NewValue:      toInfo.License.URL,
					Operation:     OpUpdate,
					Class:         ClassInfo,
				}
				changes.add(c)
			}
			diffExtensions(".info.license", changes, fromInfo.License.Extensions, toInfo.License.Extensions)
		}
	}
	diffExtensions(".info", changes, fromInfo.Extensions, toInfo.Extensions)
	return changes
}

func diffExtensions(baseLoc string, changes changesByLocation, fromExt, toExt spec.Extensions) {
	for k, v1 := range fromExt {
		value1 := v1.String()
		if v2, found := toExt[k]; found {
			value2 := v2.String()
			if value1 != value2 {
				c := Change{
					FieldLocation: fmt.Sprintf("%s.%s", baseLoc, k),
					FieldName:     k,
					OldValue:      value1,
					NewValue:      value2,
					Operation:     OpUpdate,
					Class:         ClassRoot,
				}
				changes[c.FieldLocation] = append(changes[c.FieldLocation], c)
			}
			continue
		}
		c := Change{
			FieldLocation: fmt.Sprintf("%s.%s", baseLoc, k),
			FieldName:     k,
			OldValue:      value1,
			Operation:     OpItemRemoved,
			Class:         ClassRoot,
		}
		changes[c.FieldLocation] = append(changes[c.FieldLocation], c)
	}
	for k, v1 := range toExt {
		if _, found := fromExt[k]; found {
			// updates are caught in the fromExt loop above
			continue
		} else {
			c := Change{
				FieldLocation: fmt.Sprintf("%s.%s", baseLoc, k),
				FieldName:     k,
				NewValue:      v1.String(),
				Operation:     OpItemAdded,
				Class:         ClassRoot,
			}
			changes[c.FieldLocation] = append(changes[c.FieldLocation], c)
		}
	}
}

func diffStringMapKeys[V any](from, to map[string]V) (added []string, removed []string) {
	// load removed
	for k := range from {
		if _, kept := to[k]; !kept {
			removed = append(removed, k)
		}
	}
	// load added
	for k := range to {
		if _, exists := from[k]; !exists {
			added = append(added, k)
		}
	}
	return added, removed
}

func diffStringSlice(from, to []string) (added []string, removed []string) {
	unique := make(map[string]struct{}, len(from))

	// first load from
	for _, s := range from {
		unique[s] = struct{}{}
	}
	// look for added
	for _, s := range to {
		if _, existed := unique[s]; !existed {
			added = append(added, s)
		}
	}

	// now reset unique
	unique = make(map[string]struct{}, len(to))
	// load to
	for _, s := range to {
		unique[s] = struct{}{}
	}
	// look for removed
	for _, s := range from {
		if _, kept := unique[s]; !kept {
			removed = append(removed, s)
		}
	}

	return added, removed
}
