package mspool

import "time"

type Worker struct {
	pool *Pool
	task chan func()
	//执行任务开始时间
	lastTime time.Time
}
