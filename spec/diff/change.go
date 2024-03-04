package diff

import (
	"fmt"

	"github.com/valyala/fastjson"

	"github.com/erraggy/goats/spec"
)

// Op enumeration of the possible diff operations
type Op uint8

const (
	OpNoChange Op = iota
	OpAdd
	OpRemove
	OpUpdate
	OpItemAdded
	OpItemRemoved
)

func (o Op) String() string {
	switch o {
	case OpNoChange:
		return "unchanged"
	case OpAdd:
		return "added"
	case OpRemove:
		return "removed"
	case OpUpdate:
		return "updated"
	case OpItemAdded:
		return "item-added"
	case OpItemRemoved:
		return "item-removed"
	default:
		return ""
	}
}

type Class uint8

const (
	ClassUnknown Class = iota
	ClassRoot
	ClassInfo
	ClassDefinition
	ClassPaths
	ClassOperation
)

func (c Class) String() string {
	switch c {
	case ClassRoot:
		return "Swagger Root"
	case ClassInfo:
		return "Info"
	case ClassDefinition:
		return "Definition"
	case ClassPaths:
		return "Paths"
	case ClassOperation:
		return "Operation"
	case ClassUnknown:
		return "Unknown"
	default:
		return fmt.Sprintf("Invaild Class: %d", c)
	}
}

// Change describes a single change for a single field
type Change struct {
	FieldLocation string
	FieldName     string
	OldValue      string
	NewValue      string
	Operation     Op
	Class         Class
}

// AsJSON marshals this Change as a JSON value
func (c Change) AsJSON() *fastjson.Value {
	var a fastjson.Arena
	defer a.Reset()
	v := a.NewObject()
	v.Set("diffOperation", a.NewString(c.Operation.String()))
	v.Set("class", a.NewString(c.Class.String()))
	v.Set("from", a.NewString(c.OldValue))
	v.Set("to", a.NewString(c.NewValue))
	v.Set("location", a.NewString(c.FieldLocation))
	v.Set("name", a.NewString(c.FieldName))
	return v
}

func (c Change) String() string {
	return c.AsJSON().String()
}

// Report describes all the changes detected from an analysis
type Report struct {
	Changes            map[Class]map[string][]Change
	ChangesByOperation map[spec.OperationKey][]Change
}

func (r Report) String() string {
	if len(r.Changes) == 0 {
		return "{}"
	}

	var a fastjson.Arena
	defer a.Reset()
	rptVal := a.NewObject()
	for cls, chgByLoc := range r.Changes {
		if len(chgByLoc) > 0 {
			changesVal := a.NewObject()
			for loc, changes := range chgByLoc {
				if n := len(changes); n > 0 {
					arr := a.NewArray()
					for i, chg := range changes {
						arr.SetArrayItem(i, chg.AsJSON())
					}
					changesVal.Set(loc, arr)
				}
			}
			rptVal.Set(cls.String(), changesVal)
		}
	}
	return rptVal.String()
}

// NewReport returns an initialized Report
func NewReport() *Report {
	return &Report{Changes: make(map[Class]map[string][]Change)}
}
