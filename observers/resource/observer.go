package resource

import (
	"fmt"

	v1alpha1 "github.com/cermakm/argo-await-operator/api/v1alpha1"
	"github.com/cermakm/argo-await-operator/common"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"

	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("observer")

// Observer watches for specified resources
type Observer struct {
	client dynamic.NamespaceableResourceInterface

	namespace string
	resource  *v1alpha1.Resource
	filters   []string

	K8RestConfig *rest.Config
}

// Get retrieves resources from the Observer's namespace
func (obs *Observer) Get(name string, options metav1.GetOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return obs.client.Namespace(obs.namespace).Get(name, options, subresources...)
}

// List lists resources from the Observer's namespace
func (obs *Observer) List(opts metav1.ListOptions) (*unstructured.UnstructuredList, error) {
	return obs.client.Namespace(obs.namespace).List(opts)
}

// Watch watches resources from the Observer's namespace
func (obs *Observer) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return obs.client.Namespace(obs.namespace).Watch(opts)
}

// Await awaits a resource based on given filters
func (obs *Observer) Await(callback func() error) error {
	watchInterface, err := obs.Watch(metav1.ListOptions{})
	if err != nil {
		log.Error(err, "error creating a watch for resource", "resource", *obs.resource)
		panic(err)
	}

	log.WithValues(
		"group", obs.resource.Group,
		"version", obs.resource.Version,
		"kind", obs.resource.Kind,
	).Info("watching for resources")
	for {
		select {
		case item := <-watchInterface.ResultChan():
			log := log.WithValues(
				"type", item.Type,
				"resource", item.Object.GetObjectKind().GroupVersionKind(),
			)
			log.V(1).Info("new event received", "event", item)

			gvk := item.Object.GetObjectKind().GroupVersionKind()
			if obs.resource.Kind != gvk.Kind {
				log.Info("resource does not match required kind: ", "kind", obs.resource.Kind)
				continue
			}

			unstructured, err := runtime.DefaultUnstructuredConverter.ToUnstructured(item.Object)
			if err != nil {
				log.Error(err, "Unable to convert runtime object to unstructured")
				continue
			}

			if ok, err := passFilters(unstructured, obs.filters...); ok == false {
				if err != nil {
					return fmt.Errorf("Unable to parse resource filters")
				}

				log.Info("resource dit not pass the filters")
				continue
			}

			log.Info("resource fulfilled")

			// Execute the callback function and return
			return callback()
		}
	}
}

// NewObserverForResource create a new ResourceObserver from kubernetes config
func NewObserverForResource(conf *rest.Config, res *v1alpha1.Resource, filters []string) (*Observer, error) {
	ns, err := common.GetWatchNamespace()
	if err != nil {
		panic(err)
	}

	gvr := schema.GroupVersionResource{
		Group:    res.Group,
		Version:  res.Version,
		Resource: res.Name,
	}
	resourceClient := dynamic.NewForConfigOrDie(conf).Resource(gvr)

	return &Observer{
		client:       resourceClient,
		namespace:    ns,
		resource:     res,
		filters:      filters,
		K8RestConfig: conf,
	}, nil
}
