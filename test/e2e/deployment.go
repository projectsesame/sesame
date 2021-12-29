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

package e2e

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/onsi/gomega/gexec"
	sesame_api_v1alpha1 "github.com/projectsesame/sesame/apis/projectsesame/v1alpha1"
	"github.com/projectsesame/sesame/pkg/config"
	"gopkg.in/yaml.v2"
	apps_v1 "k8s.io/api/apps/v1"
	batch_v1 "k8s.io/api/batch/v1"
	coordinationv1 "k8s.io/api/coordination/v1"
	v1 "k8s.io/api/core/v1"
	rbac_v1 "k8s.io/api/rbac/v1"
	apiextensions_v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	apimachinery_util_yaml "k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Deployment struct {
	// k8s client
	client client.Client

	// Command output is written to this writer.
	cmdOutputWriter io.Writer

	// Path to kube config to use with a local Sesame.
	kubeConfig string
	// Hostname to use when running Sesame locally.
	localSesameHost string
	// Port for local Sesame to bind to.
	localSesamePort string
	// Path to Sesame binary for use when running locally.
	SesameBin string

	// Sesame image to use in in-cluster deployment.
	SesameImage string

	Namespace                *v1.Namespace
	SesameServiceAccount     *v1.ServiceAccount
	EnvoyServiceAccount      *v1.ServiceAccount
	SesameConfigMap          *v1.ConfigMap
	ExtensionServiceCRD      *apiextensions_v1.CustomResourceDefinition
	HTTPProxyCRD             *apiextensions_v1.CustomResourceDefinition
	TLSCertDelegationCRD     *apiextensions_v1.CustomResourceDefinition
	SesameConfigurationCRD   *apiextensions_v1.CustomResourceDefinition
	SesameDeploymentCRD      *apiextensions_v1.CustomResourceDefinition
	CertgenServiceAccount    *v1.ServiceAccount
	SesameRoleBinding        *rbac_v1.RoleBinding
	CertgenRole              *rbac_v1.Role
	CertgenJob               *batch_v1.Job
	SesameClusterRoleBinding *rbac_v1.ClusterRoleBinding
	SesameClusterRole        *rbac_v1.ClusterRole
	SesameService            *v1.Service
	EnvoyService             *v1.Service
	SesameDeployment         *apps_v1.Deployment
	EnvoyDaemonSet           *apps_v1.DaemonSet

	// Optional volumes that will be attached to Envoy daemonset.
	EnvoyExtraVolumes      []v1.Volume
	EnvoyExtraVolumeMounts []v1.VolumeMount

	// Ratelimit deployment.
	RateLimitDeployment       *apps_v1.Deployment
	RateLimitService          *v1.Service
	RateLimitExtensionService *sesame_api_v1alpha1.ExtensionService
}

