package main

import (
	"sync"
	"time"
)

const (
	NoExpiration            int64         = -1
	DefaultExpiration       time.Duration = time.Hour * 24
	DefaultEvictionInterval time.Duration = time.Minute * 5
)

type Cache interface {
	Get(k string) (interface{}, bool)
	Set(k string, v interface{})
	SetWithTime(k string, v interface{}, d time.Duration)
	Delete(k string)
	DeleteExpired()
	Clear()

	Finalizable
}

type Item struct {
	Value   interface{}
	Expires int64
}

func (i Item) Expired() bool {
	return i.Expires != NoExpiration && time.Now().UnixNano() > i.Expires
}

type diskCache struct {
	sync.RWMutex
	items map[string]Item

	expiration time.Duration

	worker worker
}

func NewDiskCache(d time.Duration, e time.Duration) Cache {
	c := &diskCache{
		expiration: d,
		items:      make(map[string]Item),
		worker:     NewWorker(e),
	}

	c.worker.Run(c)

	return c
}

func (cache *diskCache) Get(k string) (interface{}, bool) {
	cache.RLock()

	item, f := cache.items[k]
	if !f {
		cache.RUnlock()
		return nil, false
	}

	if item.Expired() {
		cache.RUnlock()
		return nil, false
	}

	cache.RUnlock()

	return item.Value, true
}

func (cache *diskCache) SetWithTime(k string, v interface{}, d time.Duration) {
	cache.Lock()

	var e int64
	if d > 0 {
		e = time.Now().Add(d).UnixNano()
	}

	cache.items[k] = Item{
		Value:   v,
		Expires: e,
	}

	cache.Unlock()
}

func (cache *diskCache) Set(k string, v interface{}) {
	cache.Lock()

	e := time.Now().Add(cache.expiration).UnixNano()

	cache.items[k] = Item{
		Value:   v,
		Expires: e,
	}

	cache.Unlock()
}

func (cache *diskCache) Delete(k string) {
	cache.Lock()

	if _, f := cache.items[k]; f {
		delete(cache.items, k)
	}

	cache.Unlock()
}

func (cache *diskCache) Clear() {
	cache.Lock()

	cache.items = make(map[string]Item)

	cache.Unlock()
}

func (cache *diskCache) DeleteExpired() {
	cache.Lock()

	for k, v := range cache.items {
		if v.Expired() {
			delete(cache.items, k)
		}
	}

	cache.Unlock()
}

func (cache *diskCache) Finalize() {
	cache.worker.Stop()
	cache.Clear()
}

type worker struct {
	interval time.Duration
	stop     chan bool
}

func NewWorker(d time.Duration) worker {
	return worker{
		interval: d,
		stop:     make(chan bool),
	}
}

func (w *worker) Run(cache Cache) {
	ticker := time.NewTicker(w.interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				cache.DeleteExpired()
			case <-w.stop:
				ticker.Stop()
				return
			}
		}
	}()
}

func (w *worker) Stop() {
	w.stop <- true
}
