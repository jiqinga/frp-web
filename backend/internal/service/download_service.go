package service

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type DownloadService struct {
	githubAPI string
	mirrorID  *uint
}

type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func NewDownloadService(githubAPI string) *DownloadService {
	return &DownloadService{githubAPI: githubAPI}
}

func (ds *DownloadService) SetMirrorID(mirrorID *uint) {
	ds.mirrorID = mirrorID
}

func (ds *DownloadService) DownloadFrps(version, targetDir string) (string, error) {
	if version == "" || version == "latest" {
		var err error
		version, err = ds.getLatestVersion()
		if err != nil {
			return "", err
		}
	}

	downloadURL, fileName, err := ds.getDownloadURL(version)
	if err != nil {
		return "", err
	}

	mirrorService := NewGithubMirrorService()
	downloadURL, err = mirrorService.ConvertGithubURL(downloadURL, ds.mirrorID)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("创建目录失败: %v", err)
	}

	archivePath := filepath.Join(targetDir, fileName)
	if err := ds.downloadFile(downloadURL, archivePath); err != nil {
		return "", err
	}

	binaryPath, err := ds.extractFrps(archivePath, targetDir)
	if err != nil {
		return "", err
	}

	os.Remove(archivePath)
	return binaryPath, nil
}

func (ds *DownloadService) getLatestVersion() (string, error) {
	resp, err := http.Get(ds.githubAPI + "/releases/latest")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}
	return release.TagName, nil
}

func (ds *DownloadService) getDownloadURL(version string) (string, string, error) {
	url := fmt.Sprintf("%s/releases/tags/%s", ds.githubAPI, version)
	resp, err := http.Get(url)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", "", err
	}

	osName := runtime.GOOS
	arch := runtime.GOARCH
	pattern := fmt.Sprintf("frp_%s_%s_%s", strings.TrimPrefix(version, "v"), osName, arch)

	for _, asset := range release.Assets {
		if strings.Contains(asset.Name, pattern) {
			return asset.BrowserDownloadURL, asset.Name, nil
		}
	}
	return "", "", fmt.Errorf("未找到适合的版本")
}

func (ds *DownloadService) downloadFile(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func (ds *DownloadService) extractFrps(archivePath, targetDir string) (string, error) {
	if strings.HasSuffix(archivePath, ".zip") {
		return ds.extractZip(archivePath, targetDir)
	}
	return ds.extractTarGz(archivePath, targetDir)
}

func (ds *DownloadService) extractZip(zipPath, targetDir string) (string, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", err
	}
	defer r.Close()

	var binaryPath string
	for _, f := range r.File {
		if strings.HasSuffix(f.Name, "frps") || strings.HasSuffix(f.Name, "frps.exe") {
			rc, err := f.Open()
			if err != nil {
				return "", err
			}
			defer rc.Close()

			binaryPath = filepath.Join(targetDir, filepath.Base(f.Name))
			out, err := os.OpenFile(binaryPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
			if err != nil {
				return "", err
			}
			defer out.Close()

			if _, err = io.Copy(out, rc); err != nil {
				return "", err
			}
			break
		}
	}
	return binaryPath, nil
}

func (ds *DownloadService) extractTarGz(tarGzPath, targetDir string) (string, error) {
	file, err := os.Open(tarGzPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return "", err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	var binaryPath string

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		if strings.HasSuffix(header.Name, "frps") || strings.HasSuffix(header.Name, "frps.exe") {
			binaryPath = filepath.Join(targetDir, filepath.Base(header.Name))
			out, err := os.OpenFile(binaryPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
			if err != nil {
				return "", err
			}
			defer out.Close()

			if _, err := io.Copy(out, tr); err != nil {
				return "", err
			}
			break
		}
	}
	return binaryPath, nil
}
