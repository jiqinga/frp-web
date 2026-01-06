/*
 * FrpServerService - 本地进程管理和状态检查方法
 */
package service

import (
	"fmt"
	"frp-web-panel/internal/frp"
	"frp-web-panel/internal/model"
)

func (s *FrpServerService) TestConnection(server *model.FrpServer) error {
	dashboardHost := server.Host
	if server.ServerType == model.ServerTypeRemote && (server.Host == "" || server.Host == "0.0.0.0") {
		dashboardHost = server.SSHHost
	}

	client := frp.NewFrpsClient(dashboardHost, server.DashboardPort, server.DashboardUser, server.DashboardPwd)
	err := client.HealthCheck()

	if server.ID > 0 {
		dbServer, getErr := s.repo.GetByID(server.ID)
		if getErr == nil {
			if err == nil {
				dbServer.Status = model.StatusRunning
			} else {
				dbServer.Status = model.StatusStopped
			}
			s.repo.Update(dbServer)
		}
	}
	return err
}

func (s *FrpServerService) Start(id uint) error {
	server, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if server.ServerType == model.ServerTypeLocal {
		return fmt.Errorf("本地服务器不支持启动操作,请使用远程启动或在服务器上手动启动")
	}
	if server.BinaryPath == "" {
		return fmt.Errorf("请先下载frps二进制文件")
	}
	if server.ConfigPath == "" {
		if err := s.generateConfig(server); err != nil {
			return err
		}
	}

	server.Status = model.StatusStarting
	s.repo.Update(server)

	if err := s.processManager.Start(server); err != nil {
		server.Status = model.StatusError
		server.LastError = err.Error()
		s.repo.Update(server)
		return err
	}
	return s.repo.Update(server)
}

func (s *FrpServerService) Stop(id uint) error {
	server, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if server.ServerType == model.ServerTypeLocal {
		return fmt.Errorf("本地服务器不支持停止操作,请在服务器上手动停止")
	}

	server.Status = model.StatusStopping
	s.repo.Update(server)

	if err := s.processManager.Stop(server); err != nil {
		server.Status = model.StatusError
		server.LastError = err.Error()
		s.repo.Update(server)
		return err
	}
	return s.repo.Update(server)
}

func (s *FrpServerService) Restart(id uint) error {
	server, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if server.ServerType == model.ServerTypeLocal {
		return fmt.Errorf("本地服务器不支持重启操作,请在服务器上手动重启")
	}

	if err := s.processManager.Restart(server); err != nil {
		server.Status = model.StatusError
		server.LastError = err.Error()
		s.repo.Update(server)
		return err
	}

	server.Status = model.StatusRunning
	return s.repo.Update(server)
}

func (s *FrpServerService) GetStatus(id uint) (model.FrpServerStatus, error) {
	server, err := s.repo.GetByID(id)
	if err != nil {
		return model.StatusError, err
	}

	dashboardHost := server.Host
	if server.ServerType == model.ServerTypeRemote && (server.Host == "" || server.Host == "0.0.0.0") {
		dashboardHost = server.SSHHost
	}

	client := frp.NewFrpsClient(dashboardHost, server.DashboardPort, server.DashboardUser, server.DashboardPwd)
	if err := client.HealthCheck(); err != nil {
		server.Status = model.StatusStopped
	} else {
		server.Status = model.StatusRunning
	}

	s.repo.Update(server)
	return server.Status, nil
}

func (s *FrpServerService) GetLocalVersion(id uint) (string, error) {
	server, err := s.repo.GetByID(id)
	if err != nil {
		return "", err
	}
	if server.ServerType != model.ServerTypeLocal {
		return "", fmt.Errorf("只有本地服务器支持此操作")
	}

	host := server.Host
	if host == "" || host == "0.0.0.0" {
		host = server.SSHHost
	}

	client := frp.NewFrpsClient(host, server.DashboardPort, server.DashboardUser, server.DashboardPwd)
	info, err := client.GetServerInfo()
	if err != nil {
		return "", fmt.Errorf("获取服务器信息失败: %w", err)
	}

	server.Version = info.Version
	s.repo.Update(server)
	return info.Version, nil
}

func (s *FrpServerService) GetMetrics(id uint) (*frp.FrpsMetrics, error) {
	server, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	host := server.Host
	if server.ServerType == model.ServerTypeRemote && (host == "" || host == "0.0.0.0") {
		host = server.SSHHost
	}

	client := frp.NewFrpsClient(host, server.DashboardPort, server.DashboardUser, server.DashboardPwd)
	return client.GetMetrics()
}
