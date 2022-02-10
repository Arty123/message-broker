package service

import (
	"errors"
	"sync"
	"time"
)

const timeoutMaxLimit = 30 // sec
var ErrorQueueTimeoutLimit = errors.New("timeout limit exceeded")

type node struct {
	data string
	next *node
}

type Queue struct {
	head  *node
	tail  *node
	count int
	lock  *sync.Mutex
	cond  *sync.Cond
}

func NewQueue() *Queue {
	q := &Queue{}
	q.lock = &sync.Mutex{}
	q.cond = sync.NewCond(q.lock)
	return q
}

func (q *Queue) Len() int {
	return q.count
}

func (q *Queue) Enqueue(item string) {
	q.lock.Lock()
	defer q.lock.Unlock()

	n := &node{data: item}

	if q.tail == nil {
		q.tail = n
		q.head = n
	} else {
		q.tail.next = n
		q.tail = n
	}
	q.count++
	q.cond.Signal()
}

func (q *Queue) Dequeue() string {
	q.lock.Lock()
	defer q.lock.Unlock()

	return q.dequeueWithoutBlocking()
}

func (q *Queue) dequeueWithoutBlocking() string {
	if q.head == nil {
		return ""
	}

	n := q.head
	q.head = n.next

	if q.head == nil {
		q.tail = nil
	}
	q.count--

	return n.data
}

func (q *Queue) DequeueWithTimeout(limit int) (string, error) {
	if limit > timeoutMaxLimit {
		return "", ErrorQueueTimeoutLimit
	}

	q.lock.Lock()
	defer q.lock.Unlock()

	resultCh := make(chan string)
	go func() {
		for q.Len() == 0 {
			q.cond.Wait()
		}

		resultCh <- q.dequeueWithoutBlocking()
	}()

	timer := time.NewTimer(time.Duration(limit) * time.Second)

	for {
		select {
		case <-timer.C:
			q.Enqueue("timeout")
		case result := <-resultCh:
			return result, nil
		}
	}
}

func (q *Queue) Head() string {
	q.lock.Lock()
	defer q.lock.Unlock()

	n := q.head
	if n == nil {
		return ""
	}

	return n.data
}
