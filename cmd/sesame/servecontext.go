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
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/projectsesame/sesame/internal/k8s"
	"k8s.io/utils/pointer"

	sesame_api_v1alpha1 "github.com/projectsesame/sesame/apis/projectsesame/v1alpha1"
	envoy_v3 "github.com/projectsesame/sesame/internal/envoy/v3"
	xdscache_v3 "github.com/projectsesame/sesame/internal/xdscache/v3"
	"github.com/projectsesame/sesame/pkg/config"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

type serveContext struct {
	// Name of the SesameConfiguration CRD to use for configuration.
	sesameConfigurationName string

	Config config.Parameters

	ServerConfig

	// Enable Kubernetes client-go debugging.
	KubernetesDebug uint

	// sesame's debug handler parameters
	debugAddr string
	debugPort int

	// sesame's metrics handler parameters
	metricsAddr string
	metricsPort int

	// Sesame's health handler parameters.
	healthAddr string
	healthPort int

	// httpproxy root namespaces
	rootNamespaces string

	// ingress class
	ingressClassName string

	// envoy's stats listener parameters
	statsAddr string
	statsPort int

	// envoy's listener parameters
	useProxyProto bool

	// envoy's http listener parameters
	httpAddr      string
	httpPort      int
	httpAccessLog string

	// envoy's https listener parameters
	httpsAddr      string
	httpsPort      int
	httpsAccessLog string

	// PermitInsecureGRPC disables TLS on Sesame's gRPC listener.
	PermitInsecureGRPC bool

	// DisableLeaderElection can only be set by command line flag.
	DisableLeaderElection bool
}

type ServerConfig struct {
	// sesame's xds service parameters
	xdsAddr                       string
	xdsPort                       int
	caFile, SesameCert, SesameKey string
}

// newServeContext returns a serveContext initialized to defaults.
func newServeContext() *serveContext {
	// Set defaults for parameters which are then overridden via flags, ENV, or ConfigFile
	return &serveContext{
		Config:                config.Defaults(),
		statsAddr:             "0.0.0.0",
		statsPort:             8002,
		debugAddr:             "127.0.0.1",
		debugPort:             6060,
		healthAddr:            "0.0.0.0",
		healthPort:            8000,
		metricsAddr:           "0.0.0.0",
		metricsPort:           8000,
		httpAccessLog:         xdscache_v3.DEFAULT_HTTP_ACCESS_LOG,
		httpsAccessLog:        xdscache_v3.DEFAULT_HTTPS_ACCESS_LOG,
		httpAddr:              "0.0.0.0",
		httpsAddr:             "0.0.0.0",
		httpPort:              8080,
		httpsPort:             8443,
		PermitInsecureGRPC:    false,
		DisableLeaderElection: false,
		ServerConfig: ServerConfig{
			xdsAddr:    "127.0.0.1",
			xdsPort:    8001,
			caFile:     "",
			SesameCert: "",
			SesameKey:  "",
		},
	}
}

// grpcOptions returns a slice of grpc.ServerOptions.
// if ctx.PermitInsecureGRPC is false, the option set will
// include TLS configuration.
func grpcOptions(log logrus.FieldLogger, SesameXDSConfig *sesame_api_v1alpha1.TLS) []grpc.ServerOption {
	opts := []grpc.ServerOption{
		// By default the Go grpc library defaults to a value of ~100 streams per
		// connection. This number is likely derived from the HTTP/2 spec:
		// https://http2.github.io/http2-spec/#SettingValues
		// We need to raise this value because Envoy will open one EDS stream per
		// CDS entry. There doesn't seem to be a penalty for increasing this value,
		// so set it the limit similar to envoyproxy/go-control-plane#70.
		//
		// Somewhat arbitrary limit to handle many, many, EDS streams.
		grpc.MaxConcurrentStreams(1 << 20),
		// Set gRPC keepalive params.
		// See https://github.com/projectsesame/sesame/issues/1756 for background.
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			PermitWithoutStream: true,
		}),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    60 * time.Second,
			Timeout: 20 * time.Second,
		}),
	}
	if SesameXDSConfig != nil && !SesameXDSConfig.Insecure {
		tlsconfig := tlsconfig(log, SesameXDSConfig)
		creds := credentials.NewTLS(tlsconfig)
		opts = append(opts, grpc.Creds(creds))
	}
	return opts
}

