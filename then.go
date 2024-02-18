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

// Transform transforms the value of a successful [Future] synchronously into another, enabling i.e. unwrapping of
// values.
func Transform[R, S any](f Future[R], fn func(R, error) (S, error)) Future[S] {
	ps, fs := New[S]()

	f.OnComplete(func(r result.Result[R]) {
		ps.Do(func() (S, error) { return fn(r.V()) })
	})

	return fs
}

// AndThen executes fn asynchronously when future f completes, enabling chaining of operations.
func AndThen[R, S any](f Future[R], fn func(R, error) (S, error)) Future[S] {
	ps, fs := New[S]()

	f.OnComplete(func(r result.Result[R]) {
		go ps.Do(func() (S, error) { return fn(r.V()) })
	})

	return fs
}
