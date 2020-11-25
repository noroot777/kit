package kit

import (
	"fmt"
	"testing"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// Watch s
func TestWatch(t *testing.T) {
	config, _ := clientcmd.BuildConfigFromFlags("", "./test.kubeconfig")
	// f := cmdtesting.NewTestFactory()
	// config, err := f.ToRESTConfig()
	// if err != nil {
	// 	t.Fatal(err)
	// }
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		t.Fatal(err)
	}

	reader := watch(clientSet, make(chan struct{}))
	for {
		select {
		case event, _ := <-reader:
			fmt.Println(event.Message)
		}
	}
}
