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

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	awaitv1alpha1 "github.com/cermakm/api/v1alpha1"
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
		For(&awaitv1alpha1.Await{}).
		Complete(r)
}
