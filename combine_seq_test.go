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

//go:build goexperiment.rangefunc

package async_test

import (
	"context"
	"testing"

	"fillmore-labs.com/exp/async"
	"github.com/stretchr/testify/assert"
)

// All returns the results of all completed futures. If the context is canceled, it returns early with an error.
func TestAll(t *testing.T) {
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
	var results [3]async.Result[int]
	ctx := context.Background()
	for i, r := range async.All(ctx, memoizers...) { //nolint:typecheck
		results[i] = r
	}

	// then
	v0, err0 := results[0].V()
	_, err1 := results[1].V()
	_, err2 := results[2].V()

	if assert.NoError(t, err0) {
		assert.Equal(t, 1, v0)
	}
	assert.ErrorIs(t, err1, errTest)
	assert.ErrorIs(t, err2, async.ErrNoResult)
}
