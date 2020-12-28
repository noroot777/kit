package kit

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	appsv1beta2 "k8s.io/api/apps/v1beta2"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	batchv2alpha1 "k8s.io/api/batch/v2alpha1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	coreclient "k8s.io/client-go/kubernetes/typed/core/v1"
)

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
	switch t := obj.(type) {
	case *corev1.Pod:
		return &t.ObjectMeta
	case *corev1.ReplicationController:
		return &t.ObjectMeta

		// Deployment
	case *extensionsv1beta1.Deployment:
		return &t.ObjectMeta
	case *appsv1beta1.Deployment:
		return &t.ObjectMeta
	case *appsv1beta2.Deployment:
		return &t.ObjectMeta
	case *appsv1.Deployment:
		return &t.ObjectMeta

		// DaemonSet
	case *extensionsv1beta1.DaemonSet:
		return &t.ObjectMeta
	case *appsv1beta2.DaemonSet:
		return &t.ObjectMeta
	case *appsv1.DaemonSet:
		return &t.ObjectMeta

		// ReplicaSet
	case *extensionsv1beta1.ReplicaSet:
		return &t.ObjectMeta
	case *appsv1beta2.ReplicaSet:
		return &t.ObjectMeta
	case *appsv1.ReplicaSet:
		return &t.ObjectMeta

		// StatefulSet
	case *appsv1beta1.StatefulSet:
		return &t.ObjectMeta
	case *appsv1beta2.StatefulSet:
		return &t.ObjectMeta
	case *appsv1.StatefulSet:
		return &t.ObjectMeta

		// Job
	case *batchv1.Job:
		return &t.ObjectMeta

		// CronJob
	case *batchv1beta1.CronJob:
		return &t.ObjectMeta
	case *batchv2alpha1.CronJob:
		return &t.ObjectMeta

	default:
		// "no match type for %T", t
		return nil
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
