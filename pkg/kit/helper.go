package kit

import (
	ui "github.com/noroot777/clui"
)

// UIWriter write to UI
type UIWriter struct {
	errorWriter bool
	TxtView     *ui.TextView
}

func (t *UIWriter) Write(p []byte) (n int, err error) {
	if t.errorWriter {
		t.TxtView.SetTextColor(ui.ColorRed)
		t.TxtView.AddText([]string{"  ✖️  " + string(p)})
	} else {
		t.TxtView.AddText([]string{"  ✅  " + string(p)})
	}
	return len(p), nil
}

// NewUIWriter creates writer
func NewUIWriter(o *Options) *UIWriter {
	w := &UIWriter{
		errorWriter: false,
		TxtView:     o.TextView,
	}
	o.writer = w
	return w
}

// NewUIErrorWriter creates error writer
func NewUIErrorWriter(o *Options) *UIWriter {
	w := &UIWriter{
		errorWriter: true,
		TxtView:     o.TextView,
	}
	o.errorWriter = w
	return w
}
