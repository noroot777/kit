// If the involved object's kind is workload, kit will do more job to analysis
// it and display the process on UI window.
// k8s workload:
//   Deployment, ReplicaSet, StatefulSet, DaemonSet, Job, CronJob,
//   ~~ReplicationController~~, Pod

package kit

import (
	"context"
	"fmt"

	appsv1beta1 "k8s.io/api/apps/v1beta1"
	appsv1beta2 "k8s.io/api/apps/v1beta2"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	batchv2alpha1 "k8s.io/api/batch/v2alpha1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/resource"
	coreclient "k8s.io/client-go/kubernetes/typed/core/v1"
	deploymentutil "k8s.io/kubectl/pkg/util/deployment"
)

// 1. 生成相关的资源组合
// 2. 根据involvedObject.id过滤events
// 3. 在活动区展示event object+reson
//    如果type != Normal(core.EventTypeNormal) 则都认为是错误
// 此种方式会有一个问题：watch时尚不知道pod的名字，从而错过event
// 考虑以下方式：
// 0. 维护一个map，存有此次命令相关的object，初始值为info
// 1. watch得到event后，根据其involvedObject向上追溯，是否与传入的map相关
// 2. 若相关，且map中无此object，则将involvedObject存入map，并做相应展示
// 3. 若不相关，则放弃

func activities(event *corev1.Event, opts *Options) {
	eventInvolvedName := event.InvolvedObject.Name
	eventInvolvedNamespace := event.InvolvedObject.Namespace
	eventInvolvedKind := event.InvolvedObject.Kind

	// opts.writer.Write([]byte(fmt.Sprintf("name: %v, kind: %v\n", eventInvolvedName, eventInvolvedKind)))

	// if the involved obj of event was containes in the map, show message in the Activity View.
	involvedObj := opts.involvedObjects[eventInvolvedName]
	if involvedObj != nil {
		// 在event list中展示
		// 在活动list中展示
		if event.Type != corev1.EventTypeNormal {
			// 展示错误
			s := fmt.Sprintf(" ✖️   %v/%v/%v", event.Reason, event.InvolvedObject.Kind, eventInvolvedName)
			opts.writer.Write([]byte(s))
			return
		}
	}

	// TODO if event.InvolvedObject.Kind not in workload range, return

	for _, v := range opts.involvedObjects {
		if eventInvolvedNamespace != v.Namespace {
			continue
		}

		kind := v.Object.GetObjectKind().GroupVersionKind().Kind
		if eventInvolvedKind == "Pod" && kind != "Deployment" {
			pod, err := opts.ClientSet.CoreV1().Pods(eventInvolvedNamespace).Get(context.TODO(), eventInvolvedName, metav1.GetOptions{})
			if err != nil || pod == nil {
				// 展示错误？
				continue
			}

			metaObj := switch2Object(v.Object)
			if metav1.IsControlledBy(pod, metaObj) {
				opts.involvedObjects[eventInvolvedName] = &resource.Info{Object: pod, Namespace: eventInvolvedNamespace}

				s := fmt.Sprintf("%v/%v %v", eventInvolvedKind, eventInvolvedName, event.Reason)
				opts.writer.Write([]byte(s))
				return
			} else if string(pod.UID) == string(metaObj.GetUID()) {
				s := fmt.Sprintf("%v/%v %v", eventInvolvedKind, eventInvolvedName, event.Reason)
				opts.writer.Write([]byte(s))
				return
			}
		} else if eventInvolvedKind == "ReplicaSet" && kind == "Deployment" {
			rs, err := opts.ClientSet.AppsV1().ReplicaSets(eventInvolvedNamespace).Get(context.TODO(), eventInvolvedName, metav1.GetOptions{})
			if err != nil || rs == nil {
				// 展示错误？
				continue
			}

			if metav1.IsControlledBy(rs, switch2Object(v.Object)) {
				opts.involvedObjects[eventInvolvedName] = &resource.Info{Object: rs, Namespace: eventInvolvedNamespace}

				s := fmt.Sprintf("%v/%v %v", eventInvolvedKind, eventInvolvedName, event.Reason)
				opts.writer.Write([]byte(s))
				return
			}
		}
	}
}