// UnmarshalResources unmarshals resources from rendered Sesame manifest in
// order.
// Note: This will need to be updated if any new resources are added to the
// rendered deployment manifest.
func (d *Deployment) UnmarshalResources() error {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return errors.New("could not get path to this source file (test/e2e/deployment.go)")
	}
	renderedManifestPath := filepath.Join(filepath.Dir(thisFile), "..", "..", "examples", "render", "sesame.yaml")
	file, err := os.Open(renderedManifestPath)
	if err != nil {
		return err
	}
	defer file.Close()
	decoder := apimachinery_util_yaml.NewYAMLToJSONDecoder(file)

	// Discard empty document.
	if err := decoder.Decode(new(struct{})); err != nil {
		return err
	}

	d.Namespace = new(v1.Namespace)
	d.SesameServiceAccount = new(v1.ServiceAccount)
	d.EnvoyServiceAccount = new(v1.ServiceAccount)
	d.SesameConfigMap = new(v1.ConfigMap)
	d.ExtensionServiceCRD = new(apiextensions_v1.CustomResourceDefinition)
	d.HTTPProxyCRD = new(apiextensions_v1.CustomResourceDefinition)
	d.TLSCertDelegationCRD = new(apiextensions_v1.CustomResourceDefinition)
	d.SesameConfigurationCRD = new(apiextensions_v1.CustomResourceDefinition)
	d.SesameDeploymentCRD = new(apiextensions_v1.CustomResourceDefinition)
	d.CertgenServiceAccount = new(v1.ServiceAccount)
	d.SesameRoleBinding = new(rbac_v1.RoleBinding)
	d.CertgenRole = new(rbac_v1.Role)
	d.CertgenJob = new(batch_v1.Job)
	d.SesameClusterRoleBinding = new(rbac_v1.ClusterRoleBinding)
	d.SesameClusterRole = new(rbac_v1.ClusterRole)
	d.SesameService = new(v1.Service)
	d.EnvoyService = new(v1.Service)
	d.SesameDeployment = new(apps_v1.Deployment)
	d.EnvoyDaemonSet = new(apps_v1.DaemonSet)
	objects := []interface{}{
		d.Namespace,
		d.SesameServiceAccount,
		d.EnvoyServiceAccount,
		d.SesameConfigMap,
		d.ExtensionServiceCRD,
		d.HTTPProxyCRD,
		d.TLSCertDelegationCRD,
		d.SesameConfigurationCRD,
		d.SesameDeploymentCRD,
		d.CertgenServiceAccount,
		d.SesameRoleBinding,
		d.CertgenRole,
		d.CertgenJob,
		d.SesameClusterRoleBinding,
		d.SesameClusterRole,
		d.SesameService,
		d.EnvoyService,
		d.SesameDeployment,
		d.EnvoyDaemonSet,
	}
	for _, o := range objects {
		if err := decoder.Decode(o); err != nil {
			return err
		}
	}

	rateLimitExamplePath := filepath.Join(filepath.Dir(thisFile), "..", "..", "examples", "ratelimit")
	rateLimitDeploymentFile := filepath.Join(rateLimitExamplePath, "02-ratelimit.yaml")
	rateLimitExtSvcFile := filepath.Join(rateLimitExamplePath, "03-ratelimit-extsvc.yaml")

	rLDFile, err := os.Open(rateLimitDeploymentFile)
	if err != nil {
		return err
	}
	defer rLDFile.Close()
	decoder = apimachinery_util_yaml.NewYAMLToJSONDecoder(rLDFile)
	d.RateLimitDeployment = new(apps_v1.Deployment)
	if err := decoder.Decode(d.RateLimitDeployment); err != nil {
		return err
	}
	d.RateLimitService = new(v1.Service)
	if err := decoder.Decode(d.RateLimitService); err != nil {
		return err
	}

	rLESFile, err := os.Open(rateLimitExtSvcFile)
	if err != nil {
		return err
	}
	defer rLESFile.Close()
	decoder = apimachinery_util_yaml.NewYAMLToJSONDecoder(rLESFile)
	d.RateLimitExtensionService = new(sesame_api_v1alpha1.ExtensionService)

	return decoder.Decode(d.RateLimitExtensionService)
}

// Common case of updating object if exists, create otherwise.
func (d *Deployment) ensureResource(new, existing client.Object) error {
	if err := d.client.Get(context.TODO(), client.ObjectKeyFromObject(new), existing); err != nil {
		if api_errors.IsNotFound(err) {
			return d.client.Create(context.TODO(), new)
		}
		return err
	}
	new.SetResourceVersion(existing.GetResourceVersion())
	// If a v1.Service, pass along existing cluster IP and healthcheck node port.
	if newS, ok := new.(*v1.Service); ok {
		existingS := existing.(*v1.Service)
		newS.Spec.ClusterIP = existingS.Spec.ClusterIP
		newS.Spec.ClusterIPs = existingS.Spec.ClusterIPs
		newS.Spec.HealthCheckNodePort = existingS.Spec.HealthCheckNodePort
	}
	return d.client.Update(context.TODO(), new)
}

