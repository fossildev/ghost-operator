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
	ghostv1alpha1 "fossil.or.id/ghost-operator/pkg/apis/ghost/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func commonLabelFromCR(cr *ghostv1alpha1.GhostApp) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":     "ghostapp",
		"app.kubernetes.io/instance": cr.GetName(),
	}
}

func commonLabelSelectorFromCR(cr *ghostv1alpha1.GhostApp) *metav1.LabelSelector {
	return &metav1.LabelSelector{
		MatchLabels: commonLabelFromCR(cr),
	}
}

func configMapNameFromCR(cr *ghostv1alpha1.GhostApp) string { return cr.GetName() + "-ghost-config" }
func persistentVolumeClaimNameFromCR(cr *ghostv1alpha1.GhostApp) string {
	return cr.GetName() + "-ghost-content-pvc"
}
