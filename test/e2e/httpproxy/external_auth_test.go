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

package httpproxy

import (
	"context"

	sesamev1 "github.com/projectsesame/sesame/apis/projectsesame/v1"
	sesamev1alpha1 "github.com/projectsesame/sesame/apis/projectsesame/v1alpha1"
	"github.com/projectsesame/sesame/test/e2e"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func testExternalAuth(namespace string) {
	Specify("external auth can be configured on an HTTPRoute", func() {
		t := f.T()

		f.Fixtures.Echo.Deploy(namespace, "echo")
		f.Certs.CreateSelfSignedCert(namespace, "echo", "echo", "externalauth.projectsesame.io")

		f.Certs.CreateSelfSignedCert(namespace, "testserver-cert", "testserver-cert", "testserver")

		// auth testserver
		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      "testserver",
				Labels: map[string]string{
					"app.kubernetes.io/name": "testserver",
				},
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"app.kubernetes.io/name": "testserver"},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{"app.kubernetes.io/name": "testserver"},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:            "testserver",
								Image:           "docker.io/projectsesame/sesame-authserver:v2",
								ImagePullPolicy: corev1.PullIfNotPresent,
								Command: []string{
									"/sesame-authserver",
								},
								Args: []string{
									"testserver",
									"--address=:9443",
									"--tls-ca-path=/tls/ca.crt",
									"--tls-cert-path=/tls/tls.crt",
									"--tls-key-path=/tls/tls.key",
								},
								Ports: []corev1.ContainerPort{
									{
										Name:          "auth",
										ContainerPort: 9443,
										Protocol:      corev1.ProtocolTCP,
									},
								},
								VolumeMounts: []corev1.VolumeMount{
									{
										Name:      "tls",
										MountPath: "/tls",
										ReadOnly:  true,
									},
								},
							},
						},
						Volumes: []corev1.Volume{
							{
								Name: "tls",
								VolumeSource: corev1.VolumeSource{
									Secret: &corev1.SecretVolumeSource{
										SecretName: "testserver-cert",
									},
								},
							},
						},
					},
				},
			},
		}
		require.NoError(t, f.Client.Create(context.TODO(), deployment))

		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "testserver",
				Namespace: namespace,
				Labels: map[string]string{
					"app.kubernetes.io/name": "testserver",
				},
			},
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{
					{
						Name:       "auth",
						Protocol:   corev1.ProtocolTCP,
						Port:       9443,
						TargetPort: intstr.FromInt(9443),
					},
				},
				Selector: map[string]string{
					"app.kubernetes.io/name": "testserver",
				},
				Type: corev1.ServiceTypeClusterIP,
			},
		}
		require.NoError(t, f.Client.Create(context.TODO(), svc))

		extSvc := &sesamev1alpha1.ExtensionService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "testserver",
				Namespace: namespace,
			},
			Spec: sesamev1alpha1.ExtensionServiceSpec{
				Services: []sesamev1alpha1.ExtensionServiceTarget{
					{
						Name: "testserver",
						Port: 9443,
					},
				},
			},
		}
		require.NoError(t, f.Client.Create(context.TODO(), extSvc))

		p := &sesamev1.HTTPProxy{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      "external-auth",
			},
			Spec: sesamev1.HTTPProxySpec{
				VirtualHost: &sesamev1.VirtualHost{
					Fqdn: "externalauth.projectsesame.io",
					TLS: &sesamev1.TLS{
						SecretName: "echo",
					},
					Authorization: &sesamev1.AuthorizationServer{
						ResponseTimeout: "500ms",
						ExtensionServiceRef: sesamev1.ExtensionServiceReference{
							Name:      extSvc.Name,
							Namespace: extSvc.Namespace,
						},
						AuthPolicy: &sesamev1.AuthorizationPolicy{
							Context: map[string]string{
								"hostname": "externalauth.projectsesame.io",
							},
						},
					},
				},
				Routes: []sesamev1.Route{
					{
						Conditions: []sesamev1.MatchCondition{
							{
								Prefix: "/first",
							},
						},
						AuthPolicy: &sesamev1.AuthorizationPolicy{
							Context: map[string]string{
								"target": "first",
							},
						},
						Services: []sesamev1.Service{
							{
								Name: "echo",
								Port: 80,
							},
						},
					},

					{
						Conditions: []sesamev1.MatchCondition{
							{
								Prefix: "/second",
							},
						},
						AuthPolicy: &sesamev1.AuthorizationPolicy{
							Disabled: true,
						},
						Services: []sesamev1.Service{
							{
								Name: "echo",
								Port: 80,
							},
						},
					},

					{
						AuthPolicy: &sesamev1.AuthorizationPolicy{
							Context: map[string]string{
								"target": "default",
							},
						},
						Services: []sesamev1.Service{
							{
								Name: "echo",
								Port: 80,
							},
						},
					},
				},
			},
		}
		f.CreateHTTPProxyAndWaitFor(p, e2e.HTTPProxyValid)

		// By default requests to /first should not be authorized.
		res, ok := f.HTTP.SecureRequestUntil(&e2e.HTTPSRequestOpts{
			Host:      p.Spec.VirtualHost.Fqdn,
			Path:      "/first",
			Condition: e2e.HasStatusCode(401),
		})
		require.NotNil(t, res, "request never succeeded")
		require.Truef(t, ok, "expected 401 response code, got %d", res.StatusCode)

		// The `testserver` authorization server will accept any request with
		// "allow" in the path, so this request should succeed. We can tell that
		// the authorization server processed it by inspecting the context headers
		// that it injects.
		res, ok = f.HTTP.SecureRequestUntil(&e2e.HTTPSRequestOpts{
			Host:      p.Spec.VirtualHost.Fqdn,
			Path:      "/first/allow",
			Condition: e2e.HasStatusCode(200),
		})
		require.NotNil(t, res, "request never succeeded")
		require.Truef(t, ok, "expected 200 response code, got %d", res.StatusCode)

		body := f.GetEchoResponseBody(res.Body)
		assert.Equal(t, "first", body.RequestHeaders.Get("Auth-Context-Target"))
		assert.Equal(t, "externalauth.projectsesame.io", body.RequestHeaders.Get("Auth-Context-Hostname"))

		// THe /second route disables authorization so this request should succeed.
		res, ok = f.HTTP.SecureRequestUntil(&e2e.HTTPSRequestOpts{
			Host:      p.Spec.VirtualHost.Fqdn,
			Path:      "/second",
			Condition: e2e.HasStatusCode(200),
		})
		require.NotNil(t, res, "request never succeeded")
		require.Truef(t, ok, "expected 200 response code, got %d", res.StatusCode)

		// The default route should not authorize by default.
		res, ok = f.HTTP.SecureRequestUntil(&e2e.HTTPSRequestOpts{
			Host:      p.Spec.VirtualHost.Fqdn,
			Path:      "/matches-default-route",
			Condition: e2e.HasStatusCode(401),
		})
		require.NotNil(t, res, "request never succeeded")
		require.Truef(t, ok, "expected 401 response code, got %d", res.StatusCode)

		// The `testserver` authorization server will accept any request with
		// "allow" in the path, so this request should succeed. We can tell that
		// the authorization server processed it by inspecting the context headers
		// that it injects.
		res, ok = f.HTTP.SecureRequestUntil(&e2e.HTTPSRequestOpts{
			Host:      p.Spec.VirtualHost.Fqdn,
			Path:      "/matches-default-route/allow",
			Condition: e2e.HasStatusCode(200),
		})
		require.NotNil(t, res, "request never succeeded")
		require.Truef(t, ok, "expected 200 response code, got %d", res.StatusCode)

		body = f.GetEchoResponseBody(res.Body)
		assert.Equal(t, "default", body.RequestHeaders.Get("Auth-Context-Target"))
		assert.Equal(t, "externalauth.projectsesame.io", body.RequestHeaders.Get("Auth-Context-Hostname"))
	})
}
