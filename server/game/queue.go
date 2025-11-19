package game

import (
	"fmt"

	"github.com/dylanmccormick/light-cycles/protocol"
)

type Queue []protocol.TrailSegment

func (q *Queue) Enqueue(ts protocol.TrailSegment) {
	*q = append(*q, ts)
}

func (q *Queue) Dequeue() (protocol.TrailSegment, error) {
	if len(*q) == 0 {
		return protocol.TrailSegment{}, fmt.Errorf("Queue is empty")
	}
	value := (*q)[0]
	*q = (*q)[1:]

	return value, nil
}
