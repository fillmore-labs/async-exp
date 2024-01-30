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
)

// Memoizer caches results from a [Future] to enable multiple queries and avoid unnecessary recomputation.
type Memoizer[R any] interface {
	Awaitable[R]

	// TryWait returns the cached result when ready, [ErrNotReady] otherwise.
	TryWait() (R, error)
}

type memoizer[R any] struct {
	done   chan struct{} // done signals when future has completed
	future Future[R]     // future is the [Future] being cached
	value  Result[R]     // value will hold the cached result
}

// Memoize creates a new [Memoizer], consuming the [Future].
func (f Future[R]) Memoize() Memoizer[R] {
	return &memoizer[R]{done: make(chan struct{}), future: f}
}

// Result returns the cached result or blocks until a result is available or the context is canceled.
func (m *memoizer[R]) Wait(ctx context.Context) (R, error) {
	select { // wait for future completion or context cancel
	case <-ctx.Done():
		return *new(R), ctx.Err()

	case v, ok := <-m.future:
		if ok {
			m.value = v
			close(m.done)
		} else {
			<-m.done
		}
	}

	return m.value.V()
}

var ErrNotReady = errors.New("future not ready")

// TryWait returns the cached result when ready, [ErrNotReady] otherwise.
func (m *memoizer[R]) TryWait() (R, error) {
	select {
	default:
		return *new(R), ErrNotReady

	case v, ok := <-m.future:
		if ok {
			m.value = v
			close(m.done)
		} else {
			<-m.done
		}
	}

	return m.value.V()
}
