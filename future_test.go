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
	"errors"
	"sync"
	"testing"
	"time"

	"fillmore-labs.com/exp/async"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

var errTest = errors.New("test error")

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestAsyncValue(t *testing.T) {
	t.Parallel()

	// given
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// when
	f := async.NewAsync(func() (int, error) { return 1, nil })
	value, err := f.Await(ctx)

	// then
	if assert.NoError(t, err) {
		assert.Equal(t, 1, value)
	}
}

func TestAsyncError(t *testing.T) {
	t.Parallel()

	// given
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// when
	f := async.NewAsync(func() (int, error) { return 0, errTest })
	_, err := f.Await(ctx)

	// then
	assert.ErrorIs(t, err, errTest)
}

func TestCancellation(t *testing.T) {
	t.Parallel()

	// given
	_, f := async.New[int]()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// when
	_, err := f.Await(ctx)

	// then
	assert.ErrorIs(t, err, context.Canceled)
}

func TestMultiple(t *testing.T) {
	t.Parallel()

	// given
	const iterations = 1_000
	p, f := async.New[int]()

	// when
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	var values [iterations]int
	var errs [iterations]error

	var wg sync.WaitGroup
	wg.Add(iterations)
	for i := 0; i < iterations; i++ {
		go func(i int) {
			defer wg.Done()
			values[i], errs[i] = f.Await(ctx)
		}(i)
	}
	p.Resolve(1)
	wg.Wait()

	// then
	for i := 0; i < iterations; i++ {
		if assert.NoError(t, errs[i]) {
			assert.Equal(t, 1, values[i])
		}
	}
}

func TestTryWait(t *testing.T) {
	t.Parallel()

	// given
	p, f := async.New[int]()

	// when
	_, err1 := f.Try()

	p.Resolve(1)
	value2, err2 := f.Try()
	value3, err3 := f.Try()

	// then
	assert.ErrorIs(t, err1, async.ErrNotReady)
	if assert.NoError(t, err2) {
		assert.Equal(t, 1, value2)
	}
	if assert.NoError(t, err3) {
		assert.Equal(t, 1, value3)
	}
}

func TestMemoizerAllValues(t *testing.T) {
	t.Parallel()

	// given
	const iterations = 1_000
	p, f := async.New[int]()

	// when
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	futures := make([]async.Future[int], iterations)
	for i := 0; i < iterations; i++ {
		futures[i] = f
	}

	_ = time.AfterFunc(1*time.Millisecond, func() { p.Resolve(1) })
	values, err := async.AwaitAllValues(ctx, futures...)

	// then
	if assert.NoError(t, err) {
		for i := 0; i < iterations; i++ {
			assert.Equal(t, 1, values[i])
		}
	}
}

func TestToChannel(t *testing.T) {
	t.Parallel()

	// given
	p, f := async.New[int]()
	p.Resolve(1)

	// when
	ch := f.ToChannel()

	// then
	v, err := (<-ch).V()
	_, ok := <-ch
	if assert.NoError(t, err) {
		assert.Equal(t, 1, v)
	}
	assert.False(t, ok)
}
