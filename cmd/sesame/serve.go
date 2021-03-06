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
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	envoy_server_v3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	sesame_api_v1 "github.com/projectsesame/sesame/apis/projectsesame/v1"
	sesame_api_v1alpha1 "github.com/projectsesame/sesame/apis/projectsesame/v1alpha1"
	"github.com/projectsesame/sesame/internal/annotation"
	"github.com/projectsesame/sesame/internal/controller"
	"github.com/projectsesame/sesame/internal/dag"
	"github.com/projectsesame/sesame/internal/debug"
	envoy_v3 "github.com/projectsesame/sesame/internal/envoy/v3"
	"github.com/projectsesame/sesame/internal/health"
	"github.com/projectsesame/sesame/internal/httpsvc"
	"github.com/projectsesame/sesame/internal/k8s"
	"github.com/projectsesame/sesame/internal/leadership"
	"github.com/projectsesame/sesame/internal/metrics"
	"github.com/projectsesame/sesame/internal/sesame"
	"github.com/projectsesame/sesame/internal/sesameconfig"
	"github.com/projectsesame/sesame/internal/timeout"
	"github.com/projectsesame/sesame/internal/xds"
	sesame_xds_v3 "github.com/projectsesame/sesame/internal/xds/v3"
	"github.com/projectsesame/sesame/internal/xdscache"
	xdscache_v3 "github.com/projectsesame/sesame/internal/xdscache/v3"
	"github.com/projectsesame/sesame/pkg/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	corev1 "k8s.io/api/core/v1"
	networking_v1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	ctrl_cache "sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	gatewayapi_v1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

