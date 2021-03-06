// Copyright 2020 - See NOTICE file for copyright holders.
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

package errors_test

import (
	stderrors "errors"
	"testing"
	"time"

	pkgerrors "github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/pkg/context/test"
	"perun.network/go-perun/pkg/errors"
)

func TestGatherer_Failed(t *testing.T) {
	g := errors.NewGatherer()

	select {
	case <-g.Failed():
		t.Fatal("Failed must not be closed")
	default:
	}

	g.Add(stderrors.New(""))

	select {
	case <-g.Failed():
	default:
		t.Fatal("Failed must be closed")
	}
}

func TestGatherer_Go_and_Wait(t *testing.T) {
	g := errors.NewGatherer()

	const timeout = 100 * time.Millisecond

	g.Go(func() error {
		time.Sleep(timeout)
		return stderrors.New("")
	})

	test.AssertNotTerminates(t, timeout/2, func() { g.Wait() })
	var err error
	test.AssertTerminates(t, timeout, func() { err = g.Wait() })
	require.Error(t, err)
}

func TestGatherer_Add_and_Err(t *testing.T) {
	g := errors.NewGatherer()

	require.NoError(t, g.Err())

	g.Add(stderrors.New("1"))
	g.Add(stderrors.New("2"))
	require.Error(t, g.Err())
	require.Len(t, errors.Causes(g.Err()), 2)
}

func TestCauses(t *testing.T) {
	g := errors.NewGatherer()
	require.Len(t, errors.Causes(g.Err()), 0)

	g.Add(stderrors.New("1"))
	require.Len(t, errors.Causes(g.Err()), 1)

	g.Add(stderrors.New("2"))
	require.Len(t, errors.Causes(g.Err()), 2)

	g.Add(stderrors.New("3"))
	require.Len(t, errors.Causes(g.Err()), 3)

	err := stderrors.New("normal")
	causes := errors.Causes(err)
	require.Len(t, causes, 1)
	assert.Same(t, causes[0], err)
}

func TestAccumulatedError_Error(t *testing.T) {
	g := errors.NewGatherer()
	g.Add(stderrors.New("1"))
	require.Equal(t, g.Err().Error(), "(1 error)\n1): 1")

	g.Add(stderrors.New("2"))
	require.Equal(t, g.Err().Error(), "(2 errors)\n1): 1\n2): 2")
}

type stackTracer interface {
	StackTrace() pkgerrors.StackTrace
}

func TestAccumulatedError_StackTrace(t *testing.T) {
	g := errors.NewGatherer()

	g.Add(stderrors.New("1"))
	assert.Nil(t, g.Err().(stackTracer).StackTrace())
	g.Add(pkgerrors.New("2"))
	assert.NotNil(t, g.Err().(stackTracer).StackTrace())
}

func TestGatherer_OnFail(t *testing.T) {
	var (
		assert  = assert.New(t)
		g       = errors.NewGatherer()
		called  bool
		called2 bool
	)

	g.OnFail(func() {
		select {
		case <-g.Failed():
		case <-time.After(time.Second):
			assert.Fail("Failed not closed before OnFail hooks are executed")
		}

		called = true
		assert.False(called2)
	})

	g.OnFail(func() {
		assert.True(called)
		called2 = true
	})

	g.Add(nil)
	assert.False(called)
	assert.False(called2)

	g.Add(stderrors.New("error"))
	assert.True(called)
	assert.True(called2)
}
