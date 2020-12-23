package kit

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	ui "github.com/noroot777/clui"
	termbox "github.com/nsf/termbox-go"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubectl/pkg/describe"
)

var (
	// Errors errors before intercept
	Errors []error
	// values lock
	mtx sync.RWMutex
	// to keep current status
	curr *Current
	// options
	opts *Options
)

// Options TODO
type Options struct {
	ClientSet      kubernetes.Interface
	ActivityWindow *ui.TextView

	involvedObjects map[string]metav1.Object
	activities      Activities

	writer      *UIWriter
	errorWriter *UIWriter
	watcher     watch.Interface
}

// newOptions create new Options
func newOptions(clientSet *kubernetes.Clientset) *Options {
	o := &Options{
		ClientSet:       clientSet,
		involvedObjects: make(map[string]metav1.Object),
		activities:      []*Activity{},
	}
	return o
}

// HandleInfo handle with the execed info
func HandleInfo(info *resource.Info) {
	curr.AddNamespace(info.Namespace)
	metaObj := info.Object.(*unstructured.Unstructured)
	kn := info.Object.GetObjectKind().GroupVersionKind().Kind + "/" + metaObj.GetName()
	opts.involvedObjects[kn] = metaObj
	opts.activities.New(kn)
	// TODO print a message to activity view. 根据不同的命令打印不同内容，eg: apply(delete/create) imds/Deployment/imds-web
	// opts.writer.Write([]byte(fmt.Sprintf("apply %v/%v/%v", metaObj.GetNamespace(), info.Mapping.GroupVersionKind.Kind, metaObj.GetName())))
}

// Intercept intercept the kubectl command
func Intercept(fn InterceptFunc, clientSet *kubernetes.Clientset) (out io.Writer, errorOut io.Writer) {
	opts = newOptions(clientSet)
	curr = NewCurrent()
	err := initResourceVersion()
	if err != nil {
		fmt.Println(err)
		return nil, nil
	}

	if fn != nil {
		fn(opts)
	}

	drawUI()

	opts.writer = NewNormalUIWriter(opts.ActivityWindow)
	out = opts.writer
	opts.errorWriter = NewUIErrorWriter(opts.ActivityWindow)
	errorOut = opts.errorWriter

	watchEvents()

	return
}

// drawUI draw a interactive term ui
func drawUI() {
	ui.InitLibrary()
	ui.SetThemePath(".")
	ui.SetCurrentTheme("ui")

	createView()

	ui.RefreshScreen()
	// termbox.Flush()
}

// Hold show term ui
func Hold() {
	defer ui.DeinitLibrary()
	ui.MainLoop()
}

