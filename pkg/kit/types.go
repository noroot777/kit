package kit

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Activity Activity
type Activity struct {
	Obj     metav1.Object
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
