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

package ingress

import (
	"context"
	"testing"

	sesame_api_v1alpha1 "github.com/projectsesame/sesame/apis/projectsesame/v1alpha1"

	certmanagerv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	certmanagermetav1 "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	"github.com/onsi/gomega/gexec"
	"github.com/projectsesame/sesame/pkg/config"
	"github.com/projectsesame/sesame/test/e2e"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var f = e2e.NewFramework(false)

func TestIngress(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ingress tests")
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

var _ = Describe("Ingress", func() {
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

	f.NamespacedTest("ingress-tls-wildcard-host", testTLSWildcardHost)

	f.NamespacedTest("backend-tls", func(namespace string) {
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
					Namespace: namespace,
					Name:      "backend-client-cert",
				}
			})

			testBackendTLS(namespace)
		})
	})

	f.NamespacedTest("long-path-match", testLongPathMatch)
})
