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

// NewUIWriter TODO
func NewUIWriter(o *Options, errorWriter bool) *UIWriter {
	return &UIWriter{
		errorWriter: errorWriter,
		TxtView:     o.TextView,
	}
}
