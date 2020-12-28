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
	// involvedObj := opts.involvedObjects[eventInvolvedName]
	involvedObj := opts.activities.Get(eventInvolvedName)
	if involvedObj != nil {
		if event.Type != corev1.EventTypeNormal {
			s := fmt.Sprintf(" ✖️   %v/%v/%v", event.Reason, event.InvolvedObject.Kind, eventInvolvedName)
			opts.writer.Write([]byte(s))
			return
		}
	}

	for _, act := range opts.activities {
		kn, recordedObj := act.KindName, act.Obj
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
					// opts.involvedObjects[eventKindName] = eventPod
					curr.recordedEvents.Add(event.Name)

					a := opts.activities.GetOrNew(eventKindName, eventPod)
					a.AddMessage(Message{Info: event.Reason, When: event.CreationTimestamp.Time})
					// showActivites(opts)
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
					// opts.involvedObjects[eventKindName] = eventRs

					curr.recordedEvents.Add(event.Name)

					a := opts.activities.GetOrNew(eventKindName, eventRs)
					a.AddMessage(Message{Info: event.Reason, When: event.CreationTimestamp.Time})
					// showActivites(opts)
					return
				}
			}
		}
	}
}

// showActivitesSched
func showActivitesSched() {
	opts.ActivityWindow.SetText([]string{""})

	for _, activity := range opts.activities {
		opts.writer.Write([]byte(activity.KindName))
		for _, msg := range activity.Msg {
			s := fmt.Sprintf("  |- %v", msg.Info)
			opts.writer.Write([]byte(s))
		}
		checkComplete(activity)
		// if not ready, add `...` at end, else add ✅ at end
		if activity.Complete {
			opts.writer.Write([]byte(fmt.Sprintf("  |- %v", "✅")))
		} else {
			opts.writer.Write([]byte(fmt.Sprintf("  |- %v", "...")))
		}
	}
}

func checkComplete(act *Activity) {
	kn, obj := act.KindName, act.Obj

	switch t := obj.(type) {

	// Pod
	case *corev1.Pod:
		// when
		// Status:         Running
		if t.Status.Phase == corev1.PodRunning {
			opts.activities.Get(kn).Complete = true
		}

		// ReplicationController
	case *corev1.ReplicationController:
		// when
		// Replicas:       1 current / 1 desired
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
	case *extensionsv1beta1.ReplicaSet:
		if len(t.Status.Conditions) > 0 && t.Status.Conditions[0].Status == corev1.ConditionTrue {
			key := t.Kind + t.Name
			opts.activities.Get(key).Complete = true
		}
	case *appsv1beta2.ReplicaSet:
		if len(t.Status.Conditions) > 0 && t.Status.Conditions[0].Status == corev1.ConditionTrue {
			key := t.Kind + t.Name
			opts.activities.Get(key).Complete = true
		}
	case *appsv1.ReplicaSet:
		if len(t.Status.Conditions) > 0 && t.Status.Conditions[0].Status == corev1.ConditionTrue {
			key := t.Kind + t.Name
			opts.activities.Get(key).Complete = true
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
