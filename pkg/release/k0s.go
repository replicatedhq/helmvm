package release

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const k0sMetadataPreface = `#
# this file is automatically generated by buildtools. manual edits are not recommended.
# to regenerate this file, run the following commands:
#
# $ make buildtools
# $ output/bin/buildtools update images k0s
#
`

type K0sMetadata struct {
	Images map[string]AddonImage `yaml:"images"`
}

func (a *K0sMetadata) Save() error {
	buf := bytes.NewBufferString(k0sMetadataPreface)
	if err := yaml.NewEncoder(buf).Encode(a); err != nil {
		return fmt.Errorf("failed to encode k0s metadata: %w", err)
	}
	fpath := filepath.Join("pkg", "config", "static", "metadata.yaml")
	if err := os.WriteFile(fpath, buf.Bytes(), 0600); err != nil {
		return fmt.Errorf("failed to write k0s metadata: %w", err)
	}
	return nil
}
