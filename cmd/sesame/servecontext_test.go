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

package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/projectsesame/sesame/pkg/config"
	"github.com/tsaarni/certyaml"
	"k8s.io/utils/pointer"

	sesame_api_v1alpha1 "github.com/projectsesame/sesame/apis/projectsesame/v1alpha1"
	envoy_v3 "github.com/projectsesame/sesame/internal/envoy/v3"
	"github.com/projectsesame/sesame/internal/fixture"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestServeContextProxyRootNamespaces(t *testing.T) {
	tests := map[string]struct {
		ctx  serveContext
		want []string
	}{
		"empty": {
			ctx: serveContext{
				rootNamespaces: "",
			},
			want: nil,
		},
		"blank-ish": {
			ctx: serveContext{
				rootNamespaces: " \t ",
			},
			want: nil,
		},
		"one value": {
			ctx: serveContext{
				rootNamespaces: "projectsesame",
			},
			want: []string{"projectsesame"},
		},
		"multiple, easy": {
			ctx: serveContext{
				rootNamespaces: "prod1,prod2,prod3",
			},
			want: []string{"prod1", "prod2", "prod3"},
		},
		"multiple, hard": {
			ctx: serveContext{
				rootNamespaces: "prod1, prod2, prod3 ",
			},
			want: []string{"prod1", "prod2", "prod3"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := tc.ctx.proxyRootNamespaces()
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("expected: %q, got: %q", tc.want, got)
			}
		})
	}
}

func TestServeContextTLSParams(t *testing.T) {
	tests := map[string]struct {
		tls         *sesame_api_v1alpha1.TLS
		expectError bool
	}{
		"tls supplied correctly": {
			tls: &sesame_api_v1alpha1.TLS{
				CAFile:   "cacert.pem",
				CertFile: "Sesamecert.pem",
				KeyFile:  "Sesamekey.pem",
				Insecure: false,
			},
			expectError: false,
		},
		"tls partially supplied": {
			tls: &sesame_api_v1alpha1.TLS{
				CertFile: "Sesamecert.pem",
				KeyFile:  "Sesamekey.pem",
				Insecure: false,
			},
			expectError: true,
		},
		"tls not supplied": {
			tls:         &sesame_api_v1alpha1.TLS{},
			expectError: true,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := verifyTLSFlags(tc.tls)
			goterror := err != nil
			if goterror != tc.expectError {
				t.Errorf("TLS Config: %s", err)
			}
		})
	}
}

func TestServeContextCertificateHandling(t *testing.T) {
	// Create trusted CA, server and client certs.
	trustedCACert := certyaml.Certificate{
		Subject: "cn=trusted-ca",
	}
	SesameCertBeforeRotation := certyaml.Certificate{
		Subject:         "cn=sesame-before-rotation",
		SubjectAltNames: []string{"DNS:localhost"},
		Issuer:          &trustedCACert,
	}
	SesameCertAfterRotation := certyaml.Certificate{
		Subject:         "cn=sesame-after-rotation",
		SubjectAltNames: []string{"DNS:localhost"},
		Issuer:          &trustedCACert,
	}
	trustedEnvoyCert := certyaml.Certificate{
		Subject: "cn=trusted-envoy",
		Issuer:  &trustedCACert,
	}

	// Create another CA and a client cert to test that untrusted clients are denied.
	untrustedCACert := certyaml.Certificate{
		Subject: "cn=untrusted-ca",
	}
	untrustedClientCert := certyaml.Certificate{
		Subject: "cn=untrusted-client",
		Issuer:  &untrustedCACert,
	}

	caCertPool := x509.NewCertPool()
	ca, err := trustedCACert.X509Certificate()
	checkFatalErr(t, err)
	caCertPool.AddCert(&ca)

	tests := map[string]struct {
		serverCredentials *certyaml.Certificate
		clientCredentials *certyaml.Certificate
		expectError       bool
	}{
		"successful TLS connection established": {
			serverCredentials: &SesameCertBeforeRotation,
			clientCredentials: &trustedEnvoyCert,
			expectError:       false,
		},
		"rotating server credentials returns new server cert": {
			serverCredentials: &SesameCertAfterRotation,
			clientCredentials: &trustedEnvoyCert,
			expectError:       false,
		},
		"rotating server credentials again to ensure rotation can be repeated": {
			serverCredentials: &SesameCertBeforeRotation,
			clientCredentials: &trustedEnvoyCert,
			expectError:       false,
		},
		"fail to connect with client certificate which is not signed by correct CA": {
			serverCredentials: &SesameCertBeforeRotation,
			clientCredentials: &untrustedClientCert,
			expectError:       true,
		},
	}

	// Create temporary directory to store certificates and key for the server.
	configDir, err := ioutil.TempDir("", "sesame-testdata-")
	checkFatalErr(t, err)
	defer os.RemoveAll(configDir)

	SesameTLS := &sesame_api_v1alpha1.TLS{
		CAFile:   filepath.Join(configDir, "CAcert.pem"),
		CertFile: filepath.Join(configDir, "Sesamecert.pem"),
		KeyFile:  filepath.Join(configDir, "Sesamekey.pem"),
		Insecure: false,
	}

	// Initial set of credentials must be written into temp directory before
	// starting the tests to avoid error at server startup.
	err = trustedCACert.WritePEM(SesameTLS.CAFile, filepath.Join(configDir, "CAkey.pem"))
	checkFatalErr(t, err)
	err = SesameCertBeforeRotation.WritePEM(SesameTLS.CertFile, SesameTLS.KeyFile)
	checkFatalErr(t, err)

	// Start a dummy server.
	log := fixture.NewTestLogger(t)
	opts := grpcOptions(log, SesameTLS)
	g := grpc.NewServer(opts...)
	if g == nil {
		t.Error("failed to create server")
	}

	address := "localhost:8001"
	l, err := net.Listen("tcp", address)
	checkFatalErr(t, err)

	go func() {
		err := g.Serve(l)
		checkFatalErr(t, err)
	}()
	defer g.GracefulStop()

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Store certificate and key to temp dir used by serveContext.
			err = tc.serverCredentials.WritePEM(SesameTLS.CertFile, SesameTLS.KeyFile)
			checkFatalErr(t, err)
			clientCert, _ := tc.clientCredentials.TLSCertificate()
			receivedCert, err := tryConnect(address, clientCert, caCertPool)
			gotError := err != nil
			if gotError != tc.expectError {
				t.Errorf("Unexpected result when connecting to the server: %s", err)
			}
			if err == nil {
				expectedCert, _ := tc.serverCredentials.X509Certificate()
				assert.Equal(t, receivedCert, &expectedCert)
			}
		})
	}
}

