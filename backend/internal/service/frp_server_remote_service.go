/*
 * FrpServerService - 远程服务器操作方法
 */
package service

import (
	"fmt"
	"frp-web-panel/internal/model"
	"frp-web-panel/internal/util"
	"path/filepath"
	"time"
)

func (s *FrpServerService) TestSSH(id uint) error {
	server, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if server.ServerType != model.ServerTypeRemote {
		return fmt.Errorf("只有远程服务器支持SSH测试")
	}

	password, err := util.Decrypt(server.SSHPassword)
	if err != nil {
		return fmt.Errorf("解密密码失败: %w", err)
	}

	sshClient, err := NewSSHClient(SSHConfig{
		Host:     server.SSHHost,
		Port:     server.SSHPort,
		User:     server.SSHUser,
		Password: password,
	})
	if err != nil {
		return err
	}
	defer sshClient.Close()

	return sshClient.TestConnection()
}

func (s *FrpServerService) RemoteInstall(id uint, mirrorID *uint) error {
	if !s.setTaskRunning(id, "install") {
		return fmt.Errorf("该服务器正在执行其他任务")
	}

	server, err := s.repo.GetByID(id)
	if err != nil {
		s.clearTask(id)
		return err
	}

	if server.ServerType != model.ServerTypeRemote {
		s.clearTask(id)
		return fmt.Errorf("只有远程服务器支持远程安装")
	}

	if mirrorID != nil {
		server.MirrorID = mirrorID
	}

	if server.DashboardUser == "" {
		server.DashboardUser = "admin"
	}
	if server.DashboardPwd == "" {
		server.DashboardPwd = generateRandomPassword(16)
	}
	if server.Token == "" {
		token, err := util.GenerateRandomToken(48)
		if err != nil {
			return fmt.Errorf("生成token失败: %w", err)
		}
		server.Token = token
	}

	logFunc := func(msg string) {
		s.PublishSSHLog(id, "install", msg)
	}

	installer, err := NewRemoteFrpsInstaller(server, logFunc)
	if err != nil {
		s.clearTask(id)
		return err
	}
	defer func() {
		installer.Close()
		s.clearTask(id)
	}()

	if err := installer.Install(server); err != nil {
		return err
	}

	server.Status = model.StatusStopped
	server.BinaryPath = filepath.Join(server.InstallPath, "frps")
	server.ConfigPath = filepath.Join(server.InstallPath, "frps.yaml")
	return s.repo.Update(server)
}

func (s *FrpServerService) RemoteStart(id uint) error {
	server, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if server.ServerType != model.ServerTypeRemote {
		return fmt.Errorf("只有远程服务器支持远程启动")
	}

	logFunc := func(msg string) {
		s.PublishSSHLog(id, "start", msg)
	}

	installer, err := NewRemoteFrpsInstaller(server, logFunc)
	if err != nil {
		return err
	}
	defer installer.Close()

	if err := installer.Start(server.InstallPath); err != nil {
		return err
	}

	time.Sleep(2 * time.Second)
	actualStatus, err := s.GetStatus(id)
	if err == nil && actualStatus == model.StatusRunning {
		server.Status = model.StatusRunning
	} else {
		server.Status = model.StatusStopped
	}

	if err := s.repo.Update(server); err != nil {
		return err
	}

	s.PublishServerStatus(id, server.Name, string(server.Status))
	return nil
}

func (s *FrpServerService) RemoteStop(id uint) error {
	server, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if server.ServerType != model.ServerTypeRemote {
		return fmt.Errorf("只有远程服务器支持远程停止")
	}

	logFunc := func(msg string) {
		s.PublishSSHLog(id, "stop", msg)
	}

	installer, err := NewRemoteFrpsInstaller(server, logFunc)
	if err != nil {
		return err
	}
	defer installer.Close()

	if err := installer.Stop(server.InstallPath); err != nil {
		return err
	}

	server.Status = model.StatusStopped
	return s.repo.Update(server)
}

func (s *FrpServerService) RemoteRestart(id uint) error {
	server, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if server.ServerType != model.ServerTypeRemote {
		return fmt.Errorf("只有远程服务器支持远程重启")
	}

	logFunc := func(msg string) {
		s.PublishSSHLog(id, "restart", msg)
	}

	installer, err := NewRemoteFrpsInstaller(server, logFunc)
	if err != nil {
		return err
	}
	defer installer.Close()

	if err := installer.Restart(server.InstallPath); err != nil {
		return err
	}

	server.Status = model.StatusRunning
	return s.repo.Update(server)
}

