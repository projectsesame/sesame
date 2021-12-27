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

package v3

import (
	"testing"

	envoy_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	ratelimit_config_v3 "github.com/envoyproxy/go-control-plane/envoy/config/ratelimit/v3"
	envoy_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	ratelimit_filter_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ratelimit/v3"
	http "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	envoy_discovery_v3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	Sesame_api_v1 "github.com/projectsesame/sesame/apis/projectsesame/v1"
	"github.com/projectsesame/sesame/internal/dag"
	envoy_v3 "github.com/projectsesame/sesame/internal/envoy/v3"
	"github.com/projectsesame/sesame/internal/featuretests"
	"github.com/projectsesame/sesame/internal/fixture"
	"github.com/projectsesame/sesame/internal/k8s"
	"github.com/projectsesame/sesame/internal/protobuf"
	xdscache_v3 "github.com/projectsesame/sesame/internal/xdscache/v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
)

func globalRateLimitFilterExists(t *testing.T, rh cache.ResourceEventHandler, c *Sesame) {
	p := &Sesame_api_v1.HTTPProxy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "proxy1",
		},
		Spec: Sesame_api_v1.HTTPProxySpec{
			VirtualHost: &Sesame_api_v1.VirtualHost{
				Fqdn: "foo.com",
			},
			Routes: []Sesame_api_v1.Route{
				{
					Services: []Sesame_api_v1.Service{
						{
							Name: "s1",
							Port: 80,
						},
					},
				},
			},
		},
	}
	rh.OnAdd(p)

	httpListener := defaultHTTPListener()

	// replace the default filter chains with an HCM that includes the global
	// rate limit filter.
	hcm := envoy_v3.HTTPConnectionManagerBuilder().
		RouteConfigName("ingress_http").
		MetricsPrefix("ingress_http").
		AccessLoggers(envoy_v3.FileAccessLogEnvoy("/dev/stdout", "", nil)).
		DefaultFilters().
		AddFilter(&http.HttpFilter{
			Name: wellknown.HTTPRateLimit,
			ConfigType: &http.HttpFilter_TypedConfig{
				TypedConfig: protobuf.MustMarshalAny(&ratelimit_filter_v3.RateLimit{
					Domain:          "sesame",
					FailureModeDeny: true,
					RateLimitService: &ratelimit_config_v3.RateLimitServiceConfig{
						GrpcService: &envoy_core_v3.GrpcService{
							TargetSpecifier: &envoy_core_v3.GrpcService_EnvoyGrpc_{
								EnvoyGrpc: &envoy_core_v3.GrpcService_EnvoyGrpc{
									ClusterName: dag.ExtensionClusterName(k8s.NamespacedNameFrom("projectsesame/ratelimit")),
								},
							},
						},
						TransportApiVersion: envoy_core_v3.ApiVersion_V3,
					},
				}),
			},
		}).
		Get()

	httpListener.FilterChains = envoy_v3.FilterChains(hcm)

	c.Request(listenerType).Equals(&envoy_discovery_v3.DiscoveryResponse{
		TypeUrl: listenerType,
		Resources: resources(t,
			httpListener,
			statsListener()),
	}).Status(p).IsValid()
}

