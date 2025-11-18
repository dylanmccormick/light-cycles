package game

import (
	"fmt"

	"github.com/dylanmccormick/light-cycles/protocol"
)

type Queue []protocol.Coordinate

func (q *Queue) Enqueue(cord protocol.Coordinate) {
	*q = append(*q, cord)
}

func (q *Queue) Dequeue() (protocol.Coordinate, error) {
	if len(*q) == 0 {
		return protocol.Coordinate{}, fmt.Errorf("Queue is empty")
	}
	value := (*q)[0]
	*q = (*q)[1:]

	return value, nil
}