func createView() {
	// **************
	// * 1. draw ui *
	// **************
	w, h := termbox.Size()
	view := ui.AddWindow(0, 0, w, h, " Kubectl Interactive Tool ")
	view.SetBorder(ui.BorderThin)
	view.SetPack(ui.Vertical)

	frmExTips := ui.CreateFrame(view, ui.AutoSize, ui.AutoSize, ui.BorderNone, ui.AutoSize)
	// frame include txtActivity
	frmLeft := ui.CreateFrame(frmExTips, 50, ui.AutoSize, ui.BorderThin, ui.Fixed)
	frmLeft.SetTitle(" Activities ")

	// TextView show activities
	txtActivity := ui.CreateTextView(frmLeft, ui.AutoSize, ui.AutoSize, 1)
	opts.ActivityWindow = txtActivity
	txtActivity.SetShowScroll(false)
	txtActivity.SetWordWrap(true)
	for _, err := range Errors {
		txtActivity.AddText([]string{"⚠️ " + err.Error()})
	}

	// frame include frmRadio and frmEvents
	frmRight := ui.CreateFrame(frmExTips, ui.AutoSize, ui.AutoSize, ui.BorderThin, ui.AutoSize)
	frmRight.SetPack(ui.Vertical)
	frmRight.SetPaddings(1, 1)
	frmRight.SetTitle(" Kubernetes Event ")

	// frame include radioGroup
	frmRadio := ui.CreateFrame(frmRight, ui.AutoSize, ui.AutoSize, ui.BorderThin, ui.AutoSize)
	frmRadio.SetTitle(" Choose Event Scope ")
	frmRadio.SetPack(ui.Horizontal)
	frmRadio.SetPaddings(5, 1)
	frmRadio.SetAlign(ui.AlignRight)
	// radio group to select related namespace or objects
	radioGroup := ui.CreateRadioGroup()
	radioCR := ui.CreateRadio(frmRadio, ui.AutoSize, "Involved namespaces", 1)
	radioCR.SetSelected(true)
	radioCR.SetActive(false)
	radioGroup.AddItem(radioCR)
	radioAN := ui.CreateRadio(frmRadio, ui.AutoSize, "All namespaces", 1)
	radioAN.SetSelected(false)
	radioAN.SetActive(false)
	radioGroup.AddItem(radioAN)

	// frame include tabEvents
	frmEvents := ui.CreateFrame(frmRight, ui.AutoSize, ui.AutoSize, ui.BorderThin, ui.AutoSize)
	frmEvents.SetTitle(" Event List ")
	frmEvents.ResizeChildren()
	// TableView of events
	tabEvents := ui.CreateTableView(frmEvents, ui.AutoSize, ui.AutoSize, 1)
	tabEvents.SetTitle(" Event List ")
	cols := []ui.Column{
		{Title: "#", Width: 4, Alignment: ui.AlignLeft},
		{Title: "LAST_SEEN", Width: 10, Alignment: ui.AlignLeft},
		{Title: "TYPE", Width: 12, Alignment: ui.AlignLeft},
		{Title: "REASON", Width: 20, Alignment: ui.AlignLeft},
		{Title: "OBJECT", Width: 30, Alignment: ui.AlignLeft},
		{Title: "MESSAGE", Width: 100, Alignment: ui.AlignLeft},
		{Title: "NAMESPACE", Width: 10, Alignment: ui.AlignLeft},
	}
	tabEvents.SetColumns(cols)
	tabEvents.SetRowCount(0)
	tabEvents.SetShowScroll(false)
	tabEvents.SetFullRowSelect(true)
	tabEvents.SetShowLines(false)
	tabEvents.SetShowRowNumber(false)

	// frame include txtEvent
	frmRightBottom := ui.CreateFrame(frmRight, ui.AutoSize, 20, ui.BorderThin, ui.Fixed)
	frmRightBottom.SetTitle(" Event Detail ")
	frmRightBottom.SetPack(ui.Vertical)
	// TextView show event detail
	txtEvent := ui.CreateTextView(frmRightBottom, ui.AutoSize, ui.AutoSize, 1)
	txtEvent.SetShowScroll(true)
	txtEvent.SetWordWrap(true)
	txtEvent.SetActive(false)

	frmTips := ui.CreateFrame(view, ui.AutoSize, ui.AutoSize, ui.BorderNone, ui.Fixed)
	txtTips := ui.CreateTextView(frmTips, ui.AutoSize, 1, 1)
	txtTips.SetText([]string{" * ctrl+e: change interactive mode - text selection enable or mouse enable;   ctrl+q: exit;   esc: exit;   or click the top right corner to exit;"})
	txtTips.SetTextColor(termbox.ColorDarkGray | termbox.AttrBold)

	// ********************************
	// * 2. handlers of ui components *
	// ********************************

	// Key press handler
	//   1. press ctrl+e to change the input mode
	//      termbox.InputEsc | termbox.InputMouse: interactive mode, mouse enable, text selection disable
	//      termbox.InputEsc: traditional mode, mouse disable, text selection enable
	//   2. press ctl+q to exit ui
	inputEscMode := false
	view.OnKeyDown(func(e ui.Event, data interface{}) bool {
		switch e.Key {
		case termbox.KeyCtrlE:
			if inputEscMode {
				termbox.SetInputMode(termbox.InputEsc | termbox.InputMouse)
				ui.Reset()
				inputEscMode = false
			} else {
				termbox.SetInputMode(termbox.InputEsc)
				ui.Reset()
				inputEscMode = true
			}
		case termbox.KeyCtrlQ, termbox.KeyEsc:
			ui.PutEvent(ui.Event{Type: ui.EventCloseWindow})
		case 0x3F: // key:?
			ui.PutEvent(ui.Event{Type: ui.EventCloseWindow})
		default:
			fmt.Printf("%+v", e.Key)
		}
		return true
	}, inputEscMode)

	tabEvents.OnDrawCell(func(info *ui.ColumnDrawInfo) {
		// mtx.Lock()
		// defer mtx.Unlock()

		if len(curr.Events()) > 0 {
			event := curr.Events()[info.Row]
			switch info.Col {
			case 0:
				info.Text = strconv.Itoa(len(curr.Events()) - info.Row)
			case 1:
				info.Text = event.LastTimestamp.Format("15:04:05")
			case 2:
				info.Text = event.Type
			case 3:
				info.Text = event.Reason
			case 4:
				obj := event.InvolvedObject.Kind + "/" + event.InvolvedObject.Name
				if len(obj) >= 30 {
					info.Text = obj[:27] + ".. "
				} else {
					info.Text = obj
				}
			case 5:
				if len(event.Message) >= 100 {
					info.Text = event.Message[:97] + ".. "
				} else {
					info.Text = event.Message
				}
			case 6:
				info.Text = event.Namespace
			}
			// set visited row's bg color
			if curr.VisitedSet().Contains(info.Row) {
				if info.RowSelected {
					info.Fg = ui.ColorWhiteBold
				} else {
					info.Fg = ui.ColorWhite
				}
			} else {
				info.Fg = ui.ColorYellowBold
			}
		}
	})

	tabEvents.OnSelectCell(func(selectedCol int, selectedRow int) {
		// mtx.Lock()
		// defer mtx.Unlock()

		event := curr.Events()[selectedRow]
		txtEvent.SetText([]string{""})
		describeEvent(event, NewNormalUIWriter(txtEvent))

		tabEvents.SetSelectedRow(selectedRow)
		curr.VisitedSet().Add(selectedRow)
	})

	radioGroup.OnSelectItem(func(c *ui.Radio) {
		var index = -1
		for i, item := range c.Parent().Children() {
			if item == c {
				index = i
				break
			}
		}
		if int(curr.SelectedRadio()) == index {
			return
		}
		switch index {
		case 0:
			changeRadioFocus(FocusOnInvolved, tabEvents)
		case 1:
			changeRadioFocus(FocusOnAllNamespace, tabEvents)
		case -1:
			return
		}
		txtEvent.SetText([]string{""})
	})

	// ******************************************************
	// * 3. catch the events from k8s and show in tabEvents *
	// ******************************************************
	go func() {
		for {
			if opts.watcher == nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			select {
			case e, ok := <-opts.watcher.ResultChan():
				if !ok {
					continue
				}

				// mtx.Lock()
				switch e.Object.(type) {
				case *corev1.Event:
					event := e.Object.(*corev1.Event)

					// filter event by namespace manually
					ns, isAll := curr.Namespace()
					if !isAll {
						if !ns.Contains(event.Namespace) {
							continue
						}
					}

					activities(event, opts, curr)

					// do not display the old version event
					if event.ResourceVersion < curr.Version() {
						continue
					}

					curr.AppendEvent(event)
					// move down
					curr.MoveEach()
					// set the latest resource version
					curr.SetVersion(event.ResourceVersion)

					txtEvent.SetText([]string{""})
					describeEvent(event, NewNormalUIWriter(txtEvent))

					tabEvents.SetRowCount(len(curr.Events()))
					// tabEvents.Draw() here is not taking effect here, so refresh ui hardly.
					ui.PutEvent(ui.Event{Type: ui.EventRedraw})
				default:
					continue
				}

				// mtx.Unlock()
			}
		}
	}()

}