func globalRateLimitNoRateLimitsDefined(t *testing.T, rh cache.ResourceEventHandler, c *Sesame, tls tlsConfig) {
	p := &Sesame_api_v1.HTTPProxy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "proxy1",
		},
		Spec: Sesame_api_v1.HTTPProxySpec{
			VirtualHost: &Sesame_api_v1.VirtualHost{
				Fqdn: "foo.com",
			},
			Routes: []Sesame_api_v1.Route{
				{
					Services: []Sesame_api_v1.Service{
						{
							Name: "s1",
							Port: 80,
						},
					},
				},
			},
		},
	}

	if tls.enabled {
		p.Spec.VirtualHost.TLS = &Sesame_api_v1.TLS{
			SecretName:                "tls-cert",
			EnableFallbackCertificate: tls.fallbackEnabled,
		}
	}

	rh.OnAdd(p)
	c.Status(p).IsValid()

	switch tls.enabled {
	case true:
		c.Request(routeType, "https/foo.com").Equals(&envoy_discovery_v3.DiscoveryResponse{
			TypeUrl: routeType,
			Resources: resources(t,
				envoy_v3.RouteConfiguration(
					"https/foo.com",
					envoy_v3.VirtualHost("foo.com",
						&envoy_route_v3.Route{
							Match:  routePrefix("/"),
							Action: routeCluster("default/s1/80/da39a3ee5e"),
						},
					),
				),
			),
		})
		if tls.fallbackEnabled {
			c.Request(routeType, "ingress_fallbackcert").Equals(&envoy_discovery_v3.DiscoveryResponse{
				TypeUrl: routeType,
				Resources: resources(t,
					envoy_v3.RouteConfiguration(
						"ingress_fallbackcert",
						envoy_v3.VirtualHost("foo.com",
							&envoy_route_v3.Route{
								Match:  routePrefix("/"),
								Action: routeCluster("default/s1/80/da39a3ee5e"),
							},
						),
					),
				),
			})
		}
	default:
		c.Request(routeType).Equals(&envoy_discovery_v3.DiscoveryResponse{
			TypeUrl: routeType,
			Resources: resources(t,
				envoy_v3.RouteConfiguration(
					"ingress_http",
					envoy_v3.VirtualHost("foo.com",
						&envoy_route_v3.Route{
							Match:  routePrefix("/"),
							Action: routeCluster("default/s1/80/da39a3ee5e"),
						},
					),
				),
			),
		})
	}

}

func globalRateLimitVhostRateLimitDefined(t *testing.T, rh cache.ResourceEventHandler, c *Sesame, tls tlsConfig) {
	p := &Sesame_api_v1.HTTPProxy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "proxy1",
		},
		Spec: Sesame_api_v1.HTTPProxySpec{
			VirtualHost: &Sesame_api_v1.VirtualHost{
				Fqdn: "foo.com",
				RateLimitPolicy: &Sesame_api_v1.RateLimitPolicy{
					Global: &Sesame_api_v1.GlobalRateLimitPolicy{
						Descriptors: []Sesame_api_v1.RateLimitDescriptor{
							{
								Entries: []Sesame_api_v1.RateLimitDescriptorEntry{
									{
										RemoteAddress: &Sesame_api_v1.RemoteAddressDescriptor{},
									},
									{
										GenericKey: &Sesame_api_v1.GenericKeyDescriptor{Value: "generic-key-value"},
									},
								},
							},
						},
					},
				},
			},
			Routes: []Sesame_api_v1.Route{
				{
					Services: []Sesame_api_v1.Service{
						{
							Name: "s1",
							Port: 80,
						},
					},
				},
			},
		},
	}

	if tls.enabled {
		p.Spec.VirtualHost.TLS = &Sesame_api_v1.TLS{
			SecretName:                "tls-cert",
			EnableFallbackCertificate: tls.fallbackEnabled,
		}
	}

	rh.OnAdd(p)
	c.Status(p).IsValid()

	route := &envoy_route_v3.Route{
		Match:  routePrefix("/"),
		Action: routeCluster("default/s1/80/da39a3ee5e"),
	}

	vhost := envoy_v3.VirtualHost("foo.com", route)
	vhost.RateLimits = []*envoy_route_v3.RateLimit{
		{
			Actions: []*envoy_route_v3.RateLimit_Action{
				{
					ActionSpecifier: &envoy_route_v3.RateLimit_Action_RemoteAddress_{
						RemoteAddress: &envoy_route_v3.RateLimit_Action_RemoteAddress{},
					},
				},
				{
					ActionSpecifier: &envoy_route_v3.RateLimit_Action_GenericKey_{
						GenericKey: &envoy_route_v3.RateLimit_Action_GenericKey{DescriptorValue: "generic-key-value"},
					},
				},
			},
		},
	}

	switch tls.enabled {
	case true:
		c.Request(routeType, "https/foo.com").Equals(&envoy_discovery_v3.DiscoveryResponse{
			TypeUrl:   routeType,
			Resources: resources(t, envoy_v3.RouteConfiguration("https/foo.com", vhost)),
		})
		if tls.fallbackEnabled {
			c.Request(routeType, "ingress_fallbackcert").Equals(&envoy_discovery_v3.DiscoveryResponse{
				TypeUrl:   routeType,
				Resources: resources(t, envoy_v3.RouteConfiguration("ingress_fallbackcert", vhost)),
			})
		}
	default:
		c.Request(routeType).Equals(&envoy_discovery_v3.DiscoveryResponse{
			TypeUrl:   routeType,
			Resources: resources(t, envoy_v3.RouteConfiguration("ingress_http", vhost)),
		})
	}
}

