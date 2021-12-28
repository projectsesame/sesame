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

package infra

import (
	"testing"

	sesame_api_v1alpha1 "github.com/projectsesame/sesame/apis/projectsesame/v1alpha1"
	v1 "k8s.io/api/core/v1"

	"github.com/onsi/gomega/gexec"
	"github.com/projectsesame/sesame/pkg/config"
	"github.com/projectsesame/sesame/test/e2e"
	"github.com/stretchr/testify/require"
)

var (
	f = e2e.NewFramework(false)

	// Functions called after suite to clean up resources.
	cleanup []func()
)

func TestInfra(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Infra tests")
}

var _ = BeforeSuite(func() {
	// Add volume mount for the Envoy deployment for certificate and key,
	// used only for testing metrics over HTTPS.
	f.Deployment.EnvoyExtraVolumeMounts = []v1.VolumeMount{{
		Name:      "metrics-certs",
		MountPath: "/metrics-certs",
	}}
	f.Deployment.EnvoyExtraVolumes = []v1.Volume{{
		Name: "metrics-certs",
		VolumeSource: v1.VolumeSource{
			Secret: &v1.SecretVolumeSource{
				SecretName: "metrics-server",
			}},
	}}

	require.NoError(f.T(), f.Deployment.EnsureResourcesForLocalSesame())

	// Create certificate and key for metrics over HTTPS.
	cleanup = append(cleanup,
		f.Certs.CreateCA("projectsesame", "metrics-ca"),
		f.Certs.CreateCert("projectsesame", "metrics-server", "metrics-ca", "localhost"),
		f.Certs.CreateCert("projectsesame", "metrics-client", "metrics-ca"),
	)
})

var _ = AfterSuite(func() {
	// Delete resources individually instead of deleting the entire sesame
	// namespace as a performance optimization, because deleting non-empty
	// namespaces can take up to a couple of minutes to complete.
	for _, c := range cleanup {
		c()
	}
	require.NoError(f.T(), f.Deployment.DeleteResourcesForLocalSesame())
	gexec.CleanupBuildArtifacts()
})

var _ = Describe("Infra", func() {
	var (
		SesameCmd            *gexec.Session
		kubectlCmd           *gexec.Session
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

		kubectlCmd, err = f.Kubectl.StartKubectlPortForward(19001, 9001, "projectsesame", "daemonset/envoy", additionalSesameArgs...)
		require.NoError(f.T(), err)
	})

	AfterEach(func() {
		f.Kubectl.StopKubectlPortForward(kubectlCmd)
		require.NoError(f.T(), f.Deployment.StopLocalSesame(SesameCmd, SesameConfigFile))
	})

	f.Test(testMetrics)
	f.Test(testReady)

	Context("when serving metrics over HTTPS", func() {
		BeforeEach(func() {
			SesameConfig.Metrics.Envoy = config.MetricsServerParameters{
				Address:    "0.0.0.0",
				Port:       8003,
				ServerCert: "/metrics-certs/tls.crt",
				ServerKey:  "/metrics-certs/tls.key",
				CABundle:   "/metrics-certs/ca.crt",
			}

			SesameConfiguration.Spec.Envoy.Metrics = sesame_api_v1alpha1.MetricsConfig{
				Address: "0.0.0.0",
				Port:    8003,
				TLS: &sesame_api_v1alpha1.MetricsTLS{
					CertFile: "/metrics-certs/tls.crt",
					KeyFile:  "/metrics-certs/tls.key",
					CAFile:   "/metrics-certs/ca.crt",
				},
			}
		})
		f.Test(testEnvoyMetricsOverHTTPS)
	})

	f.Test(testAdminInterface)
})
