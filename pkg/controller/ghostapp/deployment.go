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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *ReconcileGhostApp) CreateOrUpdateDeployment(cr *ghostv1alpha1.GhostApp) error {
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.GetName(),
			Namespace: cr.GetNamespace(),
			Labels:    commonLabelFromCR(cr),
		},
	}

	op, err := controllerutil.CreateOrUpdate(context.TODO(), r.client, dep, func() error {
		if dep.ObjectMeta.CreationTimestamp.IsZero() {
			// Set label selector only when deployment has never been created
			dep.Spec.Selector = commonLabelSelectorFromCR(cr)
		}

		if err := controllerutil.SetControllerReference(cr, dep, r.scheme); err != nil {
			return err
		}

		dep.Spec.Replicas = cr.Spec.Replicas
		dep.Spec.Template = corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: commonLabelFromCR(cr),
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
		}
		return nil
	})

	r.logger.Info("Reconciling Deployment", "Operation.Result", op)
	return err
}

func (r *ReconcileGhostApp) newVolumeForCR(cr *ghostv1alpha1.GhostApp) []corev1.Volume {
	var volume []corev1.Volume
	volume = append(volume, corev1.Volume{
		Name: "ghost-config",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: configMapNameFromCR(cr),
				},
			},
		},
	})

	var ghostContentSource corev1.VolumeSource
	if cr.Spec.Persistent.Enabled {
		ghostContentSource = corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: persistentVolumeClaimNameFromCR(cr),
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
