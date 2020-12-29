package kit

import (
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	appsv1beta2 "k8s.io/api/apps/v1beta2"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	batchv2alpha1 "k8s.io/api/batch/v2alpha1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

// Activities Activities
type Activities []*Activity

// Activity Activity
type Activity struct {
	KindName string
	Obj      metav1.Object
	// Controller *Controller
	Msg      []Message
	Complete bool
}

// Message Message
type Message struct {
	Info string
	When time.Time
}

// Controller Controller
// type Controller struct {
// 	Informer cache.SharedIndexInformer
// 	Stopper  chan struct{}
// }

// AddMessage AddMessage
func (t *Activity) AddMessage(m Message) {
	t.Msg = append(t.Msg, m)
}

// Get Get
func (t *Activities) Get(kindName string) *Activity {
	for _, a := range *t {
		if a.KindName == kindName {
			return a
		}
	}
	return nil
}

// Exists Exists
func (t *Activities) Exists(kindName string) bool {
	if t.Get(kindName) != nil {
		return true
	}
	return false
}

// GetOrNew return a new one if not exists
func (t *Activities) GetOrNew(kindName string, obj metav1.Object) *Activity {
	a := t.Get(kindName)
	if a != nil {
		return a
	}
	return t.New(kindName, obj)
}

// New New
func (t *Activities) New(kindName string, obj metav1.Object) *Activity {
	n := &Activity{
		KindName: kindName,
		Obj:      obj,
		// Controller: newController(obj),
		Msg:      []Message{},
		Complete: false,
	}
	*t = append(*t, n)
	startWatch(n)

	return n
}

// func newController(obj metav1.Object) *Controller {
// 	c := &Controller{
// 		Informer: newInformer(obj),
// 		Stopper:  make(chan struct{}),
// 	}

// 	return c
// }

func startWatch(act *Activity) {
	defer runtime.HandleCrash()
	stopper := make(chan struct{})

	var siOpts = make([]informers.SharedInformerOption, 2)
	op := func(o *metav1.ListOptions) {
		o.FieldSelector = "metadata.name=" + act.Obj.GetName()
	}
	siOpts[0] = informers.WithTweakListOptions(op)
	siOpts[1] = informers.WithNamespace(act.Obj.GetNamespace())
	f := informers.NewSharedInformerFactoryWithOptions(opts.ClientSet, 0, siOpts...)
	var informer cache.SharedIndexInformer

	switch act.Obj.(type) {
	case *corev1.Pod:
		informer = f.Core().V1().Pods().Informer()
		informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				// Pod
				// when
				// Status:         Running
				no := newObj.(*corev1.Pod)
				if no.Status.Phase == corev1.PodRunning {
					act.Complete = true
					showActivites()
				}
			},
			DeleteFunc: func(obj interface{}) {
			},
		})
	// case *corev1.ReplicationController:
	// 	informer = f.Core().V1().ReplicationControllers().Informer()
	// 	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
	// 		AddFunc: func(obj interface{}) {
	// 		},
	// 		UpdateFunc: func(oldObj, newObj interface{}) {
	// 			// ReplicationController
	// 			// when
	// 			// Replicas:       1 current / 1 desired
	// 			no := newObj.(*corev1.ReplicationController)
	// 			if *no.Spec.Replicas == no.Status.AvailableReplicas {
	// 				act.Complete = true
	// 				showActivites()
	// 			}
	// 		},
	// 		DeleteFunc: func(obj interface{}) {
	// 		},
	// 	})

	// Deployment
	case *extensionsv1beta1.Deployment:
		informer = f.Extensions().V1beta1().Deployments().Informer()
		informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {},
			UpdateFunc: func(oldObj, newObj interface{}) {
				no := newObj.(*extensionsv1beta1.Deployment)
				fmt.Printf("dep want: %v, act: %v\n", *no.Spec.Replicas, no.Status.AvailableReplicas)
				if *no.Spec.Replicas == no.Status.ReadyReplicas {
					act.Complete = true
					showActivites()
				}
			},
			DeleteFunc: func(obj interface{}) {},
		})
	case *appsv1beta1.Deployment:
		informer = f.Apps().V1beta1().Deployments().Informer()
		informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {},
			UpdateFunc: func(oldObj, newObj interface{}) {
				no := newObj.(*appsv1beta1.Deployment)
				if *no.Spec.Replicas == no.Status.ReadyReplicas {
					act.Complete = true
					showActivites()
				}
			},
			DeleteFunc: func(obj interface{}) {},
		})
	case *appsv1beta2.Deployment:
		informer = f.Apps().V1beta2().Deployments().Informer()
		informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {},
			UpdateFunc: func(oldObj, newObj interface{}) {
				no := newObj.(*appsv1beta2.Deployment)
				if *no.Spec.Replicas == no.Status.ReadyReplicas {
					act.Complete = true
					showActivites()
				}
			},
			DeleteFunc: func(obj interface{}) {},
		})
	case *appsv1.Deployment:
		informer = f.Apps().V1().Deployments().Informer()
		informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {},
			UpdateFunc: func(oldObj, newObj interface{}) {
				no := newObj.(*appsv1.Deployment)
				if *no.Spec.Replicas == no.Status.ReadyReplicas {
					act.Complete = true
					showActivites()
				}
			},
			DeleteFunc: func(obj interface{}) {},
		})

		// DaemonSet
	// case *extensionsv1beta1.DaemonSet:
	// 	informer = f.Extensions().V1beta1().DaemonSets().Informer()
	// case *appsv1beta2.DaemonSet:
	// 	informer = f.Apps().V1beta2().DaemonSets().Informer()
	// case *appsv1.DaemonSet:
	// 	informer = f.Apps().V1().DaemonSets().Informer()

	// ReplicaSet
	case *extensionsv1beta1.ReplicaSet:
		informer = f.Extensions().V1beta1().ReplicaSets().Informer()
		informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {},
			UpdateFunc: func(oldObj, newObj interface{}) {
				// ReplicaSet
				// when
				// Replicas:       1 current / 1 desired
				no := newObj.(*extensionsv1beta1.ReplicaSet)
				if *no.Spec.Replicas == no.Status.ReadyReplicas {
					act.Complete = true
					showActivites()
				}
			},
			DeleteFunc: func(obj interface{}) {},
		})
	case *appsv1beta2.ReplicaSet:
		informer = f.Apps().V1beta2().ReplicaSets().Informer()
		informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {},
			UpdateFunc: func(oldObj, newObj interface{}) {
				// ReplicaSet
				// when
				// Replicas:       1 current / 1 desired
				no := newObj.(*appsv1beta2.ReplicaSet)
				if *no.Spec.Replicas == no.Status.ReadyReplicas {
					act.Complete = true
					showActivites()
				}
			},
			DeleteFunc: func(obj interface{}) {},
		})
	case *appsv1.ReplicaSet:
		informer = f.Apps().V1().ReplicaSets().Informer()
		informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {},
			UpdateFunc: func(oldObj, newObj interface{}) {
				// ReplicaSet
				// when
				// Replicas:       1 current / 1 desired
				no := newObj.(*appsv1.ReplicaSet)
				if *no.Spec.Replicas == no.Status.ReadyReplicas {
					act.Complete = true
					showActivites()
				}
			},
			DeleteFunc: func(obj interface{}) {},
		})

		// StatefulSet
		// case *appsv1beta1.StatefulSet:
		// 	informer = f.Apps().V1beta1().StatefulSets().Informer()
		// case *appsv1beta2.StatefulSet:
		// 	informer = f.Apps().V1beta2().StatefulSets().Informer()
		// case *appsv1.StatefulSet:
		// 	informer = f.Apps().V1().StatefulSets().Informer()

		// Job
		// case *batchv1.Job:
		// 	informer = f.Batch().V1().Jobs().Informer()

		// CronJob
		// case *batchv1beta1.CronJob:
		// 	informer = f.Batch().V1beta1().CronJobs().Informer()
		// case *batchv2alpha1.CronJob:
		// 	informer = f.Batch().V2alpha1().CronJobs().Informer()
	}

	go f.Start(stopper)

}

