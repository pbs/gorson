package io

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/pbs/gorson/internal/gorson/util"
)

type mockedPutParameterReturnPair struct {
	Resp ssm.PutParameterOutput
	Err  awserr.Error
}

type mockedGetParameterReturnPair struct {
	Resp ssm.GetParametersByPathOutput
	Err  awserr.Error
}

type mockedPutParameter struct {
	ssmiface.SSMAPI
	retVals   []mockedPutParameterReturnPair
	callCount *int
}

type mockedGetParameter struct {
	ssmiface.SSMAPI
	retVal mockedGetParameterReturnPair
}

func (m mockedPutParameter) PutParameter(in *ssm.PutParameterInput) (*ssm.PutParameterOutput, error) {
	// we have to increment a pointer to an integer because we can't update struct internal state here:
	// ssmiface.SSMAPI requires PutParameter to be a receiver on a struct value, so this method gets a copy
	// of the struct, not a pointer to a mockedPutParameter that we can mutate.
	time.Sleep(time.Duration(1) * time.Microsecond) // We have to wait some amount to reliably validate timeouts
	callCount := *m.callCount
	resp := m.retVals[callCount]
	*m.callCount = callCount + 1
	return &resp.Resp, resp.Err
}

func (m mockedGetParameter) GetParametersByPath(input *ssm.GetParametersByPathInput) (*ssm.GetParametersByPathOutput, error) {
	return &m.retVal.Resp, m.retVal.Err
}

type WriteSingleParamTestCase struct {
	RetVals  []mockedPutParameterReturnPair
	Expected error
}

type WriteToParameterStoreTestCase struct {
	RetVals  []mockedPutParameterReturnPair
	Timeout  time.Duration
	Expected error
}

type ReadFromParameterStoreTestCase struct {
	RetVal   mockedGetParameterReturnPair
	Expected map[string]string
}

func TestReadFromParameterStore(t *testing.T) {
	cases := []ReadFromParameterStoreTestCase{
		{
			RetVal: mockedGetParameterReturnPair{
				Resp: ssm.GetParametersByPathOutput{
					Parameters: []*ssm.Parameter{
						&ssm.Parameter{
							Name:  aws.String("name"),
							Value: aws.String("value"),
						},
					},
				},
				Err: nil,
			},
			Expected: map[string]string{
				"name": "value",
			},
		},
	}

	path := util.NewParameterStorePath("/path/parameter")

	for i, c := range cases {
		parameters := ReadFromParameterStore(*path, mockedGetParameter{retVal: c.RetVal})
		if !reflect.DeepEqual(c.Expected, parameters) {
			t.Fatalf("%v expected %v, got %v", i, c.Expected, parameters)
		}
	}
}

func TestWriteSingleParameter(t *testing.T) {
	cases := []WriteSingleParamTestCase{
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

func TestWriteToParameterStore(t *testing.T) {
	cases := []WriteToParameterStoreTestCase{
		// Plenty of time to pretend to put parameters
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
			Timeout:  time.Duration(1) * time.Minute,
			Expected: nil,
		},
		// Not enough time to pretend to put parameters
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
			Timeout:  time.Duration(0) * time.Minute,
			Expected: errors.New("timeout"),
		},
	}

	path := util.NewParameterStorePath("/path/")
	for i, c := range cases {
		callCount := 0
		err := WriteToParameterStore(map[string]string{"path": "value"}, *path, c.Timeout, mockedPutParameter{retVals: c.RetVals, callCount: &callCount})
		if c.Expected != nil {
			if err == nil {
				t.Fatalf("%d expected %d, got %d", i, c.Expected, err)
			}
			if err.Error() != c.Expected.Error() {
				t.Fatalf("%d expected %d, got %d", i, c.Expected, err)
			}
		} else {
			if err != nil {
				t.Fatalf("%d expected %d, got %d", i, c.Expected, err)
			}
		}
	}
}
