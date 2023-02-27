package msg

import (
	"container/heap"
)

func NewPriorityQueue(f LevelFunc) Queue {
	q := &priorityQueue{pml: &priorityMetadataList{levels: f()}, ch: make(chan *Metadata, 1)}
	heap.Init(q.pml)
	return q
}

type priorityQueue struct {
	pml *priorityMetadataList
	ch  chan *Metadata
}

func (q *priorityQueue) Write(x *Metadata) {
	heap.Push(q.pml, x)
}

func (q *priorityQueue) Read() <-chan *Metadata {
	for {
		v, ok := heap.Pop(q.pml).(*Metadata)
		if !ok {
			continue
		}
		q.ch <- v
		return q.ch
	}
}

func (q *priorityQueue) ListAndClear() []*Metadata {
	result := make([]*Metadata, len(q.pml.value))
	copy(result, q.pml.value)
	q.pml.value = q.pml.value[:0]
	q.pml.levels = nil
	return result
}

func (q *priorityQueue) Length() int {
	return q.pml.Len()
}

func (q *priorityQueue) Close() error {
	close(q.ch)
	return nil
}

type priorityMetadataList struct {
	value  []*Metadata
	levels map[string]uint8
}

func (p *priorityMetadataList) Push(x interface{}) {
	p.value = append(p.value, x.(*Metadata))
}

func (p *priorityMetadataList) Pop() interface{} {
	n := len(p.value)
	if n > 0 {
		x := (p.value)[n-1]
		p.value = p.value[:n-1]
		return x
	}
	return nil
}

func (p *priorityMetadataList) Len() int {
	return len(p.value)
}

func (p *priorityMetadataList) Less(i, j int) bool {
	return p.levels[p.value[i].ServiceName] < p.levels[p.value[j].ServiceName]
}

func (p *priorityMetadataList) Swap(i, j int) {
	p.value[i], p.value[j] = p.value[j], p.value[i]
}
