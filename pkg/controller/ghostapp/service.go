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
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *ReconcileGhostApp) CreateOrUpdateService(cr *ghostv1alpha1.GhostApp) error {
	serviceType := corev1.ServiceTypeClusterIP
	// set serviceType to NodePort when ingress enabled
	// we need this type for some ingress-class eg: gce
	// default is ClusterIP
	// TODO (prksu): Consider adding ServiceType field in GhostAppSpec
	if cr.Spec.Ingress.Enabled {
		serviceType = corev1.ServiceTypeNodePort
	}

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.GetName(),
			Namespace: cr.GetNamespace(),
			Labels:    commonLabelFromCR(cr),
		},
	}

	op, err := controllerutil.CreateOrUpdate(context.TODO(), r.client, svc, func() error {
		// We don't accept any update for service
		if !svc.ObjectMeta.CreationTimestamp.IsZero() {
			return nil
		}

		if err := controllerutil.SetControllerReference(cr, svc, r.scheme); err != nil {
			return err
		}

		svc.Spec = corev1.ServiceSpec{
			Selector: commonLabelFromCR(cr),
			Type:     serviceType,
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Protocol:   "TCP",
					Port:       int32(2368),
					TargetPort: intstr.FromInt(int(2368)),
				},
			},
		}
		return nil
	})

	r.logger.Info("Reconciling Service", "Operation.Result", op)
	return err
}