func (d *Deployment) EnsureNamespace() error {
	return d.ensureResource(d.Namespace, new(v1.Namespace))
}

func (d *Deployment) EnsureSesameServiceAccount() error {
	return d.ensureResource(d.SesameServiceAccount, new(v1.ServiceAccount))
}

func (d *Deployment) EnsureEnvoyServiceAccount() error {
	return d.ensureResource(d.EnvoyServiceAccount, new(v1.ServiceAccount))
}

func (d *Deployment) EnsureSesameConfigMap() error {
	return d.ensureResource(d.SesameConfigMap, new(v1.ConfigMap))
}

func (d *Deployment) EnsureExtensionServiceCRD() error {
	return d.ensureResource(d.ExtensionServiceCRD, new(apiextensions_v1.CustomResourceDefinition))
}

func (d *Deployment) EnsureHTTPProxyCRD() error {
	return d.ensureResource(d.HTTPProxyCRD, new(apiextensions_v1.CustomResourceDefinition))
}

func (d *Deployment) EnsureTLSCertDelegationCRD() error {
	return d.ensureResource(d.TLSCertDelegationCRD, new(apiextensions_v1.CustomResourceDefinition))
}

func (d *Deployment) EnsureSesameConfigurationCRD() error {
	return d.ensureResource(d.SesameConfigurationCRD, new(apiextensions_v1.CustomResourceDefinition))
}

func (d *Deployment) EnsureSesameDeploymentCRD() error {
	return d.ensureResource(d.SesameDeploymentCRD, new(apiextensions_v1.CustomResourceDefinition))
}

func (d *Deployment) EnsureCertgenServiceAccount() error {
	return d.ensureResource(d.CertgenServiceAccount, new(v1.ServiceAccount))
}

func (d *Deployment) EnsureSesameRoleBinding() error {
	return d.ensureResource(d.SesameRoleBinding, new(rbac_v1.RoleBinding))
}

func (d *Deployment) EnsureCertgenRole() error {
	return d.ensureResource(d.CertgenRole, new(rbac_v1.Role))
}

func (d *Deployment) EnsureCertgenJob() error {
	// Delete job if exists with same name, then create.
	tempJ := new(batch_v1.Job)
	jobDeleted := func() (bool, error) {
		return api_errors.IsNotFound(d.client.Get(context.TODO(), client.ObjectKeyFromObject(d.CertgenJob), tempJ)), nil
	}
	if ok, _ := jobDeleted(); !ok {
		if err := d.client.Delete(context.TODO(), tempJ); err != nil {
			return err
		}
	}
	if err := wait.PollImmediate(time.Millisecond*50, time.Minute, jobDeleted); err != nil {
		return err
	}
	return d.client.Create(context.TODO(), d.CertgenJob)
}

func (d *Deployment) EnsureSesameClusterRoleBinding() error {
	return d.ensureResource(d.SesameClusterRoleBinding, new(rbac_v1.ClusterRoleBinding))
}

func (d *Deployment) EnsureSesameClusterRole() error {
	return d.ensureResource(d.SesameClusterRole, new(rbac_v1.ClusterRole))
}

func (d *Deployment) EnsureSesameService() error {
	return d.ensureResource(d.SesameService, new(v1.Service))
}

func (d *Deployment) EnsureEnvoyService() error {
	return d.ensureResource(d.EnvoyService, new(v1.Service))
}

func (d *Deployment) EnsureSesameDeployment() error {
	return d.ensureResource(d.SesameDeployment, new(apps_v1.Deployment))
}

