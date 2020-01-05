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

	ghostv1alpha1 "fossil.or.id/ghost-operator/pkg/apis/ghost/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *ReconcileGhostApp) CreateOrUpdateConfigMap(cr *ghostv1alpha1.GhostApp) error {
	// Default ghost.server.host config is 127.0.0.1
	// we need change this configuration to 0.0.0.0, so ingress controller can resolve this application.
	// TODO (prksu): Consider create a defaulter to create default value from api.
	cr.Spec.Config.Server.Host = "0.0.0.0"
	cr.Spec.Config.Server.Port = intstr.FromInt(int(2368))

	configdata := make(map[string]string)
	config, _ := json.MarshalIndent(cr.Spec.Config, "", "  ")
	configdata["config.json"] = string(config)

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapNameFromCR(cr),
			Namespace: cr.GetNamespace(),
			Labels:    commonLabelFromCR(cr),
		},
	}

	op, err := controllerutil.CreateOrUpdate(context.TODO(), r.client, cm, func() error {
		if err := controllerutil.SetControllerReference(cr, cm, r.scheme); err != nil {
			return err
		}

		cm.Data = configdata
		return nil
	})

	r.logger.Info("Reconciling ConfigMap", "Operation.Result", op)
	return err
}
