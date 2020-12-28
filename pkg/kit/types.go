package kit

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Activities Activities
type Activities []*Activity

// Activity Activity
type Activity struct {
	KindName string
	Obj      metav1.Object
	Msg      []Message
	Complete bool
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
func (t *Activities) GetOrNew(kindName string, obj metav1.Object) *Activity {
	a := t.Get(kindName)
	if a != nil {
		return a
	}
	return t.New(kindName, obj)
}

// New New
func (t *Activities) New(kindName string, obj metav1.Object) *Activity {
	n := &Activity{KindName: kindName, Obj: obj, Msg: []Message{}, Complete: false}
	*t = append(*t, n)

	return n
}
