package kit

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"
	"unicode"

	"github.com/fatih/camelcase"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/dynamic"
	clientset "k8s.io/client-go/kubernetes"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/kubectl/pkg/describe"
	"k8s.io/kubectl/pkg/scheme"
	"k8s.io/kubectl/pkg/util/slice"
)

const (
	LEVEL_0 = iota
	LEVEL_1
	LEVEL_2
	LEVEL_3
	LEVEL_4
)

var (
	// globally skipped annotations
	skipAnnotations  = sets.NewString(corev1.LastAppliedConfigAnnotation)
	maxAnnotationLen = 140
)

// GenericDescriberFor returns a generic describer for the specified mapping
// that uses only information available from runtime.Unstructured
func GenericDescriberFor(mapping *meta.RESTMapping, clientConfig *rest.Config) (describe.ResourceDescriber, bool) {
	// used to fetch the resource
	dynamicClient, err := dynamic.NewForConfig(clientConfig)
	if err != nil {
		return nil, false
	}

	// used to get events for the resource
	clientSet, err := clientset.NewForConfig(clientConfig)
	if err != nil {
		return nil, false
	}
	eventsClient := clientSet.CoreV1()

	return &GenericProgressiveDescriber{mapping, dynamicClient, eventsClient}, true
}

func progressive(f func(io.Writer) error) (string, error) {
	out := new(tabwriter.Writer)
	buf := &bytes.Buffer{}
	out.Init(buf, 0, 8, 2, ' ', 0)

	err := f(out)
	if err != nil {
		return "", err
	}

	out.Flush()
	str := string(buf.String())
	return str, nil
}

// GenericProgressiveDescriber TODO
type GenericProgressiveDescriber struct {
	mapping *meta.RESTMapping
	dynamic dynamic.Interface
	events  corev1client.EventsGetter
}

// Describe Describe
func (g *GenericProgressiveDescriber) Describe(namespace, name string, describerSettings describe.DescriberSettings) (output string, err error) {
	obj, err := g.dynamic.Resource(g.mapping.Resource).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	var events *corev1.EventList
	if describerSettings.ShowEvents {
		events, _ = g.events.Events(namespace).Search(scheme.Scheme, obj)
	}

	return progressive(func(out io.Writer) error {
		w := describe.NewPrefixWriter(out)
		w.Write(LEVEL_0, "Name:\t%s\n", obj.GetName())
		w.Flush()
		w.Write(LEVEL_0, "Namespace:\t%s\n", obj.GetNamespace())
		w.Flush()
		printLabelsMultiline(w, "Labels", obj.GetLabels())
		printAnnotationsMultiline(w, "Annotations", obj.GetAnnotations())
		printUnstructuredContent(w, LEVEL_0, obj.UnstructuredContent(), "", ".metadata.name", ".metadata.namespace", ".metadata.labels", ".metadata.annotations")
		if events != nil {
			describe.DescribeEvents(events, w)
		}
		return nil
	})
}

func printUnstructuredContent(w describe.PrefixWriter, level int, content map[string]interface{}, skipPrefix string, skip ...string) {
	fields := []string{}
	for field := range content {
		fields = append(fields, field)
	}
	sort.Strings(fields)

	for _, field := range fields {
		value := content[field]
		switch typedValue := value.(type) {
		case map[string]interface{}:
			skipExpr := fmt.Sprintf("%s.%s", skipPrefix, field)
			if slice.ContainsString(skip, skipExpr, nil) {
				continue
			}
			w.Write(level, "%s:\n", smartLabelFor(field))
			printUnstructuredContent(w, level+1, typedValue, skipExpr, skip...)

		case []interface{}:
			skipExpr := fmt.Sprintf("%s.%s", skipPrefix, field)
			if slice.ContainsString(skip, skipExpr, nil) {
				continue
			}
			w.Write(level, "%s:\n", smartLabelFor(field))
			for _, child := range typedValue {
				switch typedChild := child.(type) {
				case map[string]interface{}:
					printUnstructuredContent(w, level+1, typedChild, skipExpr, skip...)
				default:
					w.Write(level+1, "%v\n", typedChild)
				}
			}

		default:
			skipExpr := fmt.Sprintf("%s.%s", skipPrefix, field)
			if slice.ContainsString(skip, skipExpr, nil) {
				continue
			}
			w.Write(level, "%s:\t%v\n", smartLabelFor(field), typedValue)
		}
	}
}

