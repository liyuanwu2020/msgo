package mspool

import (
	"sync"
	"sync/atomic"
	"time"
)

type signal struct {
}

const DefaultExpire = 3

type Pool struct {
	//池容量
	cap int32
	//正在运行的数量
	running int32
	//空闲 worker
	workers []*Worker
	//过期时间
	expire time.Duration
	//释放资源
	release chan signal
	//保证操作worker的安全性
	lock sync.Mutex
	//保证释放只能调用一次
	once sync.Once
}

func NewPool(cap int) (*Pool, error) {
	return NewExpirePool(cap, DefaultExpire)
}

func NewExpirePool(cap int, expire int) (*Pool, error) {
	if cap < 1 {

	}
	return &Pool{
		cap:     int32(cap),
		expire:  time.Duration(expire) * time.Second,
		release: make(chan signal, 1),
	}, nil
}

// Submit 提交任务
func (p *Pool) Submit(task func()) error {
	//从空闲 worker 池中获取一个 worker
	w := p.GetWorker()
	//使用 worker 执行 task
	w.task <- task
	w.pool.incrRunning()
	return nil
}

// GetWorker 核心代码
func (p *Pool) GetWorker() *Worker {
	//1. 目的获取pool里面的worker
	//2. 如果 有空闲的worker 直接获取
	//3. 如果没有空闲的worker，要新建一个worker
	//4. 如果正在运行的workers 如果大于pool容量，阻塞等待，worker释放
	var w *Worker
	i := len(p.workers)
	if i > 0 {
		p.lock.Lock()
		w = p.workers[0]
		p.workers = p.workers[1:]
		p.lock.Unlock()
	} else {
		if p.cap < p.running {
			w = &Worker{
				pool: p,
				task: make(chan func(), 1),
			}
			w.Run()
		} else {
			for {
				p.lock.Lock()
				idleWorkers := p.workers
				n := len(idleWorkers) - 1
				if n > 0 {
					w = idleWorkers[n]
					idleWorkers[n] = nil
					p.workers = idleWorkers[:n]
				}
				p.lock.Unlock()
			}
		}
	}
	return w
}

func (p *Pool) incrRunning() {
	atomic.AddInt32(&p.running, 1)
}

func (p *Pool) decrRunning() {
	atomic.AddInt32(&p.running, -1)
}

func (p *Pool) PutWorker(w *Worker) {
	w.lastTime = time.Now()
	p.lock.Lock()
	p.workers = append(p.workers, w)
	p.lock.Unlock()
}
