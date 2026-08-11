package builtin

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/prjvvl/exe/internal/config"
	"github.com/prjvvl/exe/internal/dispatch"
	"github.com/prjvvl/exe/internal/version"
)

const (
	latestReleaseAPI   = "https://api.github.com/repos/prjvvl/exe/releases/latest"
	checksumsAssetName = "checksums.txt"
)

var (
	metadataClient = &http.Client{Timeout: 15 * time.Second}
	downloadClient = &http.Client{Timeout: 2 * time.Minute}
)

func UpdateBuiltin() Builtin {
	return Builtin{
		Name:        "update",
		Description: "Check for and install the latest exe release",
		Handler: func(cfg *config.Config, args []string) (dispatch.Resolved, error) {
			msg, err := runUpdate()
			if err != nil {
				return dispatch.Resolved{}, err
			}
			return dispatch.Resolved{Kind: dispatch.KindPrint, Text: msg}, nil
		},
	}
}

type ghAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type ghRelease struct {
	TagName string    `json:"tag_name"`
	Assets  []ghAsset `json:"assets"`
}

func runUpdate() (string, error) {
	if version.Version == "dev" {
		return "Running a dev build (no version embedded), so there's nothing to compare against. " +
			"Install a tagged release instead, or build via goreleaser, which embeds a real version.", nil
	}

	release, err := fetchLatestRelease()
	if err != nil {
		return "", fmt.Errorf("checking latest release: %w", err)
	}

	if release.TagName == version.Version {
		return fmt.Sprintf("Already up to date (%s).", version.Version), nil
	}

	assetName := archiveName(runtime.GOOS, runtime.GOARCH)
	assetURL, err := findAssetURL(release, assetName)
	if err != nil {
		return "", err
	}

	checksumsURL, err := findAssetURL(release, checksumsAssetName)
	if err != nil {
		return "", err
	}

	expectedSum, err := fetchExpectedChecksum(checksumsURL, assetName)
	if err != nil {
		return "", fmt.Errorf("fetching checksums for %s: %w", release.TagName, err)
	}

	currentExe, err := os.Executable()
	if err != nil {
		return "", err
	}

	// Extract into the same directory as the running binary: the final
	// install step is an os.Rename, which fails with "cross-device link"
	// if source and destination are on different filesystems/drives (e.g.
	// the extracted file landing in the OS temp dir while exe lives
	// elsewhere).
	newBinary, err := downloadAndExtract(assetURL, expectedSum, filepath.Dir(currentExe))
	if err != nil {
		return "", fmt.Errorf("downloading %s: %w", release.TagName, err)
	}
	defer os.Remove(newBinary)

	if err := replaceBinary(currentExe, newBinary); err != nil {
		return "", fmt.Errorf("installing %s: %w", release.TagName, err)
	}

	return fmt.Sprintf("Updated exe %s -> %s.", version.Version, release.TagName), nil
}

func archiveName(goos, arch string) string {
	if goos == "windows" {
		return fmt.Sprintf("exe_windows_%s.zip", arch)
	}
	return fmt.Sprintf("exe_%s_%s.tar.gz", goos, arch)
}

func findAssetURL(release *ghRelease, name string) (string, error) {
	for _, a := range release.Assets {
		if a.Name == name {
			return a.BrowserDownloadURL, nil
		}
	}
	return "", fmt.Errorf("no release asset named %q in %s", name, release.TagName)
}

func fetchLatestRelease() (*ghRelease, error) {
	resp, err := metadataClient.Get(latestReleaseAPI)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api returned %s", resp.Status)
	}
	var release ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}
	return &release, nil
}

// fetchExpectedChecksum looks up assetName's sha256 in goreleaser's
// checksums.txt (format: "<sha256>  <filename>" per line).
func fetchExpectedChecksum(checksumsURL, assetName string) (string, error) {
	resp, err := metadataClient.Get(checksumsURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("checksums.txt returned %s", resp.Status)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 2 && fields[1] == assetName {
			return fields[0], nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("%s not listed in checksums.txt", assetName)
}

func downloadAndExtract(url, expectedSum, destDir string) (string, error) {
	archivePath, err := downloadToTemp(url)
	if err != nil {
		return "", err
	}
	defer os.Remove(archivePath)

	if err := verifyChecksum(archivePath, expectedSum); err != nil {
		return "", err
	}

	binaryName := "exe"
	if runtime.GOOS == "windows" {
		binaryName = "exe.exe"
	}

	if runtime.GOOS == "windows" {
		return extractFromZip(archivePath, binaryName, destDir)
	}
	return extractFromTarGz(archivePath, binaryName, destDir)
}

func verifyChecksum(path, expectedSum string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}

	got := hex.EncodeToString(h.Sum(nil))
	if !strings.EqualFold(got, expectedSum) {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedSum, got)
	}
	return nil
}

func downloadToTemp(url string) (string, error) {
	resp, err := downloadClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download returned %s", resp.Status)
	}

	f, err := os.CreateTemp("", "exe-update-archive-*")
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		os.Remove(f.Name())
		return "", err
	}
	return f.Name(), nil
}

func extractFromZip(archivePath, binaryName, destDir string) (string, error) {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", err
	}
	defer r.Close()

	for _, zf := range r.File {
		if filepath.Base(zf.Name) != binaryName {
			continue
		}
		rc, err := zf.Open()
		if err != nil {
			return "", err
		}
		defer rc.Close()
		return copyToTempFile(rc, destDir)
	}
	return "", fmt.Errorf("binary %q not found in archive", binaryName)
}

func extractFromTarGz(archivePath, binaryName, destDir string) (string, error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return "", err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		if filepath.Base(hdr.Name) != binaryName {
			continue
		}
		return copyToTempFile(tr, destDir)
	}
	return "", fmt.Errorf("binary %q not found in archive", binaryName)
}

func copyToTempFile(r io.Reader, dir string) (string, error) {
	out, err := os.CreateTemp(dir, "exe-update-bin-*")
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err := io.Copy(out, r); err != nil {
		os.Remove(out.Name())
		return "", err
	}
	return out.Name(), nil
}

// Windows won't allow overwriting a running exe directly, so it's moved
// aside first; Unix allows a plain rename over the running binary.
func replaceBinary(currentPath, newPath string) error {
	if err := os.Chmod(newPath, 0o755); err != nil {
		return err
	}

	if runtime.GOOS != "windows" {
		return os.Rename(newPath, currentPath)
	}

	oldPath := currentPath + ".old"
	os.Remove(oldPath)
	if err := os.Rename(currentPath, oldPath); err != nil {
		return err
	}
	if err := os.Rename(newPath, currentPath); err != nil {
		os.Rename(oldPath, currentPath) // best-effort restore
		return err
	}
	_ = os.Remove(oldPath)
	return nil
}
