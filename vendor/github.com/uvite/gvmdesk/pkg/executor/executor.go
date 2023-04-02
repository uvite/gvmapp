package executor

import (
	"context"
	"fmt"
	"github.com/uvite/gvmdesk/pkg/backend"
	"github.com/uvite/gvmdesk/pkg/bot"
	"github.com/uvite/gvmdesk/pkg/platform"

	"github.com/influxdata/influxdb/v2/kit/tracing"
	"github.com/influxdata/influxdb/v2/query"
	taskmodel "github.com/uvite/gvmdesk/pkg/model"
	"go.uber.org/zap"
	"sync"
	"time"
)

type Promise interface {
	ID() platform.ID
	Cancel(ctx context.Context)
	Done() <-chan struct{}
	Error() error
}

const (
	maxPromises       = 1000
	defaultMaxWorkers = 100

	lastSuccessOption = "tasks.lastSuccessTime"
)

// promise represents a promise the executor makes to finish a run's execution asynchronously.
type promise struct {
	Id    platform.ID
	run   *taskmodel.Run
	task  *taskmodel.Task
	Exbot *bot.Exbot
	done  chan struct{}
	err   error
	bot   *bot.Deskbot

	createdAt time.Time
	startedAt time.Time

	ctx        context.Context
	cancelFunc context.CancelFunc
	botcancel  context.CancelFunc
}

// ID is the id of the run that was created
func (p *promise) ID() platform.ID {
	return p.run.ID
}
func (p *promise) initBot() {
	p.bot = bot.NewBot(p.ctx, *p.task, p.Exbot)

}

// Cancel is used to cancel a executing query
func (p *promise) Cancel(ctx context.Context) {
	// call cancelfunc
	p.cancelFunc()

	// wait for ctx.Done or p.Done
	select {
	case <-p.Done():
	case <-ctx.Done():
	}
}

// Done provides a channel that closes on completion of a promise
func (p *promise) Done() <-chan struct{} {
	return p.done
}

// Error returns the error resulting from a run execution.
// If the execution is not complete error waits on Done().
func (p *promise) Error() error {
	<-p.done
	return p.err
}

// Executor it a task specific executor that works with the new scheduler system.
type Executor struct {
	log *zap.Logger
	ts  taskmodel.TaskService
	tcs backend.TaskControlService
	qs  query.QueryService

	// currentPromises are all the promises we are made that have not been fulfilled
	currentPromises sync.Map

	// futurePromises are promises that are scheduled to be executed in the future
	currentCancel sync.Map

	// keep a pool of promise's we have in queue
	promiseQueue chan *promise

	// keep a pool of execution workers.
	workerPool  sync.Pool
	workerLimit chan struct{}

	Exbot *bot.Exbot
}

// NewExecutor creates a new task executor
func NewExecutor(log *zap.Logger, ts taskmodel.TaskService, eb *bot.Exbot) *Executor {

	e := &Executor{
		log:   log,
		ts:    ts,
		Exbot: eb,

		currentPromises: sync.Map{},
		currentCancel:   sync.Map{},
		promiseQueue:    make(chan *promise, maxPromises),
		workerLimit:     make(chan struct{}, 100),
	}

	wm := &workerMaker{
		e: e,
	}

	//go e.processScheduledTasks()

	e.workerPool = sync.Pool{New: wm.new}
	return e
}

func (e *Executor) startWorker() {
	// see if have available workers
	select {
	case e.workerLimit <- struct{}{}:
	default:
		// we have reached our worker limit and we cannot start any more.
		return
	}
	// fire up some workers
	worker := e.workerPool.Get().(*worker)

	fmt.Println("[worker]", worker)
	if worker != nil {
		// if the worker is nil all the workers are busy and one of them will pick up the work we enqueued.
		go func() {
			// don't forget to put the worker back when we are done
			defer e.workerPool.Put(worker)
			worker.work()

			// remove a struct from the worker limit to another worker to work
			<-e.workerLimit
		}()
	}
}

// Cancel a run of a specific task.
func (e *Executor) Cancel(ctx context.Context, runID platform.ID) error {
	// find the promise
	val, ok := e.currentPromises.Load(runID)
	if !ok {
		return nil
	}
	promise := val.(*promise)

	// call cancel on it.
	promise.Cancel(ctx)

	return nil
}

func (e *Executor) Close(ctx context.Context, runID platform.ID) error {
	// find the promise
	val, ok := e.currentPromises.Load(runID)
	ddd, _ := e.currentCancel.Load(runID)
	if !ok {
		return nil
	}

	promise := val.(*promise)
	aaa := ddd.(context.CancelFunc)

	aaa()

	promise.ctx.Done()

	return nil
}

