package dqueue

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/Yu-33/gohelper/structs/container"
	"github.com/Yu-33/gohelper/structs/minheap"
)

const (
	defaultCapacity = 64
)

type Value interface{}
type Receiver func(value Value)

// Item represents an element in priority queue.
type Item struct {
	Expiration int64 // Expiration time of nanoseconds timestamp.
	Value      Value
}

// Implements container.Comparator.
func (k1 *Item) Compare(target container.Comparator) int {
	k2 := target.(*Item)
	if k1.Expiration < k2.Expiration {
		return -1
	}
	if k1.Expiration > k2.Expiration {
		return 1
	}
	return 0
}

// DQueue implements a delay queue base on priority queue (min heap).
// Inspired by https://github.com/RussellLuo/timingwheel/blob/master/delayqueue/delayqueue.go
type DQueue struct {
	C chan Value // Notify channel

	mu *sync.Mutex
	pq *minheap.MinHeap // priority queue implemented by min heap.

	sleeping int32         // Similar to the sleeping state of runtime.timers. 1 => true, 0 => false.
	wakeupC  chan struct{} // Used to wakeup poll goroutine when item add to queue head.

	exitC chan struct{}   // Used to make poll goroutine exit.
	wg    *sync.WaitGroup // Used wait polling exit when close queue.
}

// Default creates an DQueue with default parameters.
func Default() *DQueue {
	return New(defaultCapacity)
}

// New creates an DQueue with given c(queue capacity).
func New(c int) *DQueue {
	dq := newDQueue(c)
	go dq.polling()
	return dq
}

// newDQueue is an internal helper function that really creates an DQueue.
func newDQueue(c int) *DQueue {
	return &DQueue{
		C:        make(chan Value),
		pq:       minheap.New(c),
		mu:       new(sync.Mutex),
		sleeping: 0,
		wakeupC:  make(chan struct{}),
		exitC:    make(chan struct{}),
		wg:       new(sync.WaitGroup),
	}
}

// After adds the value with specified delay time to queue.
func (dq *DQueue) After(delay time.Duration, value Value) {
	dq.offer(dq.timeNow().Add(delay).UnixNano(), value)
}

// Expire adds the value with specified expiration timestamp(in nanoseconds) to queue.
func (dq *DQueue) Expire(exp int64, value Value) {
	dq.offer(exp, value)
}

func (dq *DQueue) offer(exp int64, value Value) {
	dq.mu.Lock()
	item := &Item{Expiration: exp, Value: value}
	index := dq.pq.Push(item)
	dq.mu.Unlock()

	// A new item with the earliest expiration is added.
	if index == 0 && atomic.CompareAndSwapInt32(&dq.sleeping, 1, 0) {
		dq.wakeupC <- struct{}{}
	}
}

func (dq *DQueue) timeNow() time.Time {
	return time.Now()
}

// Receive register a func to be called if some item expires.
func (dq *DQueue) Receive(f Receiver) {
	dq.wg.Add(1)
	defer dq.wg.Done()

	for {
		select {
		case <-dq.exitC:
			return
		case value := <-dq.C:
			f(value)
		}
	}
}

// Close to notify the polling exit. can't be called repeatedly.
func (dq *DQueue) Close() {
	close(dq.exitC)
	// Waiting for polling exit.
	dq.wg.Wait()
}

func (dq *DQueue) peekAndShift() (*Item, int64) {
	element := dq.pq.Peek()
	if element == nil {
		// queue is empty
		return nil, 0
	}

	item := element.(*Item)
	delay := item.Expiration - dq.timeNow().UnixNano()
	if delay > 0 {
		return nil, delay
	}

	// Removed from queue top.
	_ = dq.pq.Pop()
	return item, 0
}

func (dq *DQueue) polling() {
	dq.wg.Add(1)
	defer func() {
		// Reset the sleeping states.
		atomic.StoreInt32(&dq.sleeping, 0)
		dq.wg.Done()
	}()

LOOP:
	for {
		dq.mu.Lock()
		item, delay := dq.peekAndShift()
		if item == nil {
			// No items left or at least one item is pending.

			// We must ensure the atomicity of the whole operation, which is
			// composed of the above PeekAndShift and the following StoreInt32,
			// to avoid possible race conditions between Offer and Poll.
			atomic.StoreInt32(&dq.sleeping, 1)
		}
		dq.mu.Unlock()

		// No items in queue. Waiting to be wakeup.
		if item == nil && delay == 0 {
			select {
			case <-dq.exitC:
				break LOOP
			case <-dq.wakeupC:
			}

			continue LOOP
		}

		// At least one item is pending. Go to sleep.
		if delay > 0 {
			select {
			case <-dq.exitC:
				break LOOP
			case <-dq.wakeupC:
				// A new item with an "earlier" expiration than the current "earliest" one is added.
			case <-time.After(time.Duration(delay)):
				// The current "earliest" item expires.

				// Reset the sleeping state since there's no need to receive from wakeupC.
				if atomic.SwapInt32(&dq.sleeping, 0) == 0 {
					// A caller of Offer() is being blocked on sending to wakeupC,
					// drain wakeupC to unblock the caller.
					<-dq.wakeupC
				}
			}

			continue LOOP
		}

		// Send expired element to channel.
		select {
		case <-dq.exitC:
			break LOOP
		case dq.C <- item.Value:
			// The expired element has been sent out successfully.
		}
	}
}
