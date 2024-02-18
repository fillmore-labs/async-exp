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
	"fillmore-labs.com/exp/async/result"
	"github.com/stretchr/testify/assert"
)

const iterations = 3

func makePromisesAndFutures[R any]() ([]async.Promise[R], []async.Future[R]) {
	var promises [iterations]async.Promise[R]
	var futures [iterations]async.Future[R]

	for i := 0; i < iterations; i++ {
		promises[i], futures[i] = async.New[R]()
	}

	return promises[:], futures[:]
}

func TestWaitAll(t *testing.T) {
	t.Parallel()

	// given
	promises, futures := makePromisesAndFutures[int]()

	promises[0].Resolve(1)
	promises[1].Reject(errTest)
	promises[2].Resolve(2)

	// when
	ctx := context.Background()
	results := async.AwaitAllResults(ctx, futures...)

	// then
	assert.Len(t, results, len(futures))
	v0, err0 := results[0].V()
	err1 := results[1].Err()
	v2, err2 := results[2].V()

	if assert.NoError(t, err0) {
		assert.Equal(t, 1, v0)
	}
	assert.ErrorIs(t, err1, errTest)
	if assert.NoError(t, err2) {
		assert.Equal(t, 2, v2)
	}
}

func TestAllValues(t *testing.T) {
	t.Parallel()

	// given
	promises, futures := makePromisesAndFutures[int]()
	for i := 0; i < iterations; i++ {
		promises[i].Resolve(i + 1)
	}

	// when
	ctx := context.Background()
	results, err := async.AwaitAllValues(ctx, futures...)

	// then
	if assert.NoError(t, err) {
		assert.Len(t, results, iterations)
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
	_, err := async.AwaitAllValues(ctx, futures...)

	// then
	assert.ErrorIs(t, err, errTest)
}

func TestFirst(t *testing.T) {
	t.Parallel()

	// given
	promises, futures := makePromisesAndFutures[int]()
	promises[1].Resolve(2)

	// when
	ctx := context.Background()
	v, err := async.AwaitFirst(ctx, futures...)

	// then
	if assert.NoError(t, err) {
		assert.Equal(t, 2, v)
	}
}

func TestCombineCancellation(t *testing.T) {
	t.Parallel()

	subTests := []struct {
		name    string
		combine func([]async.Future[int], context.Context) error
	}{
		{name: "First", combine: func(futures []async.Future[int], ctx context.Context) error {
			_, err := async.AwaitFirst(ctx, futures...)

			return err
		}},
		{name: "All", combine: func(futures []async.Future[int], ctx context.Context) error {
			r := async.AwaitAllResults(ctx, futures...)

			return r[0].Err()
		}},
		{name: "AllValues", combine: func(futures []async.Future[int], ctx context.Context) error {
			_, err := async.AwaitAllValues(ctx, futures...)

			return err
		}},
	}

	for _, tc := range subTests {
		combine := tc.combine
		test := func(t *testing.T) {
			t.Helper()
			t.Parallel()

			// given
			_, futures := makePromisesAndFutures[int]()

			// when
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			err := combine(futures, ctx)

			// then
			assert.ErrorIs(t, err, context.Canceled)
		}

		_ = t.Run(tc.name, test)
	}
}

func TestCombineMemoized(t *testing.T) {
	t.Parallel()

	subTests := []struct {
		name    string
		combine func(context.Context, []async.Future[int]) (any, error)
		expect  func(t *testing.T, actual any)
	}{
		{name: "First", combine: func(ctx context.Context, futures []async.Future[int]) (any, error) {
			return async.AwaitFirst(ctx, futures...)
		}, expect: func(t *testing.T, actual any) { t.Helper(); assert.Equal(t, 3, actual) }},
		{name: "All", combine: func(ctx context.Context, futures []async.Future[int]) (any, error) {
			return async.AwaitAllResults(ctx, futures...), nil
		}, expect: func(t *testing.T, actual any) {
			t.Helper()
			vv, ok := actual.([]result.Result[int])
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
		{name: "AllValues", combine: func(ctx context.Context, futures []async.Future[int]) (any, error) {
			return async.AwaitAllValues(ctx, futures...)
		}, expect: func(t *testing.T, actual any) { t.Helper(); assert.Equal(t, []int{3, 3, 3}, actual) }},
	}

	for _, tc := range subTests {
		combine := tc.combine
		expect := tc.expect
		_ = t.Run(tc.name, func(t *testing.T) {
			t.Helper()
			t.Parallel()

			// given
			promises, futures := makePromisesAndFutures[int]()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			for _, p := range promises {
				p.Resolve(3)
			}

			// when
			v, err := combine(ctx, futures)

			// then
			if assert.NoError(t, err) {
				expect(t, v)
			}
		})
	}
}

func TestAwaitAllEmpty(t *testing.T) {
	t.Parallel()

	// given
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// when
	results := async.AwaitAllResultsAny(ctx)

	// then
	assert.Empty(t, results)
}

func TestAwaitAllValuesEmpty(t *testing.T) {
	t.Parallel()

	// given
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// when
	results, err := async.AwaitAllValuesAny(ctx)

	// then
	if assert.NoError(t, err) {
		assert.Empty(t, results)
	}
}

func TestAwaitFirstEmpty(t *testing.T) {
	t.Parallel()

	// given
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// when
	_, err := async.AwaitFirstAny(ctx)

	// then
	assert.ErrorIs(t, err, async.ErrNoResult)
}

func TestAllAny(t *testing.T) {
	// given
	t.Parallel()
	ctx := context.Background()

	p1, f1 := async.New[int]()
	p2, f2 := async.New[string]()
	p3, f3 := async.New[struct{}]()

	p1.Resolve(1)
	p2.Resolve("test")
	p3.Resolve(struct{}{})

	// when
	results := make([]result.Result[any], 3)
	async.AwaitAllAny(ctx, f1, f2, f3)(func(i int, r result.Result[any]) bool {
		results[i] = r

		return true
	})

	// then
	for i, r := range results {
		if assert.NoError(t, r.Err()) {
			switch i {
			case 0:
				assert.Equal(t, 1, r.Value())
			case 1:
				assert.Equal(t, "test", r.Value())
			case 2:
				assert.Equal(t, struct{}{}, r.Value())
			default:
				assert.Fail(t, "unexpected index")
			}
		}
	}
}
