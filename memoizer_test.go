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

func TestCancellation(t *testing.T) {
	t.Parallel()

	// given
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	f, _ := async.NewFuture[int]()

	// when
	m := f.Memoize()
	_, err := m.Wait(ctx)

	// then
	assert.ErrorIs(t, err, context.Canceled)
}

func TestMultiple(t *testing.T) {
	t.Parallel()

	// given
	const iterations = 1_000
	f, p := async.NewFuture[int]()

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
	p.Fulfill(1)
	wg.Wait()

	// then
	for i := 0; i < iterations; i++ {
		if assert.NoError(t, errors[i]) {
			assert.Equal(t, 1, values[i])
		}
	}
}

func TestMultipleClosed(t *testing.T) {
	t.Parallel()

	// given
	const iterations = 1_000
	f, p := async.NewFuture[int]()

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
	close(p)
	wg.Wait()

	// then
	for i := 0; i < iterations; i++ {
		assert.ErrorIs(t, errors[i], async.ErrNoResult)
	}
}

func TestTryWait(t *testing.T) {
	t.Parallel()

	// given
	f, p := async.NewFuture[int]()

	// when
	m := f.Memoize()
	_, err1 := m.TryWait()
	p.Fulfill(1)

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

func TestMemoize(t *testing.T) {
	t.Parallel()

	// given
	f, _ := async.NewFuture[int]()

	// when
	m := f.Memoize()
	mm := m.Memoize()

	// then
	assert.Same(t, m, mm)
}

func TestMemoizerAllValues(t *testing.T) {
	t.Parallel()

	// given
	const iterations = 1_000
	f, p := async.NewFuture[int]()

	// when
	m := f.Memoize()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var memoizers [iterations]async.Awaitable[int]
	for i := 0; i < iterations; i++ {
		memoizers[i] = m
	}

	_ = time.AfterFunc(1*time.Millisecond, func() { p.Fulfill(1) })
	values, err := async.WaitAllValues(ctx, memoizers[:]...)

	// then
	if assert.NoError(t, err) {
		for i := 0; i < iterations; i++ {
			assert.Equal(t, 1, values[i])
		}
	}
}