func globalRateLimitRouteRateLimitDefined(t *testing.T, rh cache.ResourceEventHandler, c *Sesame, tls tlsConfig) {
	p := &Sesame_api_v1.HTTPProxy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "proxy1",
		},
		Spec: Sesame_api_v1.HTTPProxySpec{
			VirtualHost: &Sesame_api_v1.VirtualHost{
				Fqdn: "foo.com",
			},
			Routes: []Sesame_api_v1.Route{
				{
					Services: []Sesame_api_v1.Service{
						{
							Name: "s1",
							Port: 80,
						},
					},
					RateLimitPolicy: &Sesame_api_v1.RateLimitPolicy{
						Global: &Sesame_api_v1.GlobalRateLimitPolicy{
							Descriptors: []Sesame_api_v1.RateLimitDescriptor{
								{
									Entries: []Sesame_api_v1.RateLimitDescriptorEntry{
										{
											RemoteAddress: &Sesame_api_v1.RemoteAddressDescriptor{},
										},
										{
											GenericKey: &Sesame_api_v1.GenericKeyDescriptor{Value: "generic-key-value"},
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

	if tls.enabled {
		p.Spec.VirtualHost.TLS = &Sesame_api_v1.TLS{
			SecretName:                "tls-cert",
			EnableFallbackCertificate: tls.fallbackEnabled,
		}
	}

	rh.OnAdd(p)
	c.Status(p).IsValid()

	route := &envoy_route_v3.Route{
		Match: routePrefix("/"),
		Action: routeCluster("default/s1/80/da39a3ee5e", func(r *envoy_route_v3.Route_Route) {
			r.Route.RateLimits = []*envoy_route_v3.RateLimit{
				{
					Actions: []*envoy_route_v3.RateLimit_Action{
						{
							ActionSpecifier: &envoy_route_v3.RateLimit_Action_RemoteAddress_{
								RemoteAddress: &envoy_route_v3.RateLimit_Action_RemoteAddress{},
							},
						},
						{
							ActionSpecifier: &envoy_route_v3.RateLimit_Action_GenericKey_{
								GenericKey: &envoy_route_v3.RateLimit_Action_GenericKey{DescriptorValue: "generic-key-value"},
							},
						},
					},
				},
			}
		}),
	}

	vhost := envoy_v3.VirtualHost("foo.com", route)

	switch tls.enabled {
	case true:
		c.Request(routeType, "https/foo.com").Equals(&envoy_discovery_v3.DiscoveryResponse{
			TypeUrl:   routeType,
			Resources: resources(t, envoy_v3.RouteConfiguration("https/foo.com", vhost)),
		})
		if tls.fallbackEnabled {
			c.Request(routeType, "ingress_fallbackcert").Equals(&envoy_discovery_v3.DiscoveryResponse{
				TypeUrl:   routeType,
				Resources: resources(t, envoy_v3.RouteConfiguration("ingress_fallbackcert", vhost)),
			})
		}
	default:
		c.Request(routeType).Equals(&envoy_discovery_v3.DiscoveryResponse{
			TypeUrl:   routeType,
			Resources: resources(t, envoy_v3.RouteConfiguration("ingress_http", vhost)),
		})
	}
}

func globalRateLimitVhostAndRouteRateLimitDefined(t *testing.T, rh cache.ResourceEventHandler, c *Sesame, tls tlsConfig) {
	p := &Sesame_api_v1.HTTPProxy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "proxy1",
		},
		Spec: Sesame_api_v1.HTTPProxySpec{
			VirtualHost: &Sesame_api_v1.VirtualHost{
				Fqdn: "foo.com",
				RateLimitPolicy: &Sesame_api_v1.RateLimitPolicy{
					Global: &Sesame_api_v1.GlobalRateLimitPolicy{
						Descriptors: []Sesame_api_v1.RateLimitDescriptor{
							{
								Entries: []Sesame_api_v1.RateLimitDescriptorEntry{
									{
										RemoteAddress: &Sesame_api_v1.RemoteAddressDescriptor{},
									},
									{
										GenericKey: &Sesame_api_v1.GenericKeyDescriptor{Value: "generic-key-value-vhost"},
									},
								},
							},
						},
					},
				},
			},
			Routes: []Sesame_api_v1.Route{
				{
					Services: []Sesame_api_v1.Service{
						{
							Name: "s1",
							Port: 80,
						},
					},
					RateLimitPolicy: &Sesame_api_v1.RateLimitPolicy{
						Global: &Sesame_api_v1.GlobalRateLimitPolicy{
							Descriptors: []Sesame_api_v1.RateLimitDescriptor{
								{
									Entries: []Sesame_api_v1.RateLimitDescriptorEntry{
										{
											RemoteAddress: &Sesame_api_v1.RemoteAddressDescriptor{},
										},
										{
											GenericKey: &Sesame_api_v1.GenericKeyDescriptor{Value: "generic-key-value"},
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

	if tls.enabled {
		p.Spec.VirtualHost.TLS = &Sesame_api_v1.TLS{
			SecretName:                "tls-cert",
			EnableFallbackCertificate: tls.fallbackEnabled,
		}
	}

	rh.OnAdd(p)
	c.Status(p).IsValid()

	route := &envoy_route_v3.Route{
		Match: routePrefix("/"),
		Action: routeCluster("default/s1/80/da39a3ee5e", func(r *envoy_route_v3.Route_Route) {
			r.Route.RateLimits = []*envoy_route_v3.RateLimit{
				{
					Actions: []*envoy_route_v3.RateLimit_Action{
						{
							ActionSpecifier: &envoy_route_v3.RateLimit_Action_RemoteAddress_{
								RemoteAddress: &envoy_route_v3.RateLimit_Action_RemoteAddress{},
							},
						},
						{
							ActionSpecifier: &envoy_route_v3.RateLimit_Action_GenericKey_{
								GenericKey: &envoy_route_v3.RateLimit_Action_GenericKey{DescriptorValue: "generic-key-value"},
							},
						},
					},
				},
			}
		}),
	}

	vhost := envoy_v3.VirtualHost("foo.com", route)
	vhost.RateLimits = []*envoy_route_v3.RateLimit{
		{
			Actions: []*envoy_route_v3.RateLimit_Action{
				{
					ActionSpecifier: &envoy_route_v3.RateLimit_Action_RemoteAddress_{
						RemoteAddress: &envoy_route_v3.RateLimit_Action_RemoteAddress{},
					},
				},
				{
					ActionSpecifier: &envoy_route_v3.RateLimit_Action_GenericKey_{
						GenericKey: &envoy_route_v3.RateLimit_Action_GenericKey{DescriptorValue: "generic-key-value-vhost"},
					},
				},
			},
		},
	}

	switch tls.enabled {
	case true:
		c.Request(routeType, "https/foo.com").Equals(&envoy_discovery_v3.DiscoveryResponse{
			TypeUrl:   routeType,
			Resources: resources(t, envoy_v3.RouteConfiguration("https/foo.com", vhost)),
		})
		if tls.fallbackEnabled {
			c.Request(routeType, "ingress_fallbackcert").Equals(&envoy_discovery_v3.DiscoveryResponse{
				TypeUrl:   routeType,
				Resources: resources(t, envoy_v3.RouteConfiguration("ingress_fallbackcert", vhost)),
			})
		}
	default:
		c.Request(routeType).Equals(&envoy_discovery_v3.DiscoveryResponse{
			TypeUrl:   routeType,
			Resources: resources(t, envoy_v3.RouteConfiguration("ingress_http", vhost)),
		})
	}
}

func globalRateLimitMultipleDescriptorsAndEntries(t *testing.T, rh cache.ResourceEventHandler, c *Sesame) {
	p := &Sesame_api_v1.HTTPProxy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "proxy1",
		},
		Spec: Sesame_api_v1.HTTPProxySpec{
			VirtualHost: &Sesame_api_v1.VirtualHost{
				Fqdn: "foo.com",
			},
			Routes: []Sesame_api_v1.Route{
				{
					Services: []Sesame_api_v1.Service{
						{
							Name: "s1",
							Port: 80,
						},
					},
					RateLimitPolicy: &Sesame_api_v1.RateLimitPolicy{
						Global: &Sesame_api_v1.GlobalRateLimitPolicy{
							Descriptors: []Sesame_api_v1.RateLimitDescriptor{
								// first descriptor
								{
									Entries: []Sesame_api_v1.RateLimitDescriptorEntry{
										{
											RemoteAddress: &Sesame_api_v1.RemoteAddressDescriptor{},
										},
										{
											GenericKey: &Sesame_api_v1.GenericKeyDescriptor{Value: "generic-key-value"},
										},
									},
								},
								// second descriptor
								{
									Entries: []Sesame_api_v1.RateLimitDescriptorEntry{
										{
											RequestHeader: &Sesame_api_v1.RequestHeaderDescriptor{HeaderName: "X-Sesame", DescriptorKey: "header-descriptor"},
										},
										{
											GenericKey: &Sesame_api_v1.GenericKeyDescriptor{Key: "generic-key-key", Value: "generic-key-value-2"},
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

	rh.OnAdd(p)
	c.Status(p).IsValid()

	route := &envoy_route_v3.Route{
		Match: routePrefix("/"),
		Action: routeCluster("default/s1/80/da39a3ee5e", func(r *envoy_route_v3.Route_Route) {
			r.Route.RateLimits = []*envoy_route_v3.RateLimit{
				{
					Actions: []*envoy_route_v3.RateLimit_Action{
						{
							ActionSpecifier: &envoy_route_v3.RateLimit_Action_RemoteAddress_{
								RemoteAddress: &envoy_route_v3.RateLimit_Action_RemoteAddress{},
							},
						},
						{
							ActionSpecifier: &envoy_route_v3.RateLimit_Action_GenericKey_{
								GenericKey: &envoy_route_v3.RateLimit_Action_GenericKey{DescriptorValue: "generic-key-value"},
							},
						},
					},
				},
				{
					Actions: []*envoy_route_v3.RateLimit_Action{
						{
							ActionSpecifier: &envoy_route_v3.RateLimit_Action_RequestHeaders_{
								RequestHeaders: &envoy_route_v3.RateLimit_Action_RequestHeaders{
									HeaderName:    "X-Sesame",
									DescriptorKey: "header-descriptor",
								},
							},
						},
						{
							ActionSpecifier: &envoy_route_v3.RateLimit_Action_GenericKey_{
								GenericKey: &envoy_route_v3.RateLimit_Action_GenericKey{
									DescriptorKey:   "generic-key-key",
									DescriptorValue: "generic-key-value-2"},
							},
						},
					},
				},
			}
		}),
	}

	c.Request(routeType).Equals(&envoy_discovery_v3.DiscoveryResponse{
		TypeUrl:   routeType,
		Resources: resources(t, envoy_v3.RouteConfiguration("ingress_http", envoy_v3.VirtualHost("foo.com", route))),
	})

}

type tlsConfig struct {
	enabled         bool
	fallbackEnabled bool
}

func TestGlobalRateLimiting(t *testing.T) {
	var (
		tlsDisabled     = tlsConfig{}
		tlsEnabled      = tlsConfig{enabled: true}
		fallbackEnabled = tlsConfig{enabled: true, fallbackEnabled: true}
	)

	subtests := map[string]func(*testing.T, cache.ResourceEventHandler, *Sesame){
		"GlobalRateLimitFilterExists": globalRateLimitFilterExists,

		// test cases for insecure/non-TLS vhosts
		"NoRateLimitsDefined": func(t *testing.T, rh cache.ResourceEventHandler, c *Sesame) {
			globalRateLimitNoRateLimitsDefined(t, rh, c, tlsDisabled)
		},
		"VirtualHostRateLimitDefined": func(t *testing.T, rh cache.ResourceEventHandler, c *Sesame) {
			globalRateLimitVhostRateLimitDefined(t, rh, c, tlsDisabled)
		},
		"RouteRateLimitDefined": func(t *testing.T, rh cache.ResourceEventHandler, c *Sesame) {
			globalRateLimitRouteRateLimitDefined(t, rh, c, tlsDisabled)
		},
		"VirtualHostAndRouteRateLimitsDefined": func(t *testing.T, rh cache.ResourceEventHandler, c *Sesame) {
			globalRateLimitVhostAndRouteRateLimitDefined(t, rh, c, tlsDisabled)
		},

		// test cases for secure/TLS vhosts
		"TLSNoRateLimitsDefined": func(t *testing.T, rh cache.ResourceEventHandler, c *Sesame) {
			globalRateLimitNoRateLimitsDefined(t, rh, c, tlsEnabled)
		},
		"TLSVirtualHostRateLimitDefined": func(t *testing.T, rh cache.ResourceEventHandler, c *Sesame) {
			globalRateLimitVhostRateLimitDefined(t, rh, c, tlsEnabled)
		},
		"TLSRouteRateLimitDefined": func(t *testing.T, rh cache.ResourceEventHandler, c *Sesame) {
			globalRateLimitRouteRateLimitDefined(t, rh, c, tlsEnabled)
		},
		"TLSVirtualHostAndRouteRateLimitsDefined": func(t *testing.T, rh cache.ResourceEventHandler, c *Sesame) {
			globalRateLimitVhostAndRouteRateLimitDefined(t, rh, c, tlsEnabled)
		},

		// test cases for secure/TLS vhosts with fallback cert enabled
		"FallbackNoRateLimitsDefined": func(t *testing.T, rh cache.ResourceEventHandler, c *Sesame) {
			globalRateLimitNoRateLimitsDefined(t, rh, c, fallbackEnabled)
		},
		"FallbackVirtualHostRateLimitDefined": func(t *testing.T, rh cache.ResourceEventHandler, c *Sesame) {
			globalRateLimitVhostRateLimitDefined(t, rh, c, fallbackEnabled)
		},
		"FallbackRouteRateLimitDefined": func(t *testing.T, rh cache.ResourceEventHandler, c *Sesame) {
			globalRateLimitRouteRateLimitDefined(t, rh, c, fallbackEnabled)
		},
		"FallbackVirtualHostAndRouteRateLimitsDefined": func(t *testing.T, rh cache.ResourceEventHandler, c *Sesame) {
			globalRateLimitVhostAndRouteRateLimitDefined(t, rh, c, fallbackEnabled)
		},

		"MultipleDescriptorsAndEntriesDefined": globalRateLimitMultipleDescriptorsAndEntries,
	}

	for n, f := range subtests {
		f := f
		t.Run(n, func(t *testing.T) {
			rh, c, done := setup(t,
				func(cfg *xdscache_v3.ListenerConfig) {
					cfg.RateLimitConfig = &xdscache_v3.RateLimitConfig{
						ExtensionService: k8s.NamespacedNameFrom("projectsesame/ratelimit"),
						Domain:           "sesame",
					}
				},
				func(b *dag.Builder) {
					b.Processors = []dag.Processor{
						&dag.HTTPProxyProcessor{
							FallbackCertificate: &types.NamespacedName{
								Name:      "fallback-cert",
								Namespace: "default",
							},
						},
						&dag.ListenerProcessor{},
					}
				},
			)

			defer done()

			// Add common test fixtures.
			rh.OnAdd(fixture.NewService("s1").WithPorts(corev1.ServicePort{Port: 80}))
			rh.OnAdd(fixture.NewService("s2").WithPorts(corev1.ServicePort{Port: 80}))
			rh.OnAdd(&corev1.Secret{
				ObjectMeta: fixture.ObjectMeta("tls-cert"),
				Type:       "kubernetes.io/tls",
				Data:       featuretests.Secretdata(featuretests.CERTIFICATE, featuretests.RSA_PRIVATE_KEY),
			})
			rh.OnAdd(&corev1.Secret{
				ObjectMeta: fixture.ObjectMeta("fallback-cert"),
				Type:       "kubernetes.io/tls",
				Data:       featuretests.Secretdata(featuretests.CERTIFICATE, featuretests.RSA_PRIVATE_KEY),
			})

			f(t, rh, c)
		})
	}
}
