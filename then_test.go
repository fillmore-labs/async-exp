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
	"testing"

	"fillmore-labs.com/exp/async"
	"github.com/stretchr/testify/suite"
)

type ThenTestSuite struct {
	suite.Suite
	future  async.Future[int]
	promise async.Promise[int]
}

func TestThenTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(ThenTestSuite))
}

func (s *ThenTestSuite) SetupSubTest() {
	s.future, s.promise = async.NewFuture[int]()
}

func (s *ThenTestSuite) fromFuture() async.Awaitable[int]   { return s.future }
func (s *ThenTestSuite) fromMemoizer() async.Awaitable[int] { return s.future.Memoize() }

type futureMemoizerTest[R any] struct {
	name      string
	awaitable func() async.Awaitable[R]
}

func (s *ThenTestSuite) createFutureMemoizerTests() []futureMemoizerTest[int] {
	return []futureMemoizerTest[int]{
		{name: "Future", awaitable: s.fromFuture},
		{name: "Memoizer", awaitable: s.fromMemoizer},
	}
}

func add1OrError(value int) (int, error) {
	if value == 2 {
		return 0, errTest
	}

	return value + 1, nil
}

func (s *ThenTestSuite) TestThen() {
	futureMemoizerTests := s.createFutureMemoizerTests()

	for _, tc := range futureMemoizerTests {
		_ = s.Run(tc.name, func() {
			// given
			awaitable := tc.awaitable()
			s.promise.Fulfill(1)

			// when
			value, err := async.Then[int, int](context.Background(), awaitable, add1OrError)

			// then
			if s.NoError(err) {
				s.Equal(2, value)
			}
		})
	}
}

func (s *ThenTestSuite) TestThenError() {
	futureMemoizerTests := s.createFutureMemoizerTests()

	for _, tc := range futureMemoizerTests {
		_ = s.Run(tc.name, func() {
			// given
			awaitable := tc.awaitable()

			s.promise.Reject(errTest)

			// when
			_, err := async.Then(context.Background(), awaitable, add1OrError)

			// then
			s.ErrorIs(err, errTest)
		})
	}
}

func (s *ThenTestSuite) TestThenAsync() {
	futureMemoizerTests := s.createFutureMemoizerTests()

	for _, tc := range futureMemoizerTests {
		_ = s.Run(tc.name, func() {
			// given
			awaitable := tc.awaitable()
			f := async.ThenAsync(context.Background(), awaitable, add1OrError)
			s.promise.Fulfill(2)

			// when
			_, err := f.Wait(context.Background())

			// then
			s.ErrorIs(err, errTest)
		})
	}
}
