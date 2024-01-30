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
	"sync"
	"testing"
	"time"

	"fillmore-labs.com/exp/async"
	"github.com/stretchr/testify/assert"
)

func TestMemoizer(t *testing.T) {
	t.Parallel()

	// given
	f := async.NewAsyncFuture(func() (int, error) { return 1, nil })

	// when
	m := f.Memoize()
	value, err := m.Wait(context.Background())

	// then
	if assert.NoError(t, err) {
		assert.Equal(t, 1, value)
	}
}

func TestMemoizerError(t *testing.T) {
	t.Parallel()

	// given
	f := async.NewAsyncFuture(func() (int, error) { return 0, errTest })

	// when
	m := f.Memoize()
	_, err := m.Wait(context.Background())

	// then
	assert.ErrorIs(t, err, errTest)
}

func TestCancellation(t *testing.T) {
	t.Parallel()

	// given
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	f := async.NewFuture(func(_ async.Promise[int]) {})

	// when
	m := f.Memoize()
	_, err := m.Wait(ctx)

	// then
	assert.ErrorIs(t, err, context.Canceled)
}

func TestMultiple(t *testing.T) {
	t.Parallel()

	// given
	const iterations = 10

	start := make(chan struct{})
	f := async.NewAsyncFuture(func() (int, error) {
		<-start

		return 1, nil
	})

	// when
	m := f.Memoize()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	var values [iterations]int
	var errors [iterations]error

	var wg sync.WaitGroup
	wg.Add(iterations)
	for i := 0; i < iterations; i++ {
		go func(i int) {
			defer wg.Done()
			values[i], errors[i] = m.Wait(ctx)
		}(i)
	}
	close(start)
	wg.Wait()

	// then
	for i := 0; i < iterations; i++ {
		if assert.NoError(t, errors[i]) {
			assert.Equal(t, 1, values[i])
		}
	}
}

func TestTryWait(t *testing.T) {
	t.Parallel()

	// given
	start := make(chan struct{})
	done := make(chan struct{})
	f := async.NewAsyncFuture(func() (int, error) {
		defer close(done)
		<-start

		return 1, nil
	})

	// when
	m := f.Memoize()
	_, err1 := m.TryWait()
	close(start)
	<-done

	value2, err2 := m.TryWait()
	value3, err3 := m.TryWait()

	// then
	assert.ErrorIs(t, err1, async.ErrNotReady)
	if assert.NoError(t, err2) {
		assert.Equal(t, 1, value2)
	}
	if assert.NoError(t, err3) {
		assert.Equal(t, 1, value3)
	}
}
