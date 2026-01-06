/*
 * @Author              : 寂情啊
 * @Date                : 2025-12-26 15:14:25
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-26 15:14:36
 * @FilePath            : frp-web-testfrpc-daemon-wsconfig_backup.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

// BackupConfig 备份当前配置，返回备份文件路径
func BackupConfig(configPath string) (string, error) {
	log.Printf("[ConfigBackup] 开始备份配置: %s", configPath)

	// 检查源文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Printf("[ConfigBackup] 配置文件不存在，跳过备份")
		return "", nil
	}

	// 读取当前配置
	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 生成备份文件路径（带时间戳）
	backupPath := fmt.Sprintf("%s.backup.%d", configPath, time.Now().Unix())

	// 写入备份文件
	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return "", fmt.Errorf("写入备份文件失败: %v", err)
	}

	log.Printf("[ConfigBackup] ✅ 配置已备份到: %s", backupPath)
	return backupPath, nil
}

// RestoreConfig 从备份恢复配置
func RestoreConfig(backupPath, configPath string) error {
	log.Printf("[ConfigBackup] 开始恢复配置: %s -> %s", backupPath, configPath)

	if backupPath == "" {
		return fmt.Errorf("备份路径为空")
	}

	// 读取备份文件
	data, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("读取备份文件失败: %v", err)
	}

	// 写入配置文件
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("恢复配置文件失败: %v", err)
	}

	log.Printf("[ConfigBackup] ✅ 配置已恢复")
	return nil
}

// CleanupBackup 清理备份文件
func CleanupBackup(backupPath string) error {
	if backupPath == "" {
		return nil
	}

	if err := os.Remove(backupPath); err != nil && !os.IsNotExist(err) {
		log.Printf("[ConfigBackup] ⚠️ 清理备份文件失败: %v", err)
		return err
	}

	log.Printf("[ConfigBackup] ✅ 备份文件已清理: %s", backupPath)
	return nil
}
