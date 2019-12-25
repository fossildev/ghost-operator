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
	"testing"
	"time"

	ghostv1alpha1 "fossil.or.id/ghost-operator/pkg/apis/ghost/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

func TestGostAppController(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))

	var objs []runtime.Object
	var (
		name      = "test-ghostapp"
		namespace = "ghost"
		replicas  = int32(1)
	)

	ghostapp := &ghostv1alpha1.GhostApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: ghostv1alpha1.GhostAppSpec{
			Replicas: &replicas,
			Image:    "ghost:3",
			Config: ghostv1alpha1.GhostConfigSpec{
				URL: "http://example.ghostapp.test:2368",
				Database: ghostv1alpha1.GhostDatabaseSpec{
					Client: "sqlite3",
					Connection: ghostv1alpha1.GhostDatabaseConnectionSpec{
						Filename: "testghost.db",
					},
				},
			},
		},
	}

	objs = append(objs, ghostapp)
	s := scheme.Scheme
	s.AddKnownTypes(ghostv1alpha1.SchemeGroupVersion, ghostapp)

	fc := fake.NewFakeClient(objs...)
	rec := ReconcileGhostApp{fc, s}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}

	res, err := rec.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}

	// We expected to requeue
	if !res.Requeue {
		t.Error("reconcile did not requeue request as expected")
	}

	// Check if Deployment has been created and has the correct size.
	deployment := &appsv1.Deployment{}
	if err = fc.Get(context.TODO(), req.NamespacedName, deployment); err != nil && !errors.IsNotFound(err) {
		t.Fatalf("get deployment: (%v)", err)
	}

	res, err = rec.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}

	// We expected to requeue afeter 5 second
	if !res.Requeue && res.RequeueAfter != 5*time.Second {
		t.Error("reconcile requeue which is not expected")
	}

	// Check if Service has been created.
	service := &corev1.Service{}
	if err = fc.Get(context.TODO(), req.NamespacedName, service); err != nil && !errors.IsNotFound(err) {
		t.Fatalf("get service: (%v)", err)
	}

	res, err = rec.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}

	// We expected to requeue
	if !res.Requeue {
		t.Error("reconcile requeue which is not expected")
	}

	res, err = rec.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}

	// After all we expected rempty result
	if res != (reconcile.Result{}) {
		t.Error("reconcile did not return an empty Result")
	}
}
