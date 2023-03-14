package mspool

import (
	"errors"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type signal struct {
}

const DefaultExpire = 3

var (
	ErrorInValidCap    = errors.New("pool cap can not < 1")
	ErrorInValidExpire = errors.New("pool expire can not < 1")
	ErrorHasClosed     = errors.New("pool has bean released")
)

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
		return nil, ErrorInValidCap
	}
	if expire < 1 {
		return nil, ErrorInValidExpire
	}
	p := &Pool{
		cap:     int32(cap),
		expire:  time.Duration(expire) * time.Second,
		release: make(chan signal, 1),
	}
	go p.expireWorker()
	return p, nil
}

// 定时清理过期的空闲worker
func (p *Pool) expireWorker() {
	ticker := time.NewTicker(p.expire)
	for range ticker.C {
		if p.IsClosed() {
			break
		}
		idleWorkers := p.workers
		for i, w := range idleWorkers {
			log.Println(i, ":", w.lastTime, time.Now().Sub(w.lastTime))
			if time.Now().Sub(w.lastTime) > p.expire {
				w.task = nil
				w.pool = nil
				idleWorkers[i] = nil
			}
		}
	}
}

func (p *Pool) removeWorker(i int, arr []*Worker) {
	start := 0
	for index, w := range arr {
		if index != i {
			arr[start] = w
			start++
		} else {
			w.task <- nil
			w.task = nil
			w.pool = nil
		}
	}
	p.workers = arr[:start]
}

// Submit 提交任务
func (p *Pool) Submit(task func()) error {
	if len(p.release) > 0 {
		return ErrorHasClosed
	}
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
	idleWorkers := p.workers
	n := len(idleWorkers) - 1
	if n > -1 {
		p.lock.Lock()
		w := idleWorkers[n]
		p.workers = idleWorkers[:n]
		p.lock.Unlock()
		return w
	}
	if p.cap > p.running {
		//创建worker
		w := &Worker{
			pool: p,
			task: make(chan func(), 1),
		}
		w.Run()
		return w
	}
	for {
		p.lock.Lock()
		idleWorkers := p.workers
		n := len(idleWorkers) - 1
		if n > 0 {
			w := idleWorkers[n]
			idleWorkers[n] = nil
			p.workers = idleWorkers[:n]
			p.lock.Unlock()
			return w
		}
		p.lock.Unlock()
	}
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

func (p *Pool) Release() {
	p.once.Do(func() {
		//只执行一次
		p.lock.Lock()
		workers := p.workers
		for i, w := range workers {
			w.task = nil
			w.pool = nil
			workers[i] = nil
		}
		p.workers = nil
		p.lock.Unlock()
		p.release <- signal{}
	})
}

func (p *Pool) IsClosed() bool {
	return len(p.release) > 0
}

func (p *Pool) Restart() bool {
	if len(p.release) <= 0 {
		return true
	}
	_ = <-p.release
	return true
}
