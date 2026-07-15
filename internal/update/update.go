package update

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

const repo = "tanishque-suthar/musicr"

type release struct {
	TagName string `json:"tag_name"`
}

func Check(currentVersion string) (latest string, needsUpdate bool, err error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo), nil)
	if err != nil {
		return "", false, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "musicr-updater")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", false, fmt.Errorf("fetching latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", false, fmt.Errorf("GitHub API returned %s", resp.Status)
	}

	var rel release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return "", false, fmt.Errorf("parsing release info: %w", err)
	}

	if rel.TagName == "" {
		return "", false, fmt.Errorf("no releases found")
	}

	if currentVersion == "dev" {
		return rel.TagName, true, nil
	}

	needsUpdate, err = versionGreater(rel.TagName, currentVersion)
	if err != nil {
		return "", false, err
	}

	return rel.TagName, needsUpdate, nil
}

func versionGreater(a, b string) (bool, error) {
	aVer := strings.TrimPrefix(a, "v")
	bVer := strings.TrimPrefix(b, "v")

	aParts := strings.Split(aVer, ".")
	bParts := strings.Split(bVer, ".")

	maxLen := len(aParts)
	if len(bParts) > maxLen {
		maxLen = len(bParts)
	}

	for i := 0; i < maxLen; i++ {
		var an, bn int
		if i < len(aParts) {
			an, _ = strconv.Atoi(aParts[i])
		}
		if i < len(bParts) {
			bn, _ = strconv.Atoi(bParts[i])
		}
		if an > bn {
			return true, nil
		}
		if an < bn {
			return false, nil
		}
	}
	return false, nil
}

func Apply(latestTag string) error {
	version := strings.TrimPrefix(latestTag, "v")
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	switch goarch {
	case "amd64", "arm64":
	default:
		return fmt.Errorf("unsupported architecture: %s", goarch)
	}

	switch goos {
	case "linux", "darwin":
	default:
		return fmt.Errorf("unsupported OS: %s", goos)
	}

	tarballName := fmt.Sprintf("musicr_%s_%s_%s.tar.gz", version, goos, goarch)
	downloadURL := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", repo, latestTag, tarballName)
	checksumsURL := fmt.Sprintf("https://github.com/%s/releases/download/%s/checksums.txt", repo, latestTag)

	tmpDir, err := os.MkdirTemp("", "musicr-update-*")
	if err != nil {
		return fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	tarballPath := filepath.Join(tmpDir, tarballName)
	if err := downloadFile(tarballPath, downloadURL); err != nil {
		return fmt.Errorf("downloading update: %w", err)
	}

	checksumsPath := filepath.Join(tmpDir, "checksums.txt")
	if err := downloadFile(checksumsPath, checksumsURL); err != nil {
		return fmt.Errorf("downloading checksums: %w", err)
	}

	if err := verifyChecksum(tarballPath, tarballName, checksumsPath); err != nil {
		return fmt.Errorf("checksum verification failed: %w", err)
	}

	extractedPath := filepath.Join(tmpDir, "musicr")
	if err := extractTarGz(tarballPath, extractedPath); err != nil {
		return fmt.Errorf("extracting archive: %w", err)
	}

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("getting executable path: %w", err)
	}

	realPath, err := filepath.EvalSymlinks(execPath)
	if err == nil {
		execPath = realPath
	}

	dir := filepath.Dir(execPath)
	tmpNew := filepath.Join(dir, ".musicr.new")

	if err := copyFile(extractedPath, tmpNew); err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied. Try: sudo musicr --update")
		}
		return fmt.Errorf("writing new binary: %w", err)
	}

	if err := os.Chmod(tmpNew, 0755); err != nil {
		os.Remove(tmpNew)
		return err
	}

	if err := os.Rename(tmpNew, execPath); err != nil {
		os.Remove(tmpNew)
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied. Try: sudo musicr --update")
		}
		return fmt.Errorf("replacing binary: %w", err)
	}

	return nil
}

func downloadFile(dest, url string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "musicr-updater")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %s", resp.Status)
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}

func verifyChecksum(tarballPath, tarballName, checksumsPath string) error {
	data, err := os.ReadFile(checksumsPath)
	if err != nil {
		return err
	}

	var expectedSum string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[1] == tarballName {
			expectedSum = fields[0]
			break
		}
	}

	if expectedSum == "" {
		return fmt.Errorf("checksum not found for %s", tarballName)
	}

	f, err := os.Open(tarballPath)
	if err != nil {
		return err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}

	actualSum := fmt.Sprintf("%x", h.Sum(nil))
	if actualSum != expectedSum {
		return fmt.Errorf("mismatch: got %s, expected %s", actualSum, expectedSum)
	}

	return nil
}

func extractTarGz(src, dest string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if header.Name == "musicr" {
			out, err := os.Create(dest)
			if err != nil {
				return err
			}
			defer out.Close()

			_, err = io.Copy(out, tr)
			return err
		}
	}

	return fmt.Errorf("binary not found in archive")
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, info.Mode())
}
