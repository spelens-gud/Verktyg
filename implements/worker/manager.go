package worker

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/spelens-gud/Verktyg.git/kits/kserver/govern_server"

	"github.com/spelens-gud/Verktyg.git/interfaces/ilog"

	"github.com/robfig/cron/v3"

	"github.com/spelens-gud/Verktyg.git/kits/kruntime"

	"github.com/spelens-gud/Verktyg.git/interfaces/itrace"
	"github.com/spelens-gud/Verktyg.git/interfaces/iworker"
	"github.com/spelens-gud/Verktyg.git/kits/klog/logger"
)

type workerManager struct {
	cron         *cron.Cron
	cronEntries  []cron.EntryID
	ctx          context.Context
	ctxCancel    context.CancelFunc
	managerTasks []func(ctx context.Context)

	opt       *Option
	startOnce sync.Once
}

type Option struct {
	CronOptions      []cron.Option
	BaseContext      context.Context
	CloseWaitSeconds int
}

func WithCronOption(opts ...cron.Option) func(option *Option) {
	return func(option *Option) {
		option.CronOptions = append(option.CronOptions, opts...)
	}
}

func NewWorkerManager(opts ...func(option *Option)) iworker.WorkerManager {
	opt := &Option{
		CronOptions:      []cron.Option{WithLogger(logger.FromBackground())},
		CloseWaitSeconds: 10,
		BaseContext:      context.Background(),
	}

	for _, o := range opts {
		o(opt)
	}

	ctx, cf := context.WithCancel(opt.BaseContext)
	return &workerManager{opt: opt, cron: cron.New(opt.CronOptions...), ctx: ctx, ctxCancel: cf}
}

func (c *workerManager) MustRegisterTask(f func(ctx context.Context), err error) {
	if err != nil {
		panic(err)
	}
	c.RegisterTasks(f)
}

func (c *workerManager) RegisterTasks(f ...func(ctx context.Context)) {
	c.managerTasks = append(c.managerTasks, f...)
}

// 启动worker
func (c *workerManager) startCronWorker(ctx context.Context) {
	c.cron.Start()
	<-ctx.Done()
	<-c.cron.Stop().Done()
}

func (c *workerManager) start() {
	govern_server.RunGovernServer()

	var (
		lg       = logger.FromBackground().WithTag("WORKER_MANAGER").WithField("pid", os.Getpid())
		wg       = new(sync.WaitGroup)
		stopChan = kruntime.StopChan()
	)

	if len(c.cronEntries) > 0 {
		c.managerTasks = append(c.managerTasks, c.startCronWorker)
	}

	for _, f := range c.managerTasks {
		wg.Add(1)
		go func(f func(context.Context)) {
			lg := lg.WithField("task", getFuncName(f))
			lg.Info("task start")
			defer func() {
				wg.Done()
				r := recover()
				if r == nil {
					lg.Infof("task finished")
					return
				}
				c.ctxCancel()
				buf := make([]byte, 64<<10)
				buf = buf[:runtime.Stack(buf, false)]
				info := fmt.Sprintf("[PANIC] recovered: %s\n%s", r, buf)
				fmt.Print(info)
				lg.WithTag("TASK_WORKER_PANIC").Errorf(info)
				lg.WithTag("TASK_WORKER_PANIC").Errorf("task panic,force shutdown")
			}()
			f(logger.WithContext(c.ctx, lg))
		}(f)
	}

	waiting := make(chan struct{})
	// 停止所有任务
	go func() {
		wg.Wait()
		close(waiting)
	}()

	lg.Infof("manager start")

	select {
	case <-stopChan:
		// 停止context
		c.ctxCancel()
	case <-c.ctx.Done():
		// 其他情况的手动中止
	case <-waiting:
		// 所有任务已退出
	}

	select {
	case <-time.After(time.Second * time.Duration(c.opt.CloseWaitSeconds)):
		lg.Warnf("manager exit waiting timeout")
	case <-waiting:
		lg.Infof("manager exit")
	}
}

func (c *workerManager) Start() {
	c.startOnce.Do(c.start)
}

func (c *workerManager) Stop() {
	c.ctxCancel()
}

func getFuncName(f interface{}) (r string) {
	defer func() {
		_ = recover()
	}()
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}

func (c *workerManager) RegisterCron(cronSpec string, cmdFunc func(ctx context.Context) (err error)) {
	var entry cron.EntryID
	f := getFuncName(cmdFunc)
	getLogger := func(ctx context.Context) ilog.Logger {
		return logger.FromContext(ctx).WithFields(map[string]interface{}{
			"func":  f,
			"entry": entry,
		}).WithTag(logTagCronWorker, "WORKER_MANAGER")
	}

	entry, err := c.cron.AddFunc(cronSpec, func() {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		// 绑定新的链路ID
		var (
			ctx = itrace.WithContext(c.ctx, itrace.NewIDFunc())
			lg  = getLogger(ctx)
			t   = time.Now()
		)

		defer func() {
			if r := recover(); r != nil {
				buf := make([]byte, 64<<10)
				buf = buf[:runtime.Stack(buf, false)]
				info := fmt.Sprintf("[PANIC] recovered: %s\n%s", r, buf)
				fmt.Print(info)
				lg.WithTag(logTagCronWorker + "_PANIC").Errorf(info)
				return
			}
		}()

		if err := cmdFunc(ctx); err != nil {
			lg.WithField("duration", time.Since(t).String()).Errorf("%v", err)
		} else {
			lg.WithField("duration", time.Since(t).String()).Infof("done")
		}
	})
	if err != nil {
		panic(err)
	}
	c.cronEntries = append(c.cronEntries, entry)
	getLogger(c.ctx).Infof("register cron work success")
}
