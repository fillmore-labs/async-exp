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

	"fillmore-labs.com/exp/async/result"
)

// ErrNotReady is returned when a future is not complete.
var ErrNotReady = errors.New("future not ready")

// Future represents a read-only view of the result of an asynchronous operation.
type Future[R any] struct {
	AnyFuture
	*value[R]
}

type AnyFuture interface {
	Done() <-chan struct{}
	any() result.Result[any]
}

// NewAsync runs fn asynchronously, immediately returning a [Future] that can be used to retrieve the
// eventual result. This allows separating evaluating the result from computation.
func NewAsync[R any](fn func() (R, error)) Future[R] {
	p, f := New[R]()
	go p.Do(fn)

	return f
}

// Await returns the cached result or blocks until a result is available or the context is canceled.
func (f Future[R]) Await(ctx context.Context) (R, error) {
	select { // wait for future completion or context cancel
	case <-f.done:
		return f.v.V()

	case <-ctx.Done():
		return *new(R), fmt.Errorf("future await: %w", context.Cause(ctx))
	}
}

// Try returns the cached result when ready, [ErrNotReady] otherwise.
func (f Future[R]) Try() (R, error) {
	select {
	case <-f.done:
		return f.v.V()

	default:
		return *new(R), ErrNotReady
	}
}

// Done returns a channel that is closed when the future is complete.
// It enables the use of future values in select statements.
func (f Future[R]) Done() <-chan struct{} {
	return f.done
}

// OnComplete executes fn when the [Future] is fulfilled.
func (f Future[R]) OnComplete(fn func(r result.Result[R])) {
	f.onComplete(fn)
}

func (f Future[R]) ToChannel() <-chan result.Result[R] {
	ch := make(chan result.Result[R], 1)
	fn := func(r result.Result[R]) {
		ch <- r
		close(ch)
	}

	f.onComplete(fn)

	return ch
}

func (f Future[_]) any() result.Result[any] {
	return f.v.Any()
}
