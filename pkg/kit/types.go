package kit

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
	VisitedMap    map[int]int
	// row numbers map of visited
	visitedRows1 map[int]int // for tab 1: Involved Objects
	visitedRows2 map[int]int // for tab 2: Current namespace
	visitedRows3 map[int]int // for tab 3: all namespaces
}

// NewCurrent creates a new Current
func NewCurrent() *Current {
	curr := &Current{
		// row numbers map of visited
		visitedRows1: make(map[int]int),
		visitedRows2: make(map[int]int),
		visitedRows3: make(map[int]int),
	}
	// default selected radio is No.2
	curr.Set(FocusOnCurrentNamespace)
	return curr
}

// Set set which radio selected now
func (c *Current) Set(currentRadio FocusOn) {
	c.SelectedRadio = currentRadio

	switch currentRadio {
	case FocusOnInvolved:
		c.VisitedMap = c.visitedRows1
	case FocusOnCurrentNamespace:
		c.VisitedMap = c.visitedRows2
	case FocusOnAllNamespace:
		c.VisitedMap = c.visitedRows3
	}
}

func (c *Current) MoveDownOnStep() {

	// 由于tab是倒序排列，所以有新数据时，访问过的map中的key需要加1
	// 当k8schan中有值时调用此方法，此方法中将map中所有的k值加1
	// 如何动态的加1
	// for k := range c.visitedRows1 {
	// 	delete(c.visitedRows1, k)
	// 	c.visitedRows1[k+1] = k + 1
	// }
	// for k := range c.visitedRows2 {
	// 	delete(c.visitedRows2, k)
	// 	c.visitedRows2[k+1] = k + 1
	// }
	// for k := range c.visitedRows3 {
	// 	delete(c.visitedRows3, k)
	// 	c.visitedRows3[k+1] = k + 1
	// }
}
