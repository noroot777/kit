// If the involved object's kind is workload, kit will do more job to analysis
// it and display the process on UI window.
// k8s workload:
//   Deployment, ReplicaSet, StatefulSet, DaemonSet, Job, CronJob,
//   ~~ReplicationController~~, Pod

package kit

import (
	"context"
	"fmt"
	"strings"

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
	coreclient "k8s.io/client-go/kubernetes/typed/core/v1"
)

// if the event was associated with the recorded involved objects,
// show in activity window.
// Variables that with `event` prefix are from event,
// and with `recorded` prefix means from recorded object.
func activities(event *corev1.Event, opts *Options, curr *Current) {
	if curr.recordedEvents.Contains(event.Name) {
		return
	}

	eventInvolvedName := event.InvolvedObject.Name
	eventInvolvedNamespace := event.InvolvedObject.Namespace
	eventInvolvedKind := event.InvolvedObject.Kind
	eventKindName := eventInvolvedKind + "/" + eventInvolvedName

	// if the involved obj of event was containes in the map, show message in the Activity window.
	involvedObj := opts.involvedObjects[eventInvolvedName]
	if involvedObj != nil {
		if event.Type != corev1.EventTypeNormal {
			s := fmt.Sprintf(" ✖️   %v/%v/%v", event.Reason, event.InvolvedObject.Kind, eventInvolvedName)
			opts.writer.Write([]byte(s))
			return
		}
	}

	for kn, recordedObj := range opts.involvedObjects {
		if eventInvolvedNamespace != recordedObj.GetNamespace() {
			continue
		}

		recodredKind := strings.Split(kn, "/")[0]

		if eventInvolvedKind == "Pod" {
			eventPod, err := opts.ClientSet.CoreV1().Pods(eventInvolvedNamespace).Get(context.TODO(), eventInvolvedName, metav1.GetOptions{})
			if err != nil || eventPod == nil {
				continue
			}

			if recodredKind == "ReplicaSet" {
				if metav1.IsControlledBy(eventPod, recordedObj) {
					opts.involvedObjects[eventKindName] = eventPod

					curr.recordedEvents.Add(event.Name)

					a := opts.activities.GetOrNew(eventKindName)
					a.AddMessage(Message{Info: event.Reason, When: event.CreationTimestamp.Time})
					showActivites(opts)
					return
				}
			} else if recodredKind == "Pod" {
				if string(eventPod.UID) == string(recordedObj.GetUID()) {
					curr.recordedEvents.Add(event.Name)
				}
			}
		} else if eventInvolvedKind == "ReplicaSet" {
			if recodredKind == "Deployment" {
				eventRs, err := opts.ClientSet.AppsV1().ReplicaSets(eventInvolvedNamespace).Get(context.TODO(), eventInvolvedName, metav1.GetOptions{})
				if err != nil || eventRs == nil {
					continue
				}

				if metav1.IsControlledBy(eventRs, recordedObj) {
					opts.involvedObjects[eventKindName] = eventRs

					curr.recordedEvents.Add(event.Name)

					a := opts.activities.GetOrNew(eventKindName)
					a.AddMessage(Message{Info: event.Reason, When: event.CreationTimestamp.Time})
					showActivites(opts)
					return
				}
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

func showActivites(opts *Options) {
	opts.ActivityWindow.SetText([]string{""})

	for _, activity := range opts.activities {
		opts.writer.Write([]byte(activity.KindName))
		for _, msg := range activity.Msg {
			s := fmt.Sprintf("  |- %v", msg.Info)
			opts.writer.Write([]byte(s))
		}
		// if not ready, add `...` at end, else add ✅ at end
		complete := true
		if complete {
			opts.writer.Write([]byte(fmt.Sprintf("  |- %v", "✅")))
		} else {
			opts.writer.Write([]byte(fmt.Sprintf("  |- %v", "...")))
		}
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

func syncStatus(obj runtime.Object) {
	switch obj.(type) {
	case *corev1.Pod:
		t := obj.(*corev1.Pod)
		if t.Status.Conditions[0].Status == corev1.ConditionTrue {
		}
	case *corev1.ReplicationController:
		t := obj.(*corev1.ReplicationController)
		if t.Status.Conditions[0].Status == corev1.ConditionTrue {
		}

	}
}
