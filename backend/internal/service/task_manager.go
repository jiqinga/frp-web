package service

import (
	"context"
	"frp-web-panel/internal/logger"
	"sync"
	"time"
)

// Task 定义可管理的任务接口
type Task interface {
	Start(ctx context.Context)
	Name() string
}

// PeriodicTask 定时任务
type PeriodicTask struct {
	name     string
	interval time.Duration
	fn       func()
}

func NewPeriodicTask(name string, interval time.Duration, fn func()) *PeriodicTask {
	return &PeriodicTask{name: name, interval: interval, fn: fn}
}

func (t *PeriodicTask) Name() string { return t.name }

func (t *PeriodicTask) Start(ctx context.Context) {
	ticker := time.NewTicker(t.interval)
	defer ticker.Stop()

	// 立即执行一次
	t.fn()

	for {
		select {
		case <-ticker.C:
			t.fn()
		case <-ctx.Done():
			logger.Infof("TaskManager 定时任务 %s 已停止", t.name)
			return
		}
	}
}

// OneShotTask 一次性后台任务
type OneShotTask struct {
	name string
	fn   func(ctx context.Context)
}

func NewOneShotTask(name string, fn func(ctx context.Context)) *OneShotTask {
	return &OneShotTask{name: name, fn: fn}
}

func (t *OneShotTask) Name() string { return t.name }

func (t *OneShotTask) Start(ctx context.Context) {
	t.fn(ctx)
	logger.Infof("TaskManager 一次性任务 %s 已完成", t.name)
}

// TaskManager 统一管理所有后台 goroutine
type TaskManager struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	mu     sync.Mutex
	tasks  []Task
}

// NewTaskManager 创建任务管理器
func NewTaskManager() *TaskManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &TaskManager{
		ctx:    ctx,
		cancel: cancel,
		tasks:  make([]Task, 0),
	}
}

// RegisterTask 注册任务（不立即启动）
func (m *TaskManager) RegisterTask(task Task) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tasks = append(m.tasks, task)
}

// RegisterPeriodicTask 注册定时任务的便捷方法
func (m *TaskManager) RegisterPeriodicTask(name string, interval time.Duration, fn func()) {
	m.RegisterTask(NewPeriodicTask(name, interval, fn))
}

// RegisterOneShotTask 注册一次性任务的便捷方法
func (m *TaskManager) RegisterOneShotTask(name string, fn func(ctx context.Context)) {
	m.RegisterTask(NewOneShotTask(name, fn))
}

// Start 启动所有已注册的任务
func (m *TaskManager) Start() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, task := range m.tasks {
		m.wg.Add(1)
		go func(t Task) {
			defer m.wg.Done()
			logger.Infof("TaskManager 启动任务: %s", t.Name())
			t.Start(m.ctx)
		}(task)
	}
	logger.Infof("TaskManager 已启动 %d 个任务", len(m.tasks))
}

// Shutdown 优雅关闭所有任务
func (m *TaskManager) Shutdown(timeout time.Duration) error {
	logger.Info("TaskManager 开始关闭所有任务...")
	m.cancel()

	done := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("TaskManager 所有任务已优雅关闭")
		return nil
	case <-time.After(timeout):
		logger.Warn("TaskManager 关闭超时，部分任务可能未完全停止")
		return context.DeadlineExceeded
	}
}

// Context 返回任务管理器的 context，供外部服务使用
func (m *TaskManager) Context() context.Context {
	return m.ctx
}
