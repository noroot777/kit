package kit

// import (
// 	"context"
// 	"fmt"
// 	"strconv"
// 	"sync"

// 	ui "github.com/noroot777/clui"

// 	termbox "github.com/nsf/termbox-go"

// 	corev1 "k8s.io/api/core/v1"
// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// 	"k8s.io/apimachinery/pkg/util/runtime"
// 	"k8s.io/cli-runtime/pkg/resource"
// 	"k8s.io/client-go/informers"
// 	"k8s.io/client-go/kubernetes"
// 	"k8s.io/client-go/tools/cache"
// )

// var (
// 	// Errors errors before intercept
// 	Errors []error
// 	// events displaying in Event Table
// 	values []string
// 	// values lock
// 	mtx sync.RWMutex
// 	// count of Event table(include the NO. column)
// 	columnCount = 6
// 	// to keep current status
// 	curr *Current
// )

// // Options TODO
// type Options struct {
// 	Namespace string // TODO apply -f, there are multiple namespaces
// 	Objects   []*resource.Info
// 	ClientSet kubernetes.Interface
// 	TextView  *ui.TextView

// 	writer       *UIWriter
// 	errorWriter  *UIWriter
// 	stopper      chan struct{}
// 	eventsReader chan *corev1.Event

// 	latestVersion, latestVersionAllNamespace string
// }

// // NewOptions create new Options
// func NewOptions(namespace string, objects []*resource.Info, clientSet *kubernetes.Clientset) *Options {
// 	return &Options{
// 		Namespace: namespace,
// 		Objects:   objects,
// 		ClientSet: clientSet,

// 		stopper:      make(chan struct{}),
// 		eventsReader: make(chan *corev1.Event, 100),
// 	}
// }

// // Intercept intercept the command exec
// func Intercept(fn InterceptFunc, o *Options) {
// 	o.stopper = make(chan struct{})
// 	defer func() { o.stopper <- struct{}{} }()
// 	curr = NewCurrent()
// 	err := latestResourceVersion(o)
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}

// 	if fn != nil {
// 		fn(o)
// 	}

// 	drawUI(o)
// }

// // drawUI draw a interactive term ui
// func drawUI(opt *Options) {
// 	ui.InitLibrary()
// 	ui.SetThemePath(".")
// 	ui.SetCurrentTheme("ui")

// 	watchEvents(opt)
// 	createView(opt)

// 	ui.RefreshScreen()
// 	// termbox.Flush()
// }

// // Hold show term ui
// func Hold() {
// 	defer ui.DeinitLibrary()
// 	ui.MainLoop()
// }

// // TODO set cursor focus
// func createView(opt *Options) {
// 	// **************
// 	// * 1. draw ui *
// 	// **************
// 	w, h := termbox.Size()
// 	view := ui.AddWindow(0, 0, w, h, " Kubectl Interactive Tool ")
// 	view.SetBorder(ui.BorderThin)
// 	view.SetPack(ui.Vertical)

// 	frmExTips := ui.CreateFrame(view, ui.AutoSize, ui.AutoSize, ui.BorderNone, ui.AutoSize)
// 	// frame include txtActivity
// 	frmLeft := ui.CreateFrame(frmExTips, 50, ui.AutoSize, ui.BorderThin, ui.Fixed)
// 	frmLeft.SetTitle(" Activities ")

// 	// TextView show activities
// 	txtActivity := ui.CreateTextView(frmLeft, ui.AutoSize, ui.AutoSize, 1)
// 	opt.TextView = txtActivity
// 	txtActivity.SetShowScroll(false)
// 	txtActivity.SetWordWrap(true)
// 	for _, err := range Errors {
// 		txtActivity.AddText([]string{"⚠️ " + err.Error()})
// 	}
// 	// txtActivity.AddText([]string{" ⚠️   Namespace created!"})
// 	// txtActivity.AddText([]string{" ✅   Namespace created!"})
// 	// txtActivity.AddText([]string{" ✖️   Namespace created!"})

// 	// frame include frmRadio and frmEvents
// 	frmRight := ui.CreateFrame(frmExTips, ui.AutoSize, ui.AutoSize, ui.BorderThin, ui.AutoSize)
// 	frmRight.SetPack(ui.Vertical)
// 	frmRight.SetPaddings(1, 1)
// 	frmRight.SetTitle(" Kubernetes Events ")

