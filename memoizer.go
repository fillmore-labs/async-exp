// Copyright 2023-2024 Oliver Eikemeier. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package async

import (
	"context"
	"fmt"
	"runtime/trace"
	"sync/atomic"
)

// Memoizer caches results from a [Future] to enable multiple queries and avoid unnecessary recomputation.
type Memoizer[R any] struct {
	future  <-chan Result[R] // future is the [Future] being cached
	running atomic.Int32     // number of goroutines at a select
	done    chan struct{}    // done signals when future has completed
	value   Result[R]        // value will hold the cached result
}

// Wait returns the cached result or blocks until a result is available or the context is canceled.
func (m *Memoizer[R]) Wait(ctx context.Context) (R, error) {
	defer trace.StartRegion(ctx, "asyncMemoizerWait").End()
	m.addRunning()
	select { // wait for future completion or context cancel
	case r, ok := <-m.channel():
		return m.processResult(r, ok).V()

	case <-ctx.Done():
		m.releaseRunning()

		return *new(R), fmt.Errorf("memoizer wait: %w", ctx.Err())
	}
}

// TryWait returns the cached result when ready, [ErrNotReady] otherwise.
func (m *Memoizer[R]) TryWait() (R, error) {
	m.addRunning()
	select {
	case r, ok := <-m.channel():
		return m.processResult(r, ok).V()

	default:
		m.releaseRunning()

		return *new(R), ErrNotReady
	}
}

// Memoize returns this [Memoizer].
func (m *Memoizer[R]) Memoize() *Memoizer[R] {
	return m
}

// processResult handles caching the result when received on the future channel.
// It signals completion on done after updating value.
func (m *Memoizer[R]) processResult(r Result[R], ok bool) Result[R] {
	if ok { // We got a result
		m.value = r
		close(m.done)
		m.releaseRunning() // This has to be done after signalling done

		return r
	}

	if m.thereAreOthers() { // Wait for other goroutines to resolve the closed channel
		<-m.done

		return m.value
	}

	// This is the last goroutine and the channel is closed
	select {
	case <-m.done: // Some other goroutine resolved

	default: // The channel closed without a result
		m.value = errorResult[R]{ErrNoResult}
		close(m.done)
	}
	m.releaseRunning()

	return m.value
}

// channel simply returns the underlying future channel.
func (m *Memoizer[R]) channel() <-chan Result[R] {
	return m.future
}

// addRunning manage the running counter atomically.
func (m *Memoizer[R]) addRunning() {
	m.running.Add(1)
}

// releaseRunning manage the running counter atomically.
func (m *Memoizer[R]) releaseRunning() {
	m.running.Add(-1)
}

// thereAreOthers checks if this goroutine is the only remaining one after the channel is closed.
//
// How does this work?
//
// We use an atomic counter to track the number of goroutines running. We are leaving the running phase by
// decrementing the counter and wait for the others to finish and resolve the value.
//
// If after decrementing the counter is 0, we know that there are no other goroutines running (only waiting), so we have
// to resolve ourselves.
//
// If now another goroutine starts, increasing the counter to 1 again, we can not swap out the 0 count to 1 and leave
// the work to the new goroutine.
//
// If we can swap out the counter, every later started new goroutine sees that there is another running and will leave
// resolving to it.
func (m *Memoizer[R]) thereAreOthers() bool {
	return m.running.Add(-1) != 0 || !m.running.CompareAndSwap(0, 1)
}
