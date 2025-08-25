package sms

import (
	"github.com/stretchr/testify/mock"
	api "github.com/twilio/twilio-go/rest/api/v2010"
)

type MockTwilioAPI struct {
	mock.Mock
}

func (m *MockTwilioAPI) CreateMessage(params *api.CreateMessageParams) (*api.ApiV2010Message, error) {
	args := m.Called(params)
	return args.Get(0).(*api.ApiV2010Message), args.Error(1)
}

type fixedSource struct {
	values []int64
	idx    int
}

func (s *fixedSource) Int63() int64 {
	v := s.values[s.idx%len(s.values)]
	s.idx++
	return v
}

func (s *fixedSource) Seed(seed int64) {}
