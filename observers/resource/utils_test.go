package resource

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
)

const Fake watch.EventType = "FAKE"

var (
	FakeResource = &metav1.APIResource{
		Name:    "fakes",
		Group:   "fake-group",
		Version: "v1",
		Kind:    "Fake",
		Verbs:   []string{},
	}
	FakeUnstructuredResource, _ = runtime.DefaultUnstructuredConverter.ToUnstructured(FakeResource)
)

func Test_passFilters(t *testing.T) {
	type args struct {
		object  map[string]interface{}
		filters []string
	}

	FakeUnstructuredResource["metadata"] = map[string]interface{}{"name": "fake-name"}

	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "empty list of filters",
			args: args{
				object:  FakeUnstructuredResource,
				filters: []string{},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "single filter",
			args: args{
				object:  FakeUnstructuredResource,
				filters: []string{"kind==Fake"},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "nested filter",
			args: args{
				object:  FakeUnstructuredResource,
				filters: []string{"metadata.name==fake-name"},
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := passFilters(tt.args.object, tt.args.filters...)
			if (err != nil) != tt.wantErr {
				t.Errorf("passFilters() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("passFilters() = %v, want %v", got, tt.want)
			}
		})
	}
}
