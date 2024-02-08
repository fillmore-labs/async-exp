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

package async

import (
	"context"

	// experimental.
	"iter"
)

// All returns the results of all completed futures as a range function. If the context is canceled, it returns early.
func All[R any](ctx context.Context, futures ...Awaitable[R]) iter.Seq2[int, Result[R]] {
	return func(yield func(int, Result[R]) bool) {
		_ = YieldAll[R](ctx, yield, futures...)
	}
}
