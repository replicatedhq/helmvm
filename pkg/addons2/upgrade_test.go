package addons2

import (
	"testing"

	ecv1beta1 "github.com/replicatedhq/embedded-cluster/kinds/apis/v1beta1"
	ectypes "github.com/replicatedhq/embedded-cluster/kinds/types"
	"github.com/replicatedhq/embedded-cluster/pkg/addons2/adminconsole"
	"github.com/replicatedhq/embedded-cluster/pkg/addons2/embeddedclusteroperator"
	"github.com/replicatedhq/embedded-cluster/pkg/addons2/openebs"
	"github.com/replicatedhq/embedded-cluster/pkg/addons2/registry"
	"github.com/replicatedhq/embedded-cluster/pkg/addons2/seaweedfs"
	"github.com/replicatedhq/embedded-cluster/pkg/addons2/types"
	"github.com/replicatedhq/embedded-cluster/pkg/addons2/velero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_getAddOnsForUpgrade(t *testing.T) {
	meta := &ectypes.ReleaseMetadata{
		Configs: ecv1beta1.Helm{
			Charts: []ecv1beta1.Chart{
				{
					Name:      "embedded-cluster-operator",
					ChartName: "replicated/embedded-cluster-operator",
					Version:   "1.22.0+k8s-1.30",
				},
			},
		},
		Images: []string{
			"proxy.replicated.com/anonymous/replicated/embedded-cluster-operator-image:1.22.0-k8s-1.30-amd64@sha256:929b6cb42add383a69e3b26790c06320bd4eac0ecd60b509212c1864d69c6a88",
			"proxy.replicated.com/anonymous/replicated/ec-utils:latest-amd64@sha256:f499ed26bd5899bc5a1ae14d9d13853d1fc615ae21bde86fe250960772fd2c70",
		},
	}

	tests := []struct {
		name   string
		in     *ecv1beta1.Installation
		meta   *ectypes.ReleaseMetadata
		verify func(t *testing.T, addons []types.AddOn, err error)
	}{
		{
			name: "online installation",
			in: &ecv1beta1.Installation{
				Spec: ecv1beta1.InstallationSpec{
					AirGap:           false,
					HighAvailability: false,
					BinaryName:       "test-binary-name",
				},
			},
			meta: meta,
			verify: func(t *testing.T, addons []types.AddOn, err error) {
				assert.NoError(t, err)
				assert.Len(t, addons, 3)

				_, ok := addons[0].(*openebs.OpenEBS)
				require.True(t, ok, "first addon should be OpenEBS")

				eco, ok := addons[1].(*embeddedclusteroperator.EmbeddedClusterOperator)
				require.True(t, ok, "second addon should be EmbeddedClusterOperator")
				assert.False(t, eco.IsAirgap, "ECO should not be in airgap mode")
				assert.Nil(t, eco.Proxy, "ECO should not have a proxy")
				assert.Equal(t, "replicated/embedded-cluster-operator", eco.ChartLocationOverride)
				assert.Equal(t, "1.22.0+k8s-1.30", eco.ChartVersionOverride)
				assert.Equal(t, "test-binary-name", eco.BinaryNameOverride)
				assert.Equal(t, "proxy.replicated.com/anonymous/replicated/embedded-cluster-operator-image", eco.ImageRepoOverride)
				assert.Equal(t, "1.22.0-k8s-1.30-amd64@sha256:929b6cb42add383a69e3b26790c06320bd4eac0ecd60b509212c1864d69c6a88", eco.ImageTagOverride)
				assert.Equal(t, "proxy.replicated.com/anonymous/replicated/ec-utils:latest-amd64@sha256:f499ed26bd5899bc5a1ae14d9d13853d1fc615ae21bde86fe250960772fd2c70", eco.UtilsImageOverride)

				adminConsole, ok := addons[2].(*adminconsole.AdminConsole)
				require.True(t, ok, "third addon should be AdminConsole")
				assert.False(t, adminConsole.IsAirgap, "AdminConsole should not be in airgap mode")
				assert.False(t, adminConsole.IsHA, "AdminConsole should not be in high availability mode")
				assert.Nil(t, adminConsole.Proxy, "AdminConsole should not have a proxy")
				assert.Empty(t, adminConsole.ServiceCIDR, "AdminConsole should not have a service CIDR")
			},
		},
		{
			name: "airgap installation",
			in: &ecv1beta1.Installation{
				Spec: ecv1beta1.InstallationSpec{
					AirGap:           true,
					HighAvailability: false,
					Network: &ecv1beta1.NetworkSpec{
						ServiceCIDR: "10.96.0.0/12",
					},
					BinaryName: "test-binary-name",
				},
			},
			meta: meta,
			verify: func(t *testing.T, addons []types.AddOn, err error) {
				assert.NoError(t, err)
				assert.Len(t, addons, 4)

				_, ok := addons[0].(*openebs.OpenEBS)
				require.True(t, ok, "first addon should be OpenEBS")

				eco, ok := addons[1].(*embeddedclusteroperator.EmbeddedClusterOperator)
				require.True(t, ok, "second addon should be EmbeddedClusterOperator")
				assert.True(t, eco.IsAirgap, "ECO should be in airgap mode")
				assert.Nil(t, eco.Proxy, "ECO should not have a proxy")
				assert.Equal(t, "replicated/embedded-cluster-operator", eco.ChartLocationOverride)
				assert.Equal(t, "1.22.0+k8s-1.30", eco.ChartVersionOverride)
				assert.Equal(t, "test-binary-name", eco.BinaryNameOverride)
				assert.Equal(t, "proxy.replicated.com/anonymous/replicated/embedded-cluster-operator-image", eco.ImageRepoOverride)
				assert.Equal(t, "1.22.0-k8s-1.30-amd64@sha256:929b6cb42add383a69e3b26790c06320bd4eac0ecd60b509212c1864d69c6a88", eco.ImageTagOverride)
				assert.Equal(t, "proxy.replicated.com/anonymous/replicated/ec-utils:latest-amd64@sha256:f499ed26bd5899bc5a1ae14d9d13853d1fc615ae21bde86fe250960772fd2c70", eco.UtilsImageOverride)

				reg, ok := addons[2].(*registry.Registry)
				require.True(t, ok, "third addon should be Registry")
				assert.Equal(t, "10.96.0.0/12", reg.ServiceCIDR)
				assert.False(t, reg.IsHA)

				adminConsole, ok := addons[3].(*adminconsole.AdminConsole)
				require.True(t, ok, "fourth addon should be AdminConsole")
				assert.True(t, adminConsole.IsAirgap, "AdminConsole should be in airgap mode")
				assert.False(t, adminConsole.IsHA, "AdminConsole should not be in high availability mode")
				assert.Nil(t, adminConsole.Proxy, "AdminConsole should not have a proxy")
				assert.Equal(t, "10.96.0.0/12", adminConsole.ServiceCIDR)
			},
		},
		{
			name: "with disaster recovery",
			in: &ecv1beta1.Installation{
				Spec: ecv1beta1.InstallationSpec{
					AirGap:           false,
					HighAvailability: false,
					Network: &ecv1beta1.NetworkSpec{
						ServiceCIDR: "10.96.0.0/12",
					},
					LicenseInfo: &ecv1beta1.LicenseInfo{
						IsDisasterRecoverySupported: true,
					},
					BinaryName: "test-binary-name",
				},
			},
			meta: meta,
			verify: func(t *testing.T, addons []types.AddOn, err error) {
				assert.NoError(t, err)
				assert.Len(t, addons, 4)

				_, ok := addons[0].(*openebs.OpenEBS)
				require.True(t, ok, "first addon should be OpenEBS")

				eco, ok := addons[1].(*embeddedclusteroperator.EmbeddedClusterOperator)
				require.True(t, ok, "second addon should be EmbeddedClusterOperator")
				assert.False(t, eco.IsAirgap, "ECO should not be in airgap mode")
				assert.Nil(t, eco.Proxy, "ECO should not have a proxy")
				assert.Equal(t, "replicated/embedded-cluster-operator", eco.ChartLocationOverride)
				assert.Equal(t, "1.22.0+k8s-1.30", eco.ChartVersionOverride)
				assert.Equal(t, "test-binary-name", eco.BinaryNameOverride)
				assert.Equal(t, "proxy.replicated.com/anonymous/replicated/embedded-cluster-operator-image", eco.ImageRepoOverride)
				assert.Equal(t, "1.22.0-k8s-1.30-amd64@sha256:929b6cb42add383a69e3b26790c06320bd4eac0ecd60b509212c1864d69c6a88", eco.ImageTagOverride)
				assert.Equal(t, "proxy.replicated.com/anonymous/replicated/ec-utils:latest-amd64@sha256:f499ed26bd5899bc5a1ae14d9d13853d1fc615ae21bde86fe250960772fd2c70", eco.UtilsImageOverride)

				vel, ok := addons[2].(*velero.Velero)
				require.True(t, ok, "third addon should be Velero")
				assert.Nil(t, vel.Proxy, "Velero should not have a proxy")

				adminConsole, ok := addons[3].(*adminconsole.AdminConsole)
				require.True(t, ok, "fourth addon should be AdminConsole")
				assert.False(t, adminConsole.IsAirgap, "AdminConsole should not be in airgap mode")
				assert.False(t, adminConsole.IsHA, "AdminConsole should not be in high availability mode")
				assert.Nil(t, adminConsole.Proxy, "AdminConsole should not have a proxy")
				assert.Equal(t, "10.96.0.0/12", adminConsole.ServiceCIDR)
			},
		},
		{
			name: "airgap HA with proxy and disaster recovery",
			in: &ecv1beta1.Installation{
				Spec: ecv1beta1.InstallationSpec{
					AirGap:           true,
					HighAvailability: true,
					Network: &ecv1beta1.NetworkSpec{
						ServiceCIDR: "10.96.0.0/12",
					},
					LicenseInfo: &ecv1beta1.LicenseInfo{
						IsDisasterRecoverySupported: true,
					},
					Proxy: &ecv1beta1.ProxySpec{
						HTTPProxy:  "http://proxy.example.com",
						HTTPSProxy: "https://proxy.example.com",
						NoProxy:    "localhost,127.0.0.1",
					},
					BinaryName: "test-binary-name",
				},
			},
			meta: meta,
			verify: func(t *testing.T, addons []types.AddOn, err error) {
				assert.NoError(t, err)
				assert.Len(t, addons, 6)

				_, ok := addons[0].(*openebs.OpenEBS)
				require.True(t, ok, "first addon should be OpenEBS")

				eco, ok := addons[1].(*embeddedclusteroperator.EmbeddedClusterOperator)
				require.True(t, ok, "second addon should be EmbeddedClusterOperator")
				assert.True(t, eco.IsAirgap, "ECO should be in airgap mode")
				assert.Equal(t, "http://proxy.example.com", eco.Proxy.HTTPProxy)
				assert.Equal(t, "https://proxy.example.com", eco.Proxy.HTTPSProxy)
				assert.Equal(t, "localhost,127.0.0.1", eco.Proxy.NoProxy)
				assert.Equal(t, "replicated/embedded-cluster-operator", eco.ChartLocationOverride)
				assert.Equal(t, "1.22.0+k8s-1.30", eco.ChartVersionOverride)
				assert.Equal(t, "test-binary-name", eco.BinaryNameOverride)
				assert.Equal(t, "proxy.replicated.com/anonymous/replicated/embedded-cluster-operator-image", eco.ImageRepoOverride)
				assert.Equal(t, "1.22.0-k8s-1.30-amd64@sha256:929b6cb42add383a69e3b26790c06320bd4eac0ecd60b509212c1864d69c6a88", eco.ImageTagOverride)
				assert.Equal(t, "proxy.replicated.com/anonymous/replicated/ec-utils:latest-amd64@sha256:f499ed26bd5899bc5a1ae14d9d13853d1fc615ae21bde86fe250960772fd2c70", eco.UtilsImageOverride)

				reg, ok := addons[2].(*registry.Registry)
				require.True(t, ok, "third addon should be Registry")
				assert.Equal(t, "10.96.0.0/12", reg.ServiceCIDR)
				assert.True(t, reg.IsHA)

				seaweed, ok := addons[3].(*seaweedfs.SeaweedFS)
				require.True(t, ok, "fourth addon should be SeaweedFS")
				assert.Equal(t, "10.96.0.0/12", seaweed.ServiceCIDR)

				vel, ok := addons[4].(*velero.Velero)
				require.True(t, ok, "fifth addon should be Velero")
				assert.Equal(t, "http://proxy.example.com", vel.Proxy.HTTPProxy)
				assert.Equal(t, "https://proxy.example.com", vel.Proxy.HTTPSProxy)
				assert.Equal(t, "localhost,127.0.0.1", vel.Proxy.NoProxy)

				adminConsole, ok := addons[5].(*adminconsole.AdminConsole)
				require.True(t, ok, "sixth addon should be AdminConsole")
				assert.True(t, adminConsole.IsAirgap, "AdminConsole should be in airgap mode")
				assert.True(t, adminConsole.IsHA, "AdminConsole should be in high availability mode")
				assert.Equal(t, "http://proxy.example.com", adminConsole.Proxy.HTTPProxy)
				assert.Equal(t, "https://proxy.example.com", adminConsole.Proxy.HTTPSProxy)
				assert.Equal(t, "localhost,127.0.0.1", adminConsole.Proxy.NoProxy)
				assert.Equal(t, "10.96.0.0/12", adminConsole.ServiceCIDR)
			},
		},
		{
			name: "invalid metadata - missing chart",
			in: &ecv1beta1.Installation{
				Spec: ecv1beta1.InstallationSpec{},
			},
			meta: &ectypes.ReleaseMetadata{
				Configs: ecv1beta1.Helm{
					Charts: []ecv1beta1.Chart{},
				},
				Images: meta.Images,
			},
			verify: func(t *testing.T, addons []types.AddOn, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "no embedded-cluster-operator chart found")
			},
		},
		{
			name: "invalid metadata - missing images",
			in: &ecv1beta1.Installation{
				Spec: ecv1beta1.InstallationSpec{},
			},
			meta: &ectypes.ReleaseMetadata{
				Configs: meta.Configs,
				Images:  []string{},
			},
			verify: func(t *testing.T, addons []types.AddOn, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "no embedded-cluster-operator-image found")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addons, err := getAddOnsForUpgrade(tt.in, tt.meta)
			tt.verify(t, addons, err)
		})
	}
}
