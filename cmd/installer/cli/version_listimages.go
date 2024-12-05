package cli

import (
	"context"
	"fmt"

	"github.com/replicatedhq/embedded-cluster/pkg/config"
	"github.com/spf13/cobra"
)

func VersionListImagesCmd(ctx context.Context, name string) *cobra.Command {
	var (
		omitReleaseMetadata bool
	)

	cmd := &cobra.Command{
		Use:           "list-images",
		Short:         "List images embedded in the cluster",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			k0sCfg := config.RenderK0sConfig()

			metadata, err := gatherVersionMetadata(k0sCfg, !omitReleaseMetadata)
			if err != nil {
				return fmt.Errorf("failed to gather version metadata: %w", err)
			}

			for _, image := range metadata.Images {
				fmt.Println(image)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&omitReleaseMetadata, "omit-release-metadata", false, "Omit the release metadata from the output")

	return cmd
}
