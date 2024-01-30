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
	"errors"
	"testing"

	"fillmore-labs.com/exp/async"
	"github.com/stretchr/testify/suite"
)

type PromiseTestSuite struct {
	suite.Suite
	promise async.Promise[int]
	future  async.Future[int]
}

func TestResultChannelTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(PromiseTestSuite))
}

func (s *PromiseTestSuite) setPromise(promise async.Promise[int]) { s.promise = promise }

func (s *PromiseTestSuite) SetupTest() {
	s.future = async.NewFuture(s.setPromise)
}

func (s *PromiseTestSuite) TestValue() {
	// given
	s.promise.SendValue(1)

	// when
	value, err := s.future.Wait(context.Background())

	// then
	s.NoError(err)
	s.Equal(1, value)
}

var errTest = errors.New("test error")

func (s *PromiseTestSuite) TestError() {
	// given
	s.promise.SendError(errTest)

	// when
	_, err := s.future.Wait(context.Background())

	// then
	s.ErrorIs(err, errTest)
}

func (s *PromiseTestSuite) TestCancellation() {
	// given
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// when
	_, err := s.future.Wait(ctx)

	// then
	s.ErrorIs(err, context.Canceled)
}

func (s *PromiseTestSuite) TestPanic() {
	// given
	s.promise.SendValue(1)

	// when
	_, _ = s.future.Wait(context.Background())

	// then
	defer func() { _ = recover() }()
	_, _ = s.future.Wait(context.Background()) // Should panic
	s.Fail("did not panic")
}

func add1(value int) (int, error) { return value + 1, nil }

func (s *PromiseTestSuite) TestThen() {
	// given
	s.promise.SendValue(1)

	// when
	value, err := async.Then[int, int](context.Background(), s.future, add1)

	// then
	s.NoError(err)
	s.Equal(2, value)
}

func (s *PromiseTestSuite) TestThenError() {
	// given
	s.promise.SendError(errTest)

	// when
	_, err := async.Then[int, int](context.Background(), s.future, add1)

	// then
	s.ErrorIs(err, errTest)
}

func (s *PromiseTestSuite) TestThenAsync() {
	// given
	f := async.ThenAsync[int, int](context.Background(), s.future, add1)
	s.promise.SendValue(1)

	// when
	value, err := f.Wait(context.Background())

	// then
	s.NoError(err)
	s.Equal(2, value)
}
