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
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *ReconcileGhostApp) CreateOrUpdateIngress(cr *ghostv1alpha1.GhostApp) error {
	ing := &networkingv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        cr.GetName(),
			Namespace:   cr.GetNamespace(),
			Labels:      commonLabelFromCR(cr),
			Annotations: cr.Spec.Ingress.Annotations,
		},
	}

	op, err := controllerutil.CreateOrUpdate(context.TODO(), r.client, ing, func() error {
		if err := controllerutil.SetControllerReference(cr, ing, r.scheme); err != nil {
			return err
		}

		// we only create ingress rule with backend service created by this operator.
		ingressRule := networkingv1beta1.IngressRuleValue{
			HTTP: &networkingv1beta1.HTTPIngressRuleValue{
				Paths: []networkingv1beta1.HTTPIngressPath{
					{
						Backend: networkingv1beta1.IngressBackend{
							ServiceName: cr.GetName(),
							ServicePort: intstr.FromInt(int(2368)),
						},
					},
				},
			},
		}

		hosts := cr.Spec.Ingress.Hosts
		rules := []networkingv1beta1.IngressRule{}
		// if no one hosts defined, create rule without any spesific host.
		if len(hosts) == 0 {
			rules = append(rules, networkingv1beta1.IngressRule{
				IngressRuleValue: ingressRule,
			})
		} else {
			for _, host := range hosts {
				rules = append(rules, networkingv1beta1.IngressRule{
					Host:             host,
					IngressRuleValue: ingressRule,
				})
			}
		}

		ing.Spec.Rules = rules
		// if tls is enabled add `IngressTLS` with defined hosts.
		// NOTE: hosts is required when tls is enabled.
		if cr.Spec.Ingress.TLS.Enabled {
			ing.Spec.TLS = append(ing.Spec.TLS, networkingv1beta1.IngressTLS{
				Hosts:      hosts,
				SecretName: cr.Spec.Ingress.TLS.SecretName,
			})
		}

		return nil
	})

	r.logger.Info("Reconciling Ingress", "Operation.Result", op)
	return err
}