func (d *Deployment) WaitForSesameDeploymentUpdated() error {
	// List pods with app label "sesame" and check that pods are updated
	// with expected container image and in ready state.
	// We do this instead of checking Deployment status as it is possible
	// for it not to have been updated yet and replicas not yet been shut
	// down.

	if len(d.SesameDeployment.Spec.Template.Spec.Containers) != 1 {
		return errors.New("invalid Sesame Deployment containers spec")
	}
	SesamePodImage := d.SesameDeployment.Spec.Template.Spec.Containers[0].Image
	updatedPods := func() (bool, error) {
		pods := new(v1.PodList)
		labelSelectAppSesame := &client.ListOptions{
			LabelSelector: labels.SelectorFromSet(d.SesameDeployment.Spec.Selector.MatchLabels),
			Namespace:     d.SesameDeployment.Namespace,
		}
		if err := d.client.List(context.TODO(), pods, labelSelectAppSesame); err != nil {
			return false, err
		}
		if pods == nil {
			return false, errors.New("failed to fetch Sesame Deployment pods")
		}

		updatedPods := 0
		for _, pod := range pods.Items {
			if len(pod.Spec.Containers) != 1 {
				return false, errors.New("invalid Sesame Deployment pod containers")
			}
			if pod.Spec.Containers[0].Image != SesamePodImage {
				continue
			}
			for _, cond := range pod.Status.Conditions {
				if cond.Type == v1.PodReady && cond.Status == v1.ConditionTrue {
					updatedPods++
				}
			}
		}
		return updatedPods == int(*d.SesameDeployment.Spec.Replicas), nil
	}
	return wait.PollImmediate(time.Millisecond*50, time.Minute, updatedPods)
}

func (d *Deployment) EnsureEnvoyDaemonSet() error {
	return d.ensureResource(d.EnvoyDaemonSet, new(apps_v1.DaemonSet))
}

// Wait for Envoy daemonset update to go out of date due to the
// update propagating. Possibly needed before calling WaitForEnvoyDaemonSetUpdated
// to ensure that does not pass without waiting for update to propagate.
func (d *Deployment) WaitForEnvoyDaemonSetOutOfDate() error {
	daemonSetUpdated := func() (bool, error) {
		tempDS := new(apps_v1.DaemonSet)
		if err := d.client.Get(context.TODO(), client.ObjectKeyFromObject(d.EnvoyDaemonSet), tempDS); err != nil {
			return false, err
		}
		return tempDS.Status.UpdatedNumberScheduled < tempDS.Status.DesiredNumberScheduled, nil
	}
	return wait.PollImmediate(time.Millisecond*50, time.Minute*3, daemonSetUpdated)
}

func (d *Deployment) WaitForEnvoyDaemonSetUpdated() error {
	daemonSetUpdated := func() (bool, error) {
		tempDS := new(apps_v1.DaemonSet)
		if err := d.client.Get(context.TODO(), client.ObjectKeyFromObject(d.EnvoyDaemonSet), tempDS); err != nil {
			return false, err
		}
		return tempDS.Status.NumberAvailable > 0 &&
			tempDS.Status.NumberAvailable == tempDS.Status.DesiredNumberScheduled &&
			tempDS.Status.UpdatedNumberScheduled == tempDS.Status.DesiredNumberScheduled, nil
	}
	return wait.PollImmediate(time.Millisecond*50, time.Minute*3, daemonSetUpdated)
}

func (d *Deployment) EnsureRateLimitResources(namespace string, configContents string) error {
	setNamespace := d.Namespace.Name
	if len(namespace) > 0 {
		setNamespace = namespace
	}

	configMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ratelimit-config",
			Namespace: setNamespace,
		},
		Data: map[string]string{
			"ratelimit-config.yaml": configContents,
		},
	}
	if err := d.ensureResource(configMap, new(v1.ConfigMap)); err != nil {
		return err
	}

	deployment := d.RateLimitDeployment.DeepCopy()
	deployment.Namespace = setNamespace
	if err := d.ensureResource(deployment, new(apps_v1.Deployment)); err != nil {
		return err
	}

	service := d.RateLimitService.DeepCopy()
	service.Namespace = setNamespace
	if err := d.ensureResource(service, new(v1.Service)); err != nil {
		return err
	}

	extSvc := d.RateLimitExtensionService.DeepCopy()
	extSvc.Namespace = setNamespace
	return d.ensureResource(extSvc, new(sesame_api_v1alpha1.ExtensionService))
}

