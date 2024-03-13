// Package goods handles embedded static assets. Things like writing them
// down to disk, return them as a parsed list, etc.
package goods

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	"github.com/replicatedhq/embedded-cluster/pkg/defaults"
)

// K0sBinarySHA256 returns the SHA256 checksum of the embedded k0s binary.
func K0sBinarySHA256() (string, error) {
	fp, err := binfs.Open("bins/k0s")
	if err != nil {
		return "", fmt.Errorf("unable to open embedded k0s binary: %w", err)
	}
	defer fp.Close()
	hasher := sha256.New()
	if _, err := io.Copy(hasher, fp); err != nil {
		return "", fmt.Errorf("unable to copy embedded k0s binary: %w", err)
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

//go:embed bins/*
var binfs embed.FS

// materializeBinaries materializes all binary files from inside bins directory. If the
// file already exists a copy of it is made first before overwriting it, this is done
// because we can't overwrite a running binary. Copies are removed.
func materializeBinaries() error {
	entries, err := binfs.ReadDir("bins")
	if err != nil {
		return fmt.Errorf("unable to read embedded-cluster bins dir: %w", err)
	}

	var remove []string
	defer func() {
		for _, f := range remove {
			os.Remove(f)
		}
	}()

	for _, entry := range entries {
		srcpath := fmt.Sprintf("bins/%s", entry.Name())
		srcfile, err := binfs.ReadFile(srcpath)
		if err != nil {
			return fmt.Errorf("unable to read asset: %w", err)
		}

		dstpath := defaults.PathToEmbeddedClusterBinary(entry.Name())
		if _, err := os.Stat(dstpath); err == nil {
			tmp := fmt.Sprintf("%s.bkp", dstpath)
			if err := os.Rename(dstpath, tmp); err != nil {
				return fmt.Errorf("unable to rename %s to %s: %w", dstpath, tmp, err)
			}
			remove = append(remove, tmp)
		}

		if err := os.WriteFile(dstpath, srcfile, 0755); err != nil {
			return fmt.Errorf("unable to write file: %w", err)
		}
	}
	return nil
}

//go:embed support/*
var supportfs embed.FS

// materializeSupportFiles materializes all support files from inside support directory.
func materializeSupportFiles() error {
	entries, err := supportfs.ReadDir("support")
	if err != nil {
		return fmt.Errorf("unable to read embedded-cluster support dir: %w", err)
	}
	for _, entry := range entries {
		srcpath := fmt.Sprintf("support/%s", entry.Name())
		srcfile, err := supportfs.ReadFile(srcpath)
		if err != nil {
			return fmt.Errorf("unable to read asset: %w", err)
		}
		dstpath := defaults.PathToEmbeddedClusterSupportFile(entry.Name())
		if err := os.WriteFile(dstpath, srcfile, 0700); err != nil {
			return fmt.Errorf("unable to write file: %w", err)
		}
	}
	return nil
}

// Materialize writes to disk all embedded assets.
func Materialize() error {
	if err := materializeBinaries(); err != nil {
		return fmt.Errorf("unable to materialize embedded binaries: %w", err)
	}
	if err := materializeSupportFiles(); err != nil {
		return fmt.Errorf("unable to materialize embedded support files: %w", err)
	}
	return nil
}

//go:embed systemd/*
var systemdfs embed.FS

// MaterializeLocalArtifactMirrorUnitFile writes to disk the local-artifact-mirror systemd unit file.
func MaterializeLocalArtifactMirrorUnitFile() error {
	content, err := systemdfs.ReadFile("systemd/local-artifact-mirror.service")
	if err != nil {
		return fmt.Errorf("unable to open unit file: %w", err)
	}
	dstpath := "/etc/systemd/system/local-artifact-mirror.service"
	if err := os.WriteFile(dstpath, content, 0644); err != nil {
		return fmt.Errorf("unable to write file: %w", err)
	}
	return nil
}
