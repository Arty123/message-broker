package service

import "sync"

type QueueResolver interface {
	ResolveQueue(name string) *Queue
}

type queueResolver struct {
	QueueMap map[string]*Queue
	Lock     *sync.Mutex
}

const (
	NameQueue  = "name"
	ColorQueue = "color"
)

// NewQueueResolver is a structure for communicating with the broker through another transport.
// @todo enable DI in the project and inject it as a singleton everywhere
func NewQueueResolver() QueueResolver {
	queueMap := make(map[string]*Queue)
	lock := &sync.Mutex{}
	return &queueResolver{QueueMap: queueMap, Lock: lock}
}

func (q *queueResolver) ResolveQueue(name string) *Queue {
	q.Lock.Lock()
	defer q.Lock.Unlock()

	if queue, ok := q.QueueMap[name]; ok {
		return queue
	}

	q.QueueMap[name] = NewQueue()
	return q.QueueMap[name]
}
