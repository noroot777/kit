package kit

import (
	"fmt"
	"strconv"

	ui "github.com/noroot777/clui"
	termbox "github.com/nsf/termbox-go"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

var (
	// Errors errors before intercept
	Errors []error
	// events displaying in Event Table
	values []string
	// column count of Event table(include the NO. column)
	col = 6
)

// KitOptions TODO
type KitOptions struct {
	Namespace string
	Objects   []*resource.Info
	ClientSet kubernetes.Interface
	Stopper   chan struct{}
}

// UI start a interactive term ui
func UI(o KitOptions) {
	ui.InitLibrary()
	ui.SetThemePath(".")
	ui.SetCurrentTheme("kitui")
	defer ui.DeinitLibrary()

	reader := watch(o.ClientSet, o.Stopper)
	createView(reader)

	ui.MainLoop()
}

// TODO set cursor focus
func createView(reader chan *corev1.Event) {
	w, h := termbox.Size()
	view := ui.AddWindow(0, 0, w, h, "Kubectl Interactive Tool")
	view.SetBorder(ui.BorderThin)
	// view.SetBackColor(termbox.ColorDarkGray)

	frmLeft := ui.CreateFrame(view, 50, ui.AutoSize, ui.BorderThin, ui.Fixed)
	// frmLeft.SetPaddings(1, 1)
	frmLeft.SetTitle(" Activities ")

	// text show activities
	txtActivity := ui.CreateTextView(frmLeft, ui.AutoSize, ui.AutoSize, 1)
	txtActivity.SetShowScroll(false)
	txtActivity.SetWordWrap(true)
	for _, err := range Errors {
		txtActivity.AddText([]string{"⚠️ " + err.Error()})
	}
	// txtActivity.AddText([]string{" ⚠️   Namespace created!"})
	// txtActivity.AddText([]string{" ✅   Namespace created!"})
	// txtActivity.AddText([]string{" ✖️   Namespace created!"})

	frmRight := ui.CreateFrame(view, ui.AutoSize, ui.AutoSize, ui.BorderThin, ui.AutoSize)
	frmRight.SetPack(ui.Vertical)
	// frmRight.SetPaddings(1, 1)
	frmRight.SetTitle(" Kubernetes Events ")

	// table show events
	tabEvents := ui.CreateTableView(frmRight, ui.AutoSize, ui.AutoSize, 1)
	// tabEvents.SetPaddings(1, 1)
	tabEvents.SetTitle(" Event List ")
	cols := []ui.Column{
		{Title: "#", Width: 3, Alignment: ui.AlignLeft},
		{Title: "LAST_SEEN", Width: 10, Alignment: ui.AlignLeft},
		{Title: "TYPE", Width: 12, Alignment: ui.AlignLeft},
		{Title: "REASON", Width: 20, Alignment: ui.AlignLeft},
		{Title: "OBJECT", Width: 30, Alignment: ui.AlignLeft},
		{Title: "MESSAGE", Width: 100, Alignment: ui.AlignLeft},
	}
	tabEvents.SetColumns(cols)
	tabEvents.SetRowCount(0)
	tabEvents.SetShowScroll(false)
	tabEvents.SetFullRowSelect(true)
	tabEvents.SetShowLines(false)
	tabEvents.SetShowRowNumber(false)
	tabEvents.OnDrawCell(drawCell)

	frmRightBottom := ui.CreateFrame(frmRight, ui.AutoSize, 15, ui.BorderThin, ui.Fixed)
	// frmRightBottom.SetPaddings(1, 1)
	frmRightBottom.SetTitle(" Events Detail ")
	// text show event detail
	txtEvent := ui.CreateTextView(frmRightBottom, ui.AutoSize, ui.AutoSize, 1)
	txtEvent.SetShowScroll(false)
	txtEvent.SetWordWrap(true)

	tabEvents.OnSelectCell(func(column int, row int) {
		txtEvent.SetText(text(values[row*col : (row+1)*col]))
	})

	go func() {
		for {
			select {
			case event, _ := <-reader:
				values = append(
					[]string{
						strconv.Itoa(len(values)/col + 1),
						event.LastTimestamp.Format("15:04:05"),
						event.Type,
						event.Reason,
						event.InvolvedObject.Name,
						event.Message},
					values...)
				tabEvents.SetRowCount(tabEvents.RowCount() + 1)

				txtEvent.SetText(text(values[:col]))
			}
		}
	}()
}

func text(v []string) []string {
	var t = []string{
		fmt.Sprint("#        : ", v[0]),
		fmt.Sprint("LAST_SEEN: " + v[1]),
		fmt.Sprint("TYPE     : " + v[2]),
		fmt.Sprint("REASON   : " + v[3]),
		fmt.Sprint("OBJECT   : " + v[4]),
		fmt.Sprint("MESSAGE  : " + v[5]),
	}
	return t
}

func drawCell(info *ui.ColumnDrawInfo) {
	info.Text = values[info.Row*col+info.Col]
}

func watch(clientSet kubernetes.Interface, stopper chan struct{}) chan *corev1.Event {
	var reader = make(chan *corev1.Event)
	f := informers.NewSharedInformerFactoryWithOptions(clientSet, 0)
	informer := f.Core().V1().Events().Informer()
	defer runtime.HandleCrash()

	go f.Start(stopper)

	if !cache.WaitForCacheSync(stopper, informer.HasSynced) {
		fmt.Print("Timed out waiting for caches to sync")
	}

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			reader <- obj.(*corev1.Event)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			reader <- newObj.(*corev1.Event)
		},
		DeleteFunc: func(obj interface{}) {
			reader <- obj.(*corev1.Event)
		},
	})
	return reader
}
