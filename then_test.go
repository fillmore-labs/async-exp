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
	"strconv"
	"testing"

	"fillmore-labs.com/exp/async"
	"github.com/stretchr/testify/assert"
)

func itoa(i int, err error) (string, error) {
	if err != nil {
		return "", err
	}

	if i < 0 {
		return "", errTest
	}

	return strconv.Itoa(i), nil
}

func TestTransform1(t *testing.T) {
	t.Parallel()

	// given
	p, f := async.New[int]()
	p.Resolve(42)

	// when
	f1 := async.Transform(f, itoa)

	// then
	v, err := f1.Try()
	if assert.NoError(t, err) {
		assert.Equal(t, "42", v)
	}
}

func TestTransform2(t *testing.T) {
	t.Parallel()

	// given
	p, f := async.New[int]()

	// when
	f1 := async.Transform(f, itoa)
	p.Resolve(42)

	// then
	v, err := f1.Try()
	if assert.NoError(t, err) {
		assert.Equal(t, "42", v)
	}
}

func TestTransformError1(t *testing.T) {
	t.Parallel()

	// given
	p, f := async.New[int]()

	// when
	f1 := async.Transform(f, itoa)
	p.Reject(errTest)

	// then
	_, err := f1.Try()
	assert.ErrorIs(t, err, errTest)
}

func TestTransformError2(t *testing.T) {
	t.Parallel()

	// given
	p, f := async.New[int]()

	// when
	f1 := async.Transform(f, itoa)
	p.Resolve(-1)

	// then
	_, err := f1.Try()
	assert.ErrorIs(t, err, errTest)
}

func TestAndThen(t *testing.T) {
	t.Parallel()

	// given
	p, f := async.New[int]()

	// when
	f1 := async.AndThen(f, itoa)
	p.Resolve(42)

	// then
	v, err := f1.Await(context.Background())
	if assert.NoError(t, err) {
		assert.Equal(t, "42", v)
	}
}
