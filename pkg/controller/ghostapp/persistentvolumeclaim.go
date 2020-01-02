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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *ReconcileGhostApp) CreateOrUpdatePersistentVolumeClaim(cr *ghostv1alpha1.GhostApp) error {
	requestStorage := make(corev1.ResourceList)
	requestStorage[corev1.ResourceStorage] = cr.Spec.Persistent.Size
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      persistentVolumeClaimNameFromCR(cr),
			Namespace: cr.GetNamespace(),
			Labels:    commonLabelFromCR(cr),
		},
	}

	op, err := controllerutil.CreateOrUpdate(context.TODO(), r.client, pvc, func() error {
		if err := controllerutil.SetControllerReference(cr, pvc, r.scheme); err != nil {
			return err
		}

		if pvc.ObjectMeta.CreationTimestamp.IsZero() {
			pvc.Spec = corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{
					corev1.ReadWriteOnce,
				},
				StorageClassName: cr.Spec.Persistent.StorageClass,
			}
		}

		pvc.Spec.Resources = corev1.ResourceRequirements{
			Requests: requestStorage,
		}

		return nil
	})

	r.logger.Info("Reconciling PersistentVolumeClaim", "Operation.Result", op)
	return err
}