// registerServe registers the serve subcommand and flags
// with the Application provided.
func registerServe(app *kingpin.Application) (*kingpin.CmdClause, *serveContext) {
	serve := app.Command("serve", "Serve xDS API traffic.")

	// The precedence of configuration for sesame serve is as follows:
	// If SesameConfiguration resource is specified, it takes precedence,
	// otherwise config file, overridden by env vars, overridden by cli flags.
	// however, as -c is a cli flag, we don't know its value til cli flags
	// have been parsed. To correct this ordering we assign a post parse
	// action to -c, then parse cli flags twice (see main.main). On the second
	// parse our action will return early, resulting in the precedence order
	// we want.
	var (
		configFile string
		parsed     bool
	)
	ctx := newServeContext()

	parseConfig := func(_ *kingpin.ParseContext) error {

		if ctx.sesameConfigurationName != "" && configFile != "" {
			return fmt.Errorf("cannot specify both %s and %s", "--sesame-config", "-c/--config-path")
		}

		if parsed || configFile == "" {
			// if there is no config file supplied, or we've
			// already parsed it, return immediately.
			return nil
		}

		f, err := os.Open(configFile)
		if err != nil {
			return err
		}
		defer f.Close()

		params, err := config.Parse(f)
		if err != nil {
			return err
		}

		if err := params.Validate(); err != nil {
			return fmt.Errorf("invalid Sesame configuration: %w", err)
		}

		parsed = true

		ctx.Config = *params

		return nil
	}

	serve.Flag("config-path", "Path to base configuration.").Short('c').PlaceHolder("/path/to/file").Action(parseConfig).ExistingFileVar(&configFile)
	serve.Flag("sesame-config-name", "Name of SesameConfiguration CRD.").PlaceHolder("sesame").Action(parseConfig).StringVar(&ctx.sesameConfigurationName)

	serve.Flag("incluster", "Use in cluster configuration.").BoolVar(&ctx.Config.InCluster)
	serve.Flag("kubeconfig", "Path to kubeconfig (if not in running inside a cluster).").PlaceHolder("/path/to/file").StringVar(&ctx.Config.Kubeconfig)

	serve.Flag("disable-leader-election", "Disable leader election mechanism.").BoolVar(&ctx.DisableLeaderElection)
	serve.Flag("leader-election-lease-duration", "The duration of the leadership lease.").Default("15s").DurationVar(&ctx.Config.LeaderElection.LeaseDuration)
	serve.Flag("leader-election-renew-deadline", "The duration leader will retry refreshing leadership before giving up.").Default("10s").DurationVar(&ctx.Config.LeaderElection.RenewDeadline)
	serve.Flag("leader-election-retry-period", "The interval which Sesame will attempt to acquire leadership lease.").Default("2s").DurationVar(&ctx.Config.LeaderElection.RetryPeriod)
	serve.Flag("leader-election-resource-name", "The name of the resource (ConfigMap) leader election will lease.").Default("leader-elect").StringVar(&ctx.Config.LeaderElection.Name)
	serve.Flag("leader-election-resource-namespace", "The namespace of the resource (ConfigMap) leader election will lease.").Default(ctx.Config.LeaderElection.Namespace).StringVar(&ctx.Config.LeaderElection.Namespace)

	serve.Flag("xds-address", "xDS gRPC API address.").PlaceHolder("<ipaddr>").StringVar(&ctx.xdsAddr)
	serve.Flag("xds-port", "xDS gRPC API port.").PlaceHolder("<port>").IntVar(&ctx.xdsPort)

	serve.Flag("stats-address", "Envoy /stats interface address.").PlaceHolder("<ipaddr>").StringVar(&ctx.statsAddr)
	serve.Flag("stats-port", "Envoy /stats interface port.").PlaceHolder("<port>").IntVar(&ctx.statsPort)

	serve.Flag("debug-http-address", "Address the debug http endpoint will bind to.").PlaceHolder("<ipaddr>").StringVar(&ctx.debugAddr)
	serve.Flag("debug-http-port", "Port the debug http endpoint will bind to.").PlaceHolder("<port>").IntVar(&ctx.debugPort)

	serve.Flag("http-address", "Address the metrics HTTP endpoint will bind to.").PlaceHolder("<ipaddr>").StringVar(&ctx.metricsAddr)
	serve.Flag("http-port", "Port the metrics HTTP endpoint will bind to.").PlaceHolder("<port>").IntVar(&ctx.metricsPort)
	serve.Flag("health-address", "Address the health HTTP endpoint will bind to.").PlaceHolder("<ipaddr>").StringVar(&ctx.healthAddr)
	serve.Flag("health-port", "Port the health HTTP endpoint will bind to.").PlaceHolder("<port>").IntVar(&ctx.healthPort)

	serve.Flag("sesame-cafile", "CA bundle file name for serving gRPC with TLS.").Envar("SESAME_CAFILE").StringVar(&ctx.caFile)
	serve.Flag("sesame-cert-file", "Sesame certificate file name for serving gRPC over TLS.").PlaceHolder("/path/to/file").Envar("Sesame_CERT_FILE").StringVar(&ctx.SesameCert)
	serve.Flag("sesame-key-file", "Sesame key file name for serving gRPC over TLS.").PlaceHolder("/path/to/file").Envar("Sesame_KEY_FILE").StringVar(&ctx.SesameKey)
	serve.Flag("insecure", "Allow serving without TLS secured gRPC.").BoolVar(&ctx.PermitInsecureGRPC)
	serve.Flag("root-namespaces", "Restrict sesame to searching these namespaces for root ingress routes.").PlaceHolder("<ns,ns>").StringVar(&ctx.rootNamespaces)

	serve.Flag("ingress-class-name", "Sesame IngressClass name.").PlaceHolder("<name>").StringVar(&ctx.ingressClassName)
	serve.Flag("ingress-status-address", "Address to set in Ingress object status.").PlaceHolder("<address>").StringVar(&ctx.Config.IngressStatusAddress)
	serve.Flag("envoy-http-access-log", "Envoy HTTP access log.").PlaceHolder("/path/to/file").StringVar(&ctx.httpAccessLog)
	serve.Flag("envoy-https-access-log", "Envoy HTTPS access log.").PlaceHolder("/path/to/file").StringVar(&ctx.httpsAccessLog)
	serve.Flag("envoy-service-http-address", "Kubernetes Service address for HTTP requests.").PlaceHolder("<ipaddr>").StringVar(&ctx.httpAddr)
	serve.Flag("envoy-service-https-address", "Kubernetes Service address for HTTPS requests.").PlaceHolder("<ipaddr>").StringVar(&ctx.httpsAddr)
	serve.Flag("envoy-service-http-port", "Kubernetes Service port for HTTP requests.").PlaceHolder("<port>").IntVar(&ctx.httpPort)
	serve.Flag("envoy-service-https-port", "Kubernetes Service port for HTTPS requests.").PlaceHolder("<port>").IntVar(&ctx.httpsPort)
	serve.Flag("envoy-service-name", "Name of the Envoy service to inspect for Ingress status details.").PlaceHolder("<name>").StringVar(&ctx.Config.EnvoyServiceName)
	serve.Flag("envoy-service-namespace", "Envoy Service Namespace.").PlaceHolder("<namespace>").StringVar(&ctx.Config.EnvoyServiceNamespace)
	serve.Flag("use-proxy-protocol", "Use PROXY protocol for all listeners.").BoolVar(&ctx.useProxyProto)

	serve.Flag("accesslog-format", "Format for Envoy access logs.").PlaceHolder("<envoy|json>").StringVar((*string)(&ctx.Config.AccessLogFormat))

	serve.Flag("debug", "Enable debug logging.").Short('d').BoolVar(&ctx.Config.Debug)
	serve.Flag("kubernetes-debug", "Enable Kubernetes client debug logging with log level.").PlaceHolder("<log level>").UintVar(&ctx.KubernetesDebug)
	return serve, ctx
}

type Server struct {
	log        logrus.FieldLogger
	ctx        *serveContext
	coreClient *kubernetes.Clientset
	mgr        manager.Manager
	registry   *prometheus.Registry
}

// NewServer returns a Server object which contains the initial configuration
// objects required to start an instance of Sesame.
func NewServer(log logrus.FieldLogger, ctx *serveContext) (*Server, error) {

	// Establish k8s core client connection.
	restConfig, err := k8s.NewRestConfig(ctx.Config.Kubeconfig, ctx.Config.InCluster)
	if err != nil {
		return nil, fmt.Errorf("failed to create REST config for Kubernetes clients: %w", err)
	}

	coreClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes clients: %w", err)
	}

	scheme, err := k8s.NewSesameScheme()
	if err != nil {
		return nil, fmt.Errorf("unable to create scheme: %w", err)
	}

	// Instantiate a controller-runtime manager.
	options := manager.Options{
		Scheme: scheme,
	}
	if ctx.DisableLeaderElection {
		log.Info("Leader election disabled")
		options.LeaderElection = false
	} else {
		options.LeaderElection = true
		// This represents a multilock on configmaps and leases.
		// TODO: switch to solely "leases" once a release cycle has passed.
		options.LeaderElectionResourceLock = "configmapsleases"
		options.LeaderElectionNamespace = ctx.Config.LeaderElection.Namespace
		options.LeaderElectionID = ctx.Config.LeaderElection.Name
		options.LeaseDuration = &ctx.Config.LeaderElection.LeaseDuration
		options.RenewDeadline = &ctx.Config.LeaderElection.RenewDeadline
		options.RetryPeriod = &ctx.Config.LeaderElection.RetryPeriod
		options.LeaderElectionReleaseOnCancel = true
	}
	mgr, err := manager.New(restConfig, options)
	if err != nil {
		return nil, fmt.Errorf("unable to set up controller manager: %w", err)
	}

	// Set up Prometheus registry and register base metrics.
	registry := prometheus.NewRegistry()
	registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	registry.MustRegister(collectors.NewGoCollector())

	return &Server{
		log:        log,
		ctx:        ctx,
		coreClient: coreClient,
		mgr:        mgr,
		registry:   registry,
	}, nil
}

