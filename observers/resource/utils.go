package resource

import (
	"encoding/json"
	"errors"
	"fmt"

	gjson "github.com/tidwall/gjson"

	"k8s.io/apimachinery/pkg/watch"
)


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