// 	// frame include radioGroup
// 	frmRadio := ui.CreateFrame(frmRight, ui.AutoSize, ui.AutoSize, ui.BorderThin, ui.AutoSize)
// 	frmRadio.SetTitle(" Choose Events Scope ")
// 	frmRadio.SetPack(ui.Horizontal)
// 	frmRadio.SetPaddings(5, 1)
// 	frmRadio.SetAlign(ui.AlignRight)
// 	// radio group to select related namespace or objects
// 	radioGroup := ui.CreateRadioGroup()
// 	radioCR := ui.CreateRadio(frmRadio, ui.AutoSize, "Involved objects", 1)
// 	radioCR.SetSelected(false)
// 	radioCR.SetActive(false)
// 	radioGroup.AddItem(radioCR)
// 	radioCN := ui.CreateRadio(frmRadio, ui.AutoSize, "Current namespace", 1)
// 	radioCN.SetSelected(true)
// 	radioCN.SetActive(false)
// 	radioGroup.AddItem(radioCN)
// 	radioAN := ui.CreateRadio(frmRadio, ui.AutoSize, "All namespaces", 1)
// 	radioAN.SetSelected(false)
// 	radioAN.SetActive(false)
// 	radioGroup.AddItem(radioAN)

// 	// frame include tabEvents
// 	frmEvents := ui.CreateFrame(frmRight, ui.AutoSize, ui.AutoSize, ui.BorderThin, ui.AutoSize)
// 	frmEvents.SetTitle(" Events List ")
// 	frmEvents.ResizeChildren()
// 	// TableView of events
// 	tabEvents := ui.CreateTableView(frmEvents, ui.AutoSize, ui.AutoSize, 1)
// 	tabEvents.SetTitle(" Event List ")
// 	cols := []ui.Column{
// 		{Title: "#", Width: 4, Alignment: ui.AlignLeft},
// 		{Title: "LAST_SEEN", Width: 10, Alignment: ui.AlignLeft},
// 		{Title: "TYPE", Width: 12, Alignment: ui.AlignLeft},
// 		{Title: "REASON", Width: 20, Alignment: ui.AlignLeft},
// 		{Title: "OBJECT", Width: 30, Alignment: ui.AlignLeft},
// 		{Title: "MESSAGE", Width: 100, Alignment: ui.AlignLeft},
// 	}
// 	tabEvents.SetColumns(cols)
// 	tabEvents.SetRowCount(0)
// 	tabEvents.SetShowScroll(false)
// 	tabEvents.SetFullRowSelect(true)
// 	tabEvents.SetShowLines(false)
// 	tabEvents.SetShowRowNumber(false)

// 	// frame include txtEvent
// 	frmRightBottom := ui.CreateFrame(frmRight, ui.AutoSize, 10, ui.BorderThin, ui.Fixed)
// 	frmRightBottom.SetTitle(" Events Detail ")
// 	frmRightBottom.SetPack(ui.Vertical)
// 	// TextView show event detail
// 	txtEvent := ui.CreateTextView(frmRightBottom, ui.AutoSize, ui.AutoSize, 1)
// 	txtEvent.SetShowScroll(false)
// 	txtEvent.SetWordWrap(true)
// 	txtEvent.SetActive(false)

// 	frmTips := ui.CreateFrame(view, ui.AutoSize, ui.AutoSize, ui.BorderNone, ui.Fixed)
// 	txtTips := ui.CreateTextView(frmTips, ui.AutoSize, 1, 1)
// 	txtTips.SetText([]string{" * ctrl+e: change interactive mode - text selection or mouse;   ctrl+q: exit;   esc: exit;   or click the top right corner to exit;"})
// 	txtTips.SetTextColor(termbox.ColorDarkGray | termbox.AttrBold)

// 	// ********************************
// 	// * 2. handlers of ui components *
// 	// ********************************

