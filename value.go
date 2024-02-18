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

import "fillmore-labs.com/exp/async/result"

// value wraps a [Result] to enable multiple queries and avoid unnecessary recomputation.
type value[R any] struct {
	_     noCopy
	done  chan struct{}                        // signals when future has completed
	v     result.Result[R]                     // valid only when done is closed
	queue chan []func(result result.Result[R]) // list of functions to execute synchronously when completed
}

func (r *value[R]) complete(value result.Result[R]) {
	r.v = value
	close(r.done)

	queue := <-r.queue
	close(r.queue)

	for _, fn := range queue {
		fn(value)
	}
}

func (r *value[R]) onComplete(fn func(value result.Result[R])) {
	if queue, ok := <-r.queue; ok {
		queue = append(queue, fn)
		r.queue <- queue
	} else {
		fn(r.v)
	}
}
