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

	workflowv1alpha1 "github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
	argoprojv1alpha1 "github.com/argoproj/argo/pkg/client/clientset/versioned/typed/workflow/v1alpha1"
	workflowutil "github.com/argoproj/argo/workflow/util"

	v1alpha1 "github.com/cermakm/argo-await-operator/api/v1alpha1"

	"github.com/cermakm/argo-await-operator/observers/resource"
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
	log := r.Log.WithValues("request", req)

	// Fetch the Await instance
	res := &v1alpha1.Await{}
	err := r.Get(context.TODO(), req.NamespacedName, res)
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

	wf, err := r.getWorkflowResource(res.Spec.Workflow)
	if err != nil {
		log.Error(err, "the requested Workflow was not found", "workflow", res.Spec.Workflow)

		// The Workflow to be resumed does not exist, don't reque
		return ctrl.Result{Requeue: false}, err
	}

	if workflowutil.IsWorkflowSuspended(wf) != true {
		status := getWorkflowStatus(wf)
		log.Info("workflow is not suspended, reconciling", "status", status)

		// The Workflow exists, but is not suspended (possibly yet), reque
		return ctrl.Result{}, nil
	}

	observer, err := resource.NewObserverForResource(r.Config, &res.Spec.Resource, res.Spec.Filters)
	if err != nil {
		log.Error(err, "observer could not be created")
		return ctrl.Result{Requeue: false}, err
	}

	// Await the requested Resource and then resume the Workflow
	callback := r.workflowResumeCallback(wf)
	go observer.Await(callback)

	// Observer created successfully - don't requeue
	return ctrl.Result{}, nil
}

type workflowStatus struct {
	Phase workflowv1alpha1.NodePhase `json:"phase"`
	Nodes []nodeStatus               `json:"nodes"`
}

type nodeStatus struct {
	// Name is unique name in the node tree used to generate the node ID
	Name  string                     `json:"name"`
	Type  workflowv1alpha1.NodeType  `json:"type"`
	Phase workflowv1alpha1.NodePhase `json:"phase"`
}

func getWorkflowStatus(workflow *workflowv1alpha1.Workflow) *workflowStatus {
	status := &workflowStatus{Phase: workflow.Status.Phase}

	for _, node := range workflow.Status.Nodes {
		ns := nodeStatus{
			Name:  node.Name,
			Type:  node.Type,
			Phase: node.Phase,
		}
		status.Nodes = append(status.Nodes, ns)
	}

	return status
}

// getWorkflowResource retrieves the Workflow resource from the given namespace which requested the await
func (r *AwaitReconciler) getWorkflowResource(workflow v1alpha1.NamespacedWorkflow) (*workflowv1alpha1.Workflow, error) {
	clientset := argoprojv1alpha1.NewForConfigOrDie(r.Config)
	workflows := clientset.Workflows(workflow.Namespace)

	wf, err := workflows.Get(workflow.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return wf, nil
}

// workflowResumeCallback returns the callback function
// which should be run after the resource has been awaited
func (r *AwaitReconciler) workflowResumeCallback(workflow *workflowv1alpha1.Workflow) func() error {
	f := func() error {
		log := r.Log.WithValues(
			"Workflow.Name", workflow.Name, "Workflow.Namespace", workflow.Namespace)
		log.Info("resuming workflow")

		clientset := argoprojv1alpha1.NewForConfigOrDie(r.Config)
		workflows := clientset.Workflows(workflow.Namespace)

		err := workflowutil.ResumeWorkflow(workflows, workflow.Name)
		if err != nil {
			log.Error(err, "failed to resume workflow")
			return err
		}

		log.Info("workflow successfully resumed.")
		return nil
	}

	return f
}

// SetupWithManager sets up the controller
func (r *AwaitReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Await{}).
		Complete(r)
}
