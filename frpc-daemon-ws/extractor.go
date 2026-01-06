/*
 * @Author              : Kilo Code
 * @Date                : 2025-12-01
 * @Description         : 客户端更新器 - 解压功能
 */
package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// extractFrpc 解压 frpc
func (u *Updater) extractFrpc(archivePath, targetDir string) (string, error) {
	if strings.HasSuffix(archivePath, ".zip") {
		return u.extractZip(archivePath, targetDir)
	}
	return u.extractTarGz(archivePath, targetDir)
}

// extractZipFile 提取单个 zip 文件条目，避免 defer 在循环内累积
func (u *Updater) extractZipFile(f *zip.File, targetDir string) (string, error) {
	baseName := filepath.Base(f.Name)
	rc, err := f.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()

	frpcPath := filepath.Join(targetDir, baseName)
	out, err := os.OpenFile(frpcPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err = io.Copy(out, rc); err != nil {
		return "", err
	}
	return frpcPath, nil
}

// extractZip 解压 zip 文件
func (u *Updater) extractZip(zipPath, targetDir string) (string, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", err
	}
	defer r.Close()

	for _, f := range r.File {
		baseName := filepath.Base(f.Name)
		if baseName == "frpc" || baseName == "frpc.exe" {
			return u.extractZipFile(f, targetDir)
		}
	}

	return "", fmt.Errorf("在压缩包中未找到 frpc")
}

// extractTarGz 解压 tar.gz 文件
func (u *Updater) extractTarGz(tarGzPath, targetDir string) (string, error) {
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
	var frpcPath string

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		baseName := filepath.Base(header.Name)
		if baseName == "frpc" || baseName == "frpc.exe" {
			frpcPath = filepath.Join(targetDir, baseName)
			out, err := os.OpenFile(frpcPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
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

	if frpcPath == "" {
		return "", fmt.Errorf("在压缩包中未找到 frpc")
	}
	return frpcPath, nil
}
