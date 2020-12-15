package kit

import (
	"reflect"

	mapset "github.com/deckarep/golang-set"
	ui "github.com/noroot777/clui"
)

// UIWriter write to UI
type UIWriter struct {
	Type    string
	TxtView *ui.TextView
}

func (t *UIWriter) Write(p []byte) (n int, err error) {
	switch t.Type {
	case "error":
		t.TxtView.SetTextColor(ui.ColorRed)
		t.TxtView.AddText([]string{"  ✖️  " + string(p)})
	case "normal":
		t.TxtView.AddText([]string{string(p)})
	case "correct":
		t.TxtView.AddText([]string{"  ✅  " + string(p)})
	}
	return len(p), nil
}

// NewUIWriter creates writer
func NewUIWriter(tv *ui.TextView) *UIWriter {
	w := &UIWriter{
		Type:    "correct",
		TxtView: tv,
	}
	return w
}

// NewNormalUIWriter returns a normal writer
func NewNormalUIWriter(tv *ui.TextView) *UIWriter {
	w := &UIWriter{
		Type:    "normal",
		TxtView: tv,
	}
	return w
}

// NewUIErrorWriter creates error writer
func NewUIErrorWriter(tv *ui.TextView) *UIWriter {
	w := &UIWriter{
		Type:    "error",
		TxtView: tv,
	}
	return w
}

func isNilPtr(x interface{}) bool {
	v := reflect.ValueOf(x)
	return v.Kind() == reflect.Ptr && v.IsNil()
}

// FocusOn event scope. Use to filter events
type FocusOn int

const (
	// FocusOnInvolved all events
	FocusOnInvolved FocusOn = iota
	// FocusOnCurrentNamespace current namespace's events
	FocusOnCurrentNamespace
	// FocusOnAllNamespace current objects and children's events
	FocusOnAllNamespace
)

// InterceptFunc the Interceptors's main function
type InterceptFunc func(o *Options)

// Current for current status.
// All properties in Current struct are related and only related with the radio group.
// Properties will be changed when radio group selection changed.
type Current struct {
	selectedRadio FocusOn

	// events displaying in Event TableView
	events01 []string // for tab 1: Involved Objects
	events02 []string // for tab 2: Current namespace
	events03 []string // for tab 3: all namespaces

	visitedSet mapset.Set
	// row numbers set of visited
	visitedRows01 mapset.Set // for tab 1: Involved Objects
	visitedRows02 mapset.Set // for tab 2: Current namespace
	visitedRows03 mapset.Set // for tab 3: all namespaces

	// event's resourceVersion last time
	resourceVersion, resourceVersionAllNamespace string

	namespace string
}

// NewCurrent creates a new Current
func NewCurrent(ns string) *Current {
	curr := &Current{
		// row numbers map of visited
		visitedRows01: mapset.NewSet(),
		visitedRows02: mapset.NewSet(),
		visitedRows03: mapset.NewSet(),
	}
	// default selected radio is No.2
	curr.SetSelectedRadio(FocusOnCurrentNamespace)
	curr.namespace = ns
	return curr
}

// Events return the current selection Events slice
func (c *Current) Events() []string {
	switch c.selectedRadio {
	case FocusOnInvolved:
		return c.events01
	case FocusOnCurrentNamespace:
		return c.events02
	case FocusOnAllNamespace:
		return c.events03
	}
	return nil
}

// AppendEvent append event
func (c *Current) AppendEvent(event []string) {
	switch c.selectedRadio {
	case FocusOnInvolved:
		c.events01 = append(event, c.events01...)
	case FocusOnCurrentNamespace:
		c.events02 = append(event, c.events02...)
	case FocusOnAllNamespace:
		c.events03 = append(event, c.events03...)
	}
}

// SetVersion set the resource version
func (c *Current) SetVersion(version string) {
	switch c.selectedRadio {
	case FocusOnInvolved, FocusOnCurrentNamespace:
		if curr.resourceVersion < version {
			curr.resourceVersion = version
		}
	case FocusOnAllNamespace:
		if curr.resourceVersionAllNamespace < version {
			curr.resourceVersionAllNamespace = version
		}
	}
}

// Version return the resource version
func (c *Current) Version() string {
	switch c.selectedRadio {
	case FocusOnInvolved, FocusOnCurrentNamespace:
		return curr.resourceVersion
	case FocusOnAllNamespace:
		return curr.resourceVersionAllNamespace
	}
	return "0"
}

// SetNamespace set the namespace
func (c *Current) SetNamespace(ns string) {
	curr.namespace = ns
}

// Namespace return the namespace
func (c *Current) Namespace() string {
	switch c.selectedRadio {
	case FocusOnInvolved, FocusOnCurrentNamespace:
		return curr.namespace
	case FocusOnAllNamespace:
		return ""
	}
	return ""
}

// InitVersions init the versions
func (c *Current) InitVersions(resourceVersion string, resourceVersionAllNamespace string) {
	c.resourceVersion = resourceVersion
	c.resourceVersionAllNamespace = resourceVersionAllNamespace
}

// SetSelectedRadio set which radio selected now
func (c *Current) SetSelectedRadio(currentRadio FocusOn) {
	c.selectedRadio = currentRadio

	switch currentRadio {
	case FocusOnInvolved:
		c.visitedSet = c.visitedRows01
	case FocusOnCurrentNamespace:
		c.visitedSet = c.visitedRows02
	case FocusOnAllNamespace:
		c.visitedSet = c.visitedRows03
	}
}

// SelectedRadio return selectedRadio
func (c *Current) SelectedRadio() FocusOn {
	return c.selectedRadio
}

// VisitedSet return visitedSet
func (c *Current) VisitedSet() mapset.Set {
	return c.visitedSet
}

// MoveEach bcz the items in tableview is reverse order, so when a
// new event comes, the items in visitset should +1.
func (c *Current) MoveEach() {
	rows := c.visitedSet.ToSlice()
	if len(rows) == 0 {
		return
	}
	for i, row := range rows {
		rows[i] = row.(int) + 1
	}

	switch c.selectedRadio {
	case FocusOnInvolved:
		c.visitedRows01 = mapset.NewSet(rows...)
		c.visitedSet = c.visitedRows01
	case FocusOnCurrentNamespace:
		c.visitedRows02 = mapset.NewSet(rows...)
		c.visitedSet = c.visitedRows02
	case FocusOnAllNamespace:
		c.visitedRows03 = mapset.NewSet(rows...)
		c.visitedSet = c.visitedRows03
	}
}
