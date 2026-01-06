package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

// ValidateConfig 使用 frpc verify 命令验证配置
func ValidateConfig(frpcPath, configPath string) error {
	log.Printf("[ConfigValidator] 开始验证配置: %s", configPath)

	// 检查 frpc 二进制文件是否存在
	if _, err := os.Stat(frpcPath); os.IsNotExist(err) {
		return fmt.Errorf("frpc 二进制文件不存在: %s", frpcPath)
	}

	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("配置文件不存在: %s", configPath)
	}

	// 执行 frpc verify 命令
	cmd := exec.Command(frpcPath, "verify", "-c", configPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[ConfigValidator] ❌ 配置验证失败: %s", string(output))
		return fmt.Errorf("配置验证失败: %s", string(output))
	}

	log.Printf("[ConfigValidator] ✅ 配置验证通过")
	return nil
}

// ValidateConfigContent 验证配置内容（写入临时文件后验证）
func ValidateConfigContent(frpcPath, config string) error {
	// 创建临时文件
	tempFile, err := os.CreateTemp("", "frpc_verify_*.toml")
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %v", err)
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath)

	// 写入配置内容
	if _, err := tempFile.WriteString(config); err != nil {
		tempFile.Close()
		return fmt.Errorf("写入临时配置失败: %v", err)
	}
	tempFile.Close()

	// 验证配置
	return ValidateConfig(frpcPath, tempPath)
}