func (s *FrpServerService) RemoteUninstall(id uint) error {
	server, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if server.ServerType != model.ServerTypeRemote {
		return fmt.Errorf("只有远程服务器支持远程卸载")
	}

	logFunc := func(msg string) {
		s.PublishSSHLog(id, "uninstall", msg)
	}

	installer, err := NewRemoteFrpsInstaller(server, logFunc)
	if err != nil {
		return err
	}
	defer installer.Close()

	if err := installer.Uninstall(server.InstallPath); err != nil {
		return err
	}

	server.Status = model.StatusStopped
	server.BinaryPath = ""
	server.ConfigPath = ""
	return s.repo.Update(server)
}

func (s *FrpServerService) RemoteGetLogs(id uint, lines int) (string, error) {
	server, err := s.repo.GetByID(id)
	if err != nil {
		return "", err
	}
	if server.ServerType != model.ServerTypeRemote {
		return "", fmt.Errorf("只有远程服务器支持查看远程日志")
	}

	installer, err := NewRemoteFrpsInstaller(server, nil)
	if err != nil {
		return "", err
	}
	defer installer.Close()

	return installer.GetLogs(server.InstallPath, lines)
}

func (s *FrpServerService) RemoteGetVersion(id uint) (string, error) {
	server, err := s.repo.GetByID(id)
	if err != nil {
		return "", err
	}
	if server.ServerType != model.ServerTypeRemote {
		return "", fmt.Errorf("只有远程服务器支持获取远程版本")
	}

	installer, err := NewRemoteFrpsInstaller(server, nil)
	if err != nil {
		return "", err
	}
	defer installer.Close()

	version, err := installer.GetVersion(server.InstallPath)
	if err != nil {
		return "", err
	}

	server.Version = version
	s.repo.Update(server)
	return version, nil
}

func (s *FrpServerService) RemoteReinstall(id uint, regenerateAuth bool, mirrorID *uint) error {
	if !s.setTaskRunning(id, "reinstall") {
		return fmt.Errorf("该服务器正在执行其他任务")
	}

	server, err := s.repo.GetByID(id)
	if err != nil {
		s.clearTask(id)
		return err
	}

	if server.ServerType != model.ServerTypeRemote {
		s.clearTask(id)
		return fmt.Errorf("只有远程服务器支持远程重装")
	}

	if mirrorID != nil {
		server.MirrorID = mirrorID
	}

	if regenerateAuth {
		server.DashboardUser = "admin"
		server.DashboardPwd = generateRandomPassword(16)
		token, err := util.GenerateRandomToken(48)
		if err != nil {
			return fmt.Errorf("生成token失败: %w", err)
		}
		server.Token = token
	}

	logFunc := func(msg string) {
		s.PublishSSHLog(id, "reinstall", msg)
	}

	installer, err := NewRemoteFrpsInstaller(server, logFunc)
	if err != nil {
		s.clearTask(id)
		return err
	}
	defer func() {
		installer.Close()
		s.clearTask(id)
	}()

	if err := installer.Uninstall(server.InstallPath); err != nil {
		return err
	}
	if err := installer.Install(server); err != nil {
		return err
	}

	server.Status = model.StatusStopped
	server.BinaryPath = filepath.Join(server.InstallPath, "frps")
	server.ConfigPath = filepath.Join(server.InstallPath, "frps.yaml")
	return s.repo.Update(server)
}

func (s *FrpServerService) RemoteUpgrade(id uint, version string, mirrorID *uint) error {
	if !s.setTaskRunning(id, "upgrade") {
		return fmt.Errorf("该服务器正在执行其他任务")
	}

	server, err := s.repo.GetByID(id)
	if err != nil {
		s.clearTask(id)
		return err
	}

	if server.ServerType != model.ServerTypeRemote {
		s.clearTask(id)
		return fmt.Errorf("只有远程服务器支持远程升级")
	}

	if mirrorID != nil {
		server.MirrorID = mirrorID
	}
	if version != "" {
		server.Version = version
	}

	server.DashboardUser = "admin"
	server.DashboardPwd = generateRandomPassword(16)
	token, err := util.GenerateRandomToken(48)
	if err != nil {
		return fmt.Errorf("生成token失败: %w", err)
	}
	server.Token = token

	logFunc := func(msg string) {
		s.PublishSSHLog(id, "upgrade", msg)
	}

	installer, err := NewRemoteFrpsInstaller(server, logFunc)
	if err != nil {
		s.clearTask(id)
		return err
	}
	defer func() {
		installer.Close()
		s.clearTask(id)
	}()

	if err := installer.Uninstall(server.InstallPath); err != nil {
		return err
	}
	if err := installer.Install(server); err != nil {
		return err
	}

	server.Status = model.StatusStopped
	server.BinaryPath = filepath.Join(server.InstallPath, "frps")
	server.ConfigPath = filepath.Join(server.InstallPath, "frps.yaml")
	return s.repo.Update(server)
}
