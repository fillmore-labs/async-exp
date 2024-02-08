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

type AsyncTestSuite struct {
	suite.Suite
	promise async.Promise[int]
	future  async.Future[int]
}

func TestAsyncTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(AsyncTestSuite))
}

func (s *AsyncTestSuite) SetupTest() {
	s.future, s.promise = async.NewFuture[int]()
}

func (s *AsyncTestSuite) TestValue() {
	// given
	s.promise.Do(func() (int, error) { return 1, nil })

	// when
	value, err := s.future.Wait(context.Background())

	// then
	if s.NoError(err) {
		s.Equal(1, value)
	}
}

var errTest = errors.New("test error")

func (s *AsyncTestSuite) TestError() {
	// given
	ctx := context.Background()
	s.promise.Do(func() (int, error) { return 0, errTest })

	// when
	_, err := s.future.Wait(ctx)

	// then
	s.ErrorIs(err, errTest)
}

func (s *AsyncTestSuite) TestCancellation() {
	// given
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// when
	_, err := s.future.Wait(ctx)

	// then
	s.ErrorIs(err, context.Canceled)
}

func (s *AsyncTestSuite) TestNoResult() {
	// given
	ctx := context.Background()
	s.promise.Fulfill(1)
	_, _ = s.future.Wait(ctx)

	// when
	_, err := s.future.Wait(ctx)

	// then
	s.ErrorIs(err, async.ErrNoResult)
}

func (s *AsyncTestSuite) TestTryWait() {
	// given

	// when
	_, err1 := s.future.TryWait()

	s.promise.Fulfill(1)

	v2, err2 := s.future.TryWait()
	_, err3 := s.future.TryWait()

	// then
	s.ErrorIs(err1, async.ErrNotReady)
	if s.NoError(err2) {
		s.Equal(1, v2)
	}
	s.ErrorIs(err3, async.ErrNoResult)
}
