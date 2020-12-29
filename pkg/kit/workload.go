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
					curr.recordedEvents.Add(event.Name)

					a := opts.activities.GetOrNew(eventKindName, eventPod)
					a.AddMessage(Message{Info: event.Reason, When: event.CreationTimestamp.Time})
					showActivites()
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
					curr.recordedEvents.Add(event.Name)

					a := opts.activities.GetOrNew(eventKindName, eventRs)
					a.AddMessage(Message{Info: event.Reason, When: event.CreationTimestamp.Time})
					showActivites()
					return
				}
			}
		}
	}
}

func showActivites() {
	mtx.Lock()
	defer mtx.Unlock()

	opts.ActivityWindow.SetText([]string{""})

	for _, activity := range opts.activities {
		opts.writer.Write([]byte(activity.KindName))
		for _, msg := range activity.Msg {
			s := fmt.Sprintf("  |- %v", msg.Info)
			opts.writer.Write([]byte(s))
		}
		// if not ready, add `...` at end, else add ✅ at end
		if activity.Complete {
			opts.writer.Write([]byte(fmt.Sprintf("  |- %v", "✅")))
		} else {
			opts.writer.Write([]byte(fmt.Sprintf("  |- %v", "...")))
		}
	}
}
