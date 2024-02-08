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
	"reflect"
	"runtime/trace"
)

func release[R any](futures []Awaitable[R], released []bool) {
	for i, done := range released {
		if !done {
			futures[i].releaseRunning()
		}
	}
}

func YieldAll[R any](ctx context.Context, yield func(int, Result[R]) bool, futures ...Awaitable[R]) error {
	numFutures := len(futures)
	selectCases := make([]reflect.SelectCase, numFutures+1)

	for i, future := range futures {
		future.addRunning()
		selectCases[i] = reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(future.channel()),
		}
	}
	selectCases[numFutures] = reflect.SelectCase{
		Dir:  reflect.SelectRecv,
		Chan: reflect.ValueOf(ctx.Done()),
	}

	released := make([]bool, numFutures)

	for i := 0; i < numFutures; i++ {
		chosen, rcv, ok := reflect.Select(selectCases)

		if chosen == numFutures { // context channel
			release(futures, released)

			return fmt.Errorf("async wait canceled: %w", ctx.Err())
		}

		selectCases[chosen].Chan = reflect.Value{}

		r, _ := rcv.Interface().(Result[R])
		v := futures[chosen].processResult(r, ok)
		released[chosen] = true

		if !yield(chosen, v) {
			release(futures, released)

			return nil
		}
	}

	return nil
}

// WaitAll returns the results of all completed futures. If the context is canceled, it returns early with an error.
func WaitAll[R any](ctx context.Context, futures ...Awaitable[R]) ([]Result[R], error) {
	defer trace.StartRegion(ctx, "asyncWaitAll").End()
	numFutures := len(futures)

	results := make([]Result[R], numFutures)
	yield := func(i int, r Result[R]) bool {
		results[i] = r

		return true
	}

	err := YieldAll(ctx, yield, futures...)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// WaitAllValues returns the values of all completed futures.
// If any future fails or the context is canceled, it returns early with an error.
func WaitAllValues[R any](ctx context.Context, futures ...Awaitable[R]) ([]R, error) {
	defer trace.StartRegion(ctx, "asyncWaitAllValues").End()
	numFutures := len(futures)

	results := make([]R, numFutures)
	var yieldErr error
	yield := func(i int, r Result[R]) bool {
		v, err := r.V()
		if err != nil {
			yieldErr = fmt.Errorf("async WaitAllValues result %d: %w", i, err)

			return false
		}
		results[i] = v

		return true
	}

	err := YieldAll(ctx, yield, futures...)
	if yieldErr != nil {
		return nil, yieldErr
	}
	if err != nil {
		return nil, err
	}

	return results, nil
}

// WaitFirst returns the result of the first completed future.
// If the context is canceled, it returns early with an error.
func WaitFirst[R any](ctx context.Context, futures ...Awaitable[R]) (R, error) {
	defer trace.StartRegion(ctx, "asyncWaitFirst").End()

	var result Result[R]
	yield := func(i int, r Result[R]) bool {
		result = r

		return false
	}

	err := YieldAll(ctx, yield, futures...)
	if err != nil {
		return *new(R), err
	}

	return result.V()
}
