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

package async_test

import (
	"context"
	"testing"

	"fillmore-labs.com/exp/async"
	"github.com/stretchr/testify/assert"
)

const iterations = 3

func makePromisesAndFutures[R any]() ([]async.Promise[R], []async.Awaitable[R]) {
	var promises [iterations]async.Promise[R]
	var futures [iterations]async.Awaitable[R]

	for i := 0; i < iterations; i++ {
		futures[i], promises[i] = async.NewFuture[R]()
	}

	return promises[:], futures[:]
}

func TestWaitAll(t *testing.T) {
	t.Parallel()

	// given
	promises, futures := makePromisesAndFutures[int]()
	promises[0].Fulfill(1)
	promises[1].Reject(errTest)
	close(promises[2])

	memoizers := make([]async.Awaitable[int], 0, len(futures))
	for _, f := range futures {
		memoizers = append(memoizers, f.Memoize())
	}

	// when
	ctx := context.Background()
	results, err := async.WaitAll(ctx, memoizers...)

	// then
	if assert.NoError(t, err) {
		v0, err0 := results[0].V()
		_, err1 := results[1].V()
		_, err2 := results[2].V()

		if assert.NoError(t, err0) {
			assert.Equal(t, 1, v0)
		}
		assert.ErrorIs(t, err1, errTest)
		assert.ErrorIs(t, err2, async.ErrNoResult)
	}
}

func TestAllValues(t *testing.T) {
	t.Parallel()

	// given
	promises, futures := makePromisesAndFutures[int]()
	for i := 0; i < iterations; i++ {
		promises[i].Fulfill(i + 1)
	}

	// when
	ctx := context.Background()
	results, err := async.WaitAllValues(ctx, futures...)

	// then
	if assert.NoError(t, err) {
		for i := 0; i < iterations; i++ {
			assert.Equal(t, i+1, results[i])
		}
	}
}

func TestAllValuesError(t *testing.T) {
	t.Parallel()

	// given
	promises, futures := makePromisesAndFutures[int]()
	promises[1].Reject(errTest)

	// when
	ctx := context.Background()
	_, err := async.WaitAllValues(ctx, futures...)

	// then
	assert.ErrorIs(t, err, errTest)
}

func TestFirst(t *testing.T) {
	t.Parallel()

	// given
	promises, futures := makePromisesAndFutures[int]()
	promises[1].Fulfill(2)

	// when
	ctx := context.Background()
	result, err := async.WaitFirst(ctx, futures...)

	// then
	if assert.NoError(t, err) {
		assert.Equal(t, 2, result)
	}
}

func TestCombineCancellation(t *testing.T) {
	t.Parallel()

	subTests := []struct {
		name    string
		combine func(context.Context, ...async.Awaitable[int]) (any, error)
	}{
		{name: "First", combine: func(ctx context.Context, futures ...async.Awaitable[int]) (any, error) {
			return async.WaitFirst(ctx, futures...)
		}},
		{name: "All", combine: func(ctx context.Context, futures ...async.Awaitable[int]) (any, error) {
			return async.WaitAll(ctx, futures...)
		}},
		{name: "AllValues", combine: func(ctx context.Context, futures ...async.Awaitable[int]) (any, error) {
			return async.WaitAllValues(ctx, futures...)
		}},
	}

	for _, tc := range subTests {
		combine := tc.combine
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// given
			_, futures := makePromisesAndFutures[int]()

			// when
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			_, err := combine(ctx, futures...)

			// then
			assert.ErrorIs(t, err, context.Canceled)
		})
	}
}

func TestCombineMemoized(t *testing.T) { //nolint:funlen
	t.Parallel()

	subTests := []struct {
		name    string
		combine func(context.Context, ...async.Awaitable[int]) (any, error)
		expect  func(t *testing.T, actual any)
	}{
		{name: "First", combine: func(ctx context.Context, futures ...async.Awaitable[int]) (any, error) {
			return async.WaitFirst(ctx, futures...)
		}, expect: func(t *testing.T, actual any) { t.Helper(); assert.Equal(t, 3, actual) }},
		{name: "All", combine: func(ctx context.Context, futures ...async.Awaitable[int]) (any, error) {
			return async.WaitAll(ctx, futures...)
		}, expect: func(t *testing.T, actual any) {
			t.Helper()
			vv, ok := actual.([]async.Result[int])
			if !ok {
				assert.Fail(t, "Unexpected result type")

				return
			}

			for _, v := range vv {
				value, err := v.V()
				if assert.NoError(t, err) {
					assert.Equal(t, 3, value)
				}
			}
		}},
		{name: "AllValues", combine: func(ctx context.Context, futures ...async.Awaitable[int]) (any, error) {
			return async.WaitAllValues(ctx, futures...)
		}, expect: func(t *testing.T, actual any) { t.Helper(); assert.Equal(t, []int{3, 3, 3}, actual) }},
	}

	for _, tc := range subTests {
		combine := tc.combine
		expect := tc.expect
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// given
			promises, futures := makePromisesAndFutures[int]()

			for _, promise := range promises {
				promise.Fulfill(3)
			}

			memoizers := make([]async.Awaitable[int], 0, len(futures))
			for _, f := range futures {
				memoizer := f.Memoize()
				memoizers = append(memoizers, memoizer)
				_, _ = memoizer.TryWait()
			}

			// when
			ctx := context.Background()

			result, err := combine(ctx, memoizers...)

			// then
			if assert.NoError(t, err) {
				expect(t, result)
			}
		})
	}
}

func TestCombineAfterMemoized(t *testing.T) {
	t.Parallel()

	subTests := []struct {
		name    string
		combine func(context.Context, ...async.Awaitable[int]) (any, error)
		expect  func(t *testing.T, actual any)
	}{
		{name: "First", combine: func(ctx context.Context, futures ...async.Awaitable[int]) (any, error) {
			return async.WaitFirst(ctx, futures...)
		}},
		{name: "AllValues", combine: func(ctx context.Context, futures ...async.Awaitable[int]) (any, error) {
			return async.WaitAllValues(ctx, futures...)
		}},
	}

	for _, tc := range subTests {
		combine := tc.combine
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// given
			promises, futures := makePromisesAndFutures[int]()

			promises[1].Reject(errTest)

			memoizers := make([]async.Awaitable[int], 0, len(futures))
			for _, f := range futures {
				memoizer := f.Memoize()
				memoizers = append(memoizers, memoizer)
			}

			// when
			ctx := context.Background()
			_, err := combine(ctx, memoizers...)

			close(promises[0])

			_, err0 := memoizers[0].TryWait()
			_, err1 := memoizers[1].TryWait()
			_, err2 := memoizers[2].TryWait()

			// then
			assert.ErrorIs(t, err, errTest)
			assert.ErrorIs(t, err0, async.ErrNoResult)
			assert.ErrorIs(t, err1, errTest)
			assert.ErrorIs(t, err2, async.ErrNotReady)
		})
	}
}
