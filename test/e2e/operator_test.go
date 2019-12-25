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

package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	"fossil.or.id/ghost-operator/pkg/apis"
	ghostv1alpha1 "fossil.or.id/ghost-operator/pkg/apis/ghost/v1alpha1"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	retryInterval        = time.Second * 5
	timeout              = time.Second * 60
	cleanupRetryInterval = time.Second * 1
	cleanupTimeout       = time.Second * 5
)

func TestMain(m *testing.M) {
	framework.MainEntry(m)
}

func TestGhost(t *testing.T) {
	ghostappList := &ghostv1alpha1.GhostAppList{}
	if err := framework.AddToFrameworkScheme(apis.AddToScheme, ghostappList); err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}

	// run subtests
	t.Run("ghost-opeator", func(t *testing.T) {
		t.Run("GhostOpeator", GhostOperator)
	})
}

func GhostOperator(t *testing.T) {
	t.Parallel()

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()

	if err := ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval}); err != nil {
		t.Fatalf("failed to initialize cluster resources: %v", err)
	}
	t.Log("Initialized cluster resources")
	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatal(err)
	}

	// get global framework variables
	f := framework.Global
	// wait for ghost-operator to be ready
	if err = e2eutil.WaitForOperatorDeployment(t, f.KubeClient, namespace, "ghost-operator", 1, retryInterval, timeout); err != nil {
		t.Fatal(err)
	}

	if err = CreateGhostAppTest(t, f, ctx); err != nil {
		t.Fatal(err)
	}
}

func CreateGhostAppTest(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) error {
	namespace, err := ctx.GetNamespace()
	if err != nil {
		return fmt.Errorf("could not get namespace: %v", err)
	}

	exampleGhostAppReplicas := int32(1)
	exampleGhostApp := &ghostv1alpha1.GhostApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-ghost",
			Namespace: namespace,
		},
		Spec: ghostv1alpha1.GhostAppSpec{
			Replicas: &exampleGhostAppReplicas,
			Image:    "ghost",
			Config: ghostv1alpha1.GhostConfigSpec{
				URL: "http://example.ghostapp.test",
				Database: ghostv1alpha1.GhostDatabaseSpec{
					Client: "sqlite3",
					Connection: ghostv1alpha1.GhostDatabaseConnectionSpec{
						Filename: "/var/lib/ghost/content/data/ghost.db",
					},
				},
			},
		},
	}

	if err = f.Client.Create(context.TODO(), exampleGhostApp, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval}); err != nil {
		return err
	}

	if err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "example-ghost", 1, retryInterval, timeout); err != nil {
		return err
	}

	return nil
}
