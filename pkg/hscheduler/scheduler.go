// Package hscheduler 提供进程内定时任务调度能力。
package hscheduler

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/vaynedu/hollow/pkg/hlog"
)

var (
	// ErrNilLocation 表示未提供调度时区。
	ErrNilLocation = errors.New("hscheduler: nil location")
	// ErrNilJob 表示注册了空任务。
	ErrNilJob = errors.New("hscheduler: nil job")
	// ErrEmptyJobName 表示任务名称为空。
	ErrEmptyJobName = errors.New("hscheduler: empty job name")
	// ErrInvalidTimeout 表示任务超时配置无效。
	ErrInvalidTimeout = errors.New("hscheduler: timeout must be positive")
	// ErrDuplicateJob 表示任务名称重复。
	ErrDuplicateJob = errors.New("hscheduler: duplicate job name")
	// ErrAlreadyStarted 表示调度器启动后仍尝试注册任务。
	ErrAlreadyStarted = errors.New("hscheduler: cannot register after start")
)

// Job 是定时任务需要实现的最小接口。
type Job interface {
	Name() string
	Spec() string
	Run(context.Context) error
}

// JobFunc 是无需定义 Job 结构体时使用的任务函数。
type JobFunc func(context.Context) error

type funcJob struct {
	name string
	spec string
	run  JobFunc
}

func (j *funcJob) Name() string                  { return j.name }
func (j *funcJob) Spec() string                  { return j.spec }
func (j *funcJob) Run(ctx context.Context) error { return j.run(ctx) }

type registeredJob struct {
	job        Job
	timeout    time.Duration
	runOnStart bool
	running    atomic.Bool
}

// Scheduler 管理进程内定时任务。
type Scheduler struct {
	cron        *cron.Cron
	ctx         context.Context
	cancel      context.CancelFunc
	jobs        map[string]*registeredJob
	mu          sync.Mutex
	started     bool
	startupJobs sync.WaitGroup
}

// New 创建使用指定时区的调度器。
func New(location *time.Location) (*Scheduler, error) {
	if location == nil {
		return nil, ErrNilLocation
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		cron:   cron.New(cron.WithLocation(location)),
		ctx:    ctx,
		cancel: cancel,
		jobs:   make(map[string]*registeredJob),
	}, nil
}

// Register 注册任务。
func (s *Scheduler) Register(job Job, timeout time.Duration, runOnStart bool) error {
	if s == nil || job == nil {
		return ErrNilJob
	}
	if job.Name() == "" {
		return ErrEmptyJobName
	}
	if timeout <= 0 {
		return ErrInvalidTimeout
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.started {
		return ErrAlreadyStarted
	}
	if _, exists := s.jobs[job.Name()]; exists {
		return fmt.Errorf("%w: %s", ErrDuplicateJob, job.Name())
	}
	entry := &registeredJob{job: job, timeout: timeout, runOnStart: runOnStart}
	if _, err := s.cron.AddFunc(job.Spec(), func() { s.run(entry) }); err != nil {
		return fmt.Errorf("hscheduler: register %s: %w", job.Name(), err)
	}
	s.jobs[job.Name()] = entry
	return nil
}

// RegisterFunc 使用闭包注册任务，适合逻辑简单、无需独立任务类型的场景。
func (s *Scheduler) RegisterFunc(name, spec string, timeout time.Duration, runOnStart bool, run JobFunc) error {
	if run == nil {
		return ErrNilJob
	}
	return s.Register(&funcJob{name: name, spec: spec, run: run}, timeout, runOnStart)
}

// Start 启动调度器。
func (s *Scheduler) Start() {
	if s == nil {
		return
	}
	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		return
	}
	s.started = true
	startupJobs := make([]*registeredJob, 0)
	for _, entry := range s.jobs {
		if entry.runOnStart {
			startupJobs = append(startupJobs, entry)
		}
	}
	for range startupJobs {
		s.startupJobs.Add(1)
	}
	s.cron.Start()
	s.mu.Unlock()
	for _, entry := range startupJobs {
		go func(item *registeredJob) {
			defer s.startupJobs.Done()
			s.run(item)
		}(entry)
	}
}

// Startup 以 Hollow 启动 Hook 的签名启动调度器。
func (s *Scheduler) Startup() error {
	s.Start()
	return nil
}

func (s *Scheduler) run(entry *registeredJob) {
	if !entry.running.CompareAndSwap(false, true) {
		hlog.L().Named("hscheduler").Info("job skipped because previous run is still active", hlog.String("job", entry.job.Name()))
		return
	}
	defer entry.running.Store(false)
	startedAt := time.Now()
	logger := hlog.L().Named("hscheduler").With(hlog.String("job", entry.job.Name()))
	ctx, cancel := context.WithTimeout(s.ctx, entry.timeout)
	defer cancel()
	defer func() {
		if recovered := recover(); recovered != nil {
			logger.Error("job panicked",
				hlog.String("panic", fmt.Sprint(recovered)),
				hlog.String("stack", string(debug.Stack())),
				hlog.Int64("duration_ms", time.Since(startedAt).Milliseconds()))
		}
	}()
	if err := entry.job.Run(ctx); err != nil {
		logger.Error("job failed", hlog.Err(err), hlog.Int64("duration_ms", time.Since(startedAt).Milliseconds()))
		return
	}
	logger.Info("job completed", hlog.Int64("duration_ms", time.Since(startedAt).Milliseconds()))
}

// Shutdown 停止调度器并等待正在执行的任务结束。
func (s *Scheduler) Shutdown(ctx context.Context) error {
	if s == nil {
		return nil
	}
	cronStopped := s.cron.Stop()
	done := make(chan struct{})
	go func() {
		<-cronStopped.Done()
		s.startupJobs.Wait()
		close(done)
	}()
	select {
	case <-done:
		s.cancel()
		return nil
	case <-ctx.Done():
		s.cancel()
		return ctx.Err()
	}
}
