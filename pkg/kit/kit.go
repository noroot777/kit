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
	// kit options
	opts KitOptions
)

// KitOptions TODO
type KitOptions struct {
	Namespace string // TODO apply -f, there are multiple namespaces
	Objects   []*resource.Info
	ClientSet kubernetes.Interface
	stopper   chan struct{}
	focusOn   FocusOn
}

// UI start a interactive term ui
func UI(o KitOptions) {
	ui.InitLibrary()
	ui.SetThemePath(".")
	ui.SetCurrentTheme("kitui")
	defer ui.DeinitLibrary()

	opts = o
	eventsReader := watchEvents()
	createView(eventsReader)

	ui.MainLoop()
}

// TODO set cursor focus
func createView(eventsReader chan *corev1.Event) {
	w, h := termbox.Size()
	view := ui.AddWindow(0, 0, w, h, "Kubectl Interactive Tool")
	view.SetBorder(ui.BorderThin)
	// view.SetBackColor(termbox.ColorDarkGray)

	frmLeft := ui.CreateFrame(view, 50, ui.AutoSize, ui.BorderThin, ui.Fixed)
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
	frmRight.SetPaddings(1, 1)
	frmRight.SetTitle(" Kubernetes Events ")

	// radio to select focus on
	frmRadio := ui.CreateFrame(frmRight, ui.AutoSize, ui.AutoSize, ui.BorderThin, ui.AutoSize)
	frmRadio.SetTitle(" Choose Events Scope ")
	frmRadio.SetPack(ui.Horizontal)
	frmRadio.SetPaddings(5, 1)
	frmRadio.SetAlign(ui.AlignRight)
	radioGroup := ui.CreateRadioGroup()
	radioCR := ui.CreateRadio(frmRadio, ui.AutoSize, "Current related objects", 1)
	radioCR.SetSelected(true)
	radioCR.SetActive(false)
	radioGroup.AddItem(radioCR)
	radioCN := ui.CreateRadio(frmRadio, ui.AutoSize, "Current namespace", 1)
	radioCN.SetSelected(false)
	radioCN.SetActive(false)
	radioGroup.AddItem(radioCN)
	radioAN := ui.CreateRadio(frmRadio, ui.AutoSize, "All namespaces", 1)
	radioAN.SetSelected(false)
	radioAN.SetActive(false)
	radioGroup.AddItem(radioAN)
	radioCR.OnActive(func(active bool) {
		if active {
			changeFocus(FocusOnCurrentRelated)
		}
	})
	radioCN.OnActive(func(active bool) {
		if active {
			changeFocus(FocusOnCurrentNamespace)
		}
	})
	radioAN.OnActive(func(active bool) {
		if active {
			changeFocus(FocusOnAllNamespace)
		}
	})

	// table show events
	frmEvents := ui.CreateFrame(frmRight, ui.AutoSize, ui.AutoSize, ui.BorderThin, ui.AutoSize)
	frmEvents.SetTitle(" Events List ")
	tabEvents := ui.CreateTableView(frmEvents, ui.AutoSize, ui.AutoSize, 1)
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

	frmRightBottom := ui.CreateFrame(frmRight, ui.AutoSize, 10, ui.BorderThin, ui.Fixed)
	frmRightBottom.SetTitle(" Events Detail ")
	// text show event detail
	txtEvent := ui.CreateTextView(frmRightBottom, ui.AutoSize, ui.AutoSize, 1)
	txtEvent.SetShowScroll(false)
	txtEvent.SetWordWrap(true)
	txtEvent.SetActive(false)

	tabEvents.OnSelectCell(func(column int, row int) {
		txtEvent.SetText(text(values[row*col : (row+1)*col]))
	})

	go func() {
		for {
			select {
			case event, _ := <-eventsReader:
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

func changeFocus(f FocusOn) {
	values = []string{}

}

func drawCell(info *ui.ColumnDrawInfo) {
	info.Text = values[info.Row*col+info.Col]
}

func watchEvents() chan *corev1.Event {
	var siOpts []informers.SharedInformerOption
	if opts.focusOn != FocusOnAllNamespace {
		siOpts = append(siOpts, informers.WithNamespace(opts.Namespace))
	}
	f := informers.NewSharedInformerFactoryWithOptions(opts.ClientSet, 0, siOpts...)
	informer := f.Core().V1().Events().Informer()
	defer runtime.HandleCrash()

	go f.Start(opts.stopper)

	if !cache.WaitForCacheSync(opts.stopper, informer.HasSynced) {
		fmt.Print("Timed out waiting for caches to sync")
	}

	var reader = make(chan *corev1.Event)
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

func watchObjects() chan string {
	reader := make(chan string)

	// 寻找各个resource之间的关联，获取其id
	// 找出存在status的resource，并标注哪些状态是成功状态
	for _, obj := range opts.Objects {
		if obj.Mapping.GroupVersionKind.Kind == "namespace" {

		}

	}

	return reader
}
