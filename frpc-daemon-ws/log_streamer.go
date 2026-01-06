package main

import (
	"bufio"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

// LogType 日志类型
type LogType string

const (
	LogTypeFrpc   LogType = "frpc"
	LogTypeDaemon LogType = "daemon"
)

// LogStreamer 日志流读取器
type LogStreamer struct {
	logType  LogType
	filePath string
	lines    int
	stopChan chan struct{}
	onLine   func(logType LogType, line string)
	running  bool
	mu       sync.Mutex
}

// NewLogStreamer 创建日志流读取器
func NewLogStreamer(logType LogType, filePath string, lines int, onLine func(LogType, string)) *LogStreamer {
	return &LogStreamer{
		logType:  logType,
		filePath: filePath,
		lines:    lines,
		stopChan: make(chan struct{}),
		onLine:   onLine,
	}
}

// Start 开始流式读取日志
func (s *LogStreamer) Start() error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = true
	s.stopChan = make(chan struct{})
	s.mu.Unlock()

	go s.stream()
	return nil
}

// Stop 停止流式读取
func (s *LogStreamer) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.running {
		return
	}
	s.running = false
	close(s.stopChan)
}

// IsRunning 检查是否正在运行
func (s *LogStreamer) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// stream 流式读取日志文件
func (s *LogStreamer) stream() {
	log.Printf("[日志流] 开始读取日志: type=%s, path=%s, lines=%d", s.logType, s.filePath, s.lines)

	// 先读取最后N行
	lastLines, err := s.readLastLines()
	if err != nil {
		log.Printf("[日志流] 读取历史日志失败: %v", err)
	} else {
		for _, line := range lastLines {
			select {
			case <-s.stopChan:
				return
			default:
				if s.onLine != nil {
					s.onLine(s.logType, line)
				}
			}
		}
	}

	// 然后开始 tail -f 模式
	s.tailFile()
}

// readLastLines 读取文件最后N行
func (s *LogStreamer) readLastLines() ([]string, error) {
	file, err := os.Open(s.filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 获取文件大小
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	fileSize := stat.Size()
	if fileSize == 0 {
		return []string{}, nil
	}

	// 从文件末尾开始读取
	var lines []string
	bufSize := int64(4096)
	offset := fileSize

	for len(lines) < s.lines && offset > 0 {
		// 计算读取位置
		readSize := bufSize
		if offset < bufSize {
			readSize = offset
		}
		offset -= readSize

		// 读取数据块
		buf := make([]byte, readSize)
		_, err := file.ReadAt(buf, offset)
		if err != nil && err != io.EOF {
			return nil, err
		}

		// 解析行
		var currentLines []string
		start := 0
		for i := 0; i < len(buf); i++ {
			if buf[i] == '\n' {
				if i > start {
					currentLines = append(currentLines, string(buf[start:i]))
				}
				start = i + 1
			}
		}
		if start < len(buf) && offset == 0 {
			currentLines = append(currentLines, string(buf[start:]))
		}

		// 合并到结果（倒序）
		for i := len(currentLines) - 1; i >= 0; i-- {
			lines = append([]string{currentLines[i]}, lines...)
			if len(lines) >= s.lines {
				break
			}
		}
	}

	// 只返回最后N行
	if len(lines) > s.lines {
		lines = lines[len(lines)-s.lines:]
	}

	return lines, nil
}

// tailFile 实时跟踪文件变化
func (s *LogStreamer) tailFile() {
	file, err := os.Open(s.filePath)
	if err != nil {
		log.Printf("[日志流] 打开文件失败: %v", err)
		return
	}
	defer file.Close()

	// 移动到文件末尾
	file.Seek(0, io.SeekEnd)

	reader := bufio.NewReader(file)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			log.Printf("[日志流] 停止读取日志: type=%s", s.logType)
			return
		case <-ticker.C:
			for {
				line, err := reader.ReadString('\n')
				if err != nil {
					break
				}
				// 去掉换行符
				if len(line) > 0 && line[len(line)-1] == '\n' {
					line = line[:len(line)-1]
				}
				if len(line) > 0 && line[len(line)-1] == '\r' {
					line = line[:len(line)-1]
				}
				if line != "" && s.onLine != nil {
					s.onLine(s.logType, line)
				}
			}
		}
	}
}

// LogStreamManager 日志流管理器
type LogStreamManager struct {
	streamers map[LogType]*LogStreamer
	mu        sync.RWMutex
	cfg       *Config
	onLine    func(LogType, string)
}

// NewLogStreamManager 创建日志流管理器
func NewLogStreamManager(cfg *Config, onLine func(LogType, string)) *LogStreamManager {
	return &LogStreamManager{
		streamers: make(map[LogType]*LogStreamer),
		cfg:       cfg,
		onLine:    onLine,
	}
}

// StartStream 开始指定类型的日志流
func (m *LogStreamManager) StartStream(logType LogType, lines int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 如果已存在，先停止
	if existing, ok := m.streamers[logType]; ok {
		existing.Stop()
	}

	// 获取日志文件路径
	var filePath string
	switch logType {
	case LogTypeFrpc:
		// frpc 日志路径：安装目录/frpc.log
		if m.cfg.InstallDir != "" {
			filePath = m.cfg.InstallDir + "/frpc.log"
		} else {
			filePath = "/opt/frpc/frpc.log"
		}
	case LogTypeDaemon:
		filePath = m.cfg.LogFile
		if filePath == "" {
			if m.cfg.InstallDir != "" {
				filePath = m.cfg.InstallDir + "/frpc-daemon.log"
			} else {
				filePath = "/opt/frpc/frpc-daemon.log"
			}
		}
	default:
		log.Printf("[日志流] 未知的日志类型: %s", logType)
		return nil
	}

	// 创建并启动流读取器
	streamer := NewLogStreamer(logType, filePath, lines, m.onLine)
	m.streamers[logType] = streamer
	return streamer.Start()
}

// StopStream 停止指定类型的日志流
func (m *LogStreamManager) StopStream(logType LogType) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if streamer, ok := m.streamers[logType]; ok {
		streamer.Stop()
		delete(m.streamers, logType)
	}
}

// StopAll 停止所有日志流
func (m *LogStreamManager) StopAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, streamer := range m.streamers {
		streamer.Stop()
	}
	m.streamers = make(map[LogType]*LogStreamer)
}
