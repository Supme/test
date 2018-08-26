package main

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	fmt.Println("Starting...")
	rand.Seed(time.Now().UnixNano())
	wg := &sync.WaitGroup{}

	broker := NewBroker(20)
	for i := 1; i <= 300; i++ {
		if i == 100 {
			broker.SetMax(10)
		}
		if i == 200 {
			broker.SetMax(15)
		}
		broker.Next()
		wg.Add(1)
		go func(i int) {
			fmt.Println("Start worker number", i, "current workers", broker.GetCurrent(), "from maximum", broker.GetMax())
			time.Sleep(time.Duration(rand.Int63n(500000)+500000) * time.Microsecond)
			broker.Ready()
			fmt.Println("Finish worker", i)
			wg.Done()
		}(i)
	}
	wg.Wait()
}

type Broker struct {
	maxCount     int64
	currentCount int64
	next         chan struct{}
	ready        chan struct{}
}

func NewBroker(count int64) *Broker {
	b := Broker{maxCount: count, currentCount: 0}
	b.next = make(chan struct{}, 1)
	b.ready = make(chan struct{}, 1)
	b.start()
	return &b
}

func (b *Broker) start() {
	go func() {
		for range b.ready {
			atomic.AddInt64(&b.currentCount, -1)
		}
	}()
	go func() {
		for {
			if atomic.LoadInt64(&b.currentCount) < atomic.LoadInt64(&b.maxCount) {
				b.next <- struct{}{}
				atomic.AddInt64(&b.currentCount, 1)
			}
		}
	}()
}

func (b *Broker) SetMax(count int64) {
	atomic.StoreInt64(&b.maxCount, count)
}

func (b *Broker) GetCurrent() int64 {
	return atomic.LoadInt64(&b.currentCount)
}

func (b *Broker) GetMax() int64 {
	return atomic.LoadInt64(&b.maxCount)
}

func (b *Broker) Next() {
	<-b.next
}

func (b *Broker) Ready() {
	b.ready <- struct{}{}
}