// Convenience method for deploying the pieces of the deployment needed for
// testing Sesame running locally, out of cluster.
// Includes:
// - namespace
// - Envoy service account
// - CRDs
// - Envoy service
// - ConfigMap with Envoy bootstrap config
// - Envoy DaemonSet modified for local Sesame xDS server
func (d *Deployment) EnsureResourcesForLocalSesame() error {
	if err := d.EnsureNamespace(); err != nil {
		return err
	}
	if err := d.EnsureEnvoyServiceAccount(); err != nil {
		return err
	}
	if err := d.EnsureExtensionServiceCRD(); err != nil {
		return err
	}
	if err := d.EnsureHTTPProxyCRD(); err != nil {
		return err
	}
	if err := d.EnsureTLSCertDelegationCRD(); err != nil {
		return err
	}
	if err := d.EnsureSesameConfigurationCRD(); err != nil {
		return err
	}
	if err := d.EnsureSesameDeploymentCRD(); err != nil {
		return err
	}
	if err := d.EnsureEnvoyService(); err != nil {
		return err
	}

	bFile, err := ioutil.TempFile("", "bootstrap-*.json")
	if err != nil {
		return err
	}

	// Generate bootstrap config with Sesame local address and plaintext
	// client config.
	bootstrapCmd := exec.Command( // nolint:gosec
		d.SesameBin,
		"bootstrap",
		bFile.Name(),
		"--xds-address="+d.localSesameHost,
		"--xds-port="+d.localSesamePort,
		"--xds-resource-version=v3",
		"--admin-address=/admin/admin.sock",
	)

	session, err := gexec.Start(bootstrapCmd, d.cmdOutputWriter, d.cmdOutputWriter)
	if err != nil {
		return err
	}
	session.Wait()

	bootstrapContents, err := ioutil.ReadAll(bFile)
	if err != nil {
		return err
	}
	defer func() {
		bFile.Close()
		os.RemoveAll(bFile.Name())
	}()

	bootstrapConfigMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "envoy-bootstrap",
			Namespace: d.Namespace.Name,
		},
		Data: map[string]string{
			"envoy.json": string(bootstrapContents),
		},
	}
	if err := d.ensureResource(bootstrapConfigMap, new(v1.ConfigMap)); err != nil {
		return err
	}

	// Add bootstrap ConfigMap as volume and add envoy admin volume on Envoy pods (also removes cert volume).
	d.EnvoyDaemonSet.Spec.Template.Spec.Volumes = []v1.Volume{{
		Name: "envoy-config",
		VolumeSource: v1.VolumeSource{
			ConfigMap: &v1.ConfigMapVolumeSource{
				LocalObjectReference: v1.LocalObjectReference{
					Name: "envoy-bootstrap",
				},
			},
		},
	}, {
		Name: "envoy-admin",
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{},
		},
	}}

	// Remove cert volume mount.
	d.EnvoyDaemonSet.Spec.Template.Spec.Containers[1].VolumeMounts = []v1.VolumeMount{
		d.EnvoyDaemonSet.Spec.Template.Spec.Containers[1].VolumeMounts[0], // Config mount
		d.EnvoyDaemonSet.Spec.Template.Spec.Containers[1].VolumeMounts[2], // Admin mount
	}

	d.EnvoyDaemonSet.Spec.Template.Spec.Volumes = append(d.EnvoyDaemonSet.Spec.Template.Spec.Volumes, d.EnvoyExtraVolumes...)
	d.EnvoyDaemonSet.Spec.Template.Spec.Containers[1].VolumeMounts = append(d.EnvoyDaemonSet.Spec.Template.Spec.Containers[1].VolumeMounts, d.EnvoyExtraVolumeMounts...)

	// Remove init container.
	d.EnvoyDaemonSet.Spec.Template.Spec.InitContainers = nil

	// Remove shutdown-manager container.
	d.EnvoyDaemonSet.Spec.Template.Spec.Containers = d.EnvoyDaemonSet.Spec.Template.Spec.Containers[1:]

	// Expose the metrics & admin interfaces via host port to test from outside the kind cluster.
	d.EnvoyDaemonSet.Spec.Template.Spec.Containers[0].Ports = append(d.EnvoyDaemonSet.Spec.Template.Spec.Containers[0].Ports,
		v1.ContainerPort{
			Name:          "metrics",
			ContainerPort: 8002,
			HostPort:      8002,
			Protocol:      v1.ProtocolTCP,
		})
	return d.EnsureEnvoyDaemonSet()
}

