/*
Copyright 2019 Marek Cermak <macermak@redhat.com>.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	v1alpha1 "github.com/cermakm/api/v1alpha1"
	"github.com/go-logr/logr"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// AwaitReconciler reconciles a Await object
type AwaitReconciler struct {
	client.Client
	Log logr.Logger
}

// +kubebuilder:rbac:groups=await.argoproj.io,resources=awaits,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=await.argoproj.io,resources=awaits/status,verbs=get;update;patch

func (r *AwaitReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("await", req.NamespacedName)

	// your logic here

	return ctrl.Result{}, nil
}

func (r *AwaitReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Await{}).
		Complete(r)
}

// GetWorkflow retrieves the Workflow resource from the given namespace or dies
func (r *AwaitReconciler) getWorkflow(workflow v1alpha1.NamespacedWorkflow) (*v1alpha1.Workflow, error) {
	log := r.Log.WithValues(
		"Workflow.Name", workflow.Name, "Workflow.Namespace", workflow.Namespace)

	res := &v1alpha1.Workflow{}
	err := r.Get(
		context.TODO(), types.NamespacedName{Name: res.Name, Namespace: workflow.Namespace}, res)
	if err != nil {
		log.Error(err, "Error getting the Workflow Resource.")

		if errors.IsNotFound(err) {
			log.Error(err, "Resource Workflow was not found.")
		}

		return nil, err
	}

	return res, nil
}

// ResumeWorkflow resumes a Workflow identified by its Name
func (r *AwaitReconciler) resumeWorkflow(workflow *v1alpha1.Workflow) func() error {
	resumeWorkflow := func() error {
		log := r.Log.WithValues(
			"Workflow.Name", workflow.Name, "Workflow.Namespace", workflow.Namespace)
		log.Info("Resuming workflow")

		// TODO: Resume the workflow

		return nil
	}

	return resumeWorkflow
}
