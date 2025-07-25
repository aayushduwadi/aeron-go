// Copyright 2022 Steven Stern
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package aeron

import (
	"math"
	"testing"

	"github.com/lirm/aeron-go/aeron/atomic"
	"github.com/lirm/aeron-go/aeron/counters"
	"github.com/lirm/aeron-go/aeron/logbuffer"
	"github.com/lirm/aeron-go/aeron/logbuffer/term"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

const (
	Channel            = "aeron:udp?endpoint=localhost:40124"
	StreamId           = int32(1002)
	RegistrationId     = int64(10)
	ChannelStatusId    = int32(100)
	ReadBufferCapacity = 1024
	FragmentCountLimit = math.MaxInt32
)

type SubscriptionTestSuite struct {
	suite.Suite
	headerLength        int32 // Effectively a const, but DataFrameHeader declares it in a struct
	atomicReadBuffer    *atomic.Buffer
	cc                  *MockReceivingConductor
	fragmentHandlerMock *term.MockFragmentHandler
	fragmentHandler     term.FragmentHandler // References the mock's func.  Helps readability
	imageOne            *MockImage
	imageTwo            *MockImage
	header              *logbuffer.Header
	sub                 *Subscription
}

func (s *SubscriptionTestSuite) SetupTest() {
	s.headerLength = logbuffer.DataFrameHeader_Length
	s.atomicReadBuffer = atomic.NewBufferSlice(make([]byte, s.headerLength))
	s.cc = NewMockReceivingConductor(s.T())
	s.fragmentHandlerMock = term.NewMockFragmentHandler(s.T())
	s.fragmentHandler = s.fragmentHandlerMock.Execute
	s.imageOne = NewMockImage(s.T())
	s.imageTwo = NewMockImage(s.T())
	s.header = new(logbuffer.Header) // Unused so no need to initialize
	s.sub = NewSubscription(s.cc, Channel, RegistrationId, StreamId, ChannelStatusId)
}

func (s *SubscriptionTestSuite) TestShouldEnsureTheSubscriptionIsOpenWhenPolling() {
	s.cc.On("releaseSubscription", RegistrationId, mock.Anything).Return(nil)

	s.Require().NoError(s.sub.Close())
	s.Assert().True(s.sub.IsClosed())
}

func (s *SubscriptionTestSuite) TestShouldReadNothingWhenNoImages() {
	fragments := s.sub.Poll(s.fragmentHandler, 1)
	s.Assert().Equal(0, fragments)
}

func (s *SubscriptionTestSuite) TestShouldReadNothingWhenThereIsNoData() {
	s.sub.addImage(s.imageOne)
	s.imageOne.On("Poll", mock.Anything, mock.Anything).Return(0, nil)

	fragments := s.sub.Poll(s.fragmentHandler, 1)
	s.Assert().Equal(0, fragments)
}

func (s *SubscriptionTestSuite) TestShouldReadData() {
	s.sub.addImage(s.imageOne)

	// TODO: NO RETURN HERE?  remove callback below this
	s.fragmentHandlerMock.On("Execute",
		s.atomicReadBuffer, s.headerLength, ReadBufferCapacity-s.headerLength, s.header).Return(nil)

	s.imageOne.On("Poll", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		handler := args.Get(0).(term.FragmentHandler)
		handler(s.atomicReadBuffer, s.headerLength, ReadBufferCapacity-s.headerLength, s.header)
	}).Return(1, nil)

	fragments := s.sub.Poll(s.fragmentHandler, FragmentCountLimit)
	s.Assert().Equal(1, fragments)
}

func (s *SubscriptionTestSuite) TestShouldReadDataFromMultipleSources() {
	s.sub.addImage(s.imageOne)
	s.sub.addImage(s.imageTwo)

	s.imageOne.On("Poll", mock.Anything, mock.Anything).Return(1, nil)
	s.imageTwo.On("Poll", mock.Anything, mock.Anything).Return(1, nil)

	fragments := s.sub.Poll(s.fragmentHandler, FragmentCountLimit)
	s.Assert().Equal(2, fragments)
}

func (s *SubscriptionTestSuite) TestShouldCloseImageOnRemoveImage() {
	s.sub.addImage(s.imageOne)

	s.imageOne.On("CorrelationID").Return(int64(1))
	s.imageOne.On("Close").Return(nil)

	s.sub.removeImage(1)

	s.imageOne.AssertExpectations(s.T())
}

// TODO: Implement resolveChannel set of tests.

func TestSubscription(t *testing.T) {
	suite.Run(t, new(SubscriptionTestSuite))
}

// Everything below is auto generated by mockery using this command:
// mockery --name=ReceivingConductor --inpackage --structname=MockReceivingConductor --print

// MockReceivingConductor is an autogenerated mock type for the ReceivingConductor type
type MockReceivingConductor struct {
	mock.Mock
}

// AddRcvDestination provides a mock function with given fields: registrationID, endpointChannel
func (_m *MockReceivingConductor) AddRcvDestination(registrationID int64, endpointChannel string) error {
	ret := _m.Called(registrationID, endpointChannel)

	var r0 error
	if rf, ok := ret.Get(0).(func(int64, string) error); ok {
		r0 = rf(registrationID, endpointChannel)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CounterReader provides a mock function with given fields:
func (_m *MockReceivingConductor) CounterReader() *counters.Reader {
	ret := _m.Called()

	var r0 *counters.Reader
	if rf, ok := ret.Get(0).(func() *counters.Reader); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*counters.Reader)
		}
	}

	return r0
}

// RemoveRcvDestination provides a mock function with given fields: registrationID, endpointChannel
func (_m *MockReceivingConductor) RemoveRcvDestination(registrationID int64, endpointChannel string) error {
	ret := _m.Called(registrationID, endpointChannel)

	var r0 error
	if rf, ok := ret.Get(0).(func(int64, string) error); ok {
		r0 = rf(registrationID, endpointChannel)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// releaseSubscription provides a mock function with given fields: regID, images
func (_m *MockReceivingConductor) releaseSubscription(regID int64, images []Image) error {
	ret := _m.Called(regID, images)

	var r0 error
	if rf, ok := ret.Get(0).(func(int64, []Image) error); ok {
		r0 = rf(regID, images)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewMockReceivingConductor interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockReceivingConductor creates a new instance of MockReceivingConductor. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockReceivingConductor(t mockConstructorTestingTNewMockReceivingConductor) *MockReceivingConductor {
	mock := &MockReceivingConductor{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
