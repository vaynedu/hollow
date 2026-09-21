package hscheduler

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

type testJob struct {
	name       string
	spec       string
	runs       atomic.Int32
	started    chan struct{}
	release    chan struct{}
	panicOnRun bool
}

func (j *testJob) Name() string { return j.name }
func (j *testJob) Spec() string { return j.spec }
func (j *testJob) Run(ctx context.Context) error {
	j.runs.Add(1)
	select {
	case j.started <- struct{}{}:
	default:
	}
	if j.panicOnRun {
		panic("test panic")
	}
	if j.release != nil {
		select {
		case <-j.release:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func TestSchedulerRunsRegisteredJobOnStart(t *testing.T) {
	scheduler, err := New(time.Local)
	if err != nil {
		t.Fatal(err)
	}
	job := &testJob{
		name:    "startup",
		spec:    "@every 1h",
		started: make(chan struct{}, 1),
	}
	if err := scheduler.Register(job, time.Second, true); err != nil {
		t.Fatal(err)
	}

	scheduler.Start()
	select {
	case <-job.started:
	case <-time.After(time.Second):
		t.Fatal("启动补跑任务未执行")
	}
	if err := scheduler.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := job.runs.Load(); got != 1 {
		t.Fatalf("执行次数=%d，期望=1", got)
	}
}

func TestNewRejectsNilLocation(t *testing.T) {
	if _, err := New(nil); err == nil {
		t.Fatal("nil 时区应返回错误")
	}
}

func TestRegisterRejectsInvalidJobs(t *testing.T) {
	scheduler, err := New(time.Local)
	if err != nil {
		t.Fatal(err)
	}
	if err := scheduler.Register(nil, time.Second, false); !errors.Is(err, ErrNilJob) {
		t.Fatalf("nil 任务错误=%v", err)
	}
	if err := scheduler.Register(&testJob{spec: "@every 1h"}, time.Second, false); !errors.Is(err, ErrEmptyJobName) {
		t.Fatalf("空任务名错误=%v", err)
	}
	if err := scheduler.Register(&testJob{name: "timeout", spec: "@every 1h"}, 0, false); !errors.Is(err, ErrInvalidTimeout) {
		t.Fatalf("非法超时错误=%v", err)
	}
	if err := scheduler.Register(&testJob{name: "cron", spec: "not-a-cron"}, time.Second, false); err == nil {
		t.Fatal("非法 cron 表达式应返回错误")
	}
}

func TestRegisterRejectsDuplicateAndLateRegistration(t *testing.T) {
	scheduler, err := New(time.Local)
	if err != nil {
		t.Fatal(err)
	}
	job := &testJob{name: "unique", spec: "@every 1h", started: make(chan struct{}, 1)}
	if err := scheduler.Register(job, time.Second, false); err != nil {
		t.Fatal(err)
	}
	if err := scheduler.Register(job, time.Second, false); !errors.Is(err, ErrDuplicateJob) {
		t.Fatalf("重复任务错误=%v", err)
	}
	scheduler.Start()
	if err := scheduler.Register(&testJob{name: "late", spec: "@every 1h"}, time.Second, false); !errors.Is(err, ErrAlreadyStarted) {
		t.Fatalf("启动后注册错误=%v", err)
	}
	if err := scheduler.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestSchedulerSkipsOverlappingRuns(t *testing.T) {
	scheduler, err := New(time.Local)
	if err != nil {
		t.Fatal(err)
	}
	job := &testJob{
		name:    "blocking",
		spec:    "@every 1h",
		started: make(chan struct{}, 1),
		release: make(chan struct{}),
	}
	if err := scheduler.Register(job, time.Second, true); err != nil {
		t.Fatal(err)
	}
	scheduler.Start()
	select {
	case <-job.started:
	case <-time.After(time.Second):
		t.Fatal("启动补跑任务未执行")
	}

	secondRunDone := make(chan struct{})
	go func() {
		scheduler.run(scheduler.jobs[job.name])
		close(secondRunDone)
	}()
	select {
	case <-secondRunDone:
	case <-time.After(time.Second):
		t.Fatal("重叠执行没有立即跳过")
	}
	if got := job.runs.Load(); got != 1 {
		t.Fatalf("重叠任务执行次数=%d，期望=1", got)
	}
	close(job.release)
	if err := scheduler.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestSchedulerRecoversJobPanic(t *testing.T) {
	scheduler, err := New(time.Local)
	if err != nil {
		t.Fatal(err)
	}
	job := &testJob{name: "panic", spec: "@every 1h", panicOnRun: true}
	if err := scheduler.Register(job, time.Second, false); err != nil {
		t.Fatal(err)
	}

	scheduler.run(scheduler.jobs[job.name])
	if got := job.runs.Load(); got != 1 {
		t.Fatalf("panic 任务执行次数=%d，期望=1", got)
	}
}

func TestShutdownWaitsForStartupJobs(t *testing.T) {
	scheduler, err := New(time.Local)
	if err != nil {
		t.Fatal(err)
	}
	job := &testJob{
		name:    "graceful",
		spec:    "@every 1h",
		started: make(chan struct{}, 1),
		release: make(chan struct{}),
	}
	if err := scheduler.Register(job, time.Second, true); err != nil {
		t.Fatal(err)
	}
	scheduler.Start()
	select {
	case <-job.started:
	case <-time.After(time.Second):
		t.Fatal("启动补跑任务未执行")
	}

	shutdownDone := make(chan error, 1)
	go func() {
		shutdownDone <- scheduler.Shutdown(context.Background())
	}()
	select {
	case err := <-shutdownDone:
		close(job.release)
		t.Fatalf("任务结束前 Shutdown 已返回：%v", err)
	case <-time.After(50 * time.Millisecond):
	}
	close(job.release)
	select {
	case err := <-shutdownDone:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("任务结束后 Shutdown 未返回")
	}
}

func TestNilSchedulerLifecycleIsSafe(t *testing.T) {
	var scheduler *Scheduler
	scheduler.Start()
	if err := scheduler.Shutdown(context.Background()); err != nil {
		t.Fatalf("nil 调度器关闭错误=%v", err)
	}
}

func TestRegisterFuncRunsClosureOnStart(t *testing.T) {
	scheduler, err := New(time.Local)
	if err != nil {
		t.Fatal(err)
	}
	run := make(chan struct{}, 1)
	if err := scheduler.RegisterFunc(
		"closure",
		"@every 1h",
		time.Second,
		true,
		func(context.Context) error {
			run <- struct{}{}
			return nil
		},
	); err != nil {
		t.Fatal(err)
	}
	scheduler.Start()
	select {
	case <-run:
	case <-time.After(time.Second):
		t.Fatal("闭包任务未执行")
	}
	if err := scheduler.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestRegisterFuncRejectsNilClosure(t *testing.T) {
	scheduler, err := New(time.Local)
	if err != nil {
		t.Fatal(err)
	}
	if err := scheduler.RegisterFunc("nil", "@every 1h", time.Second, false, nil); !errors.Is(err, ErrNilJob) {
		t.Fatalf("nil 闭包错误=%v", err)
	}
}

func TestStartupRunsRegisteredClosure(t *testing.T) {
	scheduler, err := New(time.Local)
	if err != nil {
		t.Fatal(err)
	}
	run := make(chan struct{}, 1)
	if err := scheduler.RegisterFunc("startup-hook", "@every 1h", time.Second, true, func(context.Context) error {
		run <- struct{}{}
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	if err := scheduler.Startup(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-run:
	case <-time.After(time.Second):
		t.Fatal("Startup Hook 未启动任务")
	}
	if err := scheduler.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestSchedulerAppliesJobTimeout(t *testing.T) {
	scheduler, err := New(time.Local)
	if err != nil {
		t.Fatal(err)
	}
	job := &testJob{
		name:    "timeout",
		spec:    "@every 1h",
		started: make(chan struct{}, 1),
		release: make(chan struct{}),
	}
	if err := scheduler.Register(job, 20*time.Millisecond, true); err != nil {
		t.Fatal(err)
	}
	scheduler.Start()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := scheduler.Shutdown(ctx); err != nil {
		t.Fatalf("任务超时后关闭错误=%v", err)
	}
	if got := job.runs.Load(); got != 1 {
		t.Fatalf("执行次数=%d，期望=1", got)
	}
}
