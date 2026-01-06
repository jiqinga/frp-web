/*
 * @Author              : 寂情啊
 * @Date                : 2025-12-26 16:59:13
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-30 14:42:00
 * @FilePath            : frp-web-testbackendinternalservicefrp_server_config_service.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
/*
 * FrpServerService - 配置生成和下载方法
 */
package service

import (
	"fmt"
	"frp-web-panel/internal/config"
	"frp-web-panel/internal/model"
	"os"
	"path/filepath"
)

func (s *FrpServerService) Download(id uint, version string) error {
	server, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if server.ServerType == model.ServerTypeLocal {
		return fmt.Errorf("本地服务器不支持下载操作")
	}

	if version == "" {
		version = config.GlobalConfig.Frps.DefaultVersion
	}

	binaryDir := config.GlobalConfig.Frps.BinaryDir
	if binaryDir == "" {
		binaryDir = "./data/frps"
	}

	s.downloadService.SetMirrorID(server.MirrorID)
	targetDir := filepath.Join(binaryDir, fmt.Sprintf("server_%d", id))
	binaryPath, err := s.downloadService.DownloadFrps(version, targetDir)
	if err != nil {
		return err
	}

	server.BinaryPath = binaryPath
	server.Version = version
	return s.repo.Update(server)
}

func (s *FrpServerService) generateConfig(server *model.FrpServer) error {
	configDir := config.GlobalConfig.Frps.ConfigDir
	if configDir == "" {
		configDir = "./data/frps/configs"
	}

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	configPath := filepath.Join(configDir, fmt.Sprintf("frps_%d.yaml", server.ID))
	content := fmt.Sprintf(`bindPort: %d
enablePrometheus: true
webServer:
	 addr: "0.0.0.0"
	 port: %d
	 user: "%s"
	 password: "%s"
`, server.BindPort, server.DashboardPort, server.DashboardUser, server.DashboardPwd)

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return err
	}

	server.ConfigPath = configPath
	return nil
}