// DeleteResourcesForLocalSesame ensures deletion of all resources
// created in the projectsesame namespace for running a local sesame.
// This is done instead of deleting the entire namespace as a performance
// optimization, because deleting non-empty namespaces can take up to a
// couple minutes to complete.
func (d *Deployment) DeleteResourcesForLocalSesame() error {
	for _, r := range []client.Object{
		d.EnvoyDaemonSet,
		d.SesameConfigMap,
		d.EnvoyService,
		d.TLSCertDelegationCRD,
		d.ExtensionServiceCRD,
		d.HTTPProxyCRD,
		d.EnvoyServiceAccount,
	} {
		if err := d.EnsureDeleted(r); err != nil {
			return err
		}
	}

	return nil
}

// Starts local sesame, applying arguments and marshaling config into config
// file. Returns running Sesame command and config file so we can clean them
// up.
func (d *Deployment) StartLocalSesame(config *config.Parameters, SesameConfiguration *sesame_api_v1alpha1.SesameConfiguration, additionalArgs ...string) (*gexec.Session, string, error) {

	var content []byte
	var configReferenceName string
	var SesameServeArgs []string
	var err error

	// Look for the ENV variable to tell if this test run should use
	// the SesameConfiguration file or the SesameConfiguration CRD.
	if UsingSesameConfigCRD() {
		port, _ := strconv.Atoi(d.localSesamePort)

		SesameConfiguration.Name = randomString(14)

		// Set the xds server to the defined testing port as well as enable insecure communication.
		SesameConfiguration.Spec.XDSServer = sesame_api_v1alpha1.XDSServerConfig{
			Type:    sesame_api_v1alpha1.SesameServerType,
			Address: "0.0.0.0",
			Port:    port,
			TLS: &sesame_api_v1alpha1.TLS{
				Insecure: true,
			},
		}

		if err := d.client.Create(context.TODO(), SesameConfiguration); err != nil {
			return nil, "", fmt.Errorf("could not create SesameConfiguration: %v", err)
		}

		SesameServeArgs = append([]string{
			"serve",
			"--kubeconfig=" + d.kubeConfig,
			"--sesame-config-name=" + SesameConfiguration.Name,
			"--disable-leader-election",
		}, additionalArgs...)

		configReferenceName = SesameConfiguration.Name
	} else {

		configFile, err := ioutil.TempFile("", "sesame-config-*.yaml")
		if err != nil {
			return nil, "", err
		}
		defer configFile.Close()

		content, err = yaml.Marshal(config)
		if err != nil {
			return nil, "", err
		}
		if err := os.WriteFile(configFile.Name(), content, 0600); err != nil {
			return nil, "", err
		}

		SesameServeArgs = append([]string{
			"serve",
			"--xds-address=0.0.0.0",
			"--xds-port=" + d.localSesamePort,
			"--insecure",
			"--kubeconfig=" + d.kubeConfig,
			"--config-path=" + configFile.Name(),
			"--disable-leader-election",
		}, additionalArgs...)

		configReferenceName = configFile.Name()
	}

	session, err := gexec.Start(exec.Command(d.SesameBin, SesameServeArgs...), d.cmdOutputWriter, d.cmdOutputWriter) // nolint:gosec
	if err != nil {
		return nil, "", err
	}
	return session, configReferenceName, nil
}

