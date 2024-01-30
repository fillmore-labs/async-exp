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

import "context"

// Result defines the interface for returning results from asynchronous operations.
// It encapsulates the final value or error from the operation.
type Result[R any] interface {
	V() (R, error) // The V method returns the final value or an error.
}

// valueResult is an implementation of [Result] that simply holds a value.
type valueResult[R any] struct {
	value R
}

// V returns the stored value.
func (v valueResult[R]) V() (R, error) {
	return v.value, nil
}

// errorResult handles errors from failed operations.
type errorResult[R any] struct {
	err error
}

// V returns the stored error.
func (e errorResult[R]) V() (R, error) {
	return *new(R), e.err
}

// Promise is used to send the result of an asynchronous operation.
//
// It is a write-only channel.
// Either [Promise.SendValue] or [Promise.SendError] should be called exactly once.
type Promise[R any] chan<- Result[R]

// Future represents an asynchronous operation that will complete sometime in the future.
//
// It is a read-only channel that can be used to retrieve the final result of a [Promise] with [Future.Wait].
type Future[R any] <-chan Result[R]

// NewFuture provides a simple way to create a Future for synchronous operations.
// This allows synchronous and asynchronous code to be composed seamlessly and separating initiation from waiting.
//
// - f takes a func that accepts a Promise as a [Promise]
//
// The returned [Future] that can be used to retrieve the eventual result of the [Promise].
func NewFuture[R any](f func(promise Promise[R])) Future[R] {
	ch := make(chan Result[R], 1)
	f(ch)

	return ch
}

// NewAsyncFuture runs f asynchronously, immediately returning a [Future] that can be used to retrieve the eventual
// result. This allows separating evaluating the result from computation.
func NewAsyncFuture[R any](f func() (R, error)) Future[R] {
	return NewFuture(func(p Promise[R]) { go p.Send(f) })
}

// Send runs f synchronously, fulfilling the promise once it completes.
func (p Promise[R]) Send(f func() (R, error)) {
	if value, err := f(); err == nil {
		p.SendValue(value)
	} else {
		p.SendError(err)
	}
}

// SendValue fulfills the promise with a value once the operation completes.
func (p Promise[R]) SendValue(value R) {
	p <- valueResult[R]{value: value}
	close(p)
}

// SendError breaks the promise with an error.
func (p Promise[R]) SendError(err error) {
	p <- errorResult[R]{err: err}
	close(p)
}

// Wait returns the final result of the associated [Promise].
// It can only be called once and blocks until a result is received or the context is canceled.
// If you need to read multiple times from a [Future] wrap it with [Future.Memoize].
func (f Future[R]) Wait(ctx context.Context) (R, error) {
	select {
	case r, ok := <-f:
		if !ok {
			panic("expired future")
		}

		return r.V()

	case <-ctx.Done():
		return *new(R), ctx.Err()
	}
}

// Awaitable is the underlying interface for [Future] and [Memoizer].
// It blocks until a result is received or the context is canceled.
// Plain futures can only be queried once, while memoizers can be queried multiple times.
type Awaitable[R any] interface {
	Wait(ctx context.Context) (R, error)
}

// Then transforms the embedded result from an [Awaitable] using 'then'.
// This allows to easily handle errors embedded in the response.
// It blocks until a result is received or the context is canceled.
func Then[R, S any](ctx context.Context, f Awaitable[R], then func(R) (S, error)) (S, error) {
	reply, err := f.Wait(ctx)
	if err != nil {
		return *new(S), err
	}

	return then(reply)
}

// ThenAsync asynchronously transforms the embedded result from an [Awaitable] using 'then'.
func ThenAsync[R, S any](ctx context.Context, f Awaitable[R], then func(R) (S, error)) Future[S] {
	return NewAsyncFuture[S](func() (S, error) { return Then(ctx, f, then) })
}