// doServe runs the sesame serve subcommand.
func (s *Server) doServe() error {

	var sesameConfiguration sesame_api_v1alpha1.SesameConfigurationSpec

	// Get the SesameConfiguration CRD if specified
	if len(s.ctx.sesameConfigurationName) > 0 {
		// Determine the name/namespace of the configuration resource utilizing the environment
		// variable "Sesame_NAMESPACE" which should exist on the Sesame deployment.
		//
		// If the env variable is not present, it will default to "projectsesame".
		sesameNamespace, found := os.LookupEnv("SESAME_NAMESPACE")
		if !found {
			sesameNamespace = "projectsesame"
		}

		SesameConfig := &sesame_api_v1alpha1.SesameConfiguration{}
		key := client.ObjectKey{Namespace: sesameNamespace, Name: s.ctx.sesameConfigurationName}

		// Using GetAPIReader() here because the manager's caches won't be started yet,
		// so reads from the manager's client (which uses the caches for reads) will fail.
		if err := s.mgr.GetAPIReader().Get(context.Background(), key, SesameConfig); err != nil {
			return fmt.Errorf("error getting sesame configuration %s: %v", key, err)
		}

		// Copy the Spec from the parsed Configuration
		sesameConfiguration = SesameConfig.Spec
	} else {
		// No sesame configuration passed, so convert the ServeContext into a SesameConfigurationSpec.
		sesameConfiguration = s.ctx.convertToSesameConfigurationSpec()
	}

	err := sesameConfiguration.Validate()
	if err != nil {
		return err
	}

	// informerNamespaces is a list of namespaces that we should start informers for.
	var informerNamespaces []string

	if len(sesameConfiguration.HTTPProxy.RootNamespaces) > 0 {
		informerNamespaces = append(informerNamespaces, sesameConfiguration.HTTPProxy.RootNamespaces...)

		// Add the FallbackCertificateNamespace to informerNamespaces if it isn't present.
		if sesameConfiguration.HTTPProxy.FallbackCertificate != nil && !contains(informerNamespaces, sesameConfiguration.HTTPProxy.FallbackCertificate.Namespace) {
			informerNamespaces = append(informerNamespaces, sesameConfiguration.HTTPProxy.FallbackCertificate.Namespace)
			s.log.WithField("context", "fallback-certificate").
				Infof("fallback certificate namespace %q not defined in 'root-namespaces', adding namespace to watch",
					sesameConfiguration.HTTPProxy.FallbackCertificate.Namespace)
		}

		// Add the client certificate namespace to informerNamespaces if it isn't present.
		if sesameConfiguration.Envoy.ClientCertificate != nil && !contains(informerNamespaces, sesameConfiguration.Envoy.ClientCertificate.Namespace) {
			informerNamespaces = append(informerNamespaces, sesameConfiguration.Envoy.ClientCertificate.Namespace)
			s.log.WithField("context", "envoy-client-certificate").
				Infof("client certificate namespace %q not defined in 'root-namespaces', adding namespace to watch",
					sesameConfiguration.Envoy.ClientCertificate.Namespace)
		}
	}

	cipherSuites := []string{}
	for _, cs := range sesameConfiguration.Envoy.Listener.TLS.CipherSuites {
		cipherSuites = append(cipherSuites, string(cs))
	}

	timeouts, err := sesameconfig.ParseTimeoutPolicy(sesameConfiguration.Envoy.Timeouts)
	if err != nil {
		return err
	}

	accessLogFormatString := ""
	if sesameConfiguration.Envoy.Logging.AccessLogFormatString != nil {
		accessLogFormatString = *sesameConfiguration.Envoy.Logging.AccessLogFormatString
	}

	listenerConfig := xdscache_v3.ListenerConfig{
		UseProxyProto: sesameConfiguration.Envoy.Listener.UseProxyProto,
		HTTPListeners: map[string]xdscache_v3.Listener{
			xdscache_v3.ENVOY_HTTP_LISTENER: {
				Name:    xdscache_v3.ENVOY_HTTP_LISTENER,
				Address: sesameConfiguration.Envoy.HTTPListener.Address,
				Port:    sesameConfiguration.Envoy.HTTPListener.Port,
			},
		},
		HTTPAccessLog: sesameConfiguration.Envoy.HTTPListener.AccessLog,
		HTTPSListeners: map[string]xdscache_v3.Listener{
			xdscache_v3.ENVOY_HTTPS_LISTENER: {
				Name:    xdscache_v3.ENVOY_HTTPS_LISTENER,
				Address: sesameConfiguration.Envoy.HTTPSListener.Address,
				Port:    sesameConfiguration.Envoy.HTTPSListener.Port,
			},
		},
		HTTPSAccessLog:               sesameConfiguration.Envoy.HTTPSListener.AccessLog,
		AccessLogType:                sesameConfiguration.Envoy.Logging.AccessLogFormat,
		AccessLogFields:              sesameConfiguration.Envoy.Logging.AccessLogFields,
		AccessLogFormatString:        accessLogFormatString,
		AccessLogFormatterExtensions: AccessLogFormatterExtensions(sesameConfiguration.Envoy.Logging.AccessLogFormat, sesameConfiguration.Envoy.Logging.AccessLogFields, sesameConfiguration.Envoy.Logging.AccessLogFormatString),
		MinimumTLSVersion:            annotation.MinTLSVersion(sesameConfiguration.Envoy.Listener.TLS.MinimumProtocolVersion, "1.2"),
		CipherSuites:                 config.SanitizeCipherSuites(cipherSuites),
		Timeouts:                     timeouts,
		DefaultHTTPVersions:          parseDefaultHTTPVersions(sesameConfiguration.Envoy.DefaultHTTPVersions),
		AllowChunkedLength:           !sesameConfiguration.Envoy.Listener.DisableAllowChunkedLength,
		XffNumTrustedHops:            sesameConfiguration.Envoy.Network.XffNumTrustedHops,
		ConnectionBalancer:           sesameConfiguration.Envoy.Listener.ConnectionBalancer,
	}

	if listenerConfig.RateLimitConfig, err = s.setupRateLimitService(sesameConfiguration); err != nil {
		return err
	}

	SesameMetrics := metrics.NewMetrics(s.registry)

	// Endpoints updates are handled directly by the EndpointsTranslator
	// due to their high update rate and their orthogonal nature.
	endpointHandler := xdscache_v3.NewEndpointsTranslator(s.log.WithField("context", "endpointstranslator"))

	resources := []xdscache.ResourceCache{
		xdscache_v3.NewListenerCache(sesameConfiguration.Envoy, listenerConfig),
		xdscache_v3.NewSecretsCache(envoy_v3.StatsSecrets(sesameConfiguration.Envoy.Metrics.TLS)),
		&xdscache_v3.RouteCache{},
		&xdscache_v3.ClusterCache{},
		endpointHandler,
	}

	// snapshotHandler is used to produce new snapshots when the internal state changes for any xDS resource.
	snapshotHandler := xdscache.NewSnapshotHandler(resources, s.log.WithField("context", "snapshotHandler"))

	// register observer for endpoints updates.
	endpointHandler.Observer = sesame.ComposeObservers(snapshotHandler)

	// Log that we're using the fallback certificate if configured.
	if sesameConfiguration.HTTPProxy.FallbackCertificate != nil {
		s.log.WithField("context", "fallback-certificate").Infof("enabled fallback certificate with secret: %q", sesameConfiguration.HTTPProxy.FallbackCertificate)
	}
	if sesameConfiguration.Envoy.ClientCertificate != nil {
		s.log.WithField("context", "envoy-client-certificate").Infof("enabled client certificate with secret: %q", sesameConfiguration.Envoy.ClientCertificate)
	}

	ingressClassName := ""
	if sesameConfiguration.Ingress != nil && sesameConfiguration.Ingress.ClassName != nil {
		ingressClassName = *sesameConfiguration.Ingress.ClassName
	}

	var clientCert *types.NamespacedName
	var fallbackCert *types.NamespacedName
	if sesameConfiguration.Envoy.ClientCertificate != nil {
		clientCert = &types.NamespacedName{Name: sesameConfiguration.Envoy.ClientCertificate.Name, Namespace: sesameConfiguration.Envoy.ClientCertificate.Namespace}
	}
	if sesameConfiguration.HTTPProxy.FallbackCertificate != nil {
		fallbackCert = &types.NamespacedName{Name: sesameConfiguration.HTTPProxy.FallbackCertificate.Name, Namespace: sesameConfiguration.HTTPProxy.FallbackCertificate.Namespace}
	}

	sh := k8s.NewStatusUpdateHandler(s.log.WithField("context", "StatusUpdateHandler"), s.mgr.GetClient())
	if err := s.mgr.Add(sh); err != nil {
		return err
	}

	builder := s.getDAGBuilder(dagBuilderConfig{
		ingressClassName:          ingressClassName,
		rootNamespaces:            sesameConfiguration.HTTPProxy.RootNamespaces,
		gatewayAPIConfigured:      sesameConfiguration.Gateway != nil,
		disablePermitInsecure:     sesameConfiguration.HTTPProxy.DisablePermitInsecure,
		enableExternalNameService: sesameConfiguration.EnableExternalNameService,
		dnsLookupFamily:           sesameConfiguration.Envoy.Cluster.DNSLookupFamily,
		headersPolicy:             sesameConfiguration.Policy,
		clientCert:                clientCert,
		fallbackCert:              fallbackCert,
	})

	// Build the core Kubernetes event handler.
	observer := sesame.NewRebuildMetricsObserver(
		SesameMetrics,
		dag.ComposeObservers(append(xdscache.ObserversOf(resources), snapshotHandler)...),
	)
	SesameHandler := sesame.NewEventHandler(sesame.EventHandlerConfig{
		Logger:          s.log.WithField("context", "SesameEventHandler"),
		HoldoffDelay:    100 * time.Millisecond,
		HoldoffMaxDelay: 500 * time.Millisecond,
		Observer:        observer,
		StatusUpdater:   sh.Writer(),
		Builder:         builder,
	})

	// Wrap SesameHandler in an EventRecorder which tracks API server events.
	eventHandler := &sesame.EventRecorder{
		Next:    SesameHandler,
		Counter: SesameMetrics.EventHandlerOperations,
	}

	// Inform on default resources.
	for name, r := range map[string]client.Object{
		"httpproxies":               &sesame_api_v1.HTTPProxy{},
		"tlscertificatedelegations": &sesame_api_v1.TLSCertificateDelegation{},
		"extensionservices":         &sesame_api_v1alpha1.ExtensionService{},
		"sesameconfigurations":      &sesame_api_v1alpha1.SesameConfiguration{},
		"services":                  &corev1.Service{},
		"ingresses":                 &networking_v1.Ingress{},
		"ingressclasses":            &networking_v1.IngressClass{},
	} {
		if err := informOnResource(r, eventHandler, s.mgr.GetCache()); err != nil {
			s.log.WithError(err).WithField("resource", name).Fatal("failed to create informer")
		}
	}

	// Inform on Gateway API resources.
	needsNotification := s.setupGatewayAPI(sesameConfiguration, s.mgr, eventHandler, sh)

	// Inform on secrets, filtering by root namespaces.
	var handler cache.ResourceEventHandler = eventHandler

	// If root namespaces are defined, filter for secrets in only those namespaces.
	if len(informerNamespaces) > 0 {
		handler = k8s.NewNamespaceFilter(informerNamespaces, eventHandler)
	}

	if err := informOnResource(&corev1.Secret{}, handler, s.mgr.GetCache()); err != nil {
		s.log.WithError(err).WithField("resource", "secrets").Fatal("failed to create informer")
	}

	// Inform on endpoints.
	if err := informOnResource(&corev1.Endpoints{}, &sesame.EventRecorder{
		Next:    endpointHandler,
		Counter: SesameMetrics.EventHandlerOperations,
	}, s.mgr.GetCache()); err != nil {
		s.log.WithError(err).WithField("resource", "endpoints").Fatal("failed to create informer")
	}

	// Register our event handler with the manager.
	if err := s.mgr.Add(SesameHandler); err != nil {
		return err
	}

	// Create metrics service.
	if err := s.setupMetrics(sesameConfiguration.Metrics, sesameConfiguration.Health, s.registry); err != nil {
		return err
	}

	// Create a separate health service if required.
	if err := s.setupHealth(sesameConfiguration.Health, sesameConfiguration.Metrics); err != nil {
		return err
	}

	// Create debug service and register with workgroup.
	if err := s.setupDebugService(sesameConfiguration.Debug, builder); err != nil {
		return err
	}

	var gatewayControllerName string
	if sesameConfiguration.Gateway != nil {
		gatewayControllerName = sesameConfiguration.Gateway.ControllerName
	}

	// Set up ingress load balancer status writer.
	lbsw := &loadBalancerStatusWriter{
		log:                   s.log.WithField("context", "loadBalancerStatusWriter"),
		cache:                 s.mgr.GetCache(),
		lbStatus:              make(chan corev1.LoadBalancerStatus, 1),
		ingressClassName:      ingressClassName,
		gatewayControllerName: gatewayControllerName,
		statusUpdater:         sh.Writer(),
	}
	if err := s.mgr.Add(lbsw); err != nil {
		return err
	}

	// Register an informer to watch envoy's service if we haven't been given static details.
	if sesameConfiguration.Ingress != nil && sesameConfiguration.Ingress.StatusAddress != nil {
		s.log.WithField("loadbalancer-address", *sesameConfiguration.Ingress.StatusAddress).Info("Using supplied information for Ingress status")
		lbsw.lbStatus <- parseStatusFlag(*sesameConfiguration.Ingress.StatusAddress)
	} else {
		serviceHandler := &k8s.ServiceStatusLoadBalancerWatcher{
			ServiceName: sesameConfiguration.Envoy.Service.Name,
			LBStatus:    lbsw.lbStatus,
			Log:         s.log.WithField("context", "serviceStatusLoadBalancerWatcher"),
		}

		var handler cache.ResourceEventHandler = serviceHandler
		if sesameConfiguration.Envoy.Service.Namespace != "" {
			handler = k8s.NewNamespaceFilter([]string{sesameConfiguration.Envoy.Service.Namespace}, handler)
		}

		if err := informOnResource(&corev1.Service{}, handler, s.mgr.GetCache()); err != nil {
			s.log.WithError(err).WithField("resource", "services").Fatal("failed to create informer")
		}

		s.log.WithField("envoy-service-name", sesameConfiguration.Envoy.Service.Name).
			WithField("envoy-service-namespace", sesameConfiguration.Envoy.Service.Namespace).
			Info("Watching Service for Ingress status")
	}

	xdsServer := &xdsServer{
		log:             s.log,
		mgr:             s.mgr,
		registry:        s.registry,
		config:          sesameConfiguration.XDSServer,
		snapshotHandler: snapshotHandler,
		resources:       resources,
	}
	if err := s.mgr.Add(xdsServer); err != nil {
		return err
	}

	notifier := &leadership.Notifier{
		ToNotify: append([]leadership.NeedLeaderElectionNotification{
			SesameHandler,
			observer,
		}, needsNotification...),
	}
	if err := s.mgr.Add(notifier); err != nil {
		return err
	}

	// GO!
	return s.mgr.Start(signals.SetupSignalHandler())
}

