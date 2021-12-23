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

	envoy_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_discovery_v3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	contour_api_v1 "github.com/projectcontour/sesame/apis/projectsesame/v1"
	envoy_v3 "github.com/projectcontour/sesame/internal/envoy/v3"
	"github.com/projectsesame/sesame/internal/contour"
	"github.com/projectsesame/sesame/internal/fixture"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestMirrorPolicy(t *testing.T) {
	rh, c, done := setup(t, func(reh *contour.EventHandler) {})
	defer done()

	svc1 := fixture.NewService("kuard").
		WithPorts(v1.ServicePort{Port: 8080, TargetPort: intstr.FromInt(8080)})
	svc2 := fixture.NewService("mirror").
		WithPorts(v1.ServicePort{Port: 8080, TargetPort: intstr.FromInt(8080)})
	rh.OnAdd(svc1)
	rh.OnAdd(svc2)

	p1 := &contour_api_v1.HTTPProxy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "simple",
			Namespace: svc1.Namespace,
		},
		Spec: contour_api_v1.HTTPProxySpec{
			VirtualHost: &contour_api_v1.VirtualHost{Fqdn: "example.com"},
			Routes: []contour_api_v1.Route{{
				Conditions: matchconditions(prefixMatchCondition("/")),
				Services: []contour_api_v1.Service{{
					Name: svc1.Name,
					Port: 8080,
				}, {
					Name:   svc2.Name,
					Port:   8080,
					Mirror: true,
				}},
			}},
		},
	}
	rh.OnAdd(p1)

	c.Request(routeType).Equals(&envoy_discovery_v3.DiscoveryResponse{
		Resources: resources(t,
			envoy_v3.RouteConfiguration("ingress_http",
				envoy_v3.VirtualHost(p1.Spec.VirtualHost.Fqdn,
					&envoy_route_v3.Route{
						Match:  routePrefix("/"),
						Action: withMirrorPolicy(routeCluster("default/kuard/8080/da39a3ee5e"), "default/mirror/8080/da39a3ee5e"),
					},
				),
			),
		),
		TypeUrl: routeType,
	})

	// assert that are two clusters in CDS, one for the route service
	// and one for the mirror service.
	c.Request(clusterType).Equals(&envoy_discovery_v3.DiscoveryResponse{
		Resources: resources(t,
			cluster("default/kuard/8080/da39a3ee5e", "default/kuard", "default_kuard_8080"),
			cluster("default/mirror/8080/da39a3ee5e", "default/mirror", "default_mirror_8080"),
		),
		TypeUrl: clusterType,
	})
}