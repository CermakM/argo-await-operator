package resource

import (
	"encoding/json"
	"fmt"

	gjson "github.com/tidwall/gjson"
)

// unstructuredToJSON formats an object and returns JSON string
func unstructuredToJSON(obj interface{}) string {
	jsonified, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}

	return string(jsonified)
}

func passFilters(object map[string]interface{}, filters ...string) (bool, error) {
	resourceJSON := unstructuredToJSON([]interface{}{object})

	if !gjson.Valid(resourceJSON) {
		return false, fmt.Errorf("failed to parse the resource: invalid json")
	}

	for _, filter := range filters {
		log.V(1).Info("applying filter", "filter", filter)

		// filter needs to wrapped in order to use comparison operator
		wrappedFilter := fmt.Sprintf("#(%s)", filter)
		validResource := gjson.Get(resourceJSON, wrappedFilter)

		if !validResource.Exists() {
			return false, nil
		}
	}

	return true, nil
}
