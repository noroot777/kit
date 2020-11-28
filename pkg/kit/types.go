package kit

// FocusOn event scope. Use to filter events
type FocusOn int

const (
	// FocusOnAllNamespace all events
	FocusOnAllNamespace FocusOn = iota
	// FocusOnCurrentNamespace current namespace's events
	FocusOnCurrentNamespace
	// FocusOnCurrentRelated current objects and children's events
	FocusOnCurrentRelated
)
