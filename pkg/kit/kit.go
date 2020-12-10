package kit

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"sync"
	"time"

	ui "github.com/noroot777/clui"
	termbox "github.com/nsf/termbox-go"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes"
)

var (
	// Errors errors before intercept
	Errors []error
	// values lock
	mtx sync.RWMutex
	// count of Event table(include the NO. column)
	columnCount = 6
	// to keep current status
	curr *Current
	// options
	opts *Options
)

// Options TODO
type Options struct {
	Namespace string // TODO apply -f, there are multiple namespaces
	ClientSet kubernetes.Interface
	TextView  *ui.TextView

	involvedNamespaces map[string]string
	involvedObjects    map[string]*resource.Info

	writer      *UIWriter
	errorWriter *UIWriter
	watcher     watch.Interface
}

// newOptions create new Options
func newOptions(namespace string, clientSet *kubernetes.Clientset) *Options {
	o := &Options{
		Namespace: namespace,
		ClientSet: clientSet,
	}
	return o
}

// HandleInfo handle with the execed info
func HandleInfo(info *resource.Info) {
	opts.writer.Write([]byte(fmt.Sprintf("HandleInfo: %v", info)))

	metaObj := switchType(info.Object)
	if metaObj == nil {
		opts.writer.Write([]byte(fmt.Sprintf("have no meta object: %+v", info)))
	}
	opts.involvedObjects[metaObj.Name] = info
	// TODO print a message to activity view. 根据不同的命令打印不同内容，eg: apply(delete/create) imds/Deployment/imds-web
	opts.writer.Write([]byte(fmt.Sprintf("apply %v/%v/%v", metaObj.Name, info.Mapping.GroupVersionKind.Kind, metaObj.Name)))
	if _, has := opts.involvedNamespaces[metaObj.Namespace]; !has {
		opts.involvedNamespaces[metaObj.Namespace] = metaObj.Namespace
	}
}

