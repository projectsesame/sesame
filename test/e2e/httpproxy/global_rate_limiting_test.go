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
	"github.com/projectsesame/sesame/test/e2e"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func testGlobalRateLimitingVirtualHostNonTLS(namespace string) {
	Specify("global rate limit policy set on non-TLS virtualhost is applied", func() {
		t := f.T()

		f.Fixtures.Echo.Deploy(namespace, "echo")

		p := &sesamev1.HTTPProxy{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      "globalratelimitvhostnontls",
			},
			Spec: sesamev1.HTTPProxySpec{
				VirtualHost: &sesamev1.VirtualHost{
					Fqdn: "globalratelimitvhostnontls.projectsesame.io",
				},
				Routes: []sesamev1.Route{
					{
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
		p, _ = f.CreateHTTPProxyAndWaitFor(p, e2e.HTTPProxyValid)

		// Wait until we get a 200 from the proxy confirming
		// the pods are up and serving traffic.
		res, ok := f.HTTP.RequestUntil(&e2e.HTTPRequestOpts{
			Host:      p.Spec.VirtualHost.Fqdn,
			Condition: e2e.HasStatusCode(200),
		})
		require.NotNil(t, res, "request never succeeded")
		require.Truef(t, ok, "expected 200 response code, got %d", res.StatusCode)

		require.NoError(t, retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			if err := f.Client.Get(context.TODO(), client.ObjectKeyFromObject(p), p); err != nil {
				return err
			}

			// Add a global rate limit policy on the virtual host.
			p.Spec.VirtualHost.RateLimitPolicy = &sesamev1.RateLimitPolicy{
				Global: &sesamev1.GlobalRateLimitPolicy{
					Descriptors: []sesamev1.RateLimitDescriptor{
						{
							Entries: []sesamev1.RateLimitDescriptorEntry{
								{
									GenericKey: &sesamev1.GenericKeyDescriptor{
										Value: "vhostlimit",
									},
								},
							},
						},
					},
				},
			}

			return f.Client.Update(context.TODO(), p)
		}))

		// Make a request against the proxy, confirm a 200 response
		// is returned since we're allowed one request per hour.
		res, ok = f.HTTP.RequestUntil(&e2e.HTTPRequestOpts{
			Host:      p.Spec.VirtualHost.Fqdn,
			Condition: e2e.HasStatusCode(200),
		})
		require.NotNil(t, res, "request never succeeded")
		require.Truef(t, ok, "expected 200 response code, got %d", res.StatusCode)

		// Make another request against the proxy, confirm a 429 response
		// is now gotten since we've exceeded the rate limit.
		res, ok = f.HTTP.RequestUntil(&e2e.HTTPRequestOpts{
			Host:      p.Spec.VirtualHost.Fqdn,
			Condition: e2e.HasStatusCode(429),
		})
		require.NotNil(t, res, "request never succeeded")
		require.Truef(t, ok, "expected 429 response code, got %d", res.StatusCode)
	})
}

func testGlobalRateLimitingRouteNonTLS(namespace string) {
	Specify("global rate limit policy set on non-TLS route is applied", func() {
		t := f.T()

		f.Fixtures.Echo.Deploy(namespace, "echo")

		p := &sesamev1.HTTPProxy{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      "globalratelimitroutenontls",
			},
			Spec: sesamev1.HTTPProxySpec{
				VirtualHost: &sesamev1.VirtualHost{
					Fqdn: "globalratelimitroutenontls.projectsesame.io",
				},
				Routes: []sesamev1.Route{
					{
						Services: []sesamev1.Service{
							{
								Name: "echo",
								Port: 80,
							},
						},
					},
					{
						Services: []sesamev1.Service{
							{
								Name: "echo",
								Port: 80,
							},
						},
						Conditions: []sesamev1.MatchCondition{
							{
								Prefix: "/unlimited",
							},
						},
					},
				},
			},
		}
		p, _ = f.CreateHTTPProxyAndWaitFor(p, e2e.HTTPProxyValid)

		// Wait until we get a 200 from the proxy confirming
		// the pods are up and serving traffic.
		res, ok := f.HTTP.RequestUntil(&e2e.HTTPRequestOpts{
			Host:      p.Spec.VirtualHost.Fqdn,
			Condition: e2e.HasStatusCode(200),
		})
		require.NotNil(t, res, "request never succeeded")
		require.Truef(t, ok, "expected 200 response code, got %d", res.StatusCode)

		// Add a global rate limit policy on the first route.
		require.NoError(t, retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			if err := f.Client.Get(context.TODO(), client.ObjectKeyFromObject(p), p); err != nil {
				return err
			}

			p.Spec.Routes[0].RateLimitPolicy = &sesamev1.RateLimitPolicy{
				Global: &sesamev1.GlobalRateLimitPolicy{
					Descriptors: []sesamev1.RateLimitDescriptor{
						{
							Entries: []sesamev1.RateLimitDescriptorEntry{
								{
									GenericKey: &sesamev1.GenericKeyDescriptor{
										Key:   "route_limit_key",
										Value: "routelimit",
									},
								},
							},
						},
					},
				},
			}

			return f.Client.Update(context.TODO(), p)
		}))

		// Make a request against the proxy, confirm a 200 response
		// is returned since we're allowed one request per hour.
		res, ok = f.HTTP.RequestUntil(&e2e.HTTPRequestOpts{
			Host:      p.Spec.VirtualHost.Fqdn,
			Condition: e2e.HasStatusCode(200),
		})
		require.NotNil(t, res, "request never succeeded")
		require.Truef(t, ok, "expected 200 response code, got %d", res.StatusCode)

		// Make another request against the proxy, confirm a 429 response
		// is now gotten since we've exceeded the rate limit.
		res, ok = f.HTTP.RequestUntil(&e2e.HTTPRequestOpts{
			Host:      p.Spec.VirtualHost.Fqdn,
			Condition: e2e.HasStatusCode(429),
		})
		require.NotNil(t, res, "request never succeeded")
		require.Truef(t, ok, "expected 429 response code, got %d", res.StatusCode)

		// Make a request against the route that doesn't have rate limiting
		// to confirm we still get a 200 for that route.
		res, ok = f.HTTP.RequestUntil(&e2e.HTTPRequestOpts{
			Host:      p.Spec.VirtualHost.Fqdn,
			Path:      "/unlimited",
			Condition: e2e.HasStatusCode(200),
		})
		require.NotNil(t, res, "request never succeeded")
		require.Truef(t, ok, "expected 200 response code for non-rate-limited route, got %d", res.StatusCode)
	})
}

