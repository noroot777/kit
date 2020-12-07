package kit

import (
	mapset "github.com/deckarep/golang-set"
)

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

// Current for current status
type Current struct {
	SelectedRadio FocusOn
	VisitedSet    mapset.Set
	// row numbers set of visited
	visitedRows01 mapset.Set // for tab 1: Involved Objects
	visitedRows02 mapset.Set // for tab 2: Current namespace
	visitedRows03 mapset.Set // for tab 3: all namespaces
}

// NewCurrent creates a new Current
func NewCurrent() *Current {
	curr := &Current{
		// row numbers map of visited
		visitedRows01: mapset.NewSet(),
		visitedRows02: mapset.NewSet(),
		visitedRows03: mapset.NewSet(),
	}
	// default selected radio is No.2
	curr.SetSelectedRadio(FocusOnCurrentNamespace)
	return curr
}

// SetSelectedRadio set which radio selected now
func (c *Current) SetSelectedRadio(currentRadio FocusOn) {
	c.SelectedRadio = currentRadio

	switch currentRadio {
	case FocusOnInvolved:
		c.VisitedSet = c.visitedRows01
	case FocusOnCurrentNamespace:
		c.VisitedSet = c.visitedRows02
	case FocusOnAllNamespace:
		c.VisitedSet = c.visitedRows03
	}
}

// MoveEach bcz the items in tableview is reverse order, so when a
// new event comes, the items in visitset should +1.
func (c *Current) MoveEach() {
	rows := c.VisitedSet.ToSlice()
	if len(rows) == 0 {
		return
	}
	for i, row := range rows {
		rows[i] = row.(int) + 1
	}

	switch c.SelectedRadio {
	case FocusOnInvolved:
		c.visitedRows01 = mapset.NewSet(rows...)
		c.VisitedSet = c.visitedRows01
	case FocusOnCurrentNamespace:
		c.visitedRows02 = mapset.NewSet(rows...)
		c.VisitedSet = c.visitedRows02
	case FocusOnAllNamespace:
		c.visitedRows03 = mapset.NewSet(rows...)
		c.VisitedSet = c.visitedRows03
	}
}
