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
}

func NewQueue() *Queue {
	q := &Queue{}
	q.lock = &sync.Mutex{}
	return q
}

func (q *Queue) Len() int {
	q.lock.Lock()
	defer q.lock.Unlock()
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
}

func (q *Queue) Dequeue() string {
	q.lock.Lock()
	defer q.lock.Unlock()

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

	timer := time.NewTimer(time.Duration(limit) * time.Second)

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			value := q.Dequeue()
			if value != "" {
				if !timer.Stop() {
					<-timer.C
				}
				return value, nil
			}
		case <-timer.C:
			return "", nil
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
