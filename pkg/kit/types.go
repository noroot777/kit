package kit

import (
	"time"

	"k8s.io/cli-runtime/pkg/resource"
)

// Activity Activity
type Activity struct {
	Obj     *resource.Info
	Message []Message
}

// Message Message
type Message struct {
	Info string
	When time.Time
}

// AddMessage AddMessage
func (t *Activity) AddMessage(m Message) {
	t.Message = append(t.Message, m)
}
