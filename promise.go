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

// Promise defines the common operations for resolving a [Future] to its final value.
// Implementations allow calling on of the functions from any goroutine once. Any subsequent call will panic.
type Promise[R any] struct {
	*value[R]
}

func New[R any]() (Promise[R], Future[R]) {
	r := value[R]{
		done:  make(chan struct{}),
		queue: make(chan []func(result result.Result[R]), 1),
	}
	r.queue <- nil

	return Promise[R]{value: &r}, Future[R]{value: &r}
}

// func (p Promise[R]) Future() Future[R] { return Future[R]{value: p.value} }

// Resolve resolves the promise with a value.
func (p Promise[R]) Resolve(value R) {
	p.complete(result.OfValue(value))
}

// Reject breaks the promise with an error.
func (p Promise[R]) Reject(err error) {
	p.complete(result.OfError[R](err))
}

// Do runs fn synchronously, fulfilling the [Promise] once it completes.
func (p Promise[R]) Do(fn func() (R, error)) {
	p.complete(result.Of(fn()))
}
