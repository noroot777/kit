package kit

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

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
			// klog.V(1).Infof("%+v", obj)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			e1 := oldObj.(*corev1.Event)
			e2 := newObj.(*corev1.Event)

			fmt.Printf("old: %v\n", e1.Message)
			fmt.Printf("new: %v\n", e2.Message)
			// klog.V(1).Infof("%+v", oldObj)
			// klog.V(1).Infof("%+v", newObj)
		},
		DeleteFunc: func(obj interface{}) {
			e := obj.(*corev1.Event)

			fmt.Printf("%v\n", e.Message)
			// klog.V(1).Infof("%+v", obj)
		},
	})
	for {
	}
}
