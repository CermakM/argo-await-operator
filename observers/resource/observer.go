package resource

import (
	"errors"

	"github.com/cermakm/argo-await-operator/controllers"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"

	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("observer")

// NewObserverForResource create a new ResourceObserver from kubernetes config
func NewObserverForResource(res *metav1.APIResource, conf *rest.Config) *Observer {
	dynamicClient := dynamic.NewForConfigOrDie(conf)

	gvr := schema.GroupVersionResource{
		Group:    res.Group,
		Version:  res.Version,
		Resource: res.Name,
	}
	resourceClient := dynamicClient.Resource(gvr)

	ns, err := controllers.GetOperatorNamespace()
	if err != nil {
		log.Error(err, "Namespace must be provided to the Observer")
	}

	return &Observer{
		client:       resourceClient,
		namespace:    ns,
		K8RestConfig: conf,
	}
}

// Observer watches for specified resources
type Observer struct {
	client    dynamic.NamespaceableResourceInterface
	namespace string

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

// AwaitResource awaits a resource based on given filters
func (obs *Observer) AwaitResource(callback func() error, res *metav1.APIResource, namespace string, filters []string) error {
	watchInterface, err := obs.Watch(metav1.ListOptions{})
	if err != nil {
		log.Error(err, "error watching resource")
		panic(err)
	}

	log.WithValues(
		"group", res.Group,
		"version", res.Version,
		"kind", res.Kind,
	).Info("watching for resources")
	for {
		select {
		case item := <-watchInterface.ResultChan():
			log := log.WithValues(
				"type", item.Type,
				"resource", item.Object.GetObjectKind().GroupVersionKind(),
			)
			log.V(1).Info("new resource received")
			log.V(2).Info("data: ", "data", item)

			gvk := item.Object.GetObjectKind().GroupVersionKind()
			if res.Kind != gvk.Kind {
				log.Info("resource does not match required kind: ", "kind", res.Kind)
				continue
			}

			log.V(1).Info("applying filters: ", "filters", filters)

			if ok, err := passFilters(&item, filters...); ok == false {
				if err != nil {
					log.Error(err, "an error occured during parsing")

					return errors.New("Unable to parse resource filters")
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
