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

// Promise is used to send the result of an asynchronous operation.
//
// It is a write-only channel.
// Either [Promise.Fulfill] or [Promise.Reject] should be called exactly once.
type Promise[R any] chan<- Result[R]

// Do runs f synchronously, fulfilling the promise once it completes.
func (p Promise[R]) Do(f func() (R, error)) {
	if value, err := f(); err == nil {
		p.Fulfill(value)
	} else {
		p.Reject(err)
	}
}

// Fulfill fulfills the promise with a value.
func (p Promise[R]) Fulfill(value R) {
	p <- valueResult[R]{value: value}
	close(p)
}

// Reject breaks the promise with an error.
func (p Promise[R]) Reject(err error) {
	p <- errorResult[R]{err: err}
	close(p)
}
