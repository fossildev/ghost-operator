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
	"encoding/json"
	"reflect"
	"time"

	ghostv1alpha1 "fossil.or.id/ghost-operator/pkg/apis/ghost/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	ghostPersistentVolumeClaimSuffix = "-ghost-content-pvc"
	ghostConfigMapSuffix             = "-ghost-config"
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
}

// Reconcile reads that state of the cluster for a GhostApp object and makes changes based on the state read
// and what is in the GhostApp.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileGhostApp) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling GhostApp")

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

	// Check if the GhostApp Config ConfigMap already exists, if not create a new one
	configMap := &corev1.ConfigMap{}
	configMapName := instance.GetName() + ghostConfigMapSuffix // TODO (prksu): instead of building this naming here, consider to move it to apis
	reqLogger.Info("Getting a GhostApp Config ConfigMap.", "ConfigMap.Namespace", instance.GetNamespace(), "ConfigMap.Name", configMapName)
	if err := r.client.Get(context.TODO(), types.NamespacedName{Namespace: instance.GetNamespace(), Name: configMapName}, configMap); err != nil {
		if errors.IsNotFound(err) {
			// GhostApp Config ConfigMap not exist, let's create new one
			configMap = r.newConfigMapForCR(instance)
			reqLogger.Info("Creating a new ConfigMap for GhostApp Config", "ConfigMap.Namespace", configMap.GetNamespace(), "ConfigMap.Name", configMap.GetName())
			// Set GhostApp instance as the owner and controller for this configmap
			if err := controllerutil.SetControllerReference(instance, configMap, r.scheme); err != nil {
				return reconcile.Result{}, err
			}
			// Create new configmap for ghostapp config
			if err := r.client.Create(context.TODO(), configMap); err != nil {
				reqLogger.Error(err, "Failed to create new ConfigMap for GhostApp Config", "ConfigMap.Namespace", configMap.GetNamespace(), "ConfigMap.Name", configMap.GetName())
				return reconcile.Result{}, err
			}
			reqLogger.Info("Created a new ConfigMap for GhostApp Config", "ConfigMap.Namespace", configMap.GetNamespace(), "ConfigMap.Name", configMap.GetName())
			return reconcile.Result{Requeue: true, RequeueAfter: 5 * time.Second}, nil
		}
		reqLogger.Error(err, "Failed to get ConfigMap")
		return reconcile.Result{}, err
	}

	// Create new PersistentVolumeClaim if persistent enabled
	if instance.Spec.Persistent.Enabled {
		persistentVolumeClaim := &corev1.PersistentVolumeClaim{}
		persistentVolumeClaimName := instance.GetName() + ghostPersistentVolumeClaimSuffix // TODO (prksu): instead of building this naming here, consider to move it to apis
		reqLogger.Info("Getting a GhostApp Content PersistentVolumeClaim.", "PersistentVolumeClaim.Namespace", instance.GetNamespace(), "PersistentVolumeClaim.Name", configMapName)
		if err := r.client.Get(context.TODO(), types.NamespacedName{Namespace: instance.GetNamespace(), Name: persistentVolumeClaimName}, persistentVolumeClaim); err != nil {
			if errors.IsNotFound(err) {

				// GhostApp Content PersistentVolumeClaim not exist, let's create new one
				persistentVolumeClaim = r.newGhostAppContentPVCForCR(instance)
				reqLogger.Info("Creating a new PersistentVolumeClaim for GhostApp Content", "PersistentVolumeClaim.Namespace", persistentVolumeClaim.GetNamespace(), "PersistentVolumeClaim.Name", persistentVolumeClaim.GetName())
				// Set GhostApp instance as the owner and controller for this pvc
				if err := controllerutil.SetControllerReference(instance, persistentVolumeClaim, r.scheme); err != nil {
					return reconcile.Result{}, err
				}
				// Create new configmap for ghost config
				if err := r.client.Create(context.TODO(), persistentVolumeClaim); err != nil {
					reqLogger.Error(err, "Failed to create new PersistentVolumeClaim for GhostApp Content", "PersistentVolumeClaim.Namespace", persistentVolumeClaim.GetNamespace(), "PersistentVolumeClaim.Name", persistentVolumeClaim.GetName())
					return reconcile.Result{}, err
				}
				reqLogger.Info("Created a new PersistentVolumeClaim for GhostApp Content", "PersistentVolumeClaim.Namespace", persistentVolumeClaim.GetNamespace(), "PersistentVolumeClaim.Name", persistentVolumeClaim.GetName())
				return reconcile.Result{Requeue: true}, nil
			}
			reqLogger.Error(err, "Failed to get PersistentVolumeClaim")
			return reconcile.Result{}, err
		}
	}

	// current deployment.
	currentDeployment := &appsv1.Deployment{}
	// Check if the Deployment already exists, if not create a new one
	reqLogger.Info("Getting a Deployment.", "Deployment.Namespace", instance.GetNamespace(), "Deployment.Name", instance.GetName())
	if err := r.client.Get(context.TODO(), types.NamespacedName{Namespace: instance.GetNamespace(), Name: instance.GetName()}, currentDeployment); err != nil {
		if errors.IsNotFound(err) {
			// new depployment.
			newDeployment := r.newDeploymentForCR(instance)
			reqLogger.Info("Creating a new Deployment", "Deployment.Namespace", newDeployment.GetNamespace(), "Deployment.Name", newDeployment.GetName())
			// Set GhostApp instance as the owner and controller for this deployment
			if err := controllerutil.SetControllerReference(instance, newDeployment, r.scheme); err != nil {
				return reconcile.Result{}, err
			}
			// Create new deployment for instance
			if err := r.client.Create(context.TODO(), newDeployment); err != nil {
				reqLogger.Error(err, "Failed to create new Deployment", "Deployment.Namespace", newDeployment.GetNamespace(), "Deployment.Name", newDeployment.GetName())
				return reconcile.Result{}, err
			}
			reqLogger.Info("Created a new Deployment", "Deployment.Namespace", newDeployment.GetNamespace(), "Deployment.Name", newDeployment.GetName())
			return reconcile.Result{Requeue: true, RequeueAfter: 5 * time.Second}, nil
		}
		reqLogger.Error(err, "Failed to get Deployment")
		return reconcile.Result{}, err
	}

	// will updated deployment
	willUpdatedDeployment := currentDeployment.DeepCopy()

	// Check if we got a replica update
	if !reflect.DeepEqual(currentDeployment.Spec.Replicas, instance.Spec.Replicas) {
		willUpdatedDeployment.Spec.Replicas = instance.Spec.Replicas
	}

	// Check if we got a container update
	for i, currentdep := range currentDeployment.Spec.Template.Spec.Containers {
		if !reflect.DeepEqual(currentdep.Image, instance.Spec.Image) {
			willUpdatedDeployment.Spec.Template.Spec.Containers[i].Image = instance.Spec.Image
		}
	}

	if !reflect.DeepEqual(currentDeployment, willUpdatedDeployment) {
		reqLogger.Info("Updating current Deployment", "Deployment.Namespace", willUpdatedDeployment.GetNamespace(), "Deployment.Name", willUpdatedDeployment.GetName())
		if err := r.client.Update(context.TODO(), willUpdatedDeployment); err != nil {
			return reconcile.Result{}, err
		}
	}

	// Check if the Service already exists, if not create a new one
	service := &corev1.Service{}
	reqLogger.Info("Getting a Service.", "Service.Namespace", instance.GetNamespace(), "Service.Name", instance.GetName())
	if err := r.client.Get(context.TODO(), types.NamespacedName{Namespace: instance.GetNamespace(), Name: instance.GetName()}, service); err != nil {
		if errors.IsNotFound(err) {
			service = r.newServiceForCR(instance)
			reqLogger.Info("Creating a new Service", "Service.Namespace", service.GetNamespace(), "Service.Name", service.GetName())
			// Set GhostApp instance as the owner and controller for this service
			if err := controllerutil.SetControllerReference(instance, service, r.scheme); err != nil {
				return reconcile.Result{}, err
			}
			// Create new service for instance
			if err := r.client.Create(context.TODO(), service); err != nil {
				reqLogger.Error(err, "Failed to create new Service.", "Service.Namespace", service.GetNamespace(), "Service.Name", service.GetName())
				return reconcile.Result{}, err
			}

			reqLogger.Info("Created a new Service", "Service.Namespace", service.GetNamespace(), "Service.Name", service.GetName())
			return reconcile.Result{Requeue: true}, nil
		}
		reqLogger.Error(err, "Failed to get Service")
		return reconcile.Result{}, err
	}

	// All resource already exists - don't requeue
	reqLogger.Info("Skip reconcile: Resource already exists", "GhostApp.Namespace", instance.GetNamespace(), "GhostApp.Name", instance.GetName())
	return reconcile.Result{}, nil
}