// tlsconfig returns a new *tls.Config. If the TLS parameters passed are not properly configured
// for tls communication, tlsconfig returns nil.
func tlsconfig(log logrus.FieldLogger, SesameXDSTLS *sesame_api_v1alpha1.TLS) *tls.Config {
	err := verifyTLSFlags(SesameXDSTLS)
	if err != nil {
		log.WithError(err).Fatal("failed to verify TLS flags")
	}

	// Define a closure that lazily loads certificates and key at TLS handshake
	// to ensure that latest certificates are used in case they have been rotated.
	loadConfig := func() (*tls.Config, error) {
		if SesameXDSTLS == nil {
			return nil, nil
		}
		cert, err := tls.LoadX509KeyPair(SesameXDSTLS.CertFile, SesameXDSTLS.KeyFile)
		if err != nil {
			return nil, err
		}

		ca, err := ioutil.ReadFile(SesameXDSTLS.CAFile)
		if err != nil {
			return nil, err
		}

		certPool := x509.NewCertPool()
		if ok := certPool.AppendCertsFromPEM(ca); !ok {
			return nil, fmt.Errorf("unable to append certificate in %s to CA pool", SesameXDSTLS.CAFile)
		}

		return &tls.Config{
			Certificates: []tls.Certificate{cert},
			ClientAuth:   tls.RequireAndVerifyClientCert,
			ClientCAs:    certPool,
			MinVersion:   tls.VersionTLS13,
		}, nil
	}

	// Attempt to load certificates and key to catch configuration errors early.
	if _, lerr := loadConfig(); lerr != nil {
		log.WithError(lerr).Fatal("failed to load certificate and key")
	}

	return &tls.Config{
		MinVersion: tls.VersionTLS13,
		ClientAuth: tls.RequireAndVerifyClientCert,
		Rand:       rand.Reader,
		GetConfigForClient: func(*tls.ClientHelloInfo) (*tls.Config, error) {
			return loadConfig()
		},
	}
}

// verifyTLSFlags indicates if the TLS flags are set up correctly.
func verifyTLSFlags(SesameXDSTLS *sesame_api_v1alpha1.TLS) error {
	if SesameXDSTLS.CAFile == "" && SesameXDSTLS.CertFile == "" && SesameXDSTLS.KeyFile == "" {
		return errors.New("no TLS parameters and --insecure not supplied. You must supply one or the other")
	}
	// If one of the three TLS commands is not empty, they all must be not empty
	if !(SesameXDSTLS.CAFile != "" && SesameXDSTLS.CertFile != "" && SesameXDSTLS.KeyFile != "") {
		return errors.New("you must supply all three TLS parameters - --sesame-cafile, --sesame-cert-file, --sesame-key-file, or none of them")
	}

	return nil
}

// proxyRootNamespaces returns a slice of namespaces restricting where
// sesame should look for httpproxy roots.
func (ctx *serveContext) proxyRootNamespaces() []string {
	if strings.TrimSpace(ctx.rootNamespaces) == "" {
		return nil
	}
	var ns []string
	for _, s := range strings.Split(ctx.rootNamespaces, ",") {
		ns = append(ns, strings.TrimSpace(s))
	}
	return ns
}

// parseDefaultHTTPVersions parses a list of supported HTTP versions
//  (of the form "HTTP/xx") into a slice of unique version constants.
func parseDefaultHTTPVersions(versions []sesame_api_v1alpha1.HTTPVersionType) []envoy_v3.HTTPVersionType {
	wanted := map[envoy_v3.HTTPVersionType]struct{}{}

	for _, v := range versions {
		switch v {
		case sesame_api_v1alpha1.HTTPVersion1:
			wanted[envoy_v3.HTTPVersion1] = struct{}{}
		case sesame_api_v1alpha1.HTTPVersion2:
			wanted[envoy_v3.HTTPVersion2] = struct{}{}
		}
	}

	var parsed []envoy_v3.HTTPVersionType
	for k := range wanted {
		parsed = append(parsed, k)
	}

	return parsed
}