func (s *Server) setupRateLimitService(SesameConfiguration sesame_api_v1alpha1.SesameConfigurationSpec) (*xdscache_v3.RateLimitConfig, error) {
	if SesameConfiguration.RateLimitService == nil {
		return nil, nil
	}

	// ensure the specified ExtensionService exists
	extensionSvc := &sesame_api_v1alpha1.ExtensionService{}
	key := client.ObjectKey{
		Namespace: SesameConfiguration.RateLimitService.ExtensionService.Namespace,
		Name:      SesameConfiguration.RateLimitService.ExtensionService.Name,
	}

	// Using GetAPIReader() here because the manager's caches won't be started yet,
	// so reads from the manager's client (which uses the caches for reads) will fail.
	if err := s.mgr.GetAPIReader().Get(context.Background(), key, extensionSvc); err != nil {
		return nil, fmt.Errorf("error getting rate limit extension service %s: %v", key, err)
	}

	// get the response timeout from the ExtensionService
	var responseTimeout timeout.Setting
	var err error

	if tp := extensionSvc.Spec.TimeoutPolicy; tp != nil {
		responseTimeout, err = timeout.Parse(tp.Response)
		if err != nil {
			return nil, fmt.Errorf("error parsing rate limit extension service %s response timeout: %v", key, err)
		}
	}

	return &xdscache_v3.RateLimitConfig{
		ExtensionService:        key,
		Domain:                  SesameConfiguration.RateLimitService.Domain,
		Timeout:                 responseTimeout,
		FailOpen:                SesameConfiguration.RateLimitService.FailOpen,
		EnableXRateLimitHeaders: SesameConfiguration.RateLimitService.EnableXRateLimitHeaders,
	}, nil
}