// 	// Key press handler
// 	//   1. press ctrl+e to change the input mode
// 	//      termbox.InputEsc | termbox.InputMouse: interactive mode, mouse enable, text selection disable
// 	//      termbox.InputEsc: traditional mode, mouse disable, text selection enable
// 	//   2. press ctl+q to exit ui
// 	inputEscMode := false
// 	view.OnKeyDown(func(e ui.Event, data interface{}) bool {
// 		switch e.Key {
// 		case termbox.KeyCtrlE:
// 			if inputEscMode {
// 				termbox.SetInputMode(termbox.InputEsc | termbox.InputMouse)
// 				ui.Reset()
// 				inputEscMode = false
// 			} else {
// 				termbox.SetInputMode(termbox.InputEsc)
// 				ui.Reset()
// 				inputEscMode = true
// 			}
// 		case termbox.KeyCtrlQ, termbox.KeyEsc:
// 			ui.PutEvent(ui.Event{Type: ui.EventCloseWindow})
// 		}
// 		return true
// 	}, inputEscMode)

// 	tabEvents.OnDrawCell(drawCell)

// 	tabEvents.OnSelectCell(func(selectedCol int, selectedRow int) {
// 		mtx.Lock()
// 		defer mtx.Unlock()

// 		txtEvent.SetText(text(values[selectedRow*columnCount : (selectedRow+1)*columnCount]))
// 		tabEvents.SetSelectedRow(selectedRow)
// 		curr.VisitedSet.Add(selectedRow)
// 	})

// 	radioGroup.OnSelectItem(func(c *ui.Radio) {
// 		var index = -1
// 		for i, item := range c.Parent().Children() {
// 			if item == c {
// 				index = i
// 				break
// 			}
// 		}
// 		if int(curr.SelectedRadio) == index {
// 			return
// 		}
// 		tabEvents.SetRowCount(0)
// 		switch index {
// 		case 0:
// 			changeRadioFocus(FocusOnInvolved, opt)
// 		case 1:
// 			changeRadioFocus(FocusOnCurrentNamespace, opt)
// 		case 2:
// 			changeRadioFocus(FocusOnAllNamespace, opt)
// 		case -1:
// 			return
// 		}
// 		tabEvents.Draw()
// 		txtEvent.SetText([]string{""})
// 	})

// 	// ******************************************************
// 	// * 3. catch the events from k8s and show in tabEvents *
// 	// ******************************************************
// 	go func() {
// 		for {
// 			select {
// 			case event, _ := <-opt.eventsReader:
// 				mtx.Lock()

// 				values = append(
// 					[]string{
// 						strconv.Itoa(len(values)/columnCount + 1),
// 						event.LastTimestamp.Format("15:04:05"),
// 						event.Type,
// 						event.Reason,
// 						event.InvolvedObject.Name,
// 						event.Message},
// 					values...)

// 				tabEvents.SetRowCount(tabEvents.RowCount() + 1)
// 				curr.MoveEach()
// 				// tabEvents.Draw() here is not taking effect here, so refresh ui hardly.
// 				ui.PutEvent(ui.Event{Type: ui.EventRedraw})

// 				txtEvent.SetText(text(values[:columnCount]))

// 				mtx.Unlock()
// 			}
// 		}
// 	}()

// }

// func text(v []string) []string {
// 	var t = []string{
// 		fmt.Sprint("#        : ", v[0]),
// 		fmt.Sprint("LAST_SEEN: " + v[1]),
// 		fmt.Sprint("TYPE     : " + v[2]),
// 		fmt.Sprint("REASON   : " + v[3]),
// 		fmt.Sprint("OBJECT   : " + v[4]),
// 		fmt.Sprint("MESSAGE  : " + v[5]),
// 	}
// 	return t
// }

// func changeRadioFocus(f FocusOn, opts *Options) {
// 	mtx.Lock()
// 	defer mtx.Unlock()

// 	values = values[0:0]
// 	opts.stopper <- struct{}{}
// 	curr.SetSelectedRadio(f)

// 	watchEvents(opts)
// }

// func drawCell(info *ui.ColumnDrawInfo) {
// 	mtx.Lock()
// 	defer mtx.Unlock()

// 	if len(values) > 0 {
// 		info.Text = values[info.Row*columnCount+info.Col]
// 		// set visited row's bg color
// 		if curr.VisitedSet.Contains(info.Row) {
// 			if info.RowSelected {
// 				info.Fg = ui.ColorWhiteBold
// 			} else {
// 				info.Fg = ui.ColorWhite
// 			}
// 		} else {
// 			info.Fg = ui.ColorYellowBold
// 		}
// 	}
// }

