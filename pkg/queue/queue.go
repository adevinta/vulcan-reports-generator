/*
Copyright 2021 Adevinta
*/

package queue

import (
	"context"
	"sync"
)

// Consumer represents a consumer for a queue.
type Consumer interface {
	Start(ctx context.Context, wg *sync.WaitGroup, processor Processor)
}

// Processor represents a queue message processor.
type Processor interface {
	ProcessMessage(mssg string) error
}
