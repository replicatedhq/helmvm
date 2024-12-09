package k0s

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	k0sconfig "github.com/k0sproject/k0s/pkg/apis/k0s/v1beta1"
	"github.com/replicatedhq/embedded-cluster/pkg/airgap"
	"github.com/replicatedhq/embedded-cluster/pkg/config"
	"github.com/replicatedhq/embedded-cluster/pkg/helpers"
	"github.com/replicatedhq/embedded-cluster/pkg/netutils"
	"github.com/replicatedhq/embedded-cluster/pkg/release"
	"github.com/replicatedhq/embedded-cluster/pkg/runtimeconfig"
	k8syaml "sigs.k8s.io/yaml"
)

// Install runs the k0s install command and waits for it to finish. If no configuration
// is found one is generated.
func Install(networkInterface string) error {
	ourbin := runtimeconfig.PathToEmbeddedClusterBinary("k0s")
	hstbin := runtimeconfig.K0sBinaryPath()
	if err := helpers.MoveFile(ourbin, hstbin); err != nil {
		return fmt.Errorf("unable to move k0s binary: %w", err)
	}

	nodeIP, err := netutils.FirstValidAddress(networkInterface)
	if err != nil {
		return fmt.Errorf("unable to find first valid address: %w", err)
	}
	if _, err := helpers.RunCommand(hstbin, config.InstallFlags(nodeIP)...); err != nil {
		return fmt.Errorf("unable to install: %w", err)
	}
	if _, err := helpers.RunCommand(hstbin, "start"); err != nil {
		return fmt.Errorf("unable to start: %w", err)
	}
	return nil
}

// IsInstalled checks if the embedded cluster is already installed by looking for
// the k0s configuration file existence.
func IsInstalled() (bool, error) {
	_, err := os.Stat(runtimeconfig.PathToK0sConfig())
	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	}

	return false, fmt.Errorf("unable to check if already installed: %w", err)
}

// WriteK0sConfig creates a new k0s.yaml configuration file. The file is saved in the
// global location (as returned by runtimeconfig.PathToK0sConfig()). If a file already sits
// there, this function returns an error.
func WriteK0sConfig(ctx context.Context, networkInterface string, airgapBundle string, podCIDR string, serviceCIDR string, overrides string) (*k0sconfig.ClusterConfig, error) {
	cfgpath := runtimeconfig.PathToK0sConfig()
	if _, err := os.Stat(cfgpath); err == nil {
		return nil, fmt.Errorf("configuration file already exists")
	}
	if err := os.MkdirAll(filepath.Dir(cfgpath), 0755); err != nil {
		return nil, fmt.Errorf("unable to create directory: %w", err)
	}
	cfg := config.RenderK0sConfig()

	address, err := netutils.FirstValidAddress(networkInterface)
	if err != nil {
		return nil, fmt.Errorf("unable to find first valid address: %w", err)
	}
	cfg.Spec.API.Address = address
	cfg.Spec.Storage.Etcd.PeerAddress = address

	cfg.Spec.Network.PodCIDR = podCIDR
	cfg.Spec.Network.ServiceCIDR = serviceCIDR

	cfg, err = applyUnsupportedOverrides(ctx, overrides, cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to apply unsupported overrides: %w", err)
	}

	if airgapBundle != "" {
		// update the k0s config to install with airgap
		airgap.RemapHelm(cfg)
		airgap.SetAirgapConfig(cfg)
	}
	// This is necessary to install the previous version of k0s in e2e tests
	// TODO: remove this once the previous version is > 1.29
	unstructured, err := helpers.K0sClusterConfigTo129Compat(cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to convert cluster config to 1.29 compat: %w", err)
	}
	data, err := k8syaml.Marshal(unstructured)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal config: %w", err)
	}
	if err := os.WriteFile(cfgpath, data, 0600); err != nil {
		return nil, fmt.Errorf("unable to write config file: %w", err)
	}
	return cfg, nil
}

// applyUnsupportedOverrides applies overrides to the k0s configuration. Applies first the
// overrides embedded into the binary and after the ones provided by the user (--overrides).
// we first apply the k0s config override and then apply the built in overrides.
func applyUnsupportedOverrides(ctx context.Context, overrides string, cfg *k0sconfig.ClusterConfig) (*k0sconfig.ClusterConfig, error) {
	embcfg, err := release.GetEmbeddedClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to get embedded cluster config: %w", err)
	}
	if embcfg != nil {
		overrides := embcfg.Spec.UnsupportedOverrides.K0s
		cfg, err = config.PatchK0sConfig(cfg, overrides)
		if err != nil {
			return nil, fmt.Errorf("unable to patch k0s config: %w", err)
		}
		cfg, err = config.ApplyBuiltInExtensionsOverrides(cfg, embcfg)
		if err != nil {
			return nil, fmt.Errorf("unable to release built in overrides: %w", err)
		}
	}

	eucfg, err := helpers.ParseEndUserConfig(overrides)
	if err != nil {
		return nil, fmt.Errorf("unable to process overrides file: %w", err)
	}
	if eucfg != nil {
		overrides := eucfg.Spec.UnsupportedOverrides.K0s
		cfg, err = config.PatchK0sConfig(cfg, overrides)
		if err != nil {
			return nil, fmt.Errorf("unable to apply overrides: %w", err)
		}
		cfg, err = config.ApplyBuiltInExtensionsOverrides(cfg, eucfg)
		if err != nil {
			return nil, fmt.Errorf("unable to end user built in overrides: %w", err)
		}
	}

	return cfg, nil
}
