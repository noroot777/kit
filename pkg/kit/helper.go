package kit

import (
	ui "github.com/noroot777/clui"
)

// TextViewWriter write to TextViewWriter
type TextViewWriter struct {
	TxtView *ui.TextView
}

func (t *TextViewWriter) Write(p []byte) (n int, err error) {
	t.TxtView.AddText([]string{string(p)})
	return len(p), nil
}

// NewTextViewWriter TODO
func NewTextViewWriter(o *KitOptions) *TextViewWriter {
	return &TextViewWriter{
		TxtView: o.TextView,
	}
}
