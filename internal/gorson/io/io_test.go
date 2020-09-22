package io

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
)

type mockedPutParameterReturnPair struct {
	Resp ssm.PutParameterOutput
	Err  awserr.Error
}

type mockedPutParameter struct {
	ssmiface.SSMAPI
	retVals   []mockedPutParameterReturnPair
	callCount *int
}

func (m mockedPutParameter) PutParameter(in *ssm.PutParameterInput) (*ssm.PutParameterOutput, error) {
	// we have to increment a pointer to an integer because we can't update struct internal state here:
	// ssmiface.SSMAPI requires PutParameter to be a receiver on a struct value, so this method gets a copy
	// of the struct, not a pointer to a mockedPutParameter that we can mutate.
	callCount := *m.callCount
	resp := m.retVals[callCount]
	*m.callCount = callCount + 1
	return &resp.Resp, resp.Err
}

type TestCase struct {
	RetVals  []mockedPutParameterReturnPair
	Expected error
}

func TestWriteSingleParameter(t *testing.T) {
	cases := []TestCase{
		// happy case: no throttling, no error. smooth sailing!
		{
			RetVals: []mockedPutParameterReturnPair{
				{
					Resp: ssm.PutParameterOutput{
						Tier:    aws.String("mock"),
						Version: aws.Int64(1),
					},
					Err: nil,
				},
			},
			Expected: nil,
		},
		// if AWS gives us an error we don't know how to handle, pass it right on through for the caller to deal with
		{
			RetVals: []mockedPutParameterReturnPair{
				{
					Resp: ssm.PutParameterOutput{
						Tier:    aws.String("mock"),
						Version: aws.Int64(1),
					},
					Err: awserr.New("Catastrophe", "something terrible has happened", errors.New("from the depths I climb")),
				},
			},
			Expected: awserr.New("Catastrophe", "something terrible has happened", errors.New("from the depths I climb")),
		},
		// if AWS gives us a throttling exception, then a success, should be fine after our auto-retry
		{
			RetVals: []mockedPutParameterReturnPair{
				{
					Resp: ssm.PutParameterOutput{
						Tier:    aws.String("mock"),
						Version: aws.Int64(1),
					},
					Err: awserr.New("ThrottlingException", "slow it down", errors.New("you got throttled")),
				},
				{
					Resp: ssm.PutParameterOutput{
						Tier:    aws.String("mock"),
						Version: aws.Int64(1),
					},
					Err: nil,
				},
			},
			Expected: nil,
		},
	}

	for i, c := range cases {
		outputChannel := make(chan WriteResult, 1)
		callCount := 0
		writeSingleParameter(outputChannel, mockedPutParameter{retVals: c.RetVals, callCount: &callCount}, "key", "value", 0)
		result := <-outputChannel
		if c.Expected != nil {
			if result.Error == nil {
				t.Fatalf("%d expected %d, got %d", i, c.Expected, result.Error)
			}
			if result.Error.Error() != c.Expected.Error() {
				t.Fatalf("%d expected %d, got %d", i, c.Expected, result.Error)
			}
		} else {
			if result.Error != nil {
				t.Fatalf("%d expected %d, got %d", i, c.Expected, result.Error)
			}
		}
	}
}
