package config

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/k0sproject/k0s/pkg/airgap"
	k0sconfig "github.com/k0sproject/k0s/pkg/apis/k0s/v1beta1"
	k0sv1beta1 "github.com/k0sproject/k0s/pkg/apis/k0s/v1beta1"
	"github.com/k0sproject/k0s/pkg/constant"
	"github.com/replicatedhq/embedded-cluster/pkg/release"
	"gopkg.in/yaml.v2"
)

var (
	//go:embed static/metadata.yaml
	rawmetadata []byte
	// Metadata is the unmarshal version of rawmetadata.
	Metadata release.K0sMetadata
)

func init() {
	if err := yaml.Unmarshal(rawmetadata, &Metadata); err != nil {
		panic(fmt.Sprintf("unable to unmarshal metadata: %v", err))
	}
}

func ListK0sImages(cfg *k0sconfig.ClusterConfig) []string {
	var images []string
	for _, image := range airgap.GetImageURIs(cfg.Spec, true) {
		switch image {
		// skip these images
		case cfg.Spec.Images.KubeRouter.CNI.URI(),
			cfg.Spec.Images.KubeRouter.CNIInstaller.URI(),
			cfg.Spec.Images.Konnectivity.URI():
		default:
			if strings.Contains(image, constant.KubePauseContainerImage) {
				// there's a bug in GetImageURIs where it always returns the original pause image
				// instead of the one in the config, make sure to use the one in the config.
				images = append(images, cfg.Spec.Images.Pause.URI())
			} else {
				images = append(images, image)
			}
		}
	}
	return images
}

func overrideK0sImages(cfg *k0sv1beta1.ClusterConfig) {
	if cfg.Spec.Images == nil {
		cfg.Spec.Images = &k0sv1beta1.ClusterImages{}
	}

	cfg.Spec.Images.CoreDNS.Image = getImageName("coredns")
	cfg.Spec.Images.CoreDNS.Version = Metadata.Images["coredns"]

	cfg.Spec.Images.Calico.Node.Image = getImageName("calico-node")
	cfg.Spec.Images.Calico.Node.Version = Metadata.Images["calico-node"]

	cfg.Spec.Images.Calico.CNI.Image = getImageName("calico-cni")
	cfg.Spec.Images.Calico.CNI.Version = Metadata.Images["calico-cni"]

	cfg.Spec.Images.Calico.KubeControllers.Image = getImageName("calico-kube-controllers")
	cfg.Spec.Images.Calico.KubeControllers.Version = Metadata.Images["calico-kube-controllers"]

	cfg.Spec.Images.MetricsServer.Image = getImageName("metrics-server")
	cfg.Spec.Images.MetricsServer.Version = Metadata.Images["metrics-server"]

	cfg.Spec.Images.KubeProxy.Image = getImageName("kube-proxy")
	cfg.Spec.Images.KubeProxy.Version = Metadata.Images["kube-proxy"]

	cfg.Spec.Images.Pause.Image = getImageName("pause")
	cfg.Spec.Images.Pause.Version = Metadata.Images["pause"]

	// TODO (salah): uncomment when upstream PR for digest support is released: https://github.com/k0sproject/k0s/pull/4792
	// if cfg.Spec.Network == nil {
	// 	cfg.Spec.Network = &k0sv1beta1.Network{}
	// }
	// if cfg.Spec.Network.NodeLocalLoadBalancing == nil {
	// 	cfg.Spec.Network.NodeLocalLoadBalancing = &k0sv1beta1.NodeLocalLoadBalancing{}
	// }
	// if cfg.Spec.Network.NodeLocalLoadBalancing.EnvoyProxy == nil {
	// 	cfg.Spec.Network.NodeLocalLoadBalancing.EnvoyProxy = &k0sv1beta1.EnvoyProxy{}
	// }
	// if cfg.Spec.Network.NodeLocalLoadBalancing.EnvoyProxy.Image == nil {
	// 	cfg.Spec.Network.NodeLocalLoadBalancing.EnvoyProxy.Image = &k0sv1beta1.ImageSpec{}
	// }
	// cfg.Spec.Network.NodeLocalLoadBalancing.EnvoyProxy.Image.Image = getImageName("envoy-distroless")
	// cfg.Spec.Network.NodeLocalLoadBalancing.EnvoyProxy.Image.Version = Metadata.Images["envoy-distroless"]
}

func getImageName(component string) string {
	return fmt.Sprintf("proxy.replicated.com/anonymous/replicated/ec-%s", component)
}
