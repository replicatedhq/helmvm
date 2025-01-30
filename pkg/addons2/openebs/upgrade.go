package openebs

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/replicatedhq/embedded-cluster/pkg/helm"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (o *OpenEBS) Upgrade(ctx context.Context, kcli client.Client, hcli *helm.Helm, overrides []string) error {
	exists, err := hcli.ReleaseExists(ctx, namespace, releaseName)
	if err != nil {
		return errors.Wrap(err, "check if release exists")
	}
	if !exists {
		fmt.Printf("%s release not found in %s namespace, installing\n", releaseName, namespace)
		if err := o.Install(ctx, kcli, hcli, overrides, nil); err != nil {
			return errors.Wrap(err, "install")
		}
		return nil
	}

	if err := o.prepare(overrides); err != nil {
		return errors.Wrap(err, "prepare")
	}

	_, err = hcli.Upgrade(ctx, helm.UpgradeOptions{
		ReleaseName:  releaseName,
		ChartPath:    Metadata.Location,
		ChartVersion: Metadata.Version,
		Values:       helmValues,
		Namespace:    namespace,
		Force:        false,
	})
	if err != nil {
		return errors.Wrap(err, "upgrade")
	}

	return nil
}
