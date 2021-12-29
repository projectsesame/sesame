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
	"fmt"
	"strings"
	"testing"

	certmanagerv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	certmanagermetav1 "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	"github.com/onsi/gomega/gexec"
	sesame_api_v1alpha1 "github.com/projectsesame/sesame/apis/projectsesame/v1alpha1"
	"github.com/projectsesame/sesame/pkg/config"
	"github.com/projectsesame/sesame/test/e2e"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var f = e2e.NewFramework(false)

func TestHTTPProxy(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "HTTPProxy tests")
}

var _ = BeforeSuite(func() {
	require.NoError(f.T(), f.Deployment.EnsureResourcesForLocalSesame())
})

var _ = AfterSuite(func() {
	// Delete resources individually instead of deleting the entire sesame
	// namespace as a performance optimization, because deleting non-empty
	// namespaces can take up to a couple minutes to complete.
	require.NoError(f.T(), f.Deployment.DeleteResourcesForLocalSesame())
	gexec.CleanupBuildArtifacts()
})

// Contains specs that test that kubebuilder API validations
// work as expected, and do not require a Sesame instance to
// be running.
var _ = Describe("HTTPProxy API validation", func() {
	f.NamespacedTest("httpproxy-required-field-validation", testRequiredFieldValidation)

	f.NamespacedTest("httpproxy-invalid-wildcard-fqdn", testWildcardFQDN)

	f.NamespacedTest("invalid-cookie-rewrite-fields", testInvalidCookieRewriteFields)
})