func (r *ReconcileGhostApp) newLabelForCR(cr *ghostv1alpha1.GhostApp) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":     "ghostapp",
		"app.kubernetes.io/instance": cr.GetName(),
	}
}

func (r *ReconcileGhostApp) newConfigMapForCR(cr *ghostv1alpha1.GhostApp) *corev1.ConfigMap {
	configdata := make(map[string]string)
	ghostAppConfigCMName := cr.GetName() + ghostConfigMapSuffix
	config, _ := json.MarshalIndent(cr.Spec.Config, "", "  ")
	configdata["config.json"] = string(config)

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ghostAppConfigCMName,
			Namespace: cr.GetNamespace(),
			Labels:    r.newLabelForCR(cr),
		},
		Data: configdata,
	}
}

func (r *ReconcileGhostApp) newGhostAppContentPVCForCR(cr *ghostv1alpha1.GhostApp) *corev1.PersistentVolumeClaim {
	ghostAppContentPVCName := cr.GetName() + ghostPersistentVolumeClaimSuffix
	requestStorage := make(corev1.ResourceList)
	requestStorage[corev1.ResourceStorage] = cr.Spec.Persistent.Size
	return &corev1.PersistentVolumeClaim{

		ObjectMeta: metav1.ObjectMeta{
			Name:      ghostAppContentPVCName,
			Namespace: cr.GetNamespace(),
			Labels:    r.newLabelForCR(cr),
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			StorageClassName: cr.Spec.Persistent.StorageClass,
			Resources: corev1.ResourceRequirements{
				Requests: requestStorage,
			},
		},
	}
}

