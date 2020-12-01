package kit

import (
	ui "github.com/noroot777/clui"
)

// TextViewWriter write to TextViewWriter
type TextViewWriter struct {
	errorWriter bool
	TxtView     *ui.TextView
}

func (t *TextViewWriter) Write(p []byte) (n int, err error) {
	if t.errorWriter {
		t.TxtView.SetTextColor(ui.ColorRed)
		t.TxtView.AddText([]string{"  ✖️  " + string(p)})
	} else {
		t.TxtView.AddText([]string{"  ✅  " + string(p)})
	}
	return len(p), nil
}

// NewTextViewWriter TODO
func NewTextViewWriter(o *KitOptions, errorWriter bool) *TextViewWriter {
	return &TextViewWriter{
		errorWriter: errorWriter,
		TxtView:     o.TextView,
	}
}