func (s *Server) setupDebugService(debugConfig sesame_api_v1alpha1.DebugConfig, builder *dag.Builder) error {
	debugsvc := &debug.Service{
		Service: httpsvc.Service{
			Addr:        debugConfig.Address,
			Port:        debugConfig.Port,
			FieldLogger: s.log.WithField("context", "debugsvc"),
		},
		Builder: builder,
	}
	return s.mgr.Add(debugsvc)
}

type xdsServer struct {
	log             logrus.FieldLogger
	mgr             manager.Manager
	registry        *prometheus.Registry
	config          sesame_api_v1alpha1.XDSServerConfig
	snapshotHandler *xdscache.SnapshotHandler
	resources       []xdscache.ResourceCache
}

func (x *xdsServer) NeedLeaderElection() bool {
	return false
}

func (x *xdsServer) Start(ctx context.Context) error {
	log := x.log.WithField("context", "xds")

	log.Printf("waiting for informer caches to sync")
	if !x.mgr.GetCache().WaitForCacheSync(ctx) {
		return errors.New("informer cache failed to sync")
	}
	log.Printf("informer caches synced")

	grpcServer := xds.NewServer(x.registry, grpcOptions(log, x.config.TLS)...)

	switch x.config.Type {
	case sesame_api_v1alpha1.EnvoyServerType:
		v3cache := sesame_xds_v3.NewSnapshotCache(false, log)
		x.snapshotHandler.AddSnapshotter(v3cache)
		sesame_xds_v3.RegisterServer(envoy_server_v3.NewServer(ctx, v3cache, sesame_xds_v3.NewRequestLoggingCallbacks(log)), grpcServer)
	case sesame_api_v1alpha1.SesameServerType:
		sesame_xds_v3.RegisterServer(sesame_xds_v3.NewSesameServer(log, xdscache.ResourcesOf(x.resources)...), grpcServer)
	default:
		// This can't happen due to config validation.
		log.Fatalf("invalid xDS server type %q", x.config.Type)
	}

	addr := net.JoinHostPort(x.config.Address, strconv.Itoa(x.config.Port))
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	log = log.WithField("address", addr)
	if tls := x.config.TLS; tls != nil {
		if tls.Insecure {
			log = log.WithField("insecure", true)
		}
	}

	log.Infof("started xDS server type: %q", x.config.Type)
	defer log.Info("stopped xDS server")

	go func() {
		<-ctx.Done()

		// We don't use GracefulStop here because envoy
		// has long-lived hanging xDS requests. There's no
		// mechanism to make those pending requests fail,
		// so we forcibly terminate the TCP sessions.
		grpcServer.Stop()
	}()

	return grpcServer.Serve(l)
}