func (r *ReconcileGhostApp) newServiceForCR(cr *ghostv1alpha1.GhostApp) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.GetName(),
			Namespace: cr.GetNamespace(),
			Labels:    r.newLabelForCR(cr),
		},
		Spec: corev1.ServiceSpec{
			Selector: r.newLabelForCR(cr),
			Type:     corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Protocol:   "TCP",
					Port:       int32(2368),
					TargetPort: intstr.FromInt(int(2368)),
				},
			},
		},
	}
}

func (r *ReconcileGhostApp) newDeploymentForCR(cr *ghostv1alpha1.GhostApp) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.GetName(),
			Namespace: cr.GetNamespace(),
			Labels:    r.newLabelForCR(cr),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: cr.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: r.newLabelForCR(cr),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: r.newLabelForCR(cr),
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            "ghost",
							Image:           cr.Spec.Image,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: int32(2368),
								},
							},
							Lifecycle: &corev1.Lifecycle{
								PostStart: &corev1.Handler{
									Exec: &corev1.ExecAction{
										Command: []string{"/bin/sh", "-c", "ln -sf /etc/ghost/config/config.json /var/lib/ghost/config.production.json"},
									},
								},
							},
							VolumeMounts: r.newVolumeMountForCR(cr),
						},
					},
					Volumes: r.newVolumeForCR(cr),
				},
			},
		},
	}
}

func (r *ReconcileGhostApp) newVolumeForCR(cr *ghostv1alpha1.GhostApp) []corev1.Volume {
	var volume []corev1.Volume
	volume = append(volume, corev1.Volume{
		Name: "ghost-config",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: cr.GetName() + ghostConfigMapSuffix,
				},
			},
		},
	})

	var ghostContentSource corev1.VolumeSource
	if cr.Spec.Persistent.Enabled {
		ghostContentSource = corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: cr.GetName() + ghostPersistentVolumeClaimSuffix,
			},
		}
	} else {

		ghostContentSource = corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		}
	}

	volume = append(volume, corev1.Volume{
		Name:         "ghost-content",
		VolumeSource: ghostContentSource,
	})

	return volume
}

func (r *ReconcileGhostApp) newVolumeMountForCR(cr *ghostv1alpha1.GhostApp) []corev1.VolumeMount {
	var volumeMount []corev1.VolumeMount

	volumeMount = append(volumeMount, corev1.VolumeMount{
		Name:      "ghost-config",
		ReadOnly:  true,
		MountPath: "/etc/ghost/config",
	})
	volumeMount = append(volumeMount, corev1.VolumeMount{
		Name:      "ghost-content",
		ReadOnly:  false,
		MountPath: "/var/lib/ghost/content",
	})

	return volumeMount

}