func testGlobalRateLimitingVirtualHostTLS(namespace string) {
	Specify("global rate limit policy set on TLS virtualhost is applied", func() {
		t := f.T()

		f.Fixtures.Echo.Deploy(namespace, "echo")
		f.Certs.CreateSelfSignedCert(namespace, "echo-cert", "echo", "globalratelimitvhosttls.projectsesame.io")

		p := &sesamev1.HTTPProxy{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      "globalratelimitvhosttls",
			},
			Spec: sesamev1.HTTPProxySpec{
				VirtualHost: &sesamev1.VirtualHost{
					Fqdn: "globalratelimitvhosttls.projectsesame.io",
					TLS: &sesamev1.TLS{
						SecretName: "echo",
					},
				},
				Routes: []sesamev1.Route{
					{
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
		p, _ = f.CreateHTTPProxyAndWaitFor(p, e2e.HTTPProxyValid)

		// Wait until we get a 200 from the proxy confirming
		// the pods are up and serving traffic.
		res, ok := f.HTTP.SecureRequestUntil(&e2e.HTTPSRequestOpts{
			Host:      p.Spec.VirtualHost.Fqdn,
			Condition: e2e.HasStatusCode(200),
		})
		require.NotNil(t, res, "request never succeeded")
		require.Truef(t, ok, "expected 200 response code, got %d", res.StatusCode)

		// Add a global rate limit policy on the virtual host.
		require.NoError(t, retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			if err := f.Client.Get(context.TODO(), client.ObjectKeyFromObject(p), p); err != nil {
				return err
			}

			p.Spec.VirtualHost.RateLimitPolicy = &sesamev1.RateLimitPolicy{
				Global: &sesamev1.GlobalRateLimitPolicy{
					Descriptors: []sesamev1.RateLimitDescriptor{
						{
							Entries: []sesamev1.RateLimitDescriptorEntry{
								{
									GenericKey: &sesamev1.GenericKeyDescriptor{
										Value: "tlsvhostlimit",
									},
								},
							},
						},
					},
				},
			}

			return f.Client.Update(context.TODO(), p)
		}))

		// Make a request against the proxy, confirm a 200 response
		// is returned since we're allowed one request per hour.
		res, ok = f.HTTP.SecureRequestUntil(&e2e.HTTPSRequestOpts{
			Host:      p.Spec.VirtualHost.Fqdn,
			Condition: e2e.HasStatusCode(200),
		})
		require.NotNil(t, res, "request never succeeded")
		require.Truef(t, ok, "expected 200 response code, got %d", res.StatusCode)

		// Make another request against the proxy, confirm a 429 response
		// is now gotten since we've exceeded the rate limit.
		res, ok = f.HTTP.SecureRequestUntil(&e2e.HTTPSRequestOpts{
			Host:      p.Spec.VirtualHost.Fqdn,
			Condition: e2e.HasStatusCode(429),
		})
		require.NotNil(t, res, "request never succeeded")
		require.Truef(t, ok, "expected 429 response code, got %d", res.StatusCode)
	})
}

func testGlobalRateLimitingRouteTLS(namespace string) {
	Specify("global rate limit policy set on TLS route is applied", func() {
		t := f.T()

		f.Fixtures.Echo.Deploy(namespace, "echo")
		f.Certs.CreateSelfSignedCert(namespace, "echo-cert", "echo", "globalratelimitroutetls.projectsesame.io")

		p := &sesamev1.HTTPProxy{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      "globalratelimitroutetls",
			},
			Spec: sesamev1.HTTPProxySpec{
				VirtualHost: &sesamev1.VirtualHost{
					Fqdn: "globalratelimitroutetls.projectsesame.io",
					TLS: &sesamev1.TLS{
						SecretName: "echo",
					},
				},
				Routes: []sesamev1.Route{
					{
						Services: []sesamev1.Service{
							{
								Name: "echo",
								Port: 80,
							},
						},
					},
					{
						Services: []sesamev1.Service{
							{
								Name: "echo",
								Port: 80,
							},
						},
						Conditions: []sesamev1.MatchCondition{
							{
								Prefix: "/unlimited",
							},
						},
					},
				},
			},
		}
		p, _ = f.CreateHTTPProxyAndWaitFor(p, e2e.HTTPProxyValid)

		// Wait until we get a 200 from the proxy confirming
		// the pods are up and serving traffic.
		res, ok := f.HTTP.SecureRequestUntil(&e2e.HTTPSRequestOpts{
			Host:      p.Spec.VirtualHost.Fqdn,
			Condition: e2e.HasStatusCode(200),
		})
		require.NotNil(t, res, "request never succeeded")
		require.Truef(t, ok, "expected 200 response code, got %d", res.StatusCode)

		// Add a global rate limit policy on the first route.
		require.NoError(t, retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			if err := f.Client.Get(context.TODO(), client.ObjectKeyFromObject(p), p); err != nil {
				return err
			}

			p.Spec.Routes[0].RateLimitPolicy = &sesamev1.RateLimitPolicy{
				Global: &sesamev1.GlobalRateLimitPolicy{
					Descriptors: []sesamev1.RateLimitDescriptor{
						{
							Entries: []sesamev1.RateLimitDescriptorEntry{
								{
									GenericKey: &sesamev1.GenericKeyDescriptor{
										Value: "tlsroutelimit",
									},
								},
							},
						},
					},
				},
			}

			return f.Client.Update(context.TODO(), p)
		}))

		// Make a request against the proxy, confirm a 200 response
		// is returned since we're allowed one request per hour.
		res, ok = f.HTTP.SecureRequestUntil(&e2e.HTTPSRequestOpts{
			Host:      p.Spec.VirtualHost.Fqdn,
			Condition: e2e.HasStatusCode(200),
		})
		require.NotNil(t, res, "request never succeeded")
		require.Truef(t, ok, "expected 200 response code, got %d", res.StatusCode)

		// Make another request against the proxy, confirm a 429 response
		// is now gotten since we've exceeded the rate limit.
		res, ok = f.HTTP.SecureRequestUntil(&e2e.HTTPSRequestOpts{
			Host:      p.Spec.VirtualHost.Fqdn,
			Condition: e2e.HasStatusCode(429),
		})
		require.NotNil(t, res, "request never succeeded")
		require.Truef(t, ok, "expected 429 response code, got %d", res.StatusCode)

		// Make a request against the route that doesn't have rate limiting
		// to confirm we still get a 200 for that route.
		res, ok = f.HTTP.SecureRequestUntil(&e2e.HTTPSRequestOpts{
			Host:      p.Spec.VirtualHost.Fqdn,
			Path:      "/unlimited",
			Condition: e2e.HasStatusCode(200),
		})
		require.NotNil(t, res, "request never succeeded")
		require.Truef(t, ok, "expected 200 response code for non-rate-limited route, got %d", res.StatusCode)
	})
}
