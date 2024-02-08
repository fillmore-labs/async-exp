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
	"testing"

	"fillmore-labs.com/exp/async"
	"github.com/stretchr/testify/assert"
)

func TestErrorResult(t *testing.T) {
	// given
	f, p := async.NewFuture[int]()
	p.Reject(errTest)
	r := <-f

	// when
	v, err := r.Value(), r.Err()

	// then
	assert.ErrorIs(t, err, errTest)
	assert.Equal(t, 0, v)
}

func TestValueResult(t *testing.T) {
	// given
	f, p := async.NewFuture[int]()
	p.Fulfill(1)
	r := <-f

	// when
	v, err := r.Value(), r.Err()

	// then
	assert.NoError(t, err)
	assert.Equal(t, 1, v)
}
