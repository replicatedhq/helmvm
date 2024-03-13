package registry

import (
	"context"
	"fmt"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/k0sproject/k0s/pkg/apis/k0s/v1beta1"
	"github.com/replicatedhq/troubleshoot/pkg/apis/troubleshoot/v1beta2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/replicatedhq/embedded-cluster/pkg/kubeutils"
	"github.com/replicatedhq/embedded-cluster/pkg/spinner"
)

const (
	releaseName = "registry"
	namespace   = "registry"
)

// Overwritten by -ldflags in Makefile
var (
	ChartURL     = "https://url"
	ChartName    = "name"
	Version      = "v0.0.0"
	ImageVersion = "2.8.3"
)

var helmValues = map[string]interface{}{
	"replicaCount":     1,
	"fullnameOverride": "registry",
	"image": map[string]interface{}{
		"tag": ImageVersion,
	},
	"storage": "hostpath", // this is not a recognized option but gets us past all the storage options
	"configData": map[string]interface{}{
		"auth": map[string]interface{}{
			"htpasswd": map[string]interface{}{
				"realm": "Registry",
				"path":  "/auth/htpasswd",
			},
		},
		"storage": map[string]interface{}{
			"filesystem": map[string]interface{}{
				"rootdirectory": "/var/lib/registry",
			},
		},
	},
	"extraVolumeMounts": []map[string]interface{}{
		{
			"name":      "auth",
			"mountPath": "/auth",
		},
		{
			"name":      "registry-data",
			"mountPath": "/var/lib/registry",
		},
	},
	"extraVolumes": []map[string]interface{}{
		{
			"name": "auth",
			"secret": map[string]interface{}{
				"secretName": "registry-auth",
			},
		},
		{
			"name": "registry-data",
			"hostPath": map[string]interface{}{
				"path": "/var/lib/embedded-cluster/registry",
			},
		},
	},
}

// Registry manages the installation of the Registry helm chart.
type Registry struct {
	isAirgap bool
}

// Version returns the version of the Registry chart.
func (o *Registry) Version() (map[string]string, error) {
	return map[string]string{"Registry": "v" + Version}, nil
}

func (a *Registry) Name() string {
	return "Registry"
}

// HostPreflights returns the host preflight objects found inside the Registry
// Helm Chart, this is empty as there is no host preflight on there.
func (o *Registry) HostPreflights() (*v1beta2.HostPreflightSpec, error) {
	return nil, nil
}

// GetProtectedFields returns the protected fields for the embedded charts.
// placeholder for now.
func (o *Registry) GetProtectedFields() map[string][]string {
	protectedFields := []string{}
	return map[string][]string{releaseName: protectedFields}
}

// GenerateHelmConfig generates the helm config for the Registry chart.
func (o *Registry) GenerateHelmConfig(onlyDefaults bool) ([]v1beta1.Chart, []v1beta1.Repository, error) {
	if !o.isAirgap {
		return nil, nil, nil
	}

	chartConfig := v1beta1.Chart{
		Name:      releaseName,
		ChartName: ChartName,
		Version:   Version,
		TargetNS:  namespace,
		Order:     3,
	}

	repositoryConfig := v1beta1.Repository{
		Name: "twuni",
		URL:  ChartURL,
	}

	valuesStringData, err := yaml.Marshal(helmValues)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to marshal helm values: %w", err)
	}
	chartConfig.Values = string(valuesStringData)

	return []v1beta1.Chart{chartConfig}, []v1beta1.Repository{repositoryConfig}, nil
}

// Outro is executed after the cluster deployment.
func (o *Registry) Outro(ctx context.Context, cli client.Client) error {
	if !o.isAirgap {
		return nil
	}

	loading := spinner.Start()
	loading.Infof("Waiting for Registry to be ready")
	if err := kubeutils.WaitForNamespace(ctx, cli, namespace); err != nil {
		loading.Close()
		return err
	}

	htpasswd := corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "registry-auth",
			Namespace: namespace,
		},
		StringData: map[string]string{
			"htpasswd": "TODO",
		},
		Type: "Opaque",
	}
	err := cli.Create(ctx, &htpasswd)
	if err != nil {
		loading.Close()
		return fmt.Errorf("unable to create registry-auth secret: %w", err)
	}

	// TODO generate a htpasswd secret

	if err := kubeutils.WaitForDeployment(ctx, cli, namespace, "registry"); err != nil {
		loading.Close()
		return err
	}
	loading.Closef("Registry is ready!")
	return nil
}

// New creates a new Registry addon.
func New(isAirgap bool) (*Registry, error) {
	return &Registry{isAirgap: isAirgap}, nil
}
