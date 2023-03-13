package mspool

import "time"

type Worker struct {
	pool *Pool
	task chan func()
	//执行任务开始时间
	lastTime time.Time
}

func (w *Worker) Run() {
	go w.running()
}

func (w *Worker) running() {
	for f := range w.task {
		f()
		w.pool.PutWorker(w)
		w.pool.decrRunning()
	}
}
