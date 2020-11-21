package kit

import (
	"testing"

	"k8s.io/client-go/kubernetes"
	cmdtesting "k8s.io/kubectl/pkg/cmd/testing"
)

// Watch s
func TestWatch(t *testing.T) {
	f := cmdtesting.NewTestFactory()
	config, err := f.ToRESTConfig()
	if err != nil {
		t.Fatal(err)
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		t.Fatal(err)
	}

	Watch(clientSet, make(chan struct{}))
}
