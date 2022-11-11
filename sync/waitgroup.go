package sync

import "sync"

// WaitGroup interface of a wait group as provided by `sync.WaitGroup`.
type WaitGroup interface {
	// Add increments or decrements the wait counter by the given delta.
	Add(delta int)
	// Done decrements the wait counter by exactly one.
	Done()
	// Wait waits until the wait counter has returned to zero.
	Wait()
}

// Synchronizer is an interface to setup the wait group of a component.
type Synchronizer interface {
	WaitGroup(WaitGroup)
}

// NewWaitGroup creates a new standard wait group.
func NewWaitGroup() WaitGroup {
	return &sync.WaitGroup{}
}

// LenientWaitGroup implements a lenient wait group for testing purposes based
// on a `sync.WaitGroup` that allows breaking the barrier by consuming the wait
// group counter completely without creating panics in consuming components.
type LenientWaitGroup struct {
	wg sync.WaitGroup
}

// NewLenientWaitGroup implements a lenient wait group for testing purposes
// based on a `sync.WaitGroup` that allows breaking the barrier by consuming
// the wait group counter completely without creating panics in consuming
// components.
func NewLenientWaitGroup() WaitGroup {
	return &LenientWaitGroup{}
}

// Add increments or decrements the wait group counter leniently by the delta,
// i.e. it does not fail, if the wait group counter is already consumed.
func (wg *LenientWaitGroup) Add(delta int) {
	if delta > 0 {
		wg.wg.Add(delta)
	} else {
		defer func() {
			if err := recover(); err != nil &&
				err.(string) == "sync: negative WaitGroup counter" {
				wg.wg.Add(1)
			}
		}()
		for d := delta; d < 0; d++ {
			wg.wg.Done()
		}
	}
}

// Done decrements the wait group counter leniently by one, i.e. it does not
// fail, if the wait group counter is already consumed.
func (wg *LenientWaitGroup) Done() {
	defer func() {
		if err := recover(); err != nil &&
			err.(string) == "sync: negative WaitGroup counter" {
			wg.wg.Add(1)
		}
	}()
	wg.wg.Done()
}

// Wait waits until the work group counter is completely consumed.
func (wg *LenientWaitGroup) Wait() {
	defer func() {
		if err := recover(); err != nil &&
			err.(string) != "sync: WaitGroup is reused before previous Wait has returned" {
			panic(err)
		}
	}()
	wg.wg.Wait()
}