func checkComplete(kn string, obj interface{}) {
	switch t := obj.(type) {

	// Pod
	// when
	// Status:         Running
	case *corev1.Pod:
		if t.Status.Phase == corev1.PodRunning {
			opts.activities.Get(kn).Complete = true
		}

		// ReplicationController
		// when
		// Replicas:       1 current / 1 desired
	case *corev1.ReplicationController:
		if *t.Spec.Replicas == t.Status.AvailableReplicas {
			opts.activities.Get(kn).Complete = true
		}

		// Deployment
	case *extensionsv1beta1.Deployment:
		if *t.Spec.Replicas == t.Status.AvailableReplicas {
			opts.activities.Get(kn).Complete = true
		}
	case *appsv1beta1.Deployment:
		if *t.Spec.Replicas == t.Status.AvailableReplicas {
			opts.activities.Get(kn).Complete = true
		}
	case *appsv1beta2.Deployment:
		if *t.Spec.Replicas == t.Status.AvailableReplicas {
			opts.activities.Get(kn).Complete = true
		}
	case *appsv1.Deployment:
		if *t.Spec.Replicas == t.Status.AvailableReplicas {
			opts.activities.Get(kn).Complete = true
		}

		// DaemonSet
	case *extensionsv1beta1.DaemonSet:
		if len(t.Status.Conditions) > 0 && t.Status.Conditions[0].Status == corev1.ConditionTrue {
			key := t.Kind + t.Name
			opts.activities.Get(key).Complete = true
		}
	case *appsv1beta2.DaemonSet:
		if len(t.Status.Conditions) > 0 && t.Status.Conditions[0].Status == corev1.ConditionTrue {
			key := t.Kind + t.Name
			opts.activities.Get(key).Complete = true
		}
	case *appsv1.DaemonSet:
		if len(t.Status.Conditions) > 0 && t.Status.Conditions[0].Status == corev1.ConditionTrue {
			key := t.Kind + t.Name
			opts.activities.Get(key).Complete = true
		}

		// ReplicaSet
		// when
		// Replicas:       1 current / 1 desired
	case *extensionsv1beta1.ReplicaSet:
		if *t.Spec.Replicas == t.Status.AvailableReplicas {
			opts.activities.Get(kn).Complete = true
		}
	case *appsv1beta2.ReplicaSet:
		if *t.Spec.Replicas == t.Status.AvailableReplicas {
			opts.activities.Get(kn).Complete = true
		}
	case *appsv1.ReplicaSet:
		if *t.Spec.Replicas == t.Status.AvailableReplicas {
			opts.activities.Get(kn).Complete = true
		}

		// StatefulSet
	case *appsv1beta1.StatefulSet:
		if len(t.Status.Conditions) > 0 && t.Status.Conditions[0].Status == corev1.ConditionTrue {
			key := t.Kind + t.Name
			opts.activities.Get(key).Complete = true
		}
	case *appsv1beta2.StatefulSet:
		if len(t.Status.Conditions) > 0 && t.Status.Conditions[0].Status == corev1.ConditionTrue {
			key := t.Kind + t.Name
			opts.activities.Get(key).Complete = true
		}
	case *appsv1.StatefulSet:
		if len(t.Status.Conditions) > 0 && t.Status.Conditions[0].Status == corev1.ConditionTrue {
			key := t.Kind + t.Name
			opts.activities.Get(key).Complete = true
		}

		// Job
	case *batchv1.Job:
		if len(t.Status.Conditions) > 0 && t.Status.Conditions[0].Status == corev1.ConditionTrue {
			key := t.Kind + t.Name
			opts.activities.Get(key).Complete = true
		}

		// CronJob complete=true by default
	case *batchv1beta1.CronJob:
		key := t.Kind + t.Name
		opts.activities.Get(key).Complete = true
	case *batchv2alpha1.CronJob:
		key := t.Kind + t.Name
		opts.activities.Get(key).Complete = true
	}

}
