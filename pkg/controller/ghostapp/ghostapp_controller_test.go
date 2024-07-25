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
	"testing"

	ghostv1alpha1 "fossil.or.id/ghost-operator/pkg/apis/ghost/v1alpha1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

func TestReconciler(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))

	var (
		replicas = int32(1)
	)

	tests := []struct {
		name    string
		resouce *ghostv1alpha1.GhostApp
		wantErr bool
	}{
		{
			name: "Test Create GhostApp Instance",
			resouce: &ghostv1alpha1.GhostApp{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ghostapp",
					Namespace: "ghost",
				},
				Spec: ghostv1alpha1.GhostAppSpec{
					Replicas: &replicas,
					Image:    "ghost:3",
					Config: ghostv1alpha1.GhostConfigSpec{
						URL: "http://example.ghostapp.test",
						Database: ghostv1alpha1.GhostDatabaseSpec{
							Client: "sqlite3",
							Connection: ghostv1alpha1.GhostDatabaseConnectionSpec{
								Filename: "testghost.db",
							},
						},
					},
				},
			},
		},
		{
			name: "Test Create GhostApp Instance With Persistent Volume",
			resouce: &ghostv1alpha1.GhostApp{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ghostapp-with-persistent-volume",
					Namespace: "ghost",
				},
				Spec: ghostv1alpha1.GhostAppSpec{
					Replicas: &replicas,
					Image:    "ghost:3",
					Config: ghostv1alpha1.GhostConfigSpec{
						URL: "http://example.ghostapp.test",
						Database: ghostv1alpha1.GhostDatabaseSpec{
							Client: "sqlite3",
							Connection: ghostv1alpha1.GhostDatabaseConnectionSpec{
								Filename: "testghost.db",
							},
						},
					},
					Persistent: ghostv1alpha1.GhostPersistentSpec{
						Enabled: true,
						Size:    resource.MustParse("10Gi"),
					},
				},
			},
		},
		{
			name: "Test Create GhostApp Instance With Ingress",
			resouce: &ghostv1alpha1.GhostApp{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ghostapp-with-ingress",
					Namespace: "ghost",
				},
				Spec: ghostv1alpha1.GhostAppSpec{
					Replicas: &replicas,
					Image:    "ghost:3",
					Config: ghostv1alpha1.GhostConfigSpec{
						URL: "http://example.ghostapp.test",
						Database: ghostv1alpha1.GhostDatabaseSpec{
							Client: "sqlite3",
							Connection: ghostv1alpha1.GhostDatabaseConnectionSpec{
								Filename: "testghost.db",
							},
						},
					},
					Ingress: ghostv1alpha1.GhostIngressSpec{
						Enabled: true,
					},
				},
			},
		},
		{
			name: "Test Create GhostApp Instance With Ingress TLS",
			resouce: &ghostv1alpha1.GhostApp{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ghostapp-with-ingress-tls",
					Namespace: "ghost",
				},
				Spec: ghostv1alpha1.GhostAppSpec{
					Replicas: &replicas,
					Image:    "ghost:3",
					Config: ghostv1alpha1.GhostConfigSpec{
						URL: "http://example.ghostapp.test",
						Database: ghostv1alpha1.GhostDatabaseSpec{
							Client: "sqlite3",
							Connection: ghostv1alpha1.GhostDatabaseConnectionSpec{
								Filename: "testghost.db",
							},
						},
					},
					Ingress: ghostv1alpha1.GhostIngressSpec{
						Enabled: true,
						Hosts:   []string{"example.ghostapp.test"},
						TLS: ghostv1alpha1.GhostIngressTLSSpec{
							Enabled:    true,
							SecretName: "test-ghostapp-with-ingress-tls",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := scheme.Scheme
			s.AddKnownTypes(ghostv1alpha1.SchemeGroupVersion, tt.resouce)

			var objs []runtime.Object
			objs = append(objs, tt.resouce)
			f := fake.NewFakeClient(objs...)

			request := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      tt.resouce.Name,
					Namespace: tt.resouce.Namespace,
				},
			}

			r := ReconcileGhostApp{f, s, log}
			result, err := r.Reconcile(request)
			if err != nil && !tt.wantErr {
				t.Fatalf("reconcile: (%v)", err)
			}

			if result != (reconcile.Result{}) {
				t.Error("reconcile did not return an empty result")
			}
		})
	}
}
