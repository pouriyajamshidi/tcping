package main

import (
	"archive/tar"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMakefileDebPackagingConfig verifies that Makefile configures dpkg-deb with --root-owner-group.
func TestMakefileDebPackagingConfig(t *testing.T) {
	content, err := os.ReadFile("Makefile")
	require.NoError(t, err, "Makefile must be readable")

	makefileText := string(content)

	assert.Contains(t, makefileText, "dpkg-deb --root-owner-group --build",
		"Makefile deb target must pass --root-owner-group to dpkg-deb to ensure root:root package ownership")

	assert.Contains(t, makefileText, "chmod 755 $$PKG_DIR",
		"Makefile deb target must ensure package root staging directory has 0755 permissions")
}

// TestDebPackageOwnershipWithRootOwnerGroup verifies that dpkg-deb --root-owner-group forces uid:gid 0:0
// on packages even when staging files are owned by a non-root user (such as uid/gid 1000).
func TestDebPackageOwnershipWithRootOwnerGroup(t *testing.T) {
	dpkgDebPath, err := exec.LookPath("dpkg-deb")
	if err != nil {
		t.Skip("dpkg-deb is not available in PATH; skipping deb package ownership integration test")
	}

	tempDir := t.TempDir()

	stagingDir := filepath.Join(tempDir, "pkg-staging")
	binDir := filepath.Join(stagingDir, "usr", "bin")
	debianDir := filepath.Join(stagingDir, "DEBIAN")

	err = os.MkdirAll(binDir, 0755)
	require.NoError(t, err)
	err = os.MkdirAll(debianDir, 0755)
	require.NoError(t, err)

	binaryPath := filepath.Join(binDir, "tcping")
	err = os.WriteFile(binaryPath, []byte("#!/bin/sh\necho tcping\n"), 0755)
	require.NoError(t, err)

	controlContent := "Package: tcping\nVersion: 2.8.0\nSection: custom\nPriority: optional\nArchitecture: all\nEssential: no\nMaintainer: test\nDescription: test\n"
	err = os.WriteFile(filepath.Join(debianDir, "control"), []byte(controlContent), 0644)
	require.NoError(t, err)

	// Attempt chown to 1000:1000 to simulate a non-root build user
	canChown := os.Chown(binaryPath, 1000, 1000) == nil
	if canChown {
		_ = os.Chown(stagingDir, 1000, 1000)
		_ = os.Chown(binDir, 1000, 1000)
		_ = os.Chown(filepath.Join(stagingDir, "usr"), 1000, 1000)
	}

	// 1. Build deb package WITH --root-owner-group
	fixedDebPath := filepath.Join(tempDir, "fixed.deb")
	cmdFixed := exec.Command(dpkgDebPath, "--root-owner-group", "--build", stagingDir, fixedDebPath)
	fixedOutput, err := cmdFixed.CombinedOutput()
	require.NoError(t, err, "dpkg-deb --root-owner-group failed: %s", string(fixedOutput))

	fixedEntries := readDebEntries(t, dpkgDebPath, fixedDebPath)
	require.NotEmpty(t, fixedEntries, "deb package archive entries must not be empty")

	for _, entry := range fixedEntries {
		assert.Equal(t, 0, entry.UID, "entry %s must have UID 0 (root)", entry.Name)
		assert.Equal(t, 0, entry.GID, "entry %s must have GID 0 (root)", entry.Name)
		if entry.Uname != "" {
			assert.Equal(t, "root", entry.Uname, "entry %s user name should be root", entry.Name)
		}
		if entry.Gname != "" {
			assert.Equal(t, "root", entry.Gname, "entry %s group name should be root", entry.Name)
		}
	}

	// Verify the binary specifically
	foundBinary := false
	for _, entry := range fixedEntries {
		if strings.HasSuffix(entry.Name, "usr/bin/tcping") {
			foundBinary = true
			assert.Equal(t, 0, entry.UID, "tcping binary UID must be 0")
			assert.Equal(t, 0, entry.GID, "tcping binary GID must be 0")
			assert.Equal(t, int64(0755), entry.Mode&0777, "tcping binary mode must be 0755")
		}
	}
	assert.True(t, foundBinary, "package must contain usr/bin/tcping")

	// 2. If chown succeeded, verify that WITHOUT --root-owner-group the bug manifests (uid 1000 inherited)
	if canChown {
		unfixedDebPath := filepath.Join(tempDir, "unfixed.deb")
		cmdUnfixed := exec.Command(dpkgDebPath, "--build", stagingDir, unfixedDebPath)
		unfixedOutput, err := cmdUnfixed.CombinedOutput()
		require.NoError(t, err, "dpkg-deb --build failed: %s", string(unfixedOutput))

		unfixedEntries := readDebEntries(t, dpkgDebPath, unfixedDebPath)
		hasUID1000 := false
		for _, entry := range unfixedEntries {
			if strings.HasSuffix(entry.Name, "usr/bin/tcping") && entry.UID == 1000 {
				hasUID1000 = true
				break
			}
		}
		assert.True(t, hasUID1000, "without --root-owner-group, non-root uid 1000 is inherited by deb package")
	}
}

// TestDebPackageArtifactOwnership inspects built deb packages in output/ if present.
func TestDebPackageArtifactOwnership(t *testing.T) {
	dpkgDebPath, err := exec.LookPath("dpkg-deb")
	if err != nil {
		t.Skip("dpkg-deb is not available in PATH; skipping deb package artifact ownership test")
	}

	debFiles, err := filepath.Glob(filepath.Join("output", "*.deb"))
	if err != nil || len(debFiles) == 0 {
		t.Skip("no .deb files in output/ directory; skipping deb package artifact test")
	}

	for _, debPath := range debFiles {
		if strings.HasPrefix(filepath.Base(debPath), ".") {
			continue
		}
		t.Run(filepath.Base(debPath), func(t *testing.T) {
			entries := readDebEntries(t, dpkgDebPath, debPath)
			require.NotEmpty(t, entries, "deb package entries should not be empty")

			foundBinary := false
			for _, entry := range entries {
				assert.Equal(t, 0, entry.UID, "file %s in %s must have UID 0", entry.Name, debPath)
				assert.Equal(t, 0, entry.GID, "file %s in %s must have GID 0", entry.Name, debPath)
				if strings.HasSuffix(entry.Name, "usr/bin/tcping") {
					foundBinary = true
					assert.Equal(t, int64(0755), entry.Mode&0777, "tcping binary in %s must have mode 0755", debPath)
				}
			}
			assert.True(t, foundBinary, "%s must contain usr/bin/tcping", debPath)
		})
	}
}

type debTarEntry struct {
	Name  string
	Mode  int64
	UID   int
	GID   int
	Uname string
	Gname string
}

func readDebEntries(t *testing.T, dpkgDebPath, debPath string) []debTarEntry {
	t.Helper()

	cmd := exec.Command(dpkgDebPath, "--fsys-tarfile", debPath)
	stdout, err := cmd.StdoutPipe()
	require.NoError(t, err)

	err = cmd.Start()
	require.NoError(t, err)

	var entries []debTarEntry
	tr := tar.NewReader(stdout)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)

		entries = append(entries, debTarEntry{
			Name:  hdr.Name,
			Mode:  hdr.Mode,
			UID:   hdr.Uid,
			GID:   hdr.Gid,
			Uname: hdr.Uname,
			Gname: hdr.Gname,
		})
	}

	err = cmd.Wait()
	require.NoError(t, err)

	return entries
}