// Intercept intercept the kubectl command
func Intercept(fn InterceptFunc, namespace string, clientSet *kubernetes.Clientset) (out io.Writer, errorOut io.Writer) {
	opts = newOptions(namespace, clientSet)
	curr = NewCurrent(opts.Namespace)
	err := latestResourceVersion()
	if err != nil {
		fmt.Println(err)
		return nil, nil
	}

	if fn != nil {
		fn(opts)
	}

	drawUI()

	out = NewUIWriter(opts)
	errorOut = NewUIErrorWriter(opts)

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
	opts.TextView = txtActivity
	txtActivity.SetShowScroll(false)
	txtActivity.SetWordWrap(true)
	for _, err := range Errors {
		txtActivity.AddText([]string{"⚠️ " + err.Error()})
	}
	// txtActivity.AddText([]string{" ⚠️   Namespace created!"})
	// txtActivity.AddText([]string{" ✅   Namespace created!"})
	// txtActivity.AddText([]string{" ✖️   Namespace created!"})

	// frame include frmRadio and frmEvents
	frmRight := ui.CreateFrame(frmExTips, ui.AutoSize, ui.AutoSize, ui.BorderThin, ui.AutoSize)
	frmRight.SetPack(ui.Vertical)
	frmRight.SetPaddings(1, 1)
	frmRight.SetTitle(" Kubernetes Events ")

	// frame include radioGroup
	frmRadio := ui.CreateFrame(frmRight, ui.AutoSize, ui.AutoSize, ui.BorderThin, ui.AutoSize)
	frmRadio.SetTitle(" Choose Events Scope ")
	frmRadio.SetPack(ui.Horizontal)
	frmRadio.SetPaddings(5, 1)
	frmRadio.SetAlign(ui.AlignRight)
	// radio group to select related namespace or objects
	radioGroup := ui.CreateRadioGroup()
	radioCR := ui.CreateRadio(frmRadio, ui.AutoSize, "Involved objects", 1)
	radioCR.SetSelected(false)
	radioCR.SetActive(false)
	radioGroup.AddItem(radioCR)
	radioCN := ui.CreateRadio(frmRadio, ui.AutoSize, "Current namespace", 1)
	radioCN.SetSelected(true)
	radioCN.SetActive(false)
	radioGroup.AddItem(radioCN)
	radioAN := ui.CreateRadio(frmRadio, ui.AutoSize, "All namespaces", 1)
	radioAN.SetSelected(false)
	radioAN.SetActive(false)
	radioGroup.AddItem(radioAN)

	// frame include tabEvents
	frmEvents := ui.CreateFrame(frmRight, ui.AutoSize, ui.AutoSize, ui.BorderThin, ui.AutoSize)
	frmEvents.SetTitle(" Events List ")
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
	}
	tabEvents.SetColumns(cols)
	tabEvents.SetRowCount(0)
	tabEvents.SetShowScroll(false)
	tabEvents.SetFullRowSelect(true)
	tabEvents.SetShowLines(false)
	tabEvents.SetShowRowNumber(false)

	// frame include txtEvent
	frmRightBottom := ui.CreateFrame(frmRight, ui.AutoSize, 10, ui.BorderThin, ui.Fixed)
	frmRightBottom.SetTitle(" Events Detail ")
	frmRightBottom.SetPack(ui.Vertical)
	// TextView show event detail
	txtEvent := ui.CreateTextView(frmRightBottom, ui.AutoSize, ui.AutoSize, 1)
	txtEvent.SetShowScroll(false)
	txtEvent.SetWordWrap(true)
	txtEvent.SetActive(false)

	frmTips := ui.CreateFrame(view, ui.AutoSize, ui.AutoSize, ui.BorderNone, ui.Fixed)
	txtTips := ui.CreateTextView(frmTips, ui.AutoSize, 1, 1)
	txtTips.SetText([]string{" * ctrl+e: change interactive mode - text selection or mouse;   ctrl+q: exit;   esc: exit;   or click the top right corner to exit;"})
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
		}
		return true
	}, inputEscMode)

	tabEvents.OnDrawCell(drawCell)

	tabEvents.OnSelectCell(func(selectedCol int, selectedRow int) {
		mtx.Lock()
		defer mtx.Unlock()

		txtEvent.SetText(text(curr.Events()[selectedRow*columnCount : (selectedRow+1)*columnCount]))
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
			changeRadioFocus(FocusOnCurrentNamespace, tabEvents)
		case 2:
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
				mtx.Lock()
				switch e.Object.(type) {
				case *corev1.Event:
					event := e.Object.(*corev1.Event)

					activities(event, opts)

					// do not display the old version event
					if event.ResourceVersion < curr.Version() {
						continue
					}

					curr.AppendEvent(
						[]string{
							strconv.Itoa(len(curr.Events())/columnCount + 1),
							event.LastTimestamp.Format("15:04:05"),
							event.Type,
							event.Reason,
							event.InvolvedObject.Kind + "/" + event.InvolvedObject.Name,
							event.Message})
					// move down
					curr.MoveEach()
					// set the latest resource version
					curr.SetVersion(event.ResourceVersion)

					txtEvent.SetText(text(curr.Events()[:columnCount]))
					tabEvents.SetRowCount(len(curr.Events()) / columnCount)
					// tabEvents.Draw() here is not taking effect here, so refresh ui hardly.
					ui.PutEvent(ui.Event{Type: ui.EventRedraw})

				default:
					continue
				}

				mtx.Unlock()
			}
		}
	}()

}

func text(v []string) []string {
	if len(v[4]) >= 30 {
		v[4] = v[4][:27] + ".. "
	}
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

func changeRadioFocus(f FocusOn, tabEvents *ui.TableView) {
	tabEvents.SetRowCount(0)
	tabEvents.Draw()
	opts.watcher.Stop()

	curr.SetSelectedRadio(f)
	curr.SetNamespace(opts.Namespace)
	tabEvents.SetRowCount(len(curr.Events()) / columnCount)
	tabEvents.Draw()

	watchEvents()
}

func drawCell(info *ui.ColumnDrawInfo) {
	mtx.Lock()
	defer mtx.Unlock()

	if len(curr.Events()) > 0 {
		info.Text = curr.Events()[info.Row*columnCount+info.Col]
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
}

// can not use informer, bcz `410 Gone` happened
func watchEvents() {
	version := curr.Version()
	ns := curr.Namespace()
	listOpt := metav1.ListOptions{ResourceVersion: version, ResourceVersionMatch: metav1.ResourceVersionMatchNotOlderThan}
	watcher, err := opts.ClientSet.CoreV1().Events(ns).Watch(context.TODO(), listOpt)
	if err != nil {
		opts.errorWriter.Write([]byte(err.Error()))
	}
	opts.watcher = watcher
}

func latestResourceVersion() error {
	var latest, latestAllNamespace int

	eventList, err := opts.ClientSet.CoreV1().Events(opts.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	latest, err = latestVersion(eventList.Items)
	if err != nil {
		return err
	}

	eventList, err = opts.ClientSet.CoreV1().Events("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	latestAllNamespace, err = latestVersion(eventList.Items)
	if err != nil {
		return err
	}

	curr.InitVersions(strconv.Itoa(latest), strconv.Itoa(latestAllNamespace))
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
