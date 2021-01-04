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

type mockedGetParametersByPathReturnPair struct {
	Resp ssm.GetParametersByPathOutput
	Err  awserr.Error
}

type mockedDeleteParametersReturnPair struct {
	Resp ssm.DeleteParametersOutput
	Err  awserr.Error
}

type mockedPutParameter struct {
	ssmiface.SSMAPI
	retVals   []mockedPutParameterReturnPair
	callCount *int
}

type mockedGetParameter struct {
	ssmiface.SSMAPI
	retVal mockedGetParametersByPathReturnPair
}

type mockedDeleteDelta struct {
	ssmiface.SSMAPI
	deleteSuccessful          bool
	getParametersByPathRetVal mockedGetParametersByPathReturnPair
	deleteParametersRetVal    mockedDeleteParametersReturnPair
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

func (m mockedDeleteDelta) GetParametersByPath(input *ssm.GetParametersByPathInput) (*ssm.GetParametersByPathOutput, error) {
	return &m.getParametersByPathRetVal.Resp, m.getParametersByPathRetVal.Err
}

func (m mockedDeleteDelta) DeleteParamters(input *ssm.DeleteParametersInput) (*ssm.DeleteParametersOutput, error) {
	deletedParams := make([]*string, 0)
	invalidParams := make([]*string, 0)

	if m.deleteSuccessful {
		for _, name := range input.Names {
			deletedParams = append(deletedParams, name)
		}
	} else {
		for _, name := range input.Names {
			invalidParams = append(invalidParams, name)
		}
	}
	deleteParametersResponse := ssm.DeleteParametersOutput{
		DeletedParameters: deletedParams,
		InvalidParameters: invalidParams,
	}
	return &deleteParametersResponse, nil
}

type WriteSingleParamTestCase struct {
	PutParameterReturnRetVals []mockedPutParameterReturnPair
	Expected                  error
}

type WriteToParameterStoreTestCase struct {
	PutParameterReturnRetVals []mockedPutParameterReturnPair
	Timeout                   time.Duration
	Expected                  error
}

type ReadFromParameterStoreTestCase struct {
	GetParamsRetVal mockedGetParametersByPathReturnPair
	Expected        map[string]string
}

type DeleteDeltaFromParameterStoreTestCase struct {
	FileParams         map[string]string
	GetParamsRetVal    mockedGetParametersByPathReturnPair
	DeleteParamsRetVal mockedDeleteParametersReturnPair
	DeleteSuccessful   bool
	Expected           []string
}

func TestReadFromParameterStore(t *testing.T) {
	cases := []ReadFromParameterStoreTestCase{
		{
			GetParamsRetVal: mockedGetParametersByPathReturnPair{
				Resp: ssm.GetParametersByPathOutput{
					Parameters: []*ssm.Parameter{
						{
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
		parameters := ReadFromParameterStore(*path, mockedGetParameter{retVal: c.GetParamsRetVal})
		if !reflect.DeepEqual(c.Expected, parameters) {
			t.Fatalf("%v expected %v, got %v", i, c.Expected, parameters)
		}
	}
}

func TestWriteSingleParameter(t *testing.T) {
	cases := []WriteSingleParamTestCase{
		// happy case: no throttling, no error. smooth sailing!
		{
			PutParameterReturnRetVals: []mockedPutParameterReturnPair{
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
			PutParameterReturnRetVals: []mockedPutParameterReturnPair{
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
			PutParameterReturnRetVals: []mockedPutParameterReturnPair{
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
		writeSingleParameter(outputChannel, mockedPutParameter{retVals: c.PutParameterReturnRetVals, callCount: &callCount}, "key", "value", 0)
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
			PutParameterReturnRetVals: []mockedPutParameterReturnPair{
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
			PutParameterReturnRetVals: []mockedPutParameterReturnPair{
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
		err := WriteToParameterStore(map[string]string{"path": "value"}, *path, c.Timeout, mockedPutParameter{retVals: c.PutParameterReturnRetVals, callCount: &callCount})
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

func TestDeleteDeltaFromParameterStore(t *testing.T) {
	cases := []DeleteDeltaFromParameterStoreTestCase{
		// Nothing to delete
		{
			FileParams: map[string]string{
				"paramOne": "valueOne",
				"paramTwo": "valueTwo",
			},
			GetParamsRetVal: mockedGetParametersByPathReturnPair{
				Resp: ssm.GetParametersByPathOutput{
					Parameters: []*ssm.Parameter{
						{
							Name:  aws.String("paramOne"),
							Value: aws.String("valueOne"),
						},
						{
							Name:  aws.String("paramTwo"),
							Value: aws.String("valueTwo"),
						},
					},
				},
				Err: nil,
			},
			DeleteSuccessful: true,
			Expected:         []string{},
		},
		// Still nothing to delete. Theoretically impossible, as param three should have been pushed
		{
			FileParams: map[string]string{
				"paramOne":   "valueOne",
				"paramTwo":   "valueTwo",
				"paramThree": "valueThree",
			},
			GetParamsRetVal: mockedGetParametersByPathReturnPair{
				Resp: ssm.GetParametersByPathOutput{
					Parameters: []*ssm.Parameter{
						{
							Name:  aws.String("paramOne"),
							Value: aws.String("valueOne"),
						},
						{
							Name:  aws.String("paramTwo"),
							Value: aws.String("valueTwo"),
						},
					},
				},
				Err: nil,
			},
			DeleteSuccessful: true,
			Expected:         []string{},
		},
		// One thing to delete
		{
			FileParams: map[string]string{
				"paramOne": "valueOne",
			},
			GetParamsRetVal: mockedGetParametersByPathReturnPair{
				Resp: ssm.GetParametersByPathOutput{
					Parameters: []*ssm.Parameter{
						{
							Name:  aws.String("paramOne"),
							Value: aws.String("valueOne"),
						},
						{
							Name:  aws.String("paramTwo"),
							Value: aws.String("valueTwo"),
						},
					},
				},
				Err: nil,
			},
			DeleteSuccessful: true,
			Expected: []string{
				"/path/paramTwo",
			},
		},
		// Other thing to delete
		{
			FileParams: map[string]string{
				"paramTwo": "valueTwo",
			},
			GetParamsRetVal: mockedGetParametersByPathReturnPair{
				Resp: ssm.GetParametersByPathOutput{
					Parameters: []*ssm.Parameter{
						{
							Name:  aws.String("paramOne"),
							Value: aws.String("valueOne"),
						},
						{
							Name:  aws.String("paramTwo"),
							Value: aws.String("valueTwo"),
						},
					},
				},
				Err: nil,
			},
			DeleteSuccessful: true,
			Expected: []string{
				"/path/paramTwo",
			},
		},
		// Two things to delete
		{
			FileParams: map[string]string{},
			GetParamsRetVal: mockedGetParametersByPathReturnPair{
				Resp: ssm.GetParametersByPathOutput{
					Parameters: []*ssm.Parameter{
						{
							Name:  aws.String("paramOne"),
							Value: aws.String("valueOne"),
						},
						{
							Name:  aws.String("paramTwo"),
							Value: aws.String("valueTwo"),
						},
					},
				},
				Err: nil,
			},
			DeleteSuccessful: true,
			Expected: []string{
				"/path/paramOne",
				"/path/paramTwo",
			},
		},
	}

	path := util.NewParameterStorePath("/path/")
	for i, c := range cases {
		m := mockedDeleteDelta{
			deleteSuccessful:          c.DeleteSuccessful,
			getParametersByPathRetVal: c.GetParamsRetVal,
		}

		deletedParams, err := DeleteDeltaFromParameterStore(
			c.FileParams,
			*path,
			true,
			m,
		)

		if err != nil {
			t.Fatalf("%d unexpected error %v", i, err)
		}

		for _, param := range deletedParams {
			_, present := findStringInSice(c.Expected, param)
			if !present {
				t.Fatalf("%d expected %s in %s\n", i, c.Expected, param)
			}
		}
	}
}

func findStringInSice(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}
