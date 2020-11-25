package kit

import (
	"fmt"
	"strconv"

	ui "github.com/VladimirMarkelov/clui"
	termbox "github.com/nsf/termbox-go"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// Interact start a interactive term ui
func Interact(clientSet kubernetes.Interface, stopper chan struct{}) {
	ui.InitLibrary()
	defer ui.DeinitLibrary()

	reader := watch(clientSet, stopper)
	createView(reader)

	ui.MainLoop()

}

var values []string
var col = 6

// TODO set cursor focus
func createView(reader chan *corev1.Event) {
	w, h := termbox.Size()
	view := ui.AddWindow(0, 0, w, h, "Kubectl Interactive Tool")
	view.SetBorder(ui.BorderThin)
	// view.SetBackColor(termbox.ColorDarkGray)

	frmLeft := ui.CreateFrame(view, 30, ui.AutoSize, ui.BorderThin, ui.Fixed)
	// frmLeft.SetPaddings(1, 1)
	frmLeft.SetTitle(" Activities ")

	frmRight := ui.CreateFrame(view, ui.AutoSize, ui.AutoSize, ui.BorderThin, ui.AutoSize)
	frmRight.SetPack(ui.Vertical)
	// frmRight.SetPaddings(1, 1)
	frmRight.SetTitle(" Kubernetes Events ")

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
	tabEvents.SetFullRowSelect(true)
	tabEvents.SetShowLines(false)
	tabEvents.SetShowRowNumber(false)
	tabEvents.OnDrawCell(drawCell)

	frmRightBottom := ui.CreateFrame(frmRight, ui.AutoSize, 15, ui.BorderThin, ui.Fixed)
	// frmRightBottom.SetPaddings(1, 1)
	frmRightBottom.SetTitle(" Events Detail ")
	txtEvent := ui.CreateTextView(frmRightBottom, 120, ui.AutoSize, ui.AutoSize)
	txtEvent.SetActive(true)
	txtEvent.SetWordWrap(true)

	tabEvents.OnSelectCell(func(column int, row int) {
		txtEvent.SetText(text(values[row*col : (row+1)*col]))
	})

	go func() {
		for {
			select {
			case event, _ := <-reader:
				values = append([]string{strconv.Itoa(len(values)/col + 1), event.LastTimestamp.Format("15:04:05"), event.Type, event.Reason, event.InvolvedObject.Name, event.Message}, values...)
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
	// t[0] = fmt.Sprint("#:\t", v[0])
	// t[1] = fmt.Sprint("LAST_SEEN:\t" + v[1])
	// t[2] = fmt.Sprint("TYPE:\t" + v[2])
	// t[3] = fmt.Sprint("REASON:\t" + v[3])
	// t[4] = fmt.Sprint("OBJECT:\t" + v[4])
	// t[5] = fmt.Sprint("MESSAGE:\t" + v[5])
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