func TestTlsVersionDeprecation(t *testing.T) {
	// To get tls.Config for the gRPC XDS server, we need to arrange valid TLS certificates and keys.
	// Create temporary directory to store them for the server.
	configDir, err := ioutil.TempDir("", "sesame-testdata-")
	checkFatalErr(t, err)
	defer os.RemoveAll(configDir)

	caCert := certyaml.Certificate{
		Subject: "cn=ca",
	}
	SesameCert := certyaml.Certificate{
		Subject: "cn=SesameBeforeRotation",
		Issuer:  &caCert,
	}

	SesameTLS := &sesame_api_v1alpha1.TLS{
		CAFile:   filepath.Join(configDir, "CAcert.pem"),
		CertFile: filepath.Join(configDir, "Sesamecert.pem"),
		KeyFile:  filepath.Join(configDir, "Sesamekey.pem"),
		Insecure: false,
	}

	err = caCert.WritePEM(SesameTLS.CAFile, filepath.Join(configDir, "CAkey.pem"))
	checkFatalErr(t, err)
	err = SesameCert.WritePEM(SesameTLS.CertFile, SesameTLS.KeyFile)
	checkFatalErr(t, err)

	// Get preliminary TLS config from the serveContext.
	log := fixture.NewTestLogger(t)
	preliminaryTLSConfig := tlsconfig(log, SesameTLS)

	// Get actual TLS config that will be used during TLS handshake.
	tlsConfig, err := preliminaryTLSConfig.GetConfigForClient(nil)
	checkFatalErr(t, err)

	assert.Equal(t, tlsConfig.MinVersion, uint16(tls.VersionTLS13))
}

func checkFatalErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

// tryConnect tries to establish TLS connection to the server.
// If successful, return the server certificate.
func tryConnect(address string, clientCert tls.Certificate, caCertPool *x509.CertPool) (*x509.Certificate, error) {
	clientConfig := &tls.Config{
		ServerName:   "localhost",
		MinVersion:   tls.VersionTLS13,
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      caCertPool,
	}
	conn, err := tls.Dial("tcp", address, clientConfig)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	err = peekError(conn)
	if err != nil {
		return nil, err
	}

	return conn.ConnectionState().PeerCertificates[0], nil
}

// peekError is a workaround for TLS 1.3: due to shortened handshake, TLS alert
// from server is received at first read from the socket.
// To receive alert for bad certificate, this function tries to read one byte.
// Adapted from https://golang.org/src/crypto/tls/handshake_client_test.go
func peekError(conn net.Conn) error {
	_ = conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	_, err := conn.Read(make([]byte, 1))
	if err != nil {
		if netErr, ok := err.(net.Error); !ok || !netErr.Timeout() {
			return err
		}
	}
	return nil
}

func TestParseHTTPVersions(t *testing.T) {
	cases := map[string]struct {
		versions      []sesame_api_v1alpha1.HTTPVersionType
		parseVersions []envoy_v3.HTTPVersionType
	}{
		"empty": {
			versions:      []sesame_api_v1alpha1.HTTPVersionType{},
			parseVersions: nil,
		},
		"http/1.1": {
			versions:      []sesame_api_v1alpha1.HTTPVersionType{sesame_api_v1alpha1.HTTPVersion1},
			parseVersions: []envoy_v3.HTTPVersionType{envoy_v3.HTTPVersion1},
		},
		"http/1.1+http/2": {
			versions:      []sesame_api_v1alpha1.HTTPVersionType{sesame_api_v1alpha1.HTTPVersion1, sesame_api_v1alpha1.HTTPVersion2},
			parseVersions: []envoy_v3.HTTPVersionType{envoy_v3.HTTPVersion1, envoy_v3.HTTPVersion2},
		},
		"http/1.1+http/2 duplicated": {
			versions: []sesame_api_v1alpha1.HTTPVersionType{
				sesame_api_v1alpha1.HTTPVersion1, sesame_api_v1alpha1.HTTPVersion2,
				sesame_api_v1alpha1.HTTPVersion1, sesame_api_v1alpha1.HTTPVersion2},
			parseVersions: []envoy_v3.HTTPVersionType{envoy_v3.HTTPVersion1, envoy_v3.HTTPVersion2},
		},
	}

	for name, testcase := range cases {
		testcase := testcase
		t.Run(name, func(t *testing.T) {
			vers := parseDefaultHTTPVersions(testcase.versions)

			// parseDefaultHTTPVersions doesn't guarantee a stable result, but the order doesn't matter.
			sort.Slice(vers,
				func(i, j int) bool { return vers[i] < vers[j] })
			sort.Slice(testcase.parseVersions,
				func(i, j int) bool { return testcase.parseVersions[i] < testcase.parseVersions[j] })

			assert.Equal(t, testcase.parseVersions, vers)
		})
	}
}

