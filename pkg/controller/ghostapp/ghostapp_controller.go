// Copyright 2019 Fossil Dev
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ghostapp

import (
	"context"

	ghostv1alpha1 "fossil.or.id/ghost-operator/pkg/apis/ghost/v1alpha1"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_ghostapp")

// Add creates a new GhostApp Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileGhostApp{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("ghostapp-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource GhostApp
	if err := c.Watch(&source.Kind{Type: &ghostv1alpha1.GhostApp{}}, &handler.EnqueueRequestForObject{}); err != nil {
		return err
	}

	owner := &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &ghostv1alpha1.GhostApp{},
	}

	// Watch for changes to ConfigMap and requeue the owner GhostApp
	if err := c.Watch(&source.Kind{Type: &corev1.ConfigMap{}}, owner); err != nil {
		return err
	}

	// Watch for changes to PersistentVolumeClaim and requeue the owner GhostApp
	if err := c.Watch(&source.Kind{Type: &corev1.PersistentVolumeClaim{}}, owner); err != nil {
		return err
	}

	// Watch for changes to Deployment and requeue the owner GhostApp
	if err := c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, owner); err != nil {
		return err
	}

	// Watch for changes to Service and requeue the owner GhostApp
	if err := c.Watch(&source.Kind{Type: &corev1.Service{}}, owner); err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileGhostApp implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileGhostApp{}

// ReconcileGhostApp reconciles a GhostApp object
type ReconcileGhostApp struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
	logger logr.Logger
}

// Reconcile reads that state of the cluster for a GhostApp object and makes changes based on the state read
// and what is in the GhostApp.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileGhostApp) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling GhostApp")
	r.logger = reqLogger

	// Fetch the GhostApp instance
	instance := &ghostv1alpha1.GhostApp{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if err := r.CreateOrUpdateConfigMap(instance); err != nil {
		instance.Status.Phase = ghostv1alpha1.GhostAppPhaseFailure
		instance.Status.Reason = err.Error()
		if err := r.client.Status().Update(context.TODO(), instance); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, err
	}

	// Create new PersistentVolumeClaim if persistent is enabled
	if instance.Spec.Persistent.Enabled {
		if err := r.CreateOrUpdatePersistentVolumeClaim(instance); err != nil {
			instance.Status.Phase = ghostv1alpha1.GhostAppPhaseFailure
			instance.Status.Reason = err.Error()
			if err := r.client.Status().Update(context.TODO(), instance); err != nil {
				return reconcile.Result{}, err
			}
			return reconcile.Result{}, err
		}
	}

	if err := r.CreateOrUpdateDeployment(instance); err != nil {
		instance.Status.Phase = ghostv1alpha1.GhostAppPhaseFailure
		instance.Status.Reason = err.Error()
		if err := r.client.Status().Update(context.TODO(), instance); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, err
	}

	if err := r.CreateOrUpdateService(instance); err != nil {
		instance.Status.Phase = ghostv1alpha1.GhostAppPhaseFailure
		instance.Status.Reason = err.Error()
		if err := r.client.Status().Update(context.TODO(), instance); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, err
	}

	if instance.Spec.Ingress.Enabled {
		if err := r.CreateOrUpdateIngress(instance); err != nil {
			instance.Status.Phase = ghostv1alpha1.GhostAppPhaseFailure
			instance.Status.Reason = err.Error()
			if err := r.client.Status().Update(context.TODO(), instance); err != nil {
				return reconcile.Result{}, err
			}
			return reconcile.Result{}, err
		}
	}

	// Set status phase to Running
	instance.Status.Replicas = *instance.Spec.Replicas
	instance.Status.Phase = ghostv1alpha1.GhostAppPhaseRunning
	if err := r.client.Status().Update(context.TODO(), instance); err != nil {
		return reconcile.Result{}, err
	}

	// All resource already up to date - don't requeue
	return reconcile.Result{}, nil
}
