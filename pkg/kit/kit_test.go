package kit

import (
	"fmt"
	"testing"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/kubectl/pkg/describe"
)

// Watch s
// func TestWatch(t *testing.T) {
// 	config, _ := clientcmd.BuildConfigFromFlags("", "./test.kubeconfig")
// 	// f := cmdtesting.NewTestFactory()
// 	// config, err := f.ToRESTConfig()
// 	// if err != nil {
// 	// 	t.Fatal(err)
// 	// }
// 	clientSet, err := kubernetes.NewForConfig(config)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	o := NewOptions("", nil, clientSet)
// 	watchEvents(o)
// 	for {
// 		select {
// 		case event, _ := <-o.eventsReader:
// 			fmt.Println(event.Message)
// 		}
// 	}
// }

func TestIdea(t *testing.T) {
	getter := genericclioptions.NewConfigFlags(false)

	var Describer func(*meta.RESTMapping) (describe.ResourceDescriber, error) = func(mapping *meta.RESTMapping) (describe.ResourceDescriber, error) {
		return describe.DescriberFn(getter, mapping)
	}
	// config, _ := clientcmd.BuildConfigFromFlags("", "./test.kubeconfig")
	b := *resource.NewBuilder(getter)
	r := b.
		Unstructured().
		ContinueOnError().
		NamespaceParam("imds").DefaultNamespace().AllNamespaces(false).
		// FilenameParam(o.EnforceNamespace, o.FilenameOptions).
		LabelSelectorParam("").
		ResourceTypeOrNameArgs(true, "event", "imds-web-764f9c4bb8.1650c4ade4451b20").
		Flatten().
		Do()
	infos, _ := r.Infos()
	// errs := sets.NewString()
	// first := true

	for _, info := range infos {
		mapping := info.ResourceMapping()
		describer, _ := Describer(mapping)
		// if err != nil {
		// 	if errs.Has(err.Error()) {
		// 		continue
		// 	}
		// 	errs.Insert(err.Error())
		// 	continue
		// }
		s, _ := describer.Describe(info.Namespace, info.Name, describe.DescriberSettings{ShowEvents: false})
		// if err != nil {
		// 	if errs.Has(err.Error()) {
		// 		continue
		// 	}
		// 	errs.Insert(err.Error())
		// 	continue
		// }
		// if first {
		// 	first = false
		fmt.Printf(s)
		// } else {
		// 	fmt.Printf("\n\n%s", s)
		// }
	}
}