func TestConvertServeContext(t *testing.T) {

	defaultContext := newServeContext()
	defaultContext.ServerConfig = ServerConfig{
		xdsAddr:    "127.0.0.1",
		xdsPort:    8001,
		caFile:     "/certs/ca.crt",
		SesameCert: "/certs/cert.crt",
		SesameKey:  "/certs/cert.key",
	}

	headersPolicyContext := newServeContext()
	headersPolicyContext.Config.Policy = config.PolicyParameters{
		RequestHeadersPolicy: config.HeadersPolicy{
			Set:    map[string]string{"custom-request-header-set": "foo-bar", "Host": "request-bar.com"},
			Remove: []string{"custom-request-header-remove"},
		},
		ResponseHeadersPolicy: config.HeadersPolicy{
			Set:    map[string]string{"custom-response-header-set": "foo-bar", "Host": "response-bar.com"},
			Remove: []string{"custom-response-header-remove"},
		},
		ApplyToIngress: true,
	}

	gatewayContext := newServeContext()
	gatewayContext.Config.GatewayConfig = &config.GatewayParameters{
		ControllerName: "projectsesame.io/projectsesame/sesame",
	}

	ingressContext := newServeContext()
	ingressContext.ingressClassName = "coolclass"
	ingressContext.Config.IngressStatusAddress = "1.2.3.4"

	clientCertificate := newServeContext()
	clientCertificate.Config.TLS.ClientCertificate = config.NamespacedName{
		Name:      "cert",
		Namespace: "secretplace",
	}

	httpProxy := newServeContext()
	httpProxy.Config.DisablePermitInsecure = true
	httpProxy.Config.TLS.FallbackCertificate = config.NamespacedName{
		Name:      "fallbackname",
		Namespace: "fallbacknamespace",
	}

	rateLimit := newServeContext()
	rateLimit.Config.RateLimitService = config.RateLimitService{
		ExtensionService:        "ratens/ratelimitext",
		Domain:                  "sesame",
		FailOpen:                true,
		EnableXRateLimitHeaders: true,
	}

	defaultHTTPVersions := newServeContext()
	defaultHTTPVersions.Config.DefaultHTTPVersions = []config.HTTPVersionType{
		config.HTTPVersion1,
	}

	accessLog := newServeContext()
	accessLog.Config.AccessLogFormat = config.JSONAccessLog
	accessLog.Config.AccessLogFormatString = "foo-bar-baz"
	accessLog.Config.AccessLogFields = []string{"custom_field"}

	cases := map[string]struct {
		serveContext *serveContext
		SesameConfig sesame_api_v1alpha1.SesameConfigurationSpec
	}{
		"default ServeContext": {
			serveContext: defaultContext,
			SesameConfig: sesame_api_v1alpha1.SesameConfigurationSpec{
				XDSServer: sesame_api_v1alpha1.XDSServerConfig{
					Type:    sesame_api_v1alpha1.SesameServerType,
					Address: "127.0.0.1",
					Port:    8001,
					TLS: &sesame_api_v1alpha1.TLS{
						CAFile:   "/certs/ca.crt",
						CertFile: "/certs/cert.crt",
						KeyFile:  "/certs/cert.key",
						Insecure: false,
					},
				},
				Ingress: &sesame_api_v1alpha1.IngressConfig{
					ClassName:     nil,
					StatusAddress: nil,
				},
				Debug: sesame_api_v1alpha1.DebugConfig{
					Address:                 "127.0.0.1",
					Port:                    6060,
					DebugLogLevel:           sesame_api_v1alpha1.InfoLog,
					KubernetesDebugLogLevel: 0,
				},
				Health: sesame_api_v1alpha1.HealthConfig{
					Address: "0.0.0.0",
					Port:    8000,
				},
				Envoy: sesame_api_v1alpha1.EnvoyConfig{
					Service: sesame_api_v1alpha1.NamespacedName{
						Name:      "envoy",
						Namespace: "projectsesame",
					},
					HTTPListener: sesame_api_v1alpha1.EnvoyListener{
						Address:   "0.0.0.0",
						Port:      8080,
						AccessLog: "/dev/stdout",
					},
					HTTPSListener: sesame_api_v1alpha1.EnvoyListener{
						Address:   "0.0.0.0",
						Port:      8443,
						AccessLog: "/dev/stdout",
					},
					Health: sesame_api_v1alpha1.HealthConfig{
						Address: "0.0.0.0",
						Port:    8002,
					},
					Metrics: sesame_api_v1alpha1.MetricsConfig{
						Address: "0.0.0.0",
						Port:    8002,
					},
					ClientCertificate: nil,
					Logging: sesame_api_v1alpha1.EnvoyLogging{
						AccessLogFormat:       sesame_api_v1alpha1.EnvoyAccessLog,
						AccessLogFormatString: nil,
						AccessLogFields: sesame_api_v1alpha1.AccessLogFields([]string{
							"@timestamp",
							"authority",
							"bytes_received",
							"bytes_sent",
							"downstream_local_address",
							"downstream_remote_address",
							"duration",
							"method",
							"path",
							"protocol",
							"request_id",
							"requested_server_name",
							"response_code",
							"response_flags",
							"uber_trace_id",
							"upstream_cluster",
							"upstream_host",
							"upstream_local_address",
							"upstream_service_time",
							"user_agent",
							"x_forwarded_for",
						}),
					},
					DefaultHTTPVersions: nil,
					Timeouts: &sesame_api_v1alpha1.TimeoutParameters{
						ConnectionIdleTimeout: pointer.StringPtr("60s"),
					},
					Cluster: sesame_api_v1alpha1.ClusterParameters{
						DNSLookupFamily: sesame_api_v1alpha1.AutoClusterDNSFamily,
					},
					Network: sesame_api_v1alpha1.NetworkParameters{
						EnvoyAdminPort: 9001,
					},
				},
				Gateway: nil,
				HTTPProxy: sesame_api_v1alpha1.HTTPProxyConfig{
					DisablePermitInsecure: false,
					FallbackCertificate:   nil,
				},
				EnableExternalNameService: false,
				RateLimitService:          nil,
				Policy: &sesame_api_v1alpha1.PolicyConfig{
					RequestHeadersPolicy:  &sesame_api_v1alpha1.HeadersPolicy{},
					ResponseHeadersPolicy: &sesame_api_v1alpha1.HeadersPolicy{},
					ApplyToIngress:        false,
				},
				Metrics: sesame_api_v1alpha1.MetricsConfig{
					Address: "0.0.0.0",
					Port:    8000,
				},
			},
		},
		"headers policy": {
			serveContext: headersPolicyContext,
			SesameConfig: sesame_api_v1alpha1.SesameConfigurationSpec{
				XDSServer: sesame_api_v1alpha1.XDSServerConfig{
					Type:    sesame_api_v1alpha1.SesameServerType,
					Address: "127.0.0.1",
					Port:    8001,
					TLS: &sesame_api_v1alpha1.TLS{
						Insecure: false,
					},
				},
				Ingress: &sesame_api_v1alpha1.IngressConfig{
					ClassName:     nil,
					StatusAddress: nil,
				},
				Debug: sesame_api_v1alpha1.DebugConfig{
					Address:                 "127.0.0.1",
					Port:                    6060,
					DebugLogLevel:           sesame_api_v1alpha1.InfoLog,
					KubernetesDebugLogLevel: 0,
				},
				Health: sesame_api_v1alpha1.HealthConfig{
					Address: "0.0.0.0",
					Port:    8000,
				},
				Envoy: sesame_api_v1alpha1.EnvoyConfig{
					Service: sesame_api_v1alpha1.NamespacedName{
						Name:      "envoy",
						Namespace: "projectsesame",
					},
					HTTPListener: sesame_api_v1alpha1.EnvoyListener{
						Address:   "0.0.0.0",
						Port:      8080,
						AccessLog: "/dev/stdout",
					},
					HTTPSListener: sesame_api_v1alpha1.EnvoyListener{
						Address:   "0.0.0.0",
						Port:      8443,
						AccessLog: "/dev/stdout",
					},
					Health: sesame_api_v1alpha1.HealthConfig{
						Address: "0.0.0.0",
						Port:    8002,
					},
					Metrics: sesame_api_v1alpha1.MetricsConfig{
						Address: "0.0.0.0",
						Port:    8002,
					},
					ClientCertificate: nil,
					Logging: sesame_api_v1alpha1.EnvoyLogging{
						AccessLogFormat:       sesame_api_v1alpha1.EnvoyAccessLog,
						AccessLogFormatString: nil,
						AccessLogFields: sesame_api_v1alpha1.AccessLogFields([]string{
							"@timestamp",
							"authority",
							"bytes_received",
							"bytes_sent",
							"downstream_local_address",
							"downstream_remote_address",
							"duration",
							"method",
							"path",
							"protocol",
							"request_id",
							"requested_server_name",
							"response_code",
							"response_flags",
							"uber_trace_id",
							"upstream_cluster",
							"upstream_host",
							"upstream_local_address",
							"upstream_service_time",
							"user_agent",
							"x_forwarded_for",
						}),
					},
					DefaultHTTPVersions: nil,
					Timeouts: &sesame_api_v1alpha1.TimeoutParameters{
						ConnectionIdleTimeout: pointer.StringPtr("60s"),
					},
					Cluster: sesame_api_v1alpha1.ClusterParameters{
						DNSLookupFamily: sesame_api_v1alpha1.AutoClusterDNSFamily,
					},
					Network: sesame_api_v1alpha1.NetworkParameters{
						EnvoyAdminPort: 9001,
					},
				},
				Gateway: nil,
				HTTPProxy: sesame_api_v1alpha1.HTTPProxyConfig{
					DisablePermitInsecure: false,
					FallbackCertificate:   nil,
				},
				EnableExternalNameService: false,
				RateLimitService:          nil,
				Policy: &sesame_api_v1alpha1.PolicyConfig{
					RequestHeadersPolicy: &sesame_api_v1alpha1.HeadersPolicy{
						Set:    map[string]string{"custom-request-header-set": "foo-bar", "Host": "request-bar.com"},
						Remove: []string{"custom-request-header-remove"},
					},
					ResponseHeadersPolicy: &sesame_api_v1alpha1.HeadersPolicy{
						Set:    map[string]string{"custom-response-header-set": "foo-bar", "Host": "response-bar.com"},
						Remove: []string{"custom-response-header-remove"},
					},
					ApplyToIngress: true,
				},
				Metrics: sesame_api_v1alpha1.MetricsConfig{
					Address: "0.0.0.0",
					Port:    8000,
				},
			},
		},
		"ingress": {
			serveContext: ingressContext,
			SesameConfig: sesame_api_v1alpha1.SesameConfigurationSpec{
				XDSServer: sesame_api_v1alpha1.XDSServerConfig{
					Type:    sesame_api_v1alpha1.SesameServerType,
					Address: "127.0.0.1",
					Port:    8001,
					TLS: &sesame_api_v1alpha1.TLS{
						Insecure: false,
					},
				},
				Ingress: &sesame_api_v1alpha1.IngressConfig{
					ClassName:     pointer.StringPtr("coolclass"),
					StatusAddress: pointer.StringPtr("1.2.3.4"),
				},
				Debug: sesame_api_v1alpha1.DebugConfig{
					Address:                 "127.0.0.1",
					Port:                    6060,
					DebugLogLevel:           sesame_api_v1alpha1.InfoLog,
					KubernetesDebugLogLevel: 0,
				},
				Health: sesame_api_v1alpha1.HealthConfig{
					Address: "0.0.0.0",
					Port:    8000,
				},
				Envoy: sesame_api_v1alpha1.EnvoyConfig{
					Service: sesame_api_v1alpha1.NamespacedName{
						Name:      "envoy",
						Namespace: "projectsesame",
					},
					HTTPListener: sesame_api_v1alpha1.EnvoyListener{
						Address:   "0.0.0.0",
						Port:      8080,
						AccessLog: "/dev/stdout",
					},
					HTTPSListener: sesame_api_v1alpha1.EnvoyListener{
						Address:   "0.0.0.0",
						Port:      8443,
						AccessLog: "/dev/stdout",
					},
					Health: sesame_api_v1alpha1.HealthConfig{
						Address: "0.0.0.0",
						Port:    8002,
					},
					Metrics: sesame_api_v1alpha1.MetricsConfig{
						Address: "0.0.0.0",
						Port:    8002,
					},
					ClientCertificate: nil,
					Logging: sesame_api_v1alpha1.EnvoyLogging{
						AccessLogFormat:       sesame_api_v1alpha1.EnvoyAccessLog,
						AccessLogFormatString: nil,
						AccessLogFields: sesame_api_v1alpha1.AccessLogFields([]string{
							"@timestamp",
							"authority",
							"bytes_received",
							"bytes_sent",
							"downstream_local_address",
							"downstream_remote_address",
							"duration",
							"method",
							"path",
							"protocol",
							"request_id",
							"requested_server_name",
							"response_code",
							"response_flags",
							"uber_trace_id",
							"upstream_cluster",
							"upstream_host",
							"upstream_local_address",
							"upstream_service_time",
							"user_agent",
							"x_forwarded_for",
						}),
					},
					DefaultHTTPVersions: nil,
					Timeouts: &sesame_api_v1alpha1.TimeoutParameters{
						ConnectionIdleTimeout: pointer.StringPtr("60s"),
					},
					Cluster: sesame_api_v1alpha1.ClusterParameters{
						DNSLookupFamily: sesame_api_v1alpha1.AutoClusterDNSFamily,
					},
					Network: sesame_api_v1alpha1.NetworkParameters{
						EnvoyAdminPort: 9001,
					},
				},
				Gateway: nil,
				HTTPProxy: sesame_api_v1alpha1.HTTPProxyConfig{
					DisablePermitInsecure: false,
					FallbackCertificate:   nil,
				},
				EnableExternalNameService: false,
				RateLimitService:          nil,
				Policy: &sesame_api_v1alpha1.PolicyConfig{
					RequestHeadersPolicy:  &sesame_api_v1alpha1.HeadersPolicy{},
					ResponseHeadersPolicy: &sesame_api_v1alpha1.HeadersPolicy{},
					ApplyToIngress:        false,
				},
				Metrics: sesame_api_v1alpha1.MetricsConfig{
					Address: "0.0.0.0",
					Port:    8000,
				},
			},
		},
		"gatewayapi": {
			serveContext: gatewayContext,
			SesameConfig: sesame_api_v1alpha1.SesameConfigurationSpec{
				XDSServer: sesame_api_v1alpha1.XDSServerConfig{
					Type:    sesame_api_v1alpha1.SesameServerType,
					Address: "127.0.0.1",
					Port:    8001,
					TLS: &sesame_api_v1alpha1.TLS{
						Insecure: false,
					},
				},
				Ingress: &sesame_api_v1alpha1.IngressConfig{
					ClassName:     nil,
					StatusAddress: nil,
				},
				Debug: sesame_api_v1alpha1.DebugConfig{
					Address:                 "127.0.0.1",
					Port:                    6060,
					DebugLogLevel:           sesame_api_v1alpha1.InfoLog,
					KubernetesDebugLogLevel: 0,
				},
				Health: sesame_api_v1alpha1.HealthConfig{
					Address: "0.0.0.0",
					Port:    8000,
				},
				Envoy: sesame_api_v1alpha1.EnvoyConfig{
					Service: sesame_api_v1alpha1.NamespacedName{
						Name:      "envoy",
						Namespace: "projectsesame",
					},
					HTTPListener: sesame_api_v1alpha1.EnvoyListener{
						Address:   "0.0.0.0",
						Port:      8080,
						AccessLog: "/dev/stdout",
					},
					HTTPSListener: sesame_api_v1alpha1.EnvoyListener{
						Address:   "0.0.0.0",
						Port:      8443,
						AccessLog: "/dev/stdout",
					},
					Health: sesame_api_v1alpha1.HealthConfig{
						Address: "0.0.0.0",
						Port:    8002,
					},
					Metrics: sesame_api_v1alpha1.MetricsConfig{
						Address: "0.0.0.0",
						Port:    8002,
					},
					ClientCertificate: nil,
					Logging: sesame_api_v1alpha1.EnvoyLogging{
						AccessLogFormat:       sesame_api_v1alpha1.EnvoyAccessLog,
						AccessLogFormatString: nil,
						AccessLogFields: sesame_api_v1alpha1.AccessLogFields([]string{
							"@timestamp",
							"authority",
							"bytes_received",
							"bytes_sent",
							"downstream_local_address",
							"downstream_remote_address",
							"duration",
							"method",
							"path",
							"protocol",
							"request_id",
							"requested_server_name",
							"response_code",
							"response_flags",
							"uber_trace_id",
							"upstream_cluster",
							"upstream_host",
							"upstream_local_address",
							"upstream_service_time",
							"user_agent",
							"x_forwarded_for",
						}),
					},
					DefaultHTTPVersions: nil,
					Timeouts: &sesame_api_v1alpha1.TimeoutParameters{
						ConnectionIdleTimeout: pointer.StringPtr("60s"),
					},
					Cluster: sesame_api_v1alpha1.ClusterParameters{
						DNSLookupFamily: sesame_api_v1alpha1.AutoClusterDNSFamily,
					},
					Network: sesame_api_v1alpha1.NetworkParameters{
						EnvoyAdminPort: 9001,
					},
				},
				Gateway: &sesame_api_v1alpha1.GatewayConfig{
					ControllerName: "projectsesame.io/projectsesame/sesame",
				},
				HTTPProxy: sesame_api_v1alpha1.HTTPProxyConfig{
					DisablePermitInsecure: false,
					FallbackCertificate:   nil,
				},
				EnableExternalNameService: false,
				RateLimitService:          nil,
				Policy: &sesame_api_v1alpha1.PolicyConfig{
					RequestHeadersPolicy:  &sesame_api_v1alpha1.HeadersPolicy{},
					ResponseHeadersPolicy: &sesame_api_v1alpha1.HeadersPolicy{},
					ApplyToIngress:        false,
				},
				Metrics: sesame_api_v1alpha1.MetricsConfig{
					Address: "0.0.0.0",
					Port:    8000,
				},
			},
		},
		"client certificate": {
			serveContext: clientCertificate,
			SesameConfig: sesame_api_v1alpha1.SesameConfigurationSpec{
				XDSServer: sesame_api_v1alpha1.XDSServerConfig{
					Type:    sesame_api_v1alpha1.SesameServerType,
					Address: "127.0.0.1",
					Port:    8001,
					TLS: &sesame_api_v1alpha1.TLS{
						Insecure: false,
					},
				},
				Ingress: &sesame_api_v1alpha1.IngressConfig{
					ClassName:     nil,
					StatusAddress: nil,
				},
				Debug: sesame_api_v1alpha1.DebugConfig{
					Address:                 "127.0.0.1",
					Port:                    6060,
					DebugLogLevel:           sesame_api_v1alpha1.InfoLog,
					KubernetesDebugLogLevel: 0,
				},
				Health: sesame_api_v1alpha1.HealthConfig{
					Address: "0.0.0.0",
					Port:    8000,
				},
				Envoy: sesame_api_v1alpha1.EnvoyConfig{
					Service: sesame_api_v1alpha1.NamespacedName{
						Name:      "envoy",
						Namespace: "projectsesame",
					},
					HTTPListener: sesame_api_v1alpha1.EnvoyListener{
						Address:   "0.0.0.0",
						Port:      8080,
						AccessLog: "/dev/stdout",
					},
					HTTPSListener: sesame_api_v1alpha1.EnvoyListener{
						Address:   "0.0.0.0",
						Port:      8443,
						AccessLog: "/dev/stdout",
					},
					Health: sesame_api_v1alpha1.HealthConfig{
						Address: "0.0.0.0",
						Port:    8002,
					},
					Metrics: sesame_api_v1alpha1.MetricsConfig{
						Address: "0.0.0.0",
						Port:    8002,
					},
					ClientCertificate: &sesame_api_v1alpha1.NamespacedName{
						Name:      "cert",
						Namespace: "secretplace",
					},
					Logging: sesame_api_v1alpha1.EnvoyLogging{
						AccessLogFormat:       sesame_api_v1alpha1.EnvoyAccessLog,
						AccessLogFormatString: nil,
						AccessLogFields: sesame_api_v1alpha1.AccessLogFields([]string{
							"@timestamp",
							"authority",
							"bytes_received",
							"bytes_sent",
							"downstream_local_address",
							"downstream_remote_address",
							"duration",
							"method",
							"path",
							"protocol",
							"request_id",
							"requested_server_name",
							"response_code",
							"response_flags",
							"uber_trace_id",
							"upstream_cluster",
							"upstream_host",
							"upstream_local_address",
							"upstream_service_time",
							"user_agent",
							"x_forwarded_for",
						}),
					},
					DefaultHTTPVersions: nil,
					Timeouts: &sesame_api_v1alpha1.TimeoutParameters{
						ConnectionIdleTimeout: pointer.StringPtr("60s"),
					},
					Cluster: sesame_api_v1alpha1.ClusterParameters{
						DNSLookupFamily: sesame_api_v1alpha1.AutoClusterDNSFamily,
					},
					Network: sesame_api_v1alpha1.NetworkParameters{
						EnvoyAdminPort: 9001,
					},
				},
				Gateway: nil,
				HTTPProxy: sesame_api_v1alpha1.HTTPProxyConfig{
					DisablePermitInsecure: false,
					FallbackCertificate:   nil,
				},
				EnableExternalNameService: false,
				RateLimitService:          nil,
				Policy: &sesame_api_v1alpha1.PolicyConfig{
					RequestHeadersPolicy:  &sesame_api_v1alpha1.HeadersPolicy{},
					ResponseHeadersPolicy: &sesame_api_v1alpha1.HeadersPolicy{},
					ApplyToIngress:        false,
				},
				Metrics: sesame_api_v1alpha1.MetricsConfig{
					Address: "0.0.0.0",
					Port:    8000,
				},
			},
		},
		"httpproxy": {
			serveContext: httpProxy,
			SesameConfig: sesame_api_v1alpha1.SesameConfigurationSpec{
				XDSServer: sesame_api_v1alpha1.XDSServerConfig{
					Type:    sesame_api_v1alpha1.SesameServerType,
					Address: "127.0.0.1",
					Port:    8001,
					TLS: &sesame_api_v1alpha1.TLS{
						Insecure: false,
					},
				},
				Ingress: &sesame_api_v1alpha1.IngressConfig{
					ClassName:     nil,
					StatusAddress: nil,
				},
				Debug: sesame_api_v1alpha1.DebugConfig{
					Address:                 "127.0.0.1",
					Port:                    6060,
					DebugLogLevel:           sesame_api_v1alpha1.InfoLog,
					KubernetesDebugLogLevel: 0,
				},
				Health: sesame_api_v1alpha1.HealthConfig{
					Address: "0.0.0.0",
					Port:    8000,
				},
				Envoy: sesame_api_v1alpha1.EnvoyConfig{
					Service: sesame_api_v1alpha1.NamespacedName{
						Name:      "envoy",
						Namespace: "projectsesame",
					},
					HTTPListener: sesame_api_v1alpha1.EnvoyListener{
						Address:   "0.0.0.0",
						Port:      8080,
						AccessLog: "/dev/stdout",
					},
					HTTPSListener: sesame_api_v1alpha1.EnvoyListener{
						Address:   "0.0.0.0",
						Port:      8443,
						AccessLog: "/dev/stdout",
					},
					Health: sesame_api_v1alpha1.HealthConfig{
						Address: "0.0.0.0",
						Port:    8002,
					},
					Metrics: sesame_api_v1alpha1.MetricsConfig{
						Address: "0.0.0.0",
						Port:    8002,
					},
					ClientCertificate: nil,
					Logging: sesame_api_v1alpha1.EnvoyLogging{
						AccessLogFormat:       sesame_api_v1alpha1.EnvoyAccessLog,
						AccessLogFormatString: nil,
						AccessLogFields: sesame_api_v1alpha1.AccessLogFields([]string{
							"@timestamp",
							"authority",
							"bytes_received",
							"bytes_sent",
							"downstream_local_address",
							"downstream_remote_address",
							"duration",
							"method",
							"path",
							"protocol",
							"request_id",
							"requested_server_name",
							"response_code",
							"response_flags",
							"uber_trace_id",
							"upstream_cluster",
							"upstream_host",
							"upstream_local_address",
							"upstream_service_time",
							"user_agent",
							"x_forwarded_for",
						}),
					},
					DefaultHTTPVersions: nil,
					Timeouts: &sesame_api_v1alpha1.TimeoutParameters{
						ConnectionIdleTimeout: pointer.StringPtr("60s"),
					},
					Cluster: sesame_api_v1alpha1.ClusterParameters{
						DNSLookupFamily: sesame_api_v1alpha1.AutoClusterDNSFamily,
					},
					Network: sesame_api_v1alpha1.NetworkParameters{
						EnvoyAdminPort: 9001,
					},
				},
				Gateway: nil,
				HTTPProxy: sesame_api_v1alpha1.HTTPProxyConfig{
					DisablePermitInsecure: true,
					FallbackCertificate: &sesame_api_v1alpha1.NamespacedName{
						Name:      "fallbackname",
						Namespace: "fallbacknamespace",
					},
				},
				EnableExternalNameService: false,
				RateLimitService:          nil,
				Policy: &sesame_api_v1alpha1.PolicyConfig{
					RequestHeadersPolicy:  &sesame_api_v1alpha1.HeadersPolicy{},
					ResponseHeadersPolicy: &sesame_api_v1alpha1.HeadersPolicy{},
					ApplyToIngress:        false,
				},
				Metrics: sesame_api_v1alpha1.MetricsConfig{
					Address: "0.0.0.0",
					Port:    8000,
				},
			},
		},
		"ratelimit": {
			serveContext: rateLimit,
			SesameConfig: sesame_api_v1alpha1.SesameConfigurationSpec{
				XDSServer: sesame_api_v1alpha1.XDSServerConfig{
					Type:    sesame_api_v1alpha1.SesameServerType,
					Address: "127.0.0.1",
					Port:    8001,
					TLS: &sesame_api_v1alpha1.TLS{
						Insecure: false,
					},
				},
				Ingress: &sesame_api_v1alpha1.IngressConfig{
					ClassName:     nil,
					StatusAddress: nil,
				},
				Debug: sesame_api_v1alpha1.DebugConfig{
					Address:                 "127.0.0.1",
					Port:                    6060,
					DebugLogLevel:           sesame_api_v1alpha1.InfoLog,
					KubernetesDebugLogLevel: 0,
				},
				Health: sesame_api_v1alpha1.HealthConfig{
					Address: "0.0.0.0",
					Port:    8000,
				},
				Envoy: sesame_api_v1alpha1.EnvoyConfig{
					Service: sesame_api_v1alpha1.NamespacedName{
						Name:      "envoy",
						Namespace: "projectsesame",
					},
					HTTPListener: sesame_api_v1alpha1.EnvoyListener{
						Address:   "0.0.0.0",
						Port:      8080,
						AccessLog: "/dev/stdout",
					},
					HTTPSListener: sesame_api_v1alpha1.EnvoyListener{
						Address:   "0.0.0.0",
						Port:      8443,
						AccessLog: "/dev/stdout",
					},
					Health: sesame_api_v1alpha1.HealthConfig{
						Address: "0.0.0.0",
						Port:    8002,
					},
					Metrics: sesame_api_v1alpha1.MetricsConfig{
						Address: "0.0.0.0",
						Port:    8002,
					},
					ClientCertificate: nil,
					Logging: sesame_api_v1alpha1.EnvoyLogging{
						AccessLogFormat:       sesame_api_v1alpha1.EnvoyAccessLog,
						AccessLogFormatString: nil,
						AccessLogFields: sesame_api_v1alpha1.AccessLogFields([]string{
							"@timestamp",
							"authority",
							"bytes_received",
							"bytes_sent",
							"downstream_local_address",
							"downstream_remote_address",
							"duration",
							"method",
							"path",
							"protocol",
							"request_id",
							"requested_server_name",
							"response_code",
							"response_flags",
							"uber_trace_id",
							"upstream_cluster",
							"upstream_host",
							"upstream_local_address",
							"upstream_service_time",
							"user_agent",
							"x_forwarded_for",
						}),
					},
					DefaultHTTPVersions: nil,
					Timeouts: &sesame_api_v1alpha1.TimeoutParameters{
						ConnectionIdleTimeout: pointer.StringPtr("60s"),
					},
					Cluster: sesame_api_v1alpha1.ClusterParameters{
						DNSLookupFamily: sesame_api_v1alpha1.AutoClusterDNSFamily,
					},
					Network: sesame_api_v1alpha1.NetworkParameters{
						EnvoyAdminPort: 9001,
					},
				},
				Gateway: nil,
				HTTPProxy: sesame_api_v1alpha1.HTTPProxyConfig{
					DisablePermitInsecure: false,
					FallbackCertificate:   nil,
				},
				EnableExternalNameService: false,
				RateLimitService: &sesame_api_v1alpha1.RateLimitServiceConfig{
					ExtensionService: sesame_api_v1alpha1.NamespacedName{
						Name:      "ratelimitext",
						Namespace: "ratens",
					},
					Domain:                  "sesame",
					FailOpen:                true,
					EnableXRateLimitHeaders: true,
				},
				Policy: &sesame_api_v1alpha1.PolicyConfig{
					RequestHeadersPolicy:  &sesame_api_v1alpha1.HeadersPolicy{},
					ResponseHeadersPolicy: &sesame_api_v1alpha1.HeadersPolicy{},
					ApplyToIngress:        false,
				},
				Metrics: sesame_api_v1alpha1.MetricsConfig{
					Address: "0.0.0.0",
					Port:    8000,
				},
			},
		},
		"default http versions": {
			serveContext: defaultHTTPVersions,
			SesameConfig: sesame_api_v1alpha1.SesameConfigurationSpec{
				XDSServer: sesame_api_v1alpha1.XDSServerConfig{
					Type:    sesame_api_v1alpha1.SesameServerType,
					Address: "127.0.0.1",
					Port:    8001,
					TLS: &sesame_api_v1alpha1.TLS{
						Insecure: false,
					},
				},
				Ingress: &sesame_api_v1alpha1.IngressConfig{
					ClassName:     nil,
					StatusAddress: nil,
				},
				Debug: sesame_api_v1alpha1.DebugConfig{
					Address:                 "127.0.0.1",
					Port:                    6060,
					DebugLogLevel:           sesame_api_v1alpha1.InfoLog,
					KubernetesDebugLogLevel: 0,
				},
				Health: sesame_api_v1alpha1.HealthConfig{
					Address: "0.0.0.0",
					Port:    8000,
				},
				Envoy: sesame_api_v1alpha1.EnvoyConfig{
					Service: sesame_api_v1alpha1.NamespacedName{
						Name:      "envoy",
						Namespace: "projectsesame",
					},
					HTTPListener: sesame_api_v1alpha1.EnvoyListener{
						Address:   "0.0.0.0",
						Port:      8080,
						AccessLog: "/dev/stdout",
					},
					HTTPSListener: sesame_api_v1alpha1.EnvoyListener{
						Address:   "0.0.0.0",
						Port:      8443,
						AccessLog: "/dev/stdout",
					},
					Health: sesame_api_v1alpha1.HealthConfig{
						Address: "0.0.0.0",
						Port:    8002,
					},
					Metrics: sesame_api_v1alpha1.MetricsConfig{
						Address: "0.0.0.0",
						Port:    8002,
					},
					ClientCertificate: nil,
					Logging: sesame_api_v1alpha1.EnvoyLogging{
						AccessLogFormat:       sesame_api_v1alpha1.EnvoyAccessLog,
						AccessLogFormatString: nil,
						AccessLogFields: sesame_api_v1alpha1.AccessLogFields([]string{
							"@timestamp",
							"authority",
							"bytes_received",
							"bytes_sent",
							"downstream_local_address",
							"downstream_remote_address",
							"duration",
							"method",
							"path",
							"protocol",
							"request_id",
							"requested_server_name",
							"response_code",
							"response_flags",
							"uber_trace_id",
							"upstream_cluster",
							"upstream_host",
							"upstream_local_address",
							"upstream_service_time",
							"user_agent",
							"x_forwarded_for",
						}),
					},
					DefaultHTTPVersions: []sesame_api_v1alpha1.HTTPVersionType{
						sesame_api_v1alpha1.HTTPVersion1,
					},
					Timeouts: &sesame_api_v1alpha1.TimeoutParameters{
						ConnectionIdleTimeout: pointer.StringPtr("60s"),
					},
					Cluster: sesame_api_v1alpha1.ClusterParameters{
						DNSLookupFamily: sesame_api_v1alpha1.AutoClusterDNSFamily,
					},
					Network: sesame_api_v1alpha1.NetworkParameters{
						EnvoyAdminPort: 9001,
					},
				},
				Gateway: nil,
				HTTPProxy: sesame_api_v1alpha1.HTTPProxyConfig{
					DisablePermitInsecure: false,
					FallbackCertificate:   nil,
				},
				EnableExternalNameService: false,
				RateLimitService:          nil,
				Policy: &sesame_api_v1alpha1.PolicyConfig{
					RequestHeadersPolicy:  &sesame_api_v1alpha1.HeadersPolicy{},
					ResponseHeadersPolicy: &sesame_api_v1alpha1.HeadersPolicy{},
					ApplyToIngress:        false,
				},
				Metrics: sesame_api_v1alpha1.MetricsConfig{
					Address: "0.0.0.0",
					Port:    8000,
				},
			},
		},
		"access log": {
			serveContext: accessLog,
			SesameConfig: sesame_api_v1alpha1.SesameConfigurationSpec{
				XDSServer: sesame_api_v1alpha1.XDSServerConfig{
					Type:    sesame_api_v1alpha1.SesameServerType,
					Address: "127.0.0.1",
					Port:    8001,
					TLS: &sesame_api_v1alpha1.TLS{
						Insecure: false,
					},
				},
				Ingress: &sesame_api_v1alpha1.IngressConfig{
					ClassName:     nil,
					StatusAddress: nil,
				},
				Debug: sesame_api_v1alpha1.DebugConfig{
					Address:                 "127.0.0.1",
					Port:                    6060,
					DebugLogLevel:           sesame_api_v1alpha1.InfoLog,
					KubernetesDebugLogLevel: 0,
				},
				Health: sesame_api_v1alpha1.HealthConfig{
					Address: "0.0.0.0",
					Port:    8000,
				},
				Envoy: sesame_api_v1alpha1.EnvoyConfig{
					Service: sesame_api_v1alpha1.NamespacedName{
						Name:      "envoy",
						Namespace: "projectsesame",
					},
					HTTPListener: sesame_api_v1alpha1.EnvoyListener{
						Address:   "0.0.0.0",
						Port:      8080,
						AccessLog: "/dev/stdout",
					},
					HTTPSListener: sesame_api_v1alpha1.EnvoyListener{
						Address:   "0.0.0.0",
						Port:      8443,
						AccessLog: "/dev/stdout",
					},
					Health: sesame_api_v1alpha1.HealthConfig{
						Address: "0.0.0.0",
						Port:    8002,
					},
					Metrics: sesame_api_v1alpha1.MetricsConfig{
						Address: "0.0.0.0",
						Port:    8002,
					},
					ClientCertificate: nil,
					Logging: sesame_api_v1alpha1.EnvoyLogging{
						AccessLogFormat:       sesame_api_v1alpha1.JSONAccessLog,
						AccessLogFormatString: pointer.StringPtr("foo-bar-baz"),
						AccessLogFields: sesame_api_v1alpha1.AccessLogFields([]string{
							"custom_field",
						}),
					},
					DefaultHTTPVersions: nil,
					Timeouts: &sesame_api_v1alpha1.TimeoutParameters{
						ConnectionIdleTimeout: pointer.StringPtr("60s"),
					},
					Cluster: sesame_api_v1alpha1.ClusterParameters{
						DNSLookupFamily: sesame_api_v1alpha1.AutoClusterDNSFamily,
					},
					Network: sesame_api_v1alpha1.NetworkParameters{
						EnvoyAdminPort: 9001,
					},
				},
				Gateway: nil,
				HTTPProxy: sesame_api_v1alpha1.HTTPProxyConfig{
					DisablePermitInsecure: false,
					FallbackCertificate:   nil,
				},
				EnableExternalNameService: false,
				RateLimitService:          nil,
				Policy: &sesame_api_v1alpha1.PolicyConfig{
					RequestHeadersPolicy:  &sesame_api_v1alpha1.HeadersPolicy{},
					ResponseHeadersPolicy: &sesame_api_v1alpha1.HeadersPolicy{},
					ApplyToIngress:        false,
				},
				Metrics: sesame_api_v1alpha1.MetricsConfig{
					Address: "0.0.0.0",
					Port:    8000,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			converted := tc.serveContext.convertToSesameConfigurationSpec()
			assert.Equal(t, tc.SesameConfig, converted)
		})
	}
}
