// If the involved object's kind is workload, kit will do more job to analysis
// it and display the process on UI window.
// k8s workload:
//   Deployment, ReplicaSet, StatefulSet, DaemonSet, Job, CronJob,
//   ~~ReplicationController~~, Pod

package kit

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func activities(event corev1.Event, opts Options) {
	involvedObj := opts.involvedObjects[event.Name]
	var metav1Obj metav1.Object
	switch involvedObj.Object.GetObjectKind().GroupVersionKind().Kind {
	case "Deployment":
		metav1Obj = involvedObj.Object.(*appsv1.Deployment)
	case "ReplicaSet":
		metav1Obj = involvedObj.Object.(*appsv1.ReplicaSet)
	case "StatefulSet":
		metav1Obj = involvedObj.Object.(*appsv1.StatefulSet)
	case "DaemonSet":
	case "Job":
	case "CronJob":
	case "ReplicationController":
	case "Pod":
	default:
	}
	// 需要一个数据结构，将各种转化后的对象存储起来，并且能根据kind获取切片
	// 只有involvedObj kind == pod or replicaset时才需要对比
	if involvedObj == nil {
		for _, v := range opts.involvedObjects {
			if metav1.IsControlledBy(metav1Obj, nobug(v)) {

			}
		}
	}
}

func nobug(info *resource.Info) metav1.Object {
	return nil
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
