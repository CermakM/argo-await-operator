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
	"errors"

	workflowv1alpha1 "github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
	argoprojv1alpha1 "github.com/argoproj/argo/pkg/client/clientset/versioned/typed/workflow/v1alpha1"
	workflowutil "github.com/argoproj/argo/workflow/util"

	v1alpha1 "github.com/cermakm/argo-await-operator/api/v1alpha1"

	// "github.com/cermakm/argo-await-operator/observers/resource"
	"github.com/go-logr/logr"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// AwaitReconciler reconciles a Await object
type AwaitReconciler struct {
	client.Client

	Log    logr.Logger
	Config *rest.Config
}

// +kubebuilder:rbac:groups=await.argoproj.io,resources=awaits,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=await.argoproj.io,resources=awaits/status,verbs=get;update;patch

func (r *AwaitReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	log := r.Log.WithValues("await", req)

	// Fetch the Await instance
	awaitResource := &v1alpha1.Await{}
	err := r.Get(context.TODO(), req.NamespacedName, awaitResource)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	wf, err := r.SuspendedWorkflowResource(awaitResource.Spec.Workflow)
	if err != nil {
		log.Error(err, "The requested Workflow was not found", "workflow", awaitResource.Spec.Workflow)

		// The Workflow to be resumed does not exist anymore, don't reque
		return ctrl.Result{}, nil
	}
	log.V(1).Info("Found matching Workflow resource", "resource", wf)

	// observer := resource.NewObserverForConfig(r.Config)

	// resourceSpec := awaitResource.Spec.Resource
	// res, err := observer.Get(resourceSpec.Name)
	// if err != nil {
	// 	r.Log.Error(err, "The requested resource was not found", "resource", res)

	// 	return ctrl.Result{}, err
	// }
	// if res != nil {
	// 	// Await the resource and resume the workflow when it appears
	// 	go observer.AwaitResource(
	// 		r.resumeWorkflow(wf), res, req.Namespace, awaitResource.Spec.Filters)

	// 	// Observer created successfully - don't requeue
	// 	return ctrl.Result{}, nil
	// }

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller
func (r *AwaitReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Await{}).
		Complete(r)
}

// SuspendedWorkflowResource retrieves the Workflow resource from the given namespace which requested the await
func (r *AwaitReconciler) SuspendedWorkflowResource(workflow v1alpha1.NamespacedWorkflow) (*workflowv1alpha1.Workflow, error) {
	clientset := argoprojv1alpha1.NewForConfigOrDie(r.Config)
	workflows := clientset.Workflows(workflow.Namespace)

	wf, err := workflows.Get(workflow.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	if workflowutil.IsWorkflowSuspended(wf) != true {
		return nil, errors.New("workflow is not suspended")
	}

	return wf, nil
}

func (r *AwaitReconciler) resumeWorkflowCallback(workflow *workflowv1alpha1.Workflow) func() error {
	f := func() error {
		log := r.Log.WithValues(
			"Workflow.Name", workflow.Name, "Workflow.Namespace", workflow.Namespace)
		log.Info("Resuming workflow")

		clientset := argoprojv1alpha1.NewForConfigOrDie(r.Config)
		workflows := clientset.Workflows(workflow.Namespace)

		return workflowutil.ResumeWorkflow(workflows, workflow.Name)
	}

	return f
}