// setupMetrics creates metrics service for Sesame.
func (s *Server) setupMetrics(metricsConfig sesame_api_v1alpha1.MetricsConfig, healthConfig sesame_api_v1alpha1.HealthConfig,
	registry *prometheus.Registry) error {

	// Create metrics service and register with workgroup.
	metricsvc := &httpsvc.Service{
		Addr:        metricsConfig.Address,
		Port:        metricsConfig.Port,
		FieldLogger: s.log.WithField("context", "metricsvc"),
		ServeMux:    http.ServeMux{},
	}

	metricsvc.ServeMux.Handle("/metrics", metrics.Handler(registry))

	if metricsConfig.TLS != nil {
		metricsvc.Cert = metricsConfig.TLS.CertFile
		metricsvc.Key = metricsConfig.TLS.KeyFile
		metricsvc.CABundle = metricsConfig.TLS.CAFile
	}

	if healthConfig.Address == metricsConfig.Address && healthConfig.Port == metricsConfig.Port {
		h := health.Handler(s.coreClient)
		metricsvc.ServeMux.Handle("/health", h)
		metricsvc.ServeMux.Handle("/healthz", h)
	}

	return s.mgr.Add(metricsvc)
}

func (s *Server) setupHealth(healthConfig sesame_api_v1alpha1.HealthConfig,
	metricsConfig sesame_api_v1alpha1.MetricsConfig) error {

	if healthConfig.Address != metricsConfig.Address || healthConfig.Port != metricsConfig.Port {
		healthsvc := &httpsvc.Service{
			Addr:        healthConfig.Address,
			Port:        healthConfig.Port,
			FieldLogger: s.log.WithField("context", "healthsvc"),
		}

		h := health.Handler(s.coreClient)
		healthsvc.ServeMux.Handle("/health", h)
		healthsvc.ServeMux.Handle("/healthz", h)

		return s.mgr.Add(healthsvc)
	}

	return nil
}