func smartLabelFor(field string) string {
	// skip creating smart label if field name contains
	// special characters other than '-'
	if strings.IndexFunc(field, func(r rune) bool {
		return !unicode.IsLetter(r) && r != '-'
	}) != -1 {
		return field
	}

	commonAcronyms := []string{"API", "URL", "UID", "OSB", "GUID"}
	parts := camelcase.Split(field)
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "_" {
			continue
		}

		if slice.ContainsString(commonAcronyms, strings.ToUpper(part), nil) {
			part = strings.ToUpper(part)
		} else {
			part = strings.Title(part)
		}
		result = append(result, part)
	}

	return strings.Join(result, " ")
}

// printLabelsMultiline prints multiple labels with a proper alignment.
func printLabelsMultiline(w describe.PrefixWriter, title string, labels map[string]string) {
	printLabelsMultilineWithIndent(w, "", title, "\t", labels, sets.NewString())
}

// printLabelsMultiline prints multiple labels with a user-defined alignment.
func printLabelsMultilineWithIndent(w describe.PrefixWriter, initialIndent, title, innerIndent string, labels map[string]string, skip sets.String) {
	w.Write(LEVEL_0, "%s%s:%s", initialIndent, title, innerIndent)

	if len(labels) == 0 {
		w.WriteLine("<none>")
		w.Flush()
		return
	}

	// to print labels in the sorted order
	keys := make([]string, 0, len(labels))
	for key := range labels {
		if skip.Has(key) {
			continue
		}
		keys = append(keys, key)
	}
	if len(keys) == 0 {
		w.WriteLine("<none>")
		w.Flush()
		return
	}
	sort.Strings(keys)

	for i, key := range keys {
		if i != 0 {
			w.Write(LEVEL_0, "%s", initialIndent)
			w.Write(LEVEL_0, "%s", innerIndent)
		}
		w.Write(LEVEL_0, "%s=%s\n", key, labels[key])
		i++
	}
}

// printAnnotationsMultiline prints multiple annotations with a proper alignment.
// If annotation string is too long, we omit chars more than 200 length.
func printAnnotationsMultiline(w describe.PrefixWriter, title string, annotations map[string]string) {
	w.Write(LEVEL_0, "%s:\t", title)

	// to print labels in the sorted order
	keys := make([]string, 0, len(annotations))
	for key := range annotations {
		if skipAnnotations.Has(key) {
			continue
		}
		keys = append(keys, key)
	}
	if len(keys) == 0 {
		w.WriteLine("<none>")
		w.Flush()
		return
	}
	sort.Strings(keys)
	indent := "\t"
	for i, key := range keys {
		if i != 0 {
			w.Write(LEVEL_0, indent)
		}
		value := strings.TrimSuffix(annotations[key], "\n")
		if (len(value)+len(key)+2) > maxAnnotationLen || strings.Contains(value, "\n") {
			w.Write(LEVEL_0, "%s:\n", key)
			for _, s := range strings.Split(value, "\n") {
				w.Write(LEVEL_0, "%s  %s\n", indent, shorten(s, maxAnnotationLen-2))
			}
		} else {
			w.Write(LEVEL_0, "%s: %s\n", key, value)
		}
		i++
	}
}

func shorten(s string, maxLength int) string {
	if len(s) > maxLength {
		return s[:maxLength] + "..."
	}
	return s
}