func describeEvent(e *corev1.Event, out io.Writer) {
	getter := genericclioptions.NewConfigFlags(false)

	var Describer = func(mapping *meta.RESTMapping) (describe.ResourceDescriber, error) {
		return describe.Describer(getter, mapping)
	}

	b := *resource.NewBuilder(getter)
	r := b.Unstructured().
		ContinueOnError().
		NamespaceParam(e.Namespace).DefaultNamespace().AllNamespaces(false).
		LabelSelectorParam("").
		ResourceTypeOrNameArgs(true, "event", e.Name).
		Flatten().
		Do()
	infos, _ := r.Infos()

	if len(infos) > 0 {
		mapping := infos[0].ResourceMapping()
		describer, _ := Describer(mapping)
		s, _ := describer.Describe(infos[0].Namespace, infos[0].Name, describe.DescriberSettings{ShowEvents: false})
		lines := strings.Split(s, "\n")
		for _, line := range lines {
			fmt.Fprintf(out, line)
		}
	} else {
		fmt.Fprintf(out, "No details")
	}
}

func changeRadioFocus(f FocusOn, tabEvents *ui.TableView) {
	tabEvents.SetRowCount(0)
	tabEvents.Draw()
	opts.watcher.Stop()

	curr.SetSelectedRadio(f)
	tabEvents.SetRowCount(len(curr.Events()))
	tabEvents.Draw()

	watchEvents()
}

// can not use informer, bcz `410 Gone` happened
func watchEvents() {
	version := curr.Version()
	listOpt := metav1.ListOptions{ResourceVersion: version, ResourceVersionMatch: metav1.ResourceVersionMatchNotOlderThan}
	// watch event all namespaces, and filter event by namespace manually when read the watcher chan
	watcher, err := opts.ClientSet.CoreV1().Events("").Watch(context.TODO(), listOpt)
	if err != nil {
		opts.errorWriter.Write([]byte(err.Error()))
	}
	opts.watcher = watcher
}

func initResourceVersion() error {
	var latestAllNamespace int

	eventList, err := opts.ClientSet.CoreV1().Events("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	latestAllNamespace, err = latestVersion(eventList.Items)
	if err != nil {
		return err
	}

	curr.InitVersions(strconv.Itoa(latestAllNamespace), strconv.Itoa(latestAllNamespace))
	return nil
}

func latestVersion(events []corev1.Event) (int, error) {
	latest := 0
	for _, event := range events {
		version, err := strconv.Atoi(event.ResourceVersion)
		if err != nil {
			return 0, err
		}
		if latest < version {
			latest = version
		}
	}
	return latest, nil
}
