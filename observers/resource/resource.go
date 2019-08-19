package resource

import (
	"encoding/json"
	"errors"
	"fmt"

	gjson "github.com/tidwall/gjson"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
)

// GroupVersionResourceForAPIResource returns GroupVersionResource from metav1.APIResource
func GroupVersionResourceForAPIResource(res *metav1.APIResource) schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    res.Group,
		Version:  res.Version, // apiResource seems to be have empty Version string
		Resource: res.Name,
	}
}

// GroupVersionKindForAPIResource returns GroupVersionKind from metav1.APIResource
func GroupVersionKindForAPIResource(res *metav1.APIResource) schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   res.Group,
		Version: res.Version, // apiResource seems to be have empty Version string
		Kind:    res.Kind,
	}
}

// GroupVersionForAPIResource returns GroupVersion from metav1.APIResource
func GroupVersionForAPIResource(res *metav1.APIResource) schema.GroupVersion {
	return schema.GroupVersion{
		Group:   res.Group,
		Version: res.Version, // apiResource seems to be have empty Version string
	}
}

// FormatJSON formats an object and returns JSON string
func formatJSON(obj interface{}) string {
	jsonified, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}

	return string(jsonified)
}

func passFilters(evt *watch.Event, filters ...string) (bool, error) {
	eventJSON := formatJSON([]watch.Event{*evt})

	if !gjson.Valid(eventJSON) {
		return false, errors.New("failed parsing event: invalid json")
	}

	for _, f := range filters {
		// filter needs to wrapped
		wrappedFilter := fmt.Sprintf("#(%s)", f)
		validResource := gjson.Get(eventJSON, wrappedFilter)

		if !validResource.Exists() {
			return false, nil
		}
	}

	return true, nil
}
