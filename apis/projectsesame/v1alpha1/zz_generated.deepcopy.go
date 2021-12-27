// +build !ignore_autogenerated

/*
Copyright Project Contour Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"github.com/projectsesame/sesame/apis/projectsesame/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in AccessLogFields) DeepCopyInto(out *AccessLogFields) {
	{
		in := &in
		*out = make(AccessLogFields, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AccessLogFields.
func (in AccessLogFields) DeepCopy() AccessLogFields {
	if in == nil {
		return nil
	}
	out := new(AccessLogFields)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterParameters) DeepCopyInto(out *ClusterParameters) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterParameters.
func (in *ClusterParameters) DeepCopy() *ClusterParameters {
	if in == nil {
		return nil
	}
	out := new(ClusterParameters)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DebugConfig) DeepCopyInto(out *DebugConfig) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DebugConfig.
func (in *DebugConfig) DeepCopy() *DebugConfig {
	if in == nil {
		return nil
	}
	out := new(DebugConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EnvoyConfig) DeepCopyInto(out *EnvoyConfig) {
	*out = *in
	in.Listener.DeepCopyInto(&out.Listener)
	out.Service = in.Service
	out.HTTPListener = in.HTTPListener
	out.HTTPSListener = in.HTTPSListener
	out.Health = in.Health
	in.Metrics.DeepCopyInto(&out.Metrics)
	if in.ClientCertificate != nil {
		in, out := &in.ClientCertificate, &out.ClientCertificate
		*out = new(NamespacedName)
		**out = **in
	}
	in.Logging.DeepCopyInto(&out.Logging)
	if in.DefaultHTTPVersions != nil {
		in, out := &in.DefaultHTTPVersions, &out.DefaultHTTPVersions
		*out = make([]HTTPVersionType, len(*in))
		copy(*out, *in)
	}
	if in.Timeouts != nil {
		in, out := &in.Timeouts, &out.Timeouts
		*out = new(TimeoutParameters)
		(*in).DeepCopyInto(*out)
	}
	out.Cluster = in.Cluster
	out.Network = in.Network
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EnvoyConfig.
func (in *EnvoyConfig) DeepCopy() *EnvoyConfig {
	if in == nil {
		return nil
	}
	out := new(EnvoyConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EnvoyListener) DeepCopyInto(out *EnvoyListener) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EnvoyListener.
func (in *EnvoyListener) DeepCopy() *EnvoyListener {
	if in == nil {
		return nil
	}
	out := new(EnvoyListener)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EnvoyListenerConfig) DeepCopyInto(out *EnvoyListenerConfig) {
	*out = *in
	in.TLS.DeepCopyInto(&out.TLS)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EnvoyListenerConfig.
func (in *EnvoyListenerConfig) DeepCopy() *EnvoyListenerConfig {
	if in == nil {
		return nil
	}
	out := new(EnvoyListenerConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EnvoyLogging) DeepCopyInto(out *EnvoyLogging) {
	*out = *in
	if in.AccessLogFormatString != nil {
		in, out := &in.AccessLogFormatString, &out.AccessLogFormatString
		*out = new(string)
		**out = **in
	}
	if in.AccessLogFields != nil {
		in, out := &in.AccessLogFields, &out.AccessLogFields
		*out = make(AccessLogFields, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EnvoyLogging.
func (in *EnvoyLogging) DeepCopy() *EnvoyLogging {
	if in == nil {
		return nil
	}
	out := new(EnvoyLogging)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EnvoyTLS) DeepCopyInto(out *EnvoyTLS) {
	*out = *in
	if in.CipherSuites != nil {
		in, out := &in.CipherSuites, &out.CipherSuites
		*out = make([]TLSCipherType, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EnvoyTLS.
func (in *EnvoyTLS) DeepCopy() *EnvoyTLS {
	if in == nil {
		return nil
	}
	out := new(EnvoyTLS)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExtensionService) DeepCopyInto(out *ExtensionService) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExtensionService.
func (in *ExtensionService) DeepCopy() *ExtensionService {
	if in == nil {
		return nil
	}
	out := new(ExtensionService)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ExtensionService) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExtensionServiceList) DeepCopyInto(out *ExtensionServiceList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ExtensionService, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExtensionServiceList.
func (in *ExtensionServiceList) DeepCopy() *ExtensionServiceList {
	if in == nil {
		return nil
	}
	out := new(ExtensionServiceList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ExtensionServiceList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExtensionServiceSpec) DeepCopyInto(out *ExtensionServiceSpec) {
	*out = *in
	if in.Services != nil {
		in, out := &in.Services, &out.Services
		*out = make([]ExtensionServiceTarget, len(*in))
		copy(*out, *in)
	}
	if in.UpstreamValidation != nil {
		in, out := &in.UpstreamValidation, &out.UpstreamValidation
		*out = new(v1.UpstreamValidation)
		**out = **in
	}
	if in.Protocol != nil {
		in, out := &in.Protocol, &out.Protocol
		*out = new(string)
		**out = **in
	}
	if in.LoadBalancerPolicy != nil {
		in, out := &in.LoadBalancerPolicy, &out.LoadBalancerPolicy
		*out = new(v1.LoadBalancerPolicy)
		(*in).DeepCopyInto(*out)
	}
	if in.TimeoutPolicy != nil {
		in, out := &in.TimeoutPolicy, &out.TimeoutPolicy
		*out = new(v1.TimeoutPolicy)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExtensionServiceSpec.
func (in *ExtensionServiceSpec) DeepCopy() *ExtensionServiceSpec {
	if in == nil {
		return nil
	}
	out := new(ExtensionServiceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExtensionServiceStatus) DeepCopyInto(out *ExtensionServiceStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.DetailedCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExtensionServiceStatus.
func (in *ExtensionServiceStatus) DeepCopy() *ExtensionServiceStatus {
	if in == nil {
		return nil
	}
	out := new(ExtensionServiceStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExtensionServiceTarget) DeepCopyInto(out *ExtensionServiceTarget) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExtensionServiceTarget.
func (in *ExtensionServiceTarget) DeepCopy() *ExtensionServiceTarget {
	if in == nil {
		return nil
	}
	out := new(ExtensionServiceTarget)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GatewayConfig) DeepCopyInto(out *GatewayConfig) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GatewayConfig.
func (in *GatewayConfig) DeepCopy() *GatewayConfig {
	if in == nil {
		return nil
	}
	out := new(GatewayConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HTTPProxyConfig) DeepCopyInto(out *HTTPProxyConfig) {
	*out = *in
	if in.RootNamespaces != nil {
		in, out := &in.RootNamespaces, &out.RootNamespaces
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.FallbackCertificate != nil {
		in, out := &in.FallbackCertificate, &out.FallbackCertificate
		*out = new(NamespacedName)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HTTPProxyConfig.
func (in *HTTPProxyConfig) DeepCopy() *HTTPProxyConfig {
	if in == nil {
		return nil
	}
	out := new(HTTPProxyConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HeadersPolicy) DeepCopyInto(out *HeadersPolicy) {
	*out = *in
	if in.Set != nil {
		in, out := &in.Set, &out.Set
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Remove != nil {
		in, out := &in.Remove, &out.Remove
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HeadersPolicy.
func (in *HeadersPolicy) DeepCopy() *HeadersPolicy {
	if in == nil {
		return nil
	}
	out := new(HeadersPolicy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HealthConfig) DeepCopyInto(out *HealthConfig) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HealthConfig.
func (in *HealthConfig) DeepCopy() *HealthConfig {
	if in == nil {
		return nil
	}
	out := new(HealthConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IngressConfig) DeepCopyInto(out *IngressConfig) {
	*out = *in
	if in.ClassName != nil {
		in, out := &in.ClassName, &out.ClassName
		*out = new(string)
		**out = **in
	}
	if in.StatusAddress != nil {
		in, out := &in.StatusAddress, &out.StatusAddress
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IngressConfig.
func (in *IngressConfig) DeepCopy() *IngressConfig {
	if in == nil {
		return nil
	}
	out := new(IngressConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MetricsConfig) DeepCopyInto(out *MetricsConfig) {
	*out = *in
	if in.TLS != nil {
		in, out := &in.TLS, &out.TLS
		*out = new(MetricsTLS)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MetricsConfig.
func (in *MetricsConfig) DeepCopy() *MetricsConfig {
	if in == nil {
		return nil
	}
	out := new(MetricsConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MetricsTLS) DeepCopyInto(out *MetricsTLS) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MetricsTLS.
func (in *MetricsTLS) DeepCopy() *MetricsTLS {
	if in == nil {
		return nil
	}
	out := new(MetricsTLS)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NamespacedName) DeepCopyInto(out *NamespacedName) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NamespacedName.
func (in *NamespacedName) DeepCopy() *NamespacedName {
	if in == nil {
		return nil
	}
	out := new(NamespacedName)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NetworkParameters) DeepCopyInto(out *NetworkParameters) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NetworkParameters.
func (in *NetworkParameters) DeepCopy() *NetworkParameters {
	if in == nil {
		return nil
	}
	out := new(NetworkParameters)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PolicyConfig) DeepCopyInto(out *PolicyConfig) {
	*out = *in
	if in.RequestHeadersPolicy != nil {
		in, out := &in.RequestHeadersPolicy, &out.RequestHeadersPolicy
		*out = new(HeadersPolicy)
		(*in).DeepCopyInto(*out)
	}
	if in.ResponseHeadersPolicy != nil {
		in, out := &in.ResponseHeadersPolicy, &out.ResponseHeadersPolicy
		*out = new(HeadersPolicy)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PolicyConfig.
func (in *PolicyConfig) DeepCopy() *PolicyConfig {
	if in == nil {
		return nil
	}
	out := new(PolicyConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RateLimitServiceConfig) DeepCopyInto(out *RateLimitServiceConfig) {
	*out = *in
	out.ExtensionService = in.ExtensionService
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RateLimitServiceConfig.
func (in *RateLimitServiceConfig) DeepCopy() *RateLimitServiceConfig {
	if in == nil {
		return nil
	}
	out := new(RateLimitServiceConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SesameConfiguration) DeepCopyInto(out *SesameConfiguration) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SesameConfiguration.
func (in *SesameConfiguration) DeepCopy() *SesameConfiguration {
	if in == nil {
		return nil
	}
	out := new(SesameConfiguration)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SesameConfiguration) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SesameConfigurationList) DeepCopyInto(out *SesameConfigurationList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]SesameConfiguration, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SesameConfigurationList.
func (in *SesameConfigurationList) DeepCopy() *SesameConfigurationList {
	if in == nil {
		return nil
	}
	out := new(SesameConfigurationList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SesameConfigurationList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SesameConfigurationSpec) DeepCopyInto(out *SesameConfigurationSpec) {
	*out = *in
	in.XDSServer.DeepCopyInto(&out.XDSServer)
	if in.Ingress != nil {
		in, out := &in.Ingress, &out.Ingress
		*out = new(IngressConfig)
		(*in).DeepCopyInto(*out)
	}
	out.Debug = in.Debug
	out.Health = in.Health
	in.Envoy.DeepCopyInto(&out.Envoy)
	if in.Gateway != nil {
		in, out := &in.Gateway, &out.Gateway
		*out = new(GatewayConfig)
		**out = **in
	}
	in.HTTPProxy.DeepCopyInto(&out.HTTPProxy)
	if in.RateLimitService != nil {
		in, out := &in.RateLimitService, &out.RateLimitService
		*out = new(RateLimitServiceConfig)
		**out = **in
	}
	if in.Policy != nil {
		in, out := &in.Policy, &out.Policy
		*out = new(PolicyConfig)
		(*in).DeepCopyInto(*out)
	}
	in.Metrics.DeepCopyInto(&out.Metrics)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SesameConfigurationSpec.
func (in *SesameConfigurationSpec) DeepCopy() *SesameConfigurationSpec {
	if in == nil {
		return nil
	}
	out := new(SesameConfigurationSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SesameConfigurationStatus) DeepCopyInto(out *SesameConfigurationStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.DetailedCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SesameConfigurationStatus.
func (in *SesameConfigurationStatus) DeepCopy() *SesameConfigurationStatus {
	if in == nil {
		return nil
	}
	out := new(SesameConfigurationStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SesameDeployment) DeepCopyInto(out *SesameDeployment) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SesameDeployment.
func (in *SesameDeployment) DeepCopy() *SesameDeployment {
	if in == nil {
		return nil
	}
	out := new(SesameDeployment)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SesameDeployment) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SesameDeploymentList) DeepCopyInto(out *SesameDeploymentList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]SesameDeployment, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SesameDeploymentList.
func (in *SesameDeploymentList) DeepCopy() *SesameDeploymentList {
	if in == nil {
		return nil
	}
	out := new(SesameDeploymentList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SesameDeploymentList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SesameDeploymentSpec) DeepCopyInto(out *SesameDeploymentSpec) {
	*out = *in
	in.Config.DeepCopyInto(&out.Config)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SesameDeploymentSpec.
func (in *SesameDeploymentSpec) DeepCopy() *SesameDeploymentSpec {
	if in == nil {
		return nil
	}
	out := new(SesameDeploymentSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SesameDeploymentStatus) DeepCopyInto(out *SesameDeploymentStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.DetailedCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SesameDeploymentStatus.
func (in *SesameDeploymentStatus) DeepCopy() *SesameDeploymentStatus {
	if in == nil {
		return nil
	}
	out := new(SesameDeploymentStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TLS) DeepCopyInto(out *TLS) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TLS.
func (in *TLS) DeepCopy() *TLS {
	if in == nil {
		return nil
	}
	out := new(TLS)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TimeoutParameters) DeepCopyInto(out *TimeoutParameters) {
	*out = *in
	if in.RequestTimeout != nil {
		in, out := &in.RequestTimeout, &out.RequestTimeout
		*out = new(string)
		**out = **in
	}
	if in.ConnectionIdleTimeout != nil {
		in, out := &in.ConnectionIdleTimeout, &out.ConnectionIdleTimeout
		*out = new(string)
		**out = **in
	}
	if in.StreamIdleTimeout != nil {
		in, out := &in.StreamIdleTimeout, &out.StreamIdleTimeout
		*out = new(string)
		**out = **in
	}
	if in.MaxConnectionDuration != nil {
		in, out := &in.MaxConnectionDuration, &out.MaxConnectionDuration
		*out = new(string)
		**out = **in
	}
	if in.DelayedCloseTimeout != nil {
		in, out := &in.DelayedCloseTimeout, &out.DelayedCloseTimeout
		*out = new(string)
		**out = **in
	}
	if in.ConnectionShutdownGracePeriod != nil {
		in, out := &in.ConnectionShutdownGracePeriod, &out.ConnectionShutdownGracePeriod
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TimeoutParameters.
func (in *TimeoutParameters) DeepCopy() *TimeoutParameters {
	if in == nil {
		return nil
	}
	out := new(TimeoutParameters)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *XDSServerConfig) DeepCopyInto(out *XDSServerConfig) {
	*out = *in
	if in.TLS != nil {
		in, out := &in.TLS, &out.TLS
		*out = new(TLS)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new XDSServerConfig.
func (in *XDSServerConfig) DeepCopy() *XDSServerConfig {
	if in == nil {
		return nil
	}
	out := new(XDSServerConfig)
	in.DeepCopyInto(out)
	return out
}
