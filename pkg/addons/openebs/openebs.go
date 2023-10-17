// Package openebs manages the OpenEBS storage provisioner helm chart
// installation or upgrade in the cluster.
package openebs

import (
	"fmt"
	"github.com/k0sproject/k0s/pkg/apis/v1beta1"
	"github.com/replicatedhq/troubleshoot/pkg/apis/troubleshoot/v1beta2"
	"gopkg.in/yaml.v2"
)

const (
	releaseName = "openebs"
)

var Version = "v0.0.0"

var helmValues = map[string]interface{}{
	"ndmOperator": map[string]interface{}{
		"enabled": false,
	},
	"ndm": map[string]interface{}{
		"enabled": false,
	},
	"localprovisioner": map[string]interface{}{
		"hostpathClass": map[string]interface{}{
			"isDefaultClass": true,
		},
	},
}

type OpenEBS struct {
	namespace string
}

func (o *OpenEBS) Version() (map[string]string, error) {
	return map[string]string{"OpenEBS": "v" + Version}, nil
}

// HostPreflight returns the host preflight objects found inside the OpenEBS
// Helm Chart, this is empty as there is no host preflight on there.
func (o *OpenEBS) HostPreflights() (*v1beta2.HostPreflightSpec, error) {
	return nil, nil
}

func (o *OpenEBS) GenerateHelmConfig() ([]v1beta1.Chart, []v1beta1.Repository, error) {
	chartConfig := v1beta1.Chart{
		Name:      releaseName,
		ChartName: "openebs/openebs",
		Version:   Version,
		TargetNS:  o.namespace,
	}

	repositoryConfig := v1beta1.Repository{
		Name: "openebs",
		URL:  "https://openebs.github.io/charts",
	}

	valuesStringData, err := yaml.Marshal(helmValues)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to marshal helm values: %w", err)
	}
	chartConfig.Values = string(valuesStringData)

	return []v1beta1.Chart{chartConfig}, []v1beta1.Repository{repositoryConfig}, nil
}

func New(namespace string) (*OpenEBS, error) {
	return &OpenEBS{namespace: namespace}, nil
}
