package main

import (
	"sync"
	"time"
)

//	NOTE
//	あるstringを登録する
//	アクセス回数が多いstringほどmapから追い出されにくい
//	きちんとして仕様にするほうがよい or LRUなどの既存ライブラリを用いる?!

type AccessMap struct {
	mutex      sync.RWMutex
	m          map[string]int
	interval   time.Duration
	maxWatcher int
}

func NewAccessMap() *AccessMap {
	m := make(map[string]int)
	return &AccessMap{
		mutex:      sync.RWMutex{},
		m:          m,
		interval:   time.Hour,
		maxWatcher: 128,
	}
}

func (a *AccessMap) Append(s string) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.m[s]++
}

func (a *AccessMap) AutoDelete(f func(s string)) {
	go func() {
		for _ = range time.Tick(a.interval) {
			func() {
				a.mutex.Lock()
				defer a.mutex.Unlock()
				for len(a.m) > a.maxWatcher {
					for k, v := range a.m {
						if v == 1 {
							f(k)
							delete(a.m, k)
							continue
						}
						a.m[k]--
					}
				}
			}()
		}
	}()
}

func (a *AccessMap) CopyMap() map[string]int {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	m := make(map[string]int)
	for k, v := range a.m {
		m[k] = v
	}
	return m
}