func (d *Deployment) StopLocalSesame(SesameCmd *gexec.Session, configFile string) error {

	// Look for the ENV variable to tell if this test run should use
	// the SesameConfiguration file or the SesameConfiguration CRD.
	if useSesameConfiguration, variableFound := os.LookupEnv("USE_Sesame_CONFIGURATION_CRD"); variableFound && useSesameConfiguration == "true" {
		cc := &sesame_api_v1alpha1.SesameConfiguration{
			ObjectMeta: metav1.ObjectMeta{
				Name:      configFile,
				Namespace: "projectsesame",
			},
		}

		if err := d.client.Delete(context.TODO(), cc); err != nil {
			return fmt.Errorf("could not delete SesameConfiguration: %v", err)
		}
	}

	// Default timeout of 1s produces test flakes,
	// a minute should be more than enough to avoid them.
	SesameCmd.Terminate().Wait(time.Minute)
	return os.RemoveAll(configFile)
}

// Convenience method for deploying the pieces of the deployment needed for
// testing Sesame running in-cluster.
// Includes:
// - namespace
// - Sesame service account
// - Envoy service account
// - Sesame configmap
// - CRDs
// - Certgen service account
// - Sesame role binding
// - Certgen role
// - Certgen job
// - Sesame cluster role binding
// - Sesame cluster role
// - Sesame service
// - Envoy service
// - Sesame deployment (only started if bool passed in is true)
// - Envoy DaemonSet
func (d *Deployment) EnsureResourcesForInclusterSesame(startSesameDeployment bool) error {
	fmt.Fprintf(d.cmdOutputWriter, "Deploying Sesame with image: %s\n", d.SesameImage)

	if err := d.EnsureNamespace(); err != nil {
		return err
	}
	if err := d.EnsureSesameServiceAccount(); err != nil {
		return err
	}
	if err := d.EnsureEnvoyServiceAccount(); err != nil {
		return err
	}
	if err := d.EnsureSesameConfigMap(); err != nil {
		return err
	}
	if err := d.EnsureExtensionServiceCRD(); err != nil {
		return err
	}
	if err := d.EnsureHTTPProxyCRD(); err != nil {
		return err
	}
	if err := d.EnsureTLSCertDelegationCRD(); err != nil {
		return err
	}
	if err := d.EnsureSesameConfigurationCRD(); err != nil {
		return err
	}
	if err := d.EnsureSesameDeploymentCRD(); err != nil {
		return err
	}
	if err := d.EnsureCertgenServiceAccount(); err != nil {
		return err
	}
	if err := d.EnsureSesameRoleBinding(); err != nil {
		return err
	}
	if err := d.EnsureCertgenRole(); err != nil {
		return err
	}
	// Update container image.
	if l := len(d.CertgenJob.Spec.Template.Spec.Containers); l != 1 {
		return fmt.Errorf("invalid certgen job containers, expected 1, got %d", l)
	}
	d.CertgenJob.Spec.Template.Spec.Containers[0].Image = d.SesameImage
	d.CertgenJob.Spec.Template.Spec.Containers[0].ImagePullPolicy = v1.PullIfNotPresent
	if err := d.EnsureCertgenJob(); err != nil {
		return err
	}
	if err := d.EnsureSesameClusterRoleBinding(); err != nil {
		return err
	}
	if err := d.EnsureSesameClusterRole(); err != nil {
		return err
	}
	if err := d.EnsureSesameService(); err != nil {
		return err
	}
	if err := d.EnsureEnvoyService(); err != nil {
		return err
	}
	// Update container image.
	if l := len(d.SesameDeployment.Spec.Template.Spec.Containers); l != 1 {
		return fmt.Errorf("invalid sesame deployment containers, expected 1, got %d", l)
	}
	d.SesameDeployment.Spec.Template.Spec.Containers[0].Image = d.SesameImage
	d.SesameDeployment.Spec.Template.Spec.Containers[0].ImagePullPolicy = v1.PullIfNotPresent
	if startSesameDeployment {
		if err := d.EnsureSesameDeployment(); err != nil {
			return err
		}
	}
	// Update container image.
	if l := len(d.EnvoyDaemonSet.Spec.Template.Spec.InitContainers); l != 1 {
		return fmt.Errorf("invalid envoy daemonset init containers, expected 1, got %d", l)
	}
	d.EnvoyDaemonSet.Spec.Template.Spec.InitContainers[0].Image = d.SesameImage
	d.EnvoyDaemonSet.Spec.Template.Spec.InitContainers[0].ImagePullPolicy = v1.PullIfNotPresent
	if l := len(d.EnvoyDaemonSet.Spec.Template.Spec.Containers); l != 2 {
		return fmt.Errorf("invalid envoy daemonset containers, expected 2, got %d", l)
	}
	d.EnvoyDaemonSet.Spec.Template.Spec.Containers[0].Image = d.SesameImage
	d.EnvoyDaemonSet.Spec.Template.Spec.Containers[0].ImagePullPolicy = v1.PullIfNotPresent
	// Set shutdown check-delay to 0s to ensure cleanup is fast.
	d.EnvoyDaemonSet.Spec.Template.Spec.Containers[0].Lifecycle.PreStop.Exec.Command = append(d.EnvoyDaemonSet.Spec.Template.Spec.Containers[0].Lifecycle.PreStop.Exec.Command, "--check-delay=0s")
	return d.EnsureEnvoyDaemonSet()
}

