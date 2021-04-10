package service

import (
	"context"
	"errors"
	"sync"
	"time"
)

const timeoutLimit = 30 // sec
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

func (q *Queue) DequeueWithTimeout(t int) (string, error) {
	if t > timeoutLimit {
		return "", ErrorQueueTimeoutLimit
	}

	d := time.Now().Add(time.Duration(t) * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), d)
	defer cancel()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			value := q.Dequeue()
			if value != "" {
				return value, nil
			}
		case <-ctx.Done():
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
