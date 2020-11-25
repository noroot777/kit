package kit

import (
	"fmt"

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
	createView()

	ui.MainLoop()
}

func createView() {

	w, h := termbox.Size()
	view := ui.AddWindow(0, 0, w, h, "Kubectl Interactive Tool")
	view.SetBorder(ui.BorderThin)
	// view.SetBackColor(termbox.ColorDarkGray)

	frmLeft := ui.CreateFrame(view, 30, ui.AutoSize, ui.BorderThin, ui.Fixed)
	frmLeft.SetPaddings(1, 1)
	frmLeft.SetTitle("Activities")
	// ui.CreateTextView(frmLeft, 20, ui.AutoSize, ui.Fixed)

	frmRight := ui.CreateFrame(view, ui.AutoSize, ui.AutoSize, ui.BorderThin, ui.AutoSize)
	frmRight.SetPaddings(1, 1)
	frmRight.SetTitle("Kubernetes Events")
}

// Watch s
func Watch(clientSet kubernetes.Interface, stopper chan struct{}) {
	f := informers.NewSharedInformerFactoryWithOptions(clientSet, 0)
	informer := f.Core().V1().Events().Informer()
	defer runtime.HandleCrash()

	go f.Start(stopper)

	if !cache.WaitForCacheSync(stopper, informer.HasSynced) {
		fmt.Print("Timed out waiting for caches to sync")
	}

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			e := obj.(*corev1.Event)
			fmt.Printf("%v\n", e.Message)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			e1 := oldObj.(*corev1.Event)
			e2 := newObj.(*corev1.Event)

			fmt.Printf("old: %v\n", e1.Message)
			fmt.Printf("new: %v\n", e2.Message)
		},
		DeleteFunc: func(obj interface{}) {
			e := obj.(*corev1.Event)

			fmt.Printf("%v\n", e.Message)
		},
	})
	for {
	}
}