var _ = Describe("HTTPProxy", func() {
	var (
		SesameCmd            *gexec.Session
		SesameConfig         *config.Parameters
		SesameConfiguration  *sesame_api_v1alpha1.SesameConfiguration
		SesameConfigFile     string
		additionalSesameArgs []string
	)

	BeforeEach(func() {
		// Sesame config file contents, can be modified in nested
		// BeforeEach.
		SesameConfig = &config.Parameters{}

		// Sesame configuration crd, can be modified in nested
		// BeforeEach.
		SesameConfiguration = e2e.DefaultSesameConfiguration()

		// Default sesame serve command line arguments can be appended to in
		// nested BeforeEach.
		additionalSesameArgs = []string{}
	})

	// JustBeforeEach is called after each of the nested BeforeEach are
	// called, so it is a final setup step before running a test.
	// A nested BeforeEach may have modified Sesame config, so we wait
	// until here to start Sesame.
	JustBeforeEach(func() {
		var err error
		SesameCmd, SesameConfigFile, err = f.Deployment.StartLocalSesame(SesameConfig, SesameConfiguration, additionalSesameArgs...)
		require.NoError(f.T(), err)

		// Wait for Envoy to be healthy.
		require.NoError(f.T(), f.Deployment.WaitForEnvoyDaemonSetUpdated())
	})

	AfterEach(func() {
		require.NoError(f.T(), f.Deployment.StopLocalSesame(SesameCmd, SesameConfigFile))
	})

	f.NamespacedTest("httpproxy-request-redirect-policy", testRequestRedirectRule)
	f.NamespacedTest("httpproxy-request-redirect-policy-nosvc", testRequestRedirectRuleNoService)

	f.NamespacedTest("httpproxy-header-condition-match", testHeaderConditionMatch)

	f.NamespacedTest("httpproxy-path-condition-match", testPathConditionMatch)

	f.NamespacedTest("httpproxy-https-sni-enforcement", testHTTPSSNIEnforcement)

	f.NamespacedTest("httpproxy-pod-restart", testPodRestart)

	f.NamespacedTest("httpproxy-merge-slash", testMergeSlash)

	f.NamespacedTest("httpproxy-client-cert-auth", testClientCertAuth)

	f.NamespacedTest("httpproxy-tcproute-https-termination", testTCPRouteHTTPSTermination)

	f.NamespacedTest("httpproxy-https-misdirected-request", testHTTPSMisdirectedRequest)

	f.NamespacedTest("httpproxy-include-prefix-condition", testIncludePrefixCondition)

	f.NamespacedTest("httpproxy-retry-policy-validation", testRetryPolicyValidation)

	f.NamespacedTest("httpproxy-wildcard-subdomain-fqdn", testWildcardSubdomainFQDN)

	f.NamespacedTest("httpproxy-https-fallback-certificate", func(namespace string) {
		Context("with fallback certificate", func() {
			BeforeEach(func() {
				SesameConfig.TLS = config.TLSParameters{
					FallbackCertificate: config.NamespacedName{
						Name:      "fallback-cert",
						Namespace: namespace,
					},
				}
				SesameConfiguration.Spec.HTTPProxy.FallbackCertificate = &sesame_api_v1alpha1.NamespacedName{
					Name:      "fallback-cert",
					Namespace: namespace,
				}

				f.Certs.CreateSelfSignedCert(namespace, "fallback-cert", "fallback-cert", "fallback.projectsesame.io")
			})

			testHTTPSFallbackCertificate(namespace)
		})
	})

	f.NamespacedTest("httpproxy-backend-tls", func(namespace string) {
		Context("with backend tls", func() {
			BeforeEach(func() {
				// Top level issuer.
				selfSignedIssuer := &certmanagerv1.Issuer{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: namespace,
						Name:      "selfsigned",
					},
					Spec: certmanagerv1.IssuerSpec{
						IssuerConfig: certmanagerv1.IssuerConfig{
							SelfSigned: &certmanagerv1.SelfSignedIssuer{},
						},
					},
				}
				require.NoError(f.T(), f.Client.Create(context.TODO(), selfSignedIssuer))

				// CA to sign backend certs with.
				caCertificate := &certmanagerv1.Certificate{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: namespace,
						Name:      "ca-cert",
					},
					Spec: certmanagerv1.CertificateSpec{
						IsCA: true,
						Usages: []certmanagerv1.KeyUsage{
							certmanagerv1.UsageSigning,
							certmanagerv1.UsageCertSign,
						},
						CommonName: "ca-cert",
						SecretName: "ca-cert",
						IssuerRef: certmanagermetav1.ObjectReference{
							Name: "selfsigned",
						},
					},
				}
				require.NoError(f.T(), f.Client.Create(context.TODO(), caCertificate))

				// Issuer based on CA to generate new certs with.
				basedOnCAIssuer := &certmanagerv1.Issuer{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: namespace,
						Name:      "ca-issuer",
					},
					Spec: certmanagerv1.IssuerSpec{
						IssuerConfig: certmanagerv1.IssuerConfig{
							CA: &certmanagerv1.CAIssuer{
								SecretName: "ca-cert",
							},
						},
					},
				}
				require.NoError(f.T(), f.Client.Create(context.TODO(), basedOnCAIssuer))

				// Backend client cert, can use for upstream validation as well.
				backendClientCert := &certmanagerv1.Certificate{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: namespace,
						Name:      "backend-client-cert",
					},
					Spec: certmanagerv1.CertificateSpec{
						Usages: []certmanagerv1.KeyUsage{
							certmanagerv1.UsageClientAuth,
						},
						CommonName: "client",
						SecretName: "backend-client-cert",
						IssuerRef: certmanagermetav1.ObjectReference{
							Name: "ca-issuer",
						},
					},
				}
				require.NoError(f.T(), f.Client.Create(context.TODO(), backendClientCert))

				SesameConfig.TLS = config.TLSParameters{
					ClientCertificate: config.NamespacedName{
						Namespace: namespace,
						Name:      "backend-client-cert",
					},
				}

				SesameConfiguration.Spec.Envoy.ClientCertificate = &sesame_api_v1alpha1.NamespacedName{
					Name:      "backend-client-cert",
					Namespace: namespace,
				}
			})

			testBackendTLS(namespace)
		})
	})

	f.NamespacedTest("httpproxy-external-auth", testExternalAuth)

	f.NamespacedTest("httpproxy-http-health-checks", testHTTPHealthChecks)

	f.NamespacedTest("httpproxy-dynamic-headers", testDynamicHeaders)

	f.NamespacedTest("httpproxy-host-header-rewrite", testHostHeaderRewrite)

	f.NamespacedTest("httpproxy-external-name-service-insecure", func(namespace string) {
		Context("with ExternalName Services enabled", func() {
			BeforeEach(func() {
				SesameConfig.EnableExternalNameService = true
				SesameConfiguration.Spec.EnableExternalNameService = true
			})
			testExternalNameServiceInsecure(namespace)
		})
	})

	f.NamespacedTest("httpproxy-external-name-service-tls", func(namespace string) {
		Context("with ExternalName Services enabled", func() {
			BeforeEach(func() {
				SesameConfig.EnableExternalNameService = true
				SesameConfiguration.Spec.EnableExternalNameService = true
			})
			testExternalNameServiceTLS(namespace)
		})
	})

	f.NamespacedTest("httpproxy-external-name-service-localhost", func(namespace string) {
		Context("with ExternalName Services enabled", func() {
			BeforeEach(func() {
				SesameConfig.EnableExternalNameService = true
				SesameConfiguration.Spec.EnableExternalNameService = true
			})
			testExternalNameServiceLocalhostInvalid(namespace)
		})
	})
	f.NamespacedTest("httpproxy-local-rate-limiting-vhost", testLocalRateLimitingVirtualHost)

	f.NamespacedTest("httpproxy-local-rate-limiting-route", testLocalRateLimitingRoute)

	Context("global rate limiting", func() {
		withRateLimitService := func(body e2e.NamespacedTestBody) e2e.NamespacedTestBody {
			return func(namespace string) {
				Context("with rate limit service", func() {
					BeforeEach(func() {
						SesameConfig.RateLimitService = config.RateLimitService{
							ExtensionService: fmt.Sprintf("%s/%s", namespace, f.Deployment.RateLimitExtensionService.Name),
							Domain:           "sesame",
							FailOpen:         false,
						}
						SesameConfiguration.Spec.RateLimitService = &sesame_api_v1alpha1.RateLimitServiceConfig{
							ExtensionService: sesame_api_v1alpha1.NamespacedName{
								Name:      f.Deployment.RateLimitExtensionService.Name,
								Namespace: namespace,
							},
							Domain:                  "sesame",
							FailOpen:                false,
							EnableXRateLimitHeaders: false,
						}
						require.NoError(f.T(),
							f.Deployment.EnsureRateLimitResources(
								namespace,
								`
domain: sesame
descriptors:
  - key: generic_key
    value: vhostlimit
    rate_limit:
      unit: hour
      requests_per_unit: 1
  - key: route_limit_key
    value: routelimit
    rate_limit:
      unit: hour
      requests_per_unit: 1
  - key: generic_key
    value: tlsvhostlimit
    rate_limit:
      unit: hour
      requests_per_unit: 1
  - key: generic_key
    value: tlsroutelimit
    rate_limit:
      unit: hour
      requests_per_unit: 1`))
					})

					body(namespace)
				})
			}
		}

		f.NamespacedTest("httpproxy-global-rate-limiting-vhost-non-tls", withRateLimitService(testGlobalRateLimitingVirtualHostNonTLS))

		f.NamespacedTest("httpproxy-global-rate-limiting-route-non-tls", withRateLimitService(testGlobalRateLimitingRouteNonTLS))

		f.NamespacedTest("httpproxy-global-rate-limiting-vhost-tls", withRateLimitService(testGlobalRateLimitingVirtualHostTLS))

		f.NamespacedTest("httpproxy-global-rate-limiting-route-tls", withRateLimitService(testGlobalRateLimitingRouteTLS))
	})

	Context("cookie-rewriting", func() {
		f.NamespacedTest("app-cookie-rewrite", testAppCookieRewrite)

		f.NamespacedTest("cookie-rewrite-tls", testCookieRewriteTLS)

		Context("rewriting cookies from globally rewritten headers", func() {
			BeforeEach(func() {
				SesameConfig.Policy = config.PolicyParameters{
					ResponseHeadersPolicy: config.HeadersPolicy{
						Set: map[string]string{
							"Set-Cookie": "global=foo",
						},
					},
				}
				SesameConfiguration.Spec.Policy = &sesame_api_v1alpha1.PolicyConfig{
					ResponseHeadersPolicy: &sesame_api_v1alpha1.HeadersPolicy{
						Set: map[string]string{
							"Set-Cookie": "global=foo",
						},
					},
				}
			})

			f.NamespacedTest("global-rewrite-headers-cookie-rewrite", testHeaderGlobalRewriteCookieRewrite)
		})

		f.NamespacedTest("rewrite-headers-cookie-rewrite", testHeaderRewriteCookieRewrite)
	})

	Context("using root namespaces", func() {
		Context("configured via config CRD", func() {
			rootNamespaces := []string{
				"root-ns-crd-1",
				"root-ns-crd-2",
			}

			BeforeEach(func() {
				if !e2e.UsingSesameConfigCRD() {
					// Test only applies to sesame config CRD.
					Skip("")
				}
				for _, ns := range rootNamespaces {
					f.CreateNamespace(ns)
				}
				SesameConfiguration.Spec.HTTPProxy.RootNamespaces = rootNamespaces
			})

			AfterEach(func() {
				for _, ns := range rootNamespaces {
					f.DeleteNamespace(ns, false)
				}
			})

			f.NamespacedTest("root-ns-crd", testRootNamespaces(rootNamespaces))
		})

		Context("configured via CLI flag", func() {
			rootNamespaces := []string{
				"root-ns-cli-1",
				"root-ns-cli-2",
			}

			BeforeEach(func() {
				if e2e.UsingSesameConfigCRD() {
					// Test only applies to sesame configmap.
					Skip("")
				}
				for _, ns := range rootNamespaces {
					f.CreateNamespace(ns)
				}
				additionalSesameArgs = []string{
					"--root-namespaces=" + strings.Join(rootNamespaces, ","),
				}
			})

			AfterEach(func() {
				for _, ns := range rootNamespaces {
					f.DeleteNamespace(ns, false)
				}
			})

			f.NamespacedTest("root-ns-cli", testRootNamespaces(rootNamespaces))
		})
	})
})