func (ctx *serveContext) convertToSesameConfigurationSpec() sesame_api_v1alpha1.SesameConfigurationSpec {
	ingress := &sesame_api_v1alpha1.IngressConfig{}
	if len(ctx.ingressClassName) > 0 {
		ingress.ClassName = pointer.StringPtr(ctx.ingressClassName)
	}
	if len(ctx.Config.IngressStatusAddress) > 0 {
		ingress.StatusAddress = pointer.StringPtr(ctx.Config.IngressStatusAddress)
	}

	debugLogLevel := sesame_api_v1alpha1.InfoLog
	switch ctx.Config.Debug {
	case true:
		debugLogLevel = sesame_api_v1alpha1.DebugLog
	case false:
		debugLogLevel = sesame_api_v1alpha1.InfoLog
	}

	var gatewayConfig *sesame_api_v1alpha1.GatewayConfig
	if ctx.Config.GatewayConfig != nil {
		gatewayConfig = &sesame_api_v1alpha1.GatewayConfig{
			ControllerName: ctx.Config.GatewayConfig.ControllerName,
		}
	}

	var cipherSuites []sesame_api_v1alpha1.TLSCipherType
	for _, suite := range ctx.Config.TLS.CipherSuites {
		cipherSuites = append(cipherSuites, sesame_api_v1alpha1.TLSCipherType(suite))
	}

	var accessLogFormat sesame_api_v1alpha1.AccessLogType
	switch ctx.Config.AccessLogFormat {
	case config.EnvoyAccessLog:
		accessLogFormat = sesame_api_v1alpha1.EnvoyAccessLog
	case config.JSONAccessLog:
		accessLogFormat = sesame_api_v1alpha1.JSONAccessLog
	}

	var accessLogFields sesame_api_v1alpha1.AccessLogFields
	for _, alf := range ctx.Config.AccessLogFields {
		accessLogFields = append(accessLogFields, alf)
	}

	var defaultHTTPVersions []sesame_api_v1alpha1.HTTPVersionType
	for _, version := range ctx.Config.DefaultHTTPVersions {
		switch version {
		case config.HTTPVersion1:
			defaultHTTPVersions = append(defaultHTTPVersions, sesame_api_v1alpha1.HTTPVersion1)
		case config.HTTPVersion2:
			defaultHTTPVersions = append(defaultHTTPVersions, sesame_api_v1alpha1.HTTPVersion2)
		}
	}

	timeoutParams := &sesame_api_v1alpha1.TimeoutParameters{}
	if len(ctx.Config.Timeouts.RequestTimeout) > 0 {
		timeoutParams.RequestTimeout = pointer.StringPtr(ctx.Config.Timeouts.RequestTimeout)
	}
	if len(ctx.Config.Timeouts.ConnectionIdleTimeout) > 0 {
		timeoutParams.ConnectionIdleTimeout = pointer.StringPtr(ctx.Config.Timeouts.ConnectionIdleTimeout)
	}
	if len(ctx.Config.Timeouts.StreamIdleTimeout) > 0 {
		timeoutParams.StreamIdleTimeout = pointer.StringPtr(ctx.Config.Timeouts.StreamIdleTimeout)
	}
	if len(ctx.Config.Timeouts.MaxConnectionDuration) > 0 {
		timeoutParams.MaxConnectionDuration = pointer.StringPtr(ctx.Config.Timeouts.MaxConnectionDuration)
	}
	if len(ctx.Config.Timeouts.DelayedCloseTimeout) > 0 {
		timeoutParams.DelayedCloseTimeout = pointer.StringPtr(ctx.Config.Timeouts.DelayedCloseTimeout)
	}
	if len(ctx.Config.Timeouts.ConnectionShutdownGracePeriod) > 0 {
		timeoutParams.ConnectionShutdownGracePeriod = pointer.StringPtr(ctx.Config.Timeouts.ConnectionShutdownGracePeriod)
	}

	var dnsLookupFamily sesame_api_v1alpha1.ClusterDNSFamilyType
	switch ctx.Config.Cluster.DNSLookupFamily {
	case config.AutoClusterDNSFamily:
		dnsLookupFamily = sesame_api_v1alpha1.AutoClusterDNSFamily
	case config.IPv6ClusterDNSFamily:
		dnsLookupFamily = sesame_api_v1alpha1.IPv6ClusterDNSFamily
	case config.IPv4ClusterDNSFamily:
		dnsLookupFamily = sesame_api_v1alpha1.IPv4ClusterDNSFamily
	}

	var rateLimitService *sesame_api_v1alpha1.RateLimitServiceConfig
	if ctx.Config.RateLimitService.ExtensionService != "" {
		rateLimitService = &sesame_api_v1alpha1.RateLimitServiceConfig{
			ExtensionService: sesame_api_v1alpha1.NamespacedName{
				Name:      k8s.NamespacedNameFrom(ctx.Config.RateLimitService.ExtensionService).Name,
				Namespace: k8s.NamespacedNameFrom(ctx.Config.RateLimitService.ExtensionService).Namespace,
			},
			Domain:                  ctx.Config.RateLimitService.Domain,
			FailOpen:                ctx.Config.RateLimitService.FailOpen,
			EnableXRateLimitHeaders: ctx.Config.RateLimitService.EnableXRateLimitHeaders,
		}
	}

	policy := &sesame_api_v1alpha1.PolicyConfig{
		RequestHeadersPolicy: &sesame_api_v1alpha1.HeadersPolicy{
			Set:    ctx.Config.Policy.RequestHeadersPolicy.Set,
			Remove: ctx.Config.Policy.RequestHeadersPolicy.Remove,
		},
		ResponseHeadersPolicy: &sesame_api_v1alpha1.HeadersPolicy{
			Set:    ctx.Config.Policy.ResponseHeadersPolicy.Set,
			Remove: ctx.Config.Policy.ResponseHeadersPolicy.Remove,
		},
		ApplyToIngress: ctx.Config.Policy.ApplyToIngress,
	}

	var clientCertificate *sesame_api_v1alpha1.NamespacedName
	if len(ctx.Config.TLS.ClientCertificate.Name) > 0 {
		clientCertificate = &sesame_api_v1alpha1.NamespacedName{
			Name:      ctx.Config.TLS.ClientCertificate.Name,
			Namespace: ctx.Config.TLS.ClientCertificate.Namespace,
		}
	}

	var accessLogFormatString *string
	if len(ctx.Config.AccessLogFormatString) > 0 {
		accessLogFormatString = pointer.StringPtr(ctx.Config.AccessLogFormatString)
	}

	var fallbackCertificate *sesame_api_v1alpha1.NamespacedName
	if len(ctx.Config.TLS.FallbackCertificate.Name) > 0 {
		fallbackCertificate = &sesame_api_v1alpha1.NamespacedName{
			Name:      ctx.Config.TLS.FallbackCertificate.Name,
			Namespace: ctx.Config.TLS.FallbackCertificate.Namespace,
		}
	}

	SesameMetrics := sesame_api_v1alpha1.MetricsConfig{
		Address: ctx.metricsAddr,
		Port:    ctx.metricsPort,
	}

	envoyMetrics := sesame_api_v1alpha1.MetricsConfig{
		Address: ctx.statsAddr,
		Port:    ctx.statsPort,
	}

	// Override metrics endpoint info from config files
	//
	// Note!
	// Parameters from command line should take precedence over config file,
	// but here we cannot know anymore if value in ctx.nnn are defaults from
	// newServeContext() or from command line arguments. Therefore metrics
	// configuration from config file takes precedence over command line.
	setMetricsFromConfig(ctx.Config.Metrics.Sesame, &SesameMetrics)
	setMetricsFromConfig(ctx.Config.Metrics.Envoy, &envoyMetrics)

	// Convert serveContext to a SesameConfiguration
	SesameConfiguration := sesame_api_v1alpha1.SesameConfigurationSpec{
		Ingress: ingress,
		Debug: sesame_api_v1alpha1.DebugConfig{
			Address:                 ctx.debugAddr,
			Port:                    ctx.debugPort,
			DebugLogLevel:           debugLogLevel,
			KubernetesDebugLogLevel: ctx.KubernetesDebug,
		},
		Health: sesame_api_v1alpha1.HealthConfig{
			Address: ctx.healthAddr,
			Port:    ctx.healthPort,
		},
		Envoy: sesame_api_v1alpha1.EnvoyConfig{
			Listener: sesame_api_v1alpha1.EnvoyListenerConfig{
				UseProxyProto:             ctx.useProxyProto,
				DisableAllowChunkedLength: ctx.Config.DisableAllowChunkedLength,
				ConnectionBalancer:        ctx.Config.Listener.ConnectionBalancer,
				TLS: sesame_api_v1alpha1.EnvoyTLS{
					MinimumProtocolVersion: ctx.Config.TLS.MinimumProtocolVersion,
					CipherSuites:           cipherSuites,
				},
			},
			Service: sesame_api_v1alpha1.NamespacedName{
				Name:      ctx.Config.EnvoyServiceName,
				Namespace: ctx.Config.EnvoyServiceNamespace,
			},
			HTTPListener: sesame_api_v1alpha1.EnvoyListener{
				Address:   ctx.httpAddr,
				Port:      ctx.httpPort,
				AccessLog: ctx.httpAccessLog,
			},
			HTTPSListener: sesame_api_v1alpha1.EnvoyListener{
				Address:   ctx.httpsAddr,
				Port:      ctx.httpsPort,
				AccessLog: ctx.httpsAccessLog,
			},
			Metrics: envoyMetrics,
			Health: sesame_api_v1alpha1.HealthConfig{
				Address: ctx.statsAddr,
				Port:    ctx.statsPort,
			},
			ClientCertificate: clientCertificate,
			Logging: sesame_api_v1alpha1.EnvoyLogging{
				AccessLogFormat:       accessLogFormat,
				AccessLogFormatString: accessLogFormatString,
				AccessLogFields:       accessLogFields,
			},
			DefaultHTTPVersions: defaultHTTPVersions,
			Timeouts:            timeoutParams,
			Cluster: sesame_api_v1alpha1.ClusterParameters{
				DNSLookupFamily: dnsLookupFamily,
			},
			Network: sesame_api_v1alpha1.NetworkParameters{
				XffNumTrustedHops: ctx.Config.Network.XffNumTrustedHops,
				EnvoyAdminPort:    ctx.Config.Network.EnvoyAdminPort,
			},
		},
		Gateway: gatewayConfig,
		HTTPProxy: sesame_api_v1alpha1.HTTPProxyConfig{
			DisablePermitInsecure: ctx.Config.DisablePermitInsecure,
			RootNamespaces:        ctx.proxyRootNamespaces(),
			FallbackCertificate:   fallbackCertificate,
		},
		EnableExternalNameService: ctx.Config.EnableExternalNameService,
		RateLimitService:          rateLimitService,
		Policy:                    policy,
		Metrics:                   SesameMetrics,
	}

	xdsServerType := sesame_api_v1alpha1.SesameServerType
	if ctx.Config.Server.XDSServerType == config.EnvoyServerType {
		xdsServerType = sesame_api_v1alpha1.EnvoyServerType
	}

	SesameConfiguration.XDSServer = sesame_api_v1alpha1.XDSServerConfig{
		Type:    xdsServerType,
		Address: ctx.xdsAddr,
		Port:    ctx.xdsPort,
		TLS: &sesame_api_v1alpha1.TLS{
			CAFile:   ctx.caFile,
			CertFile: ctx.SesameCert,
			KeyFile:  ctx.SesameKey,
			Insecure: ctx.PermitInsecureGRPC,
		},
	}

	return SesameConfiguration
}

func setMetricsFromConfig(src config.MetricsServerParameters, dst *sesame_api_v1alpha1.MetricsConfig) {
	if len(src.Address) > 0 {
		dst.Address = src.Address
	}

	if src.Port > 0 {
		dst.Port = src.Port
	}

	if src.HasTLS() {
		dst.TLS = &sesame_api_v1alpha1.MetricsTLS{
			CertFile: src.ServerCert,
			KeyFile:  src.ServerKey,
			CAFile:   src.CABundle,
		}
	}

	if src.HasTLS() {
		dst.TLS = &sesame_api_v1alpha1.MetricsTLS{
			CertFile: src.ServerCert,
			KeyFile:  src.ServerKey,
			CAFile:   src.CABundle,
		}
	}
}