func (s *Server) setupGatewayAPI(SesameConfiguration sesame_api_v1alpha1.SesameConfigurationSpec,
	mgr manager.Manager, eventHandler *sesame.EventRecorder, sh *k8s.StatusUpdateHandler) []leadership.NeedLeaderElectionNotification {

	needLeadershipNotification := []leadership.NeedLeaderElectionNotification{}

	// Check if GatewayAPI is configured.
	if SesameConfiguration.Gateway != nil {
		// Create and register the gatewayclass controller with the manager.
		gatewayClassControllerName := SesameConfiguration.Gateway.ControllerName
		gwClass, err := controller.RegisterGatewayClassController(
			s.log.WithField("context", "gatewayclass-controller"),
			mgr,
			eventHandler,
			sh.Writer(),
			gatewayClassControllerName,
		)
		if err != nil {
			s.log.WithError(err).Fatal("failed to create gatewayclass-controller")
		}
		needLeadershipNotification = append(needLeadershipNotification, gwClass)

		// Create and register the NewGatewayController controller with the manager.
		gw, err := controller.RegisterGatewayController(
			s.log.WithField("context", "gateway-controller"),
			mgr,
			eventHandler,
			sh.Writer(),
			gatewayClassControllerName,
		)
		if err != nil {
			s.log.WithError(err).Fatal("failed to create gateway-controller")
		}
		needLeadershipNotification = append(needLeadershipNotification, gw)

		// Create and register the HTTPRoute controller with the manager.
		if err := controller.RegisterHTTPRouteController(s.log.WithField("context", "httproute-controller"), mgr, eventHandler); err != nil {
			s.log.WithError(err).Fatal("failed to create httproute-controller")
		}

		// Create and register the TLSRoute controller with the manager.
		if err := controller.RegisterTLSRouteController(s.log.WithField("context", "tlsroute-controller"), mgr, eventHandler); err != nil {
			s.log.WithError(err).Fatal("failed to create tlsroute-controller")
		}

		// Inform on ReferencePolicies.
		if err := informOnResource(&gatewayapi_v1alpha2.ReferencePolicy{}, eventHandler, mgr.GetCache()); err != nil {
			s.log.WithError(err).WithField("resource", "referencepolicies").Fatal("failed to create informer")
		}

		// Inform on Namespaces.
		if err := informOnResource(&corev1.Namespace{}, eventHandler, mgr.GetCache()); err != nil {
			s.log.WithError(err).WithField("resource", "namespaces").Fatal("failed to create informer")
		}
	}
	return needLeadershipNotification
}

type dagBuilderConfig struct {
	ingressClassName           string
	rootNamespaces             []string
	gatewayAPIConfigured       bool
	disablePermitInsecure      bool
	enableExternalNameService  bool
	dnsLookupFamily            sesame_api_v1alpha1.ClusterDNSFamilyType
	headersPolicy              *sesame_api_v1alpha1.PolicyConfig
	applyHeaderPolicyToIngress bool
	clientCert                 *types.NamespacedName
	fallbackCert               *types.NamespacedName
}

