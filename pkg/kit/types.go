package kit

import (
	"time"
)

// Activities Activities
type Activities []*Activity

// Activity Activity
type Activity struct {
	KindName string
	Msg      []Message
}

// Message Message
type Message struct {
	Info string
	When time.Time
}

// AddMessage AddMessage
func (t *Activity) AddMessage(m Message) {
	t.Msg = append(t.Msg, m)
}

// Get Get
func (t *Activities) Get(kindName string) *Activity {
	for _, a := range *t {
		if a.KindName == kindName {
			return a
		}
	}
	return nil
}

// Exists Exists
func (t *Activities) Exists(kindName string) bool {
	if t.Get(kindName) != nil {
		return true
	}
	return false
}

// GetOrNew return a new one if not exists
func (t *Activities) GetOrNew(kindName string) *Activity {
	a := t.Get(kindName)
	if a != nil {
		return a
	}
	return t.New(kindName)
}

// New New
func (t *Activities) New(kindName string) *Activity {
	n := &Activity{KindName: kindName, Msg: []Message{}}
	*t = append(*t, n)

	return n
}