func switch2Object(obj runtime.Object) metav1.Object {
	var ret metav1.Object
	ret = switch2ObjectMeta(obj)
	if isNilPtr(ret) {
		switch obj.(type) {
		case *unstructured.Unstructured:
			ret = obj.(*unstructured.Unstructured)
		}
	}
	return ret
}

func switch2ObjectMeta(obj runtime.Object) *metav1.ObjectMeta {
	switch obj.(type) {
	case *corev1.Pod:
		return &obj.(*corev1.Pod).ObjectMeta
	case *corev1.ReplicationController:
		return &obj.(*corev1.ReplicationController).ObjectMeta

		// Deployment
	case *extensionsv1beta1.Deployment:
		return &obj.(*extensionsv1beta1.Deployment).ObjectMeta
	case *appsv1beta1.Deployment:
		return &obj.(*appsv1beta1.Deployment).ObjectMeta
	case *appsv1beta2.Deployment:
		return &obj.(*appsv1beta2.Deployment).ObjectMeta
	case *appsv1.Deployment:
		return &obj.(*appsv1.Deployment).ObjectMeta

		// DaemonSet
	case *extensionsv1beta1.DaemonSet:
		return &obj.(*extensionsv1beta1.DaemonSet).ObjectMeta
	case *appsv1beta2.DaemonSet:
		return &obj.(*appsv1beta2.DaemonSet).ObjectMeta
	case *appsv1.DaemonSet:
		return &obj.(*appsv1.DaemonSet).ObjectMeta

		// ReplicaSet
	case *extensionsv1beta1.ReplicaSet:
		return &obj.(*extensionsv1beta1.ReplicaSet).ObjectMeta
	case *appsv1beta2.ReplicaSet:
		return &obj.(*appsv1beta2.ReplicaSet).ObjectMeta
	case *appsv1.ReplicaSet:
		return &obj.(*appsv1.ReplicaSet).ObjectMeta

		// StatefulSet
	case *appsv1beta1.StatefulSet:
		return &obj.(*appsv1beta1.StatefulSet).ObjectMeta
	case *appsv1beta2.StatefulSet:
		return &obj.(*appsv1beta2.StatefulSet).ObjectMeta
	case *appsv1.StatefulSet:
		return &obj.(*appsv1.StatefulSet).ObjectMeta

		// Job
	case *batchv1.Job:
		return &obj.(*batchv1.Job).ObjectMeta

		// CronJob
	case *batchv1beta1.CronJob:
		return &obj.(*batchv1beta1.CronJob).ObjectMeta
	case *batchv2alpha1.CronJob:
		return &obj.(*batchv2alpha1.CronJob).ObjectMeta

	default:
		// "no match type for %T", t
		return nil
	}
}

func deepin(info *resource.Info, opts Options) {
	switch info.Object.GetObjectKind().GroupVersionKind().Kind {
	case "Deployment":
		d := info.Object.(*appsv1.Deployment)
		_, _, newRS, err := deploymentutil.GetAllReplicaSets(d, opts.ClientSet.AppsV1())
		if err != nil {
			fmt.Println(err.Error())
		}
		pods, err := getControlleePods(newRS, opts.ClientSet.CoreV1())
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println(len(pods))
	case "ReplicaSet":
		rs := info.Object.(*appsv1.ReplicaSet)
		pods, err := getControlleePods(rs, opts.ClientSet.CoreV1())
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println(len(pods))
	case "StatefulSet":
		ss := info.Object.(*appsv1.StatefulSet)
		pods, err := getControlleePods(ss, opts.ClientSet.CoreV1())
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println(len(pods))
	case "DaemonSet":
		// loop pods
	case "Job":
		// loop pods
	case "CronJob":
		// loop pods
	case "ReplicationController":
		// loop pods
	case "Pod":
	default:
	}
}

func getControlleePods(owner metav1.Object, c coreclient.CoreV1Interface) ([]corev1.Pod, error) {
	pods, err := c.Pods(owner.GetNamespace()).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var allPods []corev1.Pod
	for _, pod := range pods.Items {
		if metav1.IsControlledBy(&pod, owner) {
			allPods = append(allPods, pod)
		}
	}
	return allPods, err
}
