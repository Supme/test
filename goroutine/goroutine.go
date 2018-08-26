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

	broker := NewBroker(10)
	for i := 1; i <= 500; i++ {
		if i == 150 {
			broker.SetCount(5)
		}
		if i == 350 {
			broker.SetCount(15)
		}
		broker.Next()
		wg.Add(1)
		go func(i int) {
			fmt.Println("Start worker number", i)
			time.Sleep(time.Duration(rand.Int63n(1000)) * time.Millisecond)
			fmt.Println("Finish worker", i)
			broker.Ready()
			wg.Done()
		}(i)
	}
	wg.Wait()
}

type Broker struct {
	maxCount     int64
	currentCount int64
	next         chan struct{}
	free         chan struct{}
}

func NewBroker(count int64) *Broker {
	b := Broker{maxCount: count, currentCount: 0}
	b.next = make(chan struct{}, 1)
	b.free = make(chan struct{}, 1)
	b.start()
	return &b
}

func (b *Broker) start() {
	go func() {
		for range b.free {
			atomic.AddInt64(&b.currentCount, -1)
		}
	}()
	go func() {
		for {
			if current := atomic.LoadInt64(&b.currentCount); current <= b.maxCount {
				b.next <- struct{}{}
				fmt.Println("Current count workers = ", current)
				atomic.AddInt64(&b.currentCount, 1)
			}
		}
	}()
}

func (b *Broker) SetCount(count int64) {
	b.maxCount = count
}

func (b *Broker) Next() {
	<-b.next
}

func (b *Broker) Ready() {
	b.free <- struct{}{}
}