// func watchEvents(opts *Options) {
// 	var siOpts []informers.SharedInformerOption
// 	var ns, version string
// 	switch curr.SelectedRadio {
// 	case FocusOnInvolved, FocusOnCurrentNamespace:
// 		ns = opts.Namespace
// 		version = opts.latestVersion
// 	case FocusOnAllNamespace:
// 		ns = ""
// 		version = opts.latestVersionAllNamespace
// 	}
// 	tweakListOptions := func(o *metav1.ListOptions) {
// 		o.ResourceVersion = version
// 		o.ResourceVersionMatch = metav1.ResourceVersionMatchNotOlderThan
// 	}
// 	siOpts = append(siOpts, informers.WithTweakListOptions(tweakListOptions))
// 	siOpts = append(siOpts, informers.WithNamespace(ns))

// 	f := informers.NewSharedInformerFactoryWithOptions(opts.ClientSet, 0, siOpts...)
// 	informer := f.Core().V1().Events().Informer()
// 	informer.GetIndexer()
// 	defer runtime.HandleCrash()

// 	go f.Start(opts.stopper)

// 	// if !cache.WaitForCacheSync(opts.stopper, informer.HasSynced) {
// 	// 	fmt.Print("Timed out waiting for caches to sync")
// 	// }

// 	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
// 		AddFunc: func(obj interface{}) {
// 			defer func() {
// 				if err := recover(); err != nil {
// 					opts.TextView.AddText([]string{fmt.Sprintf("%v", err)})
// 				}
// 			}()
// 			// TODO filter related objects
// 			opts.eventsReader <- obj.(*corev1.Event)
// 		},
// 		UpdateFunc: func(oldObj, newObj interface{}) {
// 			defer func() {
// 				if err := recover(); err != nil {
// 					opts.TextView.AddText([]string{fmt.Sprintf("%v", err)})
// 				}
// 			}()
// 			opts.eventsReader <- newObj.(*corev1.Event)
// 		},
// 		DeleteFunc: func(obj interface{}) {
// 			defer func() {
// 				if err := recover(); err != nil {
// 					opts.TextView.AddText([]string{fmt.Sprintf("%v", err)})
// 				}
// 			}()
// 			opts.eventsReader <- obj.(*corev1.Event)
// 		},
// 	})
// }

// func watchObjects(opts Options) chan string {
// 	reader := make(chan string)

// 	var siOpts []informers.SharedInformerOption
// 	if curr.SelectedRadio != FocusOnAllNamespace {
// 		siOpts = append(siOpts, informers.WithNamespace(opts.Namespace))
// 	}
// 	f := informers.NewSharedInformerFactoryWithOptions(opts.ClientSet, 0, siOpts...)
// 	defer runtime.HandleCrash()

// 	go f.Start(opts.stopper)

// 	// 寻找各个resource之间的关联，获取其id
// 	// 找出存在status的resource，并标注哪些状态是成功状态
// 	for _, obj := range opts.Objects {
// 		// informer, err := f.ForResource(opts.Objects[0].Mapping.Resource)

// 		if obj.Mapping.GroupVersionKind.Kind == "namespace" {

// 		}

// 	}

// 	return reader
// }

// func latestResourceVersion(opts *Options) error {
// 	var latest, latestAllNamespace int

// 	eventList, err := opts.ClientSet.CoreV1().Events(opts.Namespace).List(context.TODO(), metav1.ListOptions{})
// 	if err != nil {
// 		return err
// 	}
// 	latest, err = latestVersion(eventList.Items)
// 	if err != nil {
// 		return err
// 	}

// 	eventList, err = opts.ClientSet.CoreV1().Events("").List(context.TODO(), metav1.ListOptions{})
// 	if err != nil {
// 		return err
// 	}
// 	latestAllNamespace, err = latestVersion(eventList.Items)
// 	if err != nil {
// 		return err
// 	}

// 	opts.latestVersion = strconv.Itoa(latest)
// 	opts.latestVersionAllNamespace = strconv.Itoa(latestAllNamespace)
// 	return nil
// }

// func latestVersion(events []corev1.Event) (int, error) {
// 	latest := 0
// 	for _, event := range events {
// 		version, err := strconv.Atoi(event.ResourceVersion)
// 		if err != nil {
// 			return 0, err
// 		}
// 		if latest < version {
// 			latest = version
// 		}
// 	}
// 	return latest, nil
// }