// DeleteResourcesForInclusterSesame ensures deletion of all resources
// created in the projectsesame namespace for running a sesame incluster.
// This is done instead of deleting the entire namespace as a performance
// optimization, because deleting non-empty namespaces can take up to a
// couple minutes to complete.
func (d *Deployment) DeleteResourcesForInclusterSesame() error {
	// Also need to delete leader election resources to ensure
	// multiple test runs can be run cleanly.
	leaderElectionConfigMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "leader-elect",
			Namespace: d.Namespace.Name,
		},
	}
	leaderElectionLease := &coordinationv1.Lease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "leader-elect",
			Namespace: d.Namespace.Name,
		},
	}

	for _, r := range []client.Object{
		d.EnvoyDaemonSet,
		d.SesameDeployment,
		leaderElectionLease,
		leaderElectionConfigMap,
		d.EnvoyService,
		d.SesameService,
		d.SesameClusterRole,
		d.SesameClusterRoleBinding,
		d.CertgenJob,
		d.CertgenRole,
		d.SesameRoleBinding,
		d.CertgenServiceAccount,
		d.TLSCertDelegationCRD,
		d.ExtensionServiceCRD,
		d.HTTPProxyCRD,
		d.SesameConfigMap,
		d.EnvoyServiceAccount,
		d.SesameServiceAccount,
	} {
		if err := d.EnsureDeleted(r); err != nil {
			return err
		}
	}

	return nil
}

func (d *Deployment) EnsureDeleted(obj client.Object) error {
	// Delete the object; if it already doesn't exist,
	// then we're done.
	err := d.client.Delete(context.Background(), obj)
	if api_errors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting resource %T %s/%s: %v", obj, obj.GetNamespace(), obj.GetName(), err)
	}

	// Wait to ensure it's fully deleted.
	if err := wait.PollImmediate(100*time.Millisecond, time.Minute, func() (bool, error) {
		err := d.client.Get(context.Background(), client.ObjectKeyFromObject(obj), obj)
		if api_errors.IsNotFound(err) {
			return true, nil
		}
		return false, nil
	}); err != nil {
		return fmt.Errorf("error waiting for deletion of resource %T %s/%s: %v", obj, obj.GetNamespace(), obj.GetName(), err)
	}

	// Clear out resource version to ensure object can be used again.
	obj.SetResourceVersion("")

	return nil
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return ""
		}
		ret[i] = letters[num.Int64()]
	}

	return string(ret)
}
