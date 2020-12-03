package kit

// FocusOn event scope. Use to filter events
type FocusOn int

const (
	// FocusOnCurrentRelated all events
	FocusOnCurrentRelated FocusOn = iota
	// FocusOnCurrentNamespace current namespace's events
	FocusOnCurrentNamespace
	// FocusOnAllNamespace current objects and children's events
	FocusOnAllNamespace
)

// InterceptFunc the Interceptors's main function
type InterceptFunc func(o *Options)
