// Copyright Project Contour Authors
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

//go:build e2e
// +build e2e

package incluster

import (
	"context"
	"time"

	sesamev1 "github.com/projectsesame/sesame/apis/projectsesame/v1"
	sesamev1alpha1 "github.com/projectsesame/sesame/apis/projectsesame/v1alpha1"
	"github.com/projectsesame/sesame/test/e2e"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func testprojectsesameResourcesRBAC(namespace string) {
	Specify("Sesame ClusterRole is set up to allow access to projectsesame.io API group resources and resource status", func() {
		f.Fixtures.Echo.Deploy(namespace, "echo")

		otherNS := "another-" + namespace
		f.CreateNamespace(otherNS)
		defer f.DeleteNamespace(otherNS, false)
		f.Certs.CreateSelfSignedCert(otherNS, "delegated-cert", "delegated-cert", "rbac-test.projectsesame.io")

		// HTTPProxy and TLSCertificateDelegation
		t := &sesamev1.TLSCertificateDelegation{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: otherNS,
				Name:      "rbac",
			},
			Spec: sesamev1.TLSCertificateDelegationSpec{
				Delegations: []sesamev1.CertificateDelegation{
					{
						SecretName:       "delegated-cert",
						TargetNamespaces: []string{namespace},
					},
				},
			},
		}
		require.NoError(f.T(), f.Client.Create(context.TODO(), t))

		p := &sesamev1.HTTPProxy{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      "rbac",
			},
			Spec: sesamev1.HTTPProxySpec{
				VirtualHost: &sesamev1.VirtualHost{
					Fqdn: "rbac-test.projectsesame.io",
					TLS: &sesamev1.TLS{
						SecretName: otherNS + "/delegated-cert",
					},
				},
				Routes: []sesamev1.Route{
					{
						Services: []sesamev1.Service{
							{Name: "invalid-service", Port: 80},
						},
					},
				},
			},
		}
		_, success := f.CreateHTTPProxyAndWaitFor(p, e2e.HTTPProxyInvalid)
		assert.True(f.T(), success)

		// Update HTTPProxy to valid service.
		require.NoError(f.T(), retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			if err := f.Client.Get(context.TODO(), client.ObjectKeyFromObject(p), p); err != nil {
				return err
			}

			p.Spec.Routes[0].Services[0].Name = "echo"
			return f.Client.Update(context.TODO(), p)
		}))

		assert.Eventually(f.T(), func() bool {
			if err := f.Client.Get(context.TODO(), client.ObjectKeyFromObject(p), p); err != nil {
				return false
			}
			return e2e.HTTPProxyValid(p)
		}, time.Second*5, time.Millisecond*20)

		// Check request to app works.
		res, ok := f.HTTP.SecureRequestUntil(&e2e.HTTPSRequestOpts{
			Host:      "rbac-test.projectsesame.io",
			Condition: e2e.HasStatusCode(200),
		})
		assert.Truef(f.T(), ok, "expected %d response code, got %d", 200, res.StatusCode)

		// ExtensionService
		e := &sesamev1alpha1.ExtensionService{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      "rbac",
			},
			Spec: sesamev1alpha1.ExtensionServiceSpec{
				Services: []sesamev1alpha1.ExtensionServiceTarget{
					{Name: "invalid-service", Port: 80},
				},
			},
		}
		require.NoError(f.T(), f.Client.Create(context.TODO(), e))
		assert.Eventually(f.T(), func() bool {
			if err := f.Client.Get(context.TODO(), client.ObjectKeyFromObject(e), e); err != nil {
				return false
			}
			return e2e.DetailedConditionInvalid(e.Status.Conditions)
		}, time.Second*5, time.Millisecond*20)
	})
}

func testIngressResourceRBAC(namespace string) {
	Specify("Sesame ClusterRole is set up to allow access to Ingress v1 resources and resource status", func() {
		f.Fixtures.Echo.Deploy(namespace, "echo")

		i := &networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      "rbac",
			},
			Spec: networkingv1.IngressSpec{
				Rules: []networkingv1.IngressRule{
					{
						Host: "rbac-test-ingress.projectsesame.io",
						IngressRuleValue: networkingv1.IngressRuleValue{
							HTTP: &networkingv1.HTTPIngressRuleValue{
								Paths: []networkingv1.HTTPIngressPath{
									{
										PathType: e2e.IngressPathTypePtr(networkingv1.PathTypePrefix),
										Path:     "/",
										Backend: networkingv1.IngressBackend{
											Service: &networkingv1.IngressServiceBackend{
												Name: "echo",
												Port: networkingv1.ServiceBackendPort{Number: 80},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}
		require.NoError(f.T(), f.Client.Create(context.TODO(), i))

		// TODO: In kind, the Envoy service does not get a load
		// balancer address and currently the default
		// deployment does not yet utilize the
		// --ingress-status-address flag.
		//
		// Make sure Sesame has updated load balancer status
		// assert.Eventually(f.T(), func() bool {
		// 	if err := f.Client.Get(context.TODO(), client.ObjectKeyFromObject(i), i); err != nil {
		// 		return false
		// 	}
		// 	return len(i.Status.LoadBalancer.Ingress) > 0
		// }, time.Second*5, time.Millisecond*20)

		// Check Sesame has read Ingress resource and
		// programmed a route.
		res, ok := f.HTTP.RequestUntil(&e2e.HTTPRequestOpts{
			Host:      "rbac-test-ingress.projectsesame.io",
			Path:      "/",
			Condition: e2e.HasStatusCode(200),
		})
		require.NotNil(f.T(), res, "request never succeeded")
		require.Truef(f.T(), ok, "expected 200 response code, got %d", res.StatusCode)
	})
}
