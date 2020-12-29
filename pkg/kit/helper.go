package kit

import (
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	mapset "github.com/deckarep/golang-set"
	ui "github.com/noroot777/clui"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
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
	events01 []*corev1.Event // for tab 1: Involved Objects
	events03 []*corev1.Event // for tab 3: all namespaces

	visitedSet mapset.Set
	// row numbers set of visited
	visitedRows01 mapset.Set // for tab 1: Involved Objects
	visitedRows03 mapset.Set // for tab 3: all namespaces

	// event's latest resourceVersion
	resourceVersionOfInvolvedNS, resourceVersionOfAllNS string

	// involved namespace set
	involvedNamespaces mapset.Set

	// event names that kit recorded in Activity Window
	recordedEvents mapset.Set
}

// NewCurrent creates a new Current
func NewCurrent() *Current {
	curr := &Current{
		// row numbers map of visited
		visitedRows01:      mapset.NewSet(),
		visitedRows03:      mapset.NewSet(),
		involvedNamespaces: mapset.NewSet(),
	}
	// default selected radio is No.1
	curr.SetSelectedRadio(FocusOnInvolved)
	// curr.namespace.Add(ns)
	curr.recordedEvents = mapset.NewSet()
	return curr
}

// Events return the current selection Events slice
func (c *Current) Events() []*corev1.Event {
	switch c.selectedRadio {
	case FocusOnInvolved:
		return c.events01
	case FocusOnAllNamespace:
		return c.events03
	}
	return nil
}

// AppendEvent append event
func (c *Current) AppendEvent(event *corev1.Event) {
	switch c.selectedRadio {
	case FocusOnInvolved:
		c.events01 = append([]*corev1.Event{event}, c.events01...)
	case FocusOnAllNamespace:
		c.events03 = append([]*corev1.Event{event}, c.events03...)
	}
}

// SetVersion set the resource version
func (c *Current) SetVersion(version string) {
	switch c.selectedRadio {
	case FocusOnInvolved:
		if curr.resourceVersionOfInvolvedNS < version {
			curr.resourceVersionOfInvolvedNS = version
		}
	case FocusOnAllNamespace:
		if curr.resourceVersionOfAllNS < version {
			curr.resourceVersionOfAllNS = version
		}
	}
}

// Version return the resource version
func (c *Current) Version() string {
	switch c.selectedRadio {
	case FocusOnInvolved:
		return curr.resourceVersionOfInvolvedNS
	case FocusOnAllNamespace:
		return curr.resourceVersionOfAllNS
	}
	return "0"
}

// AddNamespace set the namespace
func (c *Current) AddNamespace(ns string) {
	curr.involvedNamespaces.Add(ns)
}

// Namespace return the namespace and isall flag
func (c *Current) Namespace() (mapset.Set, bool) {
	switch c.selectedRadio {
	case FocusOnInvolved:
		return curr.involvedNamespaces, false
	case FocusOnAllNamespace:
		return nil, true
	}
	return nil, true
}

// InitVersions init the versions
func (c *Current) InitVersions(resourceVersionOfInvolvedNS string, resourceVersionAllNamespace string) {
	c.resourceVersionOfInvolvedNS = resourceVersionOfInvolvedNS
	c.resourceVersionOfAllNS = resourceVersionAllNamespace
}

// SetSelectedRadio set which radio selected now
func (c *Current) SetSelectedRadio(currentRadio FocusOn) {
	c.selectedRadio = currentRadio

	switch currentRadio {
	case FocusOnInvolved:
		c.visitedSet = c.visitedRows01
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
	case FocusOnAllNamespace:
		c.visitedRows03 = mapset.NewSet(rows...)
		c.visitedSet = c.visitedRows03
	}
}

// GoID GoID
func GoID() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
	}
	return id
}

// KitFatal prints the message (if provided) and then exits. If V(6) or greater,
// klog.Fatal is invoked for extended information.
func KitFatal(msg string, code int) {
	if klog.V(6).Enabled() {
		klog.FatalDepth(2, msg)
	}
	if len(msg) > 0 {
		// add newline if needed
		if !strings.HasSuffix(msg, "\n") {
			msg += "\n"
		}
		fmt.Fprint(opts.errorWriter, msg)
	}
	// os.Exit(code)
}
