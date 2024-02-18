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

	"fillmore-labs.com/exp/async/result"
)

// This iterator is used to combine the results of multiple asynchronous operations waiting in parallel.
type iterator[R any, F AnyFuture] struct {
	_          noCopy
	numFutures int
	active     []F
	cases      []reflect.SelectCase
	value      func(f F) result.Result[R]
	ctx        context.Context //nolint:containedctx
}

func newIterator[R any, F AnyFuture](
	ctx context.Context, value func(f F) result.Result[R], l []F,
) *iterator[R, F] {
	numFutures := len(l)
	active := make([]F, numFutures)
	_ = copy(active, l)

	cases := make([]reflect.SelectCase, numFutures+1)
	for idx, f := range active {
		cases[idx] = reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(f.Done()),
		}
	}
	cases[numFutures] = reflect.SelectCase{
		Dir:  reflect.SelectRecv,
		Chan: reflect.ValueOf(ctx.Done()),
	}

	return &iterator[R, F]{
		numFutures: numFutures,
		active:     active,
		cases:      cases,
		value:      value,
		ctx:        ctx,
	}
}

func (i *iterator[R, F]) yieldTo(yield func(int, result.Result[R]) bool) {
	defer trace.StartRegion(i.ctx, "asyncSeq").End()
	for run := 0; run < i.numFutures; run++ {
		chosen, _, _ := reflect.Select(i.cases)

		if chosen == i.numFutures { // context channel
			err := fmt.Errorf("list yield canceled: %w", context.Cause(i.ctx))
			i.yieldErr(yield, err)

			break
		}

		i.cases[chosen].Chan = reflect.Value{} // Disable case
		v := i.value(i.active[chosen])
		if !yield(chosen, v) {
			break
		}
	}
}

func (i *iterator[R, F]) yieldErr(yield func(int, result.Result[R]) bool, err error) {
	e := result.OfError[R](err)
	for idx := 0; idx < i.numFutures; idx++ {
		if i.cases[idx].Chan.IsValid() && !yield(idx, e) {
			break
		}
	}
}