func (e *Executor) createPromise(ctx context.Context, id platform.ID) (*promise, error) {
	span, ctx := tracing.StartSpanFromContext(ctx)
	defer span.Finish()

	t, err := e.ts.FindTaskByID(ctx, id)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)
	// create promise
	p := &promise{
		Id:         id,
		Exbot:      e.Exbot,
		task:       t,
		createdAt:  time.Now().UTC(),
		done:       make(chan struct{}),
		ctx:        ctx,
		cancelFunc: cancel,
	}

	// insert promise into queue to be worked
	// when the queue gets full we will hand and apply back pressure to the scheduler
	e.promiseQueue <- p

	// insert the promise into the registry
	e.currentPromises.Store(id, p)

	return p, nil
}

// PromisedExecute begins execution for the tasks id with a specific scheduledFor time.
// When we execute we will first build a run for the scheduledFor time,
// We then want to add to the queue anything that was manually queued to run.
// If the queue is full the call to execute should hang and apply back pressure to the caller
// We then start a worker to work the newly queued jobs.
func (e *Executor) PromisedExecute(ctx context.Context, id platform.ID) (Promise, error) {

	// create a run
	p, err := e.createPromise(ctx, id)
	if err != nil {
		return nil, err
	}

	e.startWorker()
	return p, nil
}

func (e *Executor) LoadExistingScheduleRuns(ctx context.Context) error {
	tasks, _, err := e.ts.FindTasks(ctx, taskmodel.TaskFilter{})

	fmt.Println("{task}", len(tasks))
	if err != nil {
		e.log.Error("err finding tasks:", zap.Error(err))
		return err
	}
	for _, t := range tasks {

		ctx, cancel := context.WithCancel(ctx)
		// create promise
		p := &promise{
			Id:         t.ID,
			Exbot:      e.Exbot,
			task:       t,
			createdAt:  time.Now().UTC(),
			done:       make(chan struct{}),
			ctx:        ctx,
			cancelFunc: cancel,
		}
		e.promiseQueue <- p
		e.currentPromises.Store(t.ID, p)
		e.startWorker()
	}

	return nil
}

type workerMaker struct {
	e *Executor
}

func (wm *workerMaker) new() interface{} {
	return &worker{
		e: wm.e,
	}
}

type worker struct {
	e *Executor
}

func (w *worker) work() {
	fmt.Println("run=============")
	// loop until we have no more work to do in the promise queue
	for {
		var prom *promise
		// check to see if we can execute
		select {
		case p, ok := <-w.e.promiseQueue:
			fmt.Println("[p,ok]", p, ok)
			if !ok {
				///p.bot.Gvm.CancelFun()
				// the promiseQueue has been closed
				return
			}
			prom = p
			prom.initBot()
			prom.bot.OnklineClose()
			//
			//p.botcancel = p.bot.Gvm.CancelFun
			fmt.Println("pppppp", prom.Id)
			w.e.currentCancel.Store(prom.Id, prom.bot.Gvm.CancelFun)

		default:
			fmt.Println("44444")
			// if nothing is left in the queue we are done
			return
		}
		//
		//// check to make sure we are below the limits.
		for {

			// sleep
			select {
			// If done the promise was canceled
			case <-prom.ctx.Done():
				prom.bot.Gvm.CancelFun()
				fmt.Println("324324")
				//prom.bot.Gvm
				prom.err = taskmodel.ErrRunCanceled
				close(prom.done)
				return
			case <-time.After(15 * time.Second):
				prom.bot.Gvm.Run()
			}
		}

		//fmt.Println("[prom]", prom)
		// execute the promise
		//w.executeQuery(prom)

		// close promise done channel and set appropriate error
		//close(prom.done)
		//
		//// remove promise from registry
		//w.e.currentPromises.Delete(prom.run.ID)
	}
}

func (w *worker) start(p *promise) {
	// trace
	span, ctx := tracing.StartSpanFromContext(p.ctx)
	defer span.Finish()
	fmt.Println("[start]", ctx)

	p.startedAt = time.Now()
}

func (w *worker) finish(p *promise, rs taskmodel.RunStatus, err error) {
	span, ctx := tracing.StartSpanFromContext(p.ctx)
	defer span.Finish()
	fmt.Println(ctx)

}

func (w *worker) executeQuery(p *promise) {
	go p.bot.Gvm.Run()
	//nb := bot.NewBot(p.ctx, *p.task, p.Exbot)
	//fmt.Printf("333333333\n%+v", p.task)
	////nb.Onkline()
	//v, e := nb.Gvm.Run()
	fmt.Println("[=======]")

}
