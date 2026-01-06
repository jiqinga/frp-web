package service

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

type SSHClient struct {
	client *ssh.Client
}

type SSHConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Timeout  time.Duration
}

func NewSSHClient(config SSHConfig) (*SSHClient, error) {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	sshConfig := &ssh.ClientConfig{
		User: config.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(config.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         config.Timeout,
	}

	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("连接失败: %w", err)
	}

	return &SSHClient{client: client}, nil
}

func (c *SSHClient) ExecuteCommand(cmd string) (string, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("创建会话失败: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return string(output), fmt.Errorf("命令执行失败: %w", err)
	}

	return string(output), nil
}

func (c *SSHClient) ExecuteCommandWithCallback(cmd string, callback func(string)) error {
	session, err := c.client.NewSession()
	if err != nil {
		return fmt.Errorf("创建会话失败: %w", err)
	}
	defer session.Close()

	stdout, err := session.StdoutPipe()
	if err != nil {
		return fmt.Errorf("获取标准输出失败: %w", err)
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		return fmt.Errorf("获取标准错误输出失败: %w", err)
	}

	if err := session.Start(cmd); err != nil {
		return fmt.Errorf("启动命令失败: %w", err)
	}

	go c.readAndFilterOutput(stdout, callback)
	go c.readAndFilterOutput(stderr, callback)

	return session.Wait()
}

func (c *SSHClient) readAndFilterOutput(reader io.Reader, callback func(string)) {
	scanner := bufio.NewScanner(reader)
	wgetProgressRegex := regexp.MustCompile(`^\s*\d+K\s+\.+\s+\d+%`)

	for scanner.Scan() {
		line := scanner.Text()

		// 过滤wget的进度条行
		if wgetProgressRegex.MatchString(line) {
			// 提取百分比信息
			percentRegex := regexp.MustCompile(`(\d+)%`)
			if matches := percentRegex.FindStringSubmatch(line); len(matches) > 1 {
				callback(fmt.Sprintf("下载进度: %s%%", matches[1]))
			}
			continue
		}

		// 过滤只包含点号的行
		if strings.TrimSpace(line) == "" || regexp.MustCompile(`^[\s.]+$`).MatchString(line) {
			continue
		}

		callback(line)
	}
}

func (c *SSHClient) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

func (c *SSHClient) TestConnection() error {
	session, err := c.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	return session.Run("echo test")
}

func (c *SSHClient) UploadFile(localContent []byte, remotePath string) error {
	session, err := c.client.NewSession()
	if err != nil {
		return fmt.Errorf("创建会话失败: %w", err)
	}
	defer session.Close()

	stdin, err := session.StdinPipe()
	if err != nil {
		return fmt.Errorf("获取标准输入失败: %w", err)
	}

	go func() {
		defer stdin.Close()
		fmt.Fprintf(stdin, "C0644 %d %s\n", len(localContent), remotePath)
		stdin.Write(localContent)
		fmt.Fprint(stdin, "\x00")
	}()

	if err := session.Run(fmt.Sprintf("scp -t %s", remotePath)); err != nil {
		return fmt.Errorf("上传文件失败: %w", err)
	}

	return nil
}

func (c *SSHClient) DownloadFile(remotePath string) ([]byte, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("创建会话失败: %w", err)
	}
	defer session.Close()

	stdout, err := session.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("获取标准输出失败: %w", err)
	}

	if err := session.Start(fmt.Sprintf("cat %s", remotePath)); err != nil {
		return nil, fmt.Errorf("启动命令失败: %w", err)
	}

	content, err := io.ReadAll(stdout)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	if err := session.Wait(); err != nil {
		return nil, fmt.Errorf("命令执行失败: %w", err)
	}

	return content, nil
}