func (s *Server) getDAGBuilder(dbc dagBuilderConfig) *dag.Builder {

	var requestHeadersPolicy dag.HeadersPolicy
	var responseHeadersPolicy dag.HeadersPolicy

	if dbc.headersPolicy != nil {
		if dbc.headersPolicy.RequestHeadersPolicy != nil {
			if dbc.headersPolicy.RequestHeadersPolicy.Set != nil {
				requestHeadersPolicy.Set = make(map[string]string)
				for k, v := range dbc.headersPolicy.RequestHeadersPolicy.Set {
					requestHeadersPolicy.Set[k] = v
				}
			}
			if dbc.headersPolicy.RequestHeadersPolicy.Remove != nil {
				requestHeadersPolicy.Remove = make([]string, 0, len(dbc.headersPolicy.RequestHeadersPolicy.Remove))
				requestHeadersPolicy.Remove = append(requestHeadersPolicy.Remove, dbc.headersPolicy.RequestHeadersPolicy.Remove...)
			}
		}

		if dbc.headersPolicy.ResponseHeadersPolicy != nil {
			if dbc.headersPolicy.ResponseHeadersPolicy.Set != nil {
				responseHeadersPolicy.Set = make(map[string]string)
				for k, v := range dbc.headersPolicy.ResponseHeadersPolicy.Set {
					responseHeadersPolicy.Set[k] = v
				}
			}
			if dbc.headersPolicy.ResponseHeadersPolicy.Remove != nil {
				responseHeadersPolicy.Remove = make([]string, 0, len(dbc.headersPolicy.ResponseHeadersPolicy.Remove))
				responseHeadersPolicy.Remove = append(responseHeadersPolicy.Remove, dbc.headersPolicy.ResponseHeadersPolicy.Remove...)
			}
		}
	}

	var requestHeadersPolicyIngress dag.HeadersPolicy
	var responseHeadersPolicyIngress dag.HeadersPolicy
	if dbc.applyHeaderPolicyToIngress {
		requestHeadersPolicyIngress = requestHeadersPolicy
		responseHeadersPolicyIngress = responseHeadersPolicy
	}

	s.log.Debugf("EnableExternalNameService is set to %t", dbc.enableExternalNameService)

	// Get the appropriate DAG processors.
	dagProcessors := []dag.Processor{
		&dag.IngressProcessor{
			EnableExternalNameService: dbc.enableExternalNameService,
			FieldLogger:               s.log.WithField("context", "IngressProcessor"),
			ClientCertificate:         dbc.clientCert,
			RequestHeadersPolicy:      &requestHeadersPolicyIngress,
			ResponseHeadersPolicy:     &responseHeadersPolicyIngress,
		},
		&dag.ExtensionServiceProcessor{
			// Note that ExtensionService does not support ExternalName, if it does get added,
			// need to bring EnableExternalNameService in here too.
			FieldLogger:       s.log.WithField("context", "ExtensionServiceProcessor"),
			ClientCertificate: dbc.clientCert,
		},
		&dag.HTTPProxyProcessor{
			EnableExternalNameService: dbc.enableExternalNameService,
			DisablePermitInsecure:     dbc.disablePermitInsecure,
			FallbackCertificate:       dbc.fallbackCert,
			DNSLookupFamily:           dbc.dnsLookupFamily,
			ClientCertificate:         dbc.clientCert,
			RequestHeadersPolicy:      &requestHeadersPolicy,
			ResponseHeadersPolicy:     &responseHeadersPolicy,
		},
	}

	if dbc.gatewayAPIConfigured {
		dagProcessors = append(dagProcessors, &dag.GatewayAPIProcessor{
			EnableExternalNameService: dbc.enableExternalNameService,
			FieldLogger:               s.log.WithField("context", "GatewayAPIProcessor"),
		})
	}

	// The listener processor has to go last since it looks at
	// the output of the other processors.
	dagProcessors = append(dagProcessors, &dag.ListenerProcessor{})

	var configuredSecretRefs []*types.NamespacedName
	if dbc.fallbackCert != nil {
		configuredSecretRefs = append(configuredSecretRefs, dbc.fallbackCert)
	}
	if dbc.clientCert != nil {
		configuredSecretRefs = append(configuredSecretRefs, dbc.clientCert)
	}

	builder := &dag.Builder{
		Source: dag.KubernetesCache{
			RootNamespaces:       dbc.rootNamespaces,
			IngressClassName:     dbc.ingressClassName,
			ConfiguredSecretRefs: configuredSecretRefs,
			FieldLogger:          s.log.WithField("context", "KubernetesCache"),
		},
		Processors: dagProcessors,
	}

	// govet complains about copying the sync.Once that's in the dag.KubernetesCache
	// but it's safe to ignore since this function is only called once.
	// nolint:govet
	return builder
}

func contains(namespaces []string, ns string) bool {
	for _, namespace := range namespaces {
		if ns == namespace {
			return true
		}
	}
	return false
}

func informOnResource(obj client.Object, handler cache.ResourceEventHandler, cache ctrl_cache.Cache) error {
	inf, err := cache.GetInformer(context.Background(), obj)
	if err != nil {
		return err
	}

	inf.AddEventHandler(handler)
	return nil
}

// commandOperatorRegexp parses the command operators used in Envoy access log configuration
//
// Capture Groups:
// Given string "the start time is %START_TIME(%s):3% wow!"
//
//   0. Whole match "%START_TIME(%s):3%"
//   1. Full operator: "START_TIME(%s):3%"
//   2. Operator Name: "START_TIME"
//   3. Arguments: "(%s)"
//   4. Truncation length: ":3"
var commandOperatorRegexp = regexp.MustCompile(`%(([A-Z_]+)(\([^)]+\)(:[0-9]+)?)?%)?`)

// AccessLogFormatterExtensions returns a list of formatter extension names required by the access log format.
//
// Note: When adding support for new formatter, update the list of extensions here and
// add corresponding configuration in internal/envoy/v3/accesslog.go extensionConfig().
// Currently only one extension exist in Envoy.
func AccessLogFormatterExtensions(accessLogFormat sesame_api_v1alpha1.AccessLogType, accessLogFields sesame_api_v1alpha1.AccessLogFields,
	accessLogFormatString *string) []string {
	// Function that finds out if command operator is present in a format string.
	contains := func(format, command string) bool {
		tokens := commandOperatorRegexp.FindAllStringSubmatch(format, -1)
		for _, t := range tokens {
			if t[2] == command {
				return true
			}
		}
		return false
	}

	extensionsMap := make(map[string]bool)
	switch accessLogFormat {
	case sesame_api_v1alpha1.EnvoyAccessLog:
		if accessLogFormatString != nil {
			if contains(*accessLogFormatString, "REQ_WITHOUT_QUERY") {
				extensionsMap["envoy.formatter.req_without_query"] = true
			}
		}
	case sesame_api_v1alpha1.JSONAccessLog:
		for _, f := range accessLogFields.AsFieldMap() {
			if contains(f, "REQ_WITHOUT_QUERY") {
				extensionsMap["envoy.formatter.req_without_query"] = true
			}
		}
	}

	var extensions []string
	for k := range extensionsMap {
		extensions = append(extensions, k)
	}

	return extensions
}
