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
	"errors"
	"fmt"
)

// Future represents an asynchronous operation that will complete sometime in the future.
//
// It is a read-only channel that can be used to retrieve the final result of a [Promise] with [Future.Wait].
type Future[R any] <-chan Result[R]

// NewFuture provides a simple way to create a Future for synchronous operations.
// This allows synchronous and asynchronous code to be composed seamlessly and separating initiation from running.
//
// The returned [Future] can be used to retrieve the eventual result of the [Promise].
func NewFuture[R any]() (Future[R], Promise[R]) {
	ch := make(chan Result[R], 1)

	return ch, ch
}

// NewFutureAsync runs f asynchronously, immediately returning a [Future] that can be used to retrieve the eventual
// result. This allows separating evaluating the result from computation.
func NewFutureAsync[R any](fn func() (R, error)) Future[R] {
	f, p := NewFuture[R]()
	go p.Do(fn)

	return f
}

// Awaitable is the underlying interface for [Future] and [Memoizer].
// Plain futures can only be queried once, while memoizers can be queried multiple times.
type Awaitable[R any] interface {
	Wait(ctx context.Context) (R, error) // Wait returns the final result of the associated [Promise].
	TryWait() (R, error)                 // TryWait returns the result when ready, [ErrNotReady] otherwise.
	Memoize() *Memoizer[R]               // Memoizer returns a [Memoizer] which can be queried multiple times.

	channel() <-chan Result[R]
	addRunning()
	releaseRunning()
	processResult(r Result[R], ok bool) Result[R]
}

// ErrNotReady is returned when a future is not complete.
var ErrNotReady = errors.New("future not ready")

// ErrNoResult is returned when a future completes but has no defined result value.
var ErrNoResult = errors.New("no result")

// Wait returns the final result of the associated [Promise].
// It can only be called once and blocks until a result is received or the context is canceled.
// If you need to read multiple times from a [Future] wrap it with [Future.Memoize].
func (f Future[R]) Wait(ctx context.Context) (R, error) {
	select {
	case r, ok := <-f:
		return f.processResult(r, ok).V()

	case <-ctx.Done():
		return *new(R), fmt.Errorf("future wait: %w", ctx.Err())
	}
}

// TryWait returns the result when ready, [ErrNotReady] otherwise.
func (f Future[R]) TryWait() (R, error) {
	select {
	case r, ok := <-f:
		return f.processResult(r, ok).V()

	default:
		return *new(R), ErrNotReady
	}
}

// Memoize creates a new [Memoizer], consuming the [Future].
func (f Future[R]) Memoize() *Memoizer[R] {
	return &Memoizer[R]{done: make(chan struct{}), future: f}
}

func (f Future[R]) processResult(r Result[R], ok bool) Result[R] {
	if ok {
		return r
	}

	return errorResult[R]{err: ErrNoResult}
}

func (f Future[R]) channel() <-chan Result[R] { //nolint:unused
	return f
}

func (f Future[R]) addRunning() {} //nolint:unused

func (f Future[R]) releaseRunning() {} //nolint:unused
