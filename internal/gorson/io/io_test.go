package io

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/pbs/gorson/internal/gorson/util"
)

type mockedPutParameterReturnPair struct {
	Resp ssm.PutParameterOutput
	Err  error
}

type mockedGetParametersByPathReturnPair struct {
	Resp ssm.GetParametersByPathOutput
	Err  error
}

type mockedDeleteParametersReturnPair struct {
	Resp ssm.DeleteParametersOutput
	Err  error
}

type mockedPutParameter struct {
	retVals   []mockedPutParameterReturnPair
	callCount *int
}

type mockedGetParameter struct {
	retVal mockedGetParametersByPathReturnPair
}

type mockedDeleteDelta struct {
	deleteSuccessful          bool
	getParametersByPathRetVal mockedGetParametersByPathReturnPair
	deleteParametersRetVal    mockedDeleteParametersReturnPair
}

func (m mockedPutParameter) PutParameter(ctx context.Context, in *ssm.PutParameterInput, opts ...func(*ssm.Options)) (*ssm.PutParameterOutput, error) {
	// we have to increment a pointer to an integer because we can't update struct internal state here:
	// the interface requires PutParameter to be a receiver on a struct value, so this method gets a copy
	// of the struct, not a pointer to a mockedPutParameter that we can mutate.
	time.Sleep(time.Duration(1) * time.Microsecond) // We have to wait some amount to reliably validate timeouts
	callCount := *m.callCount
	resp := m.retVals[callCount]
	*m.callCount = callCount + 1
	return &resp.Resp, resp.Err
}

func (m mockedPutParameter) GetParametersByPath(ctx context.Context, input *ssm.GetParametersByPathInput, opts ...func(*ssm.Options)) (*ssm.GetParametersByPathOutput, error) {
	return nil, errors.New("not implemented")
}

func (m mockedPutParameter) DeleteParameters(ctx context.Context, input *ssm.DeleteParametersInput, opts ...func(*ssm.Options)) (*ssm.DeleteParametersOutput, error) {
	return nil, errors.New("not implemented")
}

func (m mockedGetParameter) GetParametersByPath(ctx context.Context, input *ssm.GetParametersByPathInput, opts ...func(*ssm.Options)) (*ssm.GetParametersByPathOutput, error) {
	return &m.retVal.Resp, m.retVal.Err
}

func (m mockedGetParameter) PutParameter(ctx context.Context, input *ssm.PutParameterInput, opts ...func(*ssm.Options)) (*ssm.PutParameterOutput, error) {
	return nil, errors.New("not implemented")
}

func (m mockedGetParameter) DeleteParameters(ctx context.Context, input *ssm.DeleteParametersInput, opts ...func(*ssm.Options)) (*ssm.DeleteParametersOutput, error) {
	return nil, errors.New("not implemented")
}

func (m mockedDeleteDelta) GetParametersByPath(ctx context.Context, input *ssm.GetParametersByPathInput, opts ...func(*ssm.Options)) (*ssm.GetParametersByPathOutput, error) {
	return &m.getParametersByPathRetVal.Resp, m.getParametersByPathRetVal.Err
}

func (m mockedDeleteDelta) PutParameter(ctx context.Context, input *ssm.PutParameterInput, opts ...func(*ssm.Options)) (*ssm.PutParameterOutput, error) {
	return nil, errors.New("not implemented")
}

func (m mockedDeleteDelta) DeleteParameters(ctx context.Context, input *ssm.DeleteParametersInput, opts ...func(*ssm.Options)) (*ssm.DeleteParametersOutput, error) {
	deletedParams := make([]string, 0)
	invalidParams := make([]string, 0)

	if m.deleteSuccessful {
		deletedParams = append(deletedParams, input.Names...)
	} else {
		invalidParams = append(invalidParams, input.Names...)
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
					Parameters: []types.Parameter{
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
		parameters := ReadFromParameterStore(*path, &mockedGetParameter{retVal: c.GetParamsRetVal})
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
						Tier:    types.ParameterTierStandard,
						Version: 1,
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
						Tier:    types.ParameterTierStandard,
						Version: 1,
					},
					Err: errors.New("Catastrophe: something terrible has happened"),
				},
			},
			Expected: errors.New("Catastrophe: something terrible has happened"),
		},
		// if AWS gives us a throttling exception, then a success, should be fine after our auto-retry
		{
			PutParameterReturnRetVals: []mockedPutParameterReturnPair{
				{
					Resp: ssm.PutParameterOutput{
						Tier:    types.ParameterTierStandard,
						Version: 1,
					},
					Err: &types.ThrottlingException{Message: aws.String("slow it down")},
				},
				{
					Resp: ssm.PutParameterOutput{
						Tier:    types.ParameterTierStandard,
						Version: 1,
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
		writeSingleParameter(outputChannel, &mockedPutParameter{retVals: c.PutParameterReturnRetVals, callCount: &callCount}, "key", "value", 0)
		result := <-outputChannel
		if c.Expected != nil {
			if result.Error == nil {
				t.Fatalf("%d expected %v, got %v", i, c.Expected, result.Error)
			}
			if result.Error.Error() != c.Expected.Error() {
				t.Fatalf("%d expected %v, got %v", i, c.Expected, result.Error)
			}
		} else {
			if result.Error != nil {
				t.Fatalf("%d expected %v, got %v", i, c.Expected, result.Error)
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
						Tier:    types.ParameterTierStandard,
						Version: 1,
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
						Tier:    types.ParameterTierStandard,
						Version: 1,
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
		err := WriteToParameterStore(map[string]string{"path": "value"}, *path, c.Timeout, &mockedPutParameter{retVals: c.PutParameterReturnRetVals, callCount: &callCount})
		if c.Expected != nil {
			if err == nil {
				t.Fatalf("%d expected %v, got %v", i, c.Expected, err)
			}
			if err.Error() != c.Expected.Error() {
				t.Fatalf("%d expected %v, got %v", i, c.Expected, err)
			}
		} else {
			if err != nil {
				t.Fatalf("%d expected %v, got %v", i, c.Expected, err)
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
					Parameters: []types.Parameter{
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
					Parameters: []types.Parameter{
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
					Parameters: []types.Parameter{
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
					Parameters: []types.Parameter{
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
			},
		},
		// Two things to delete
		{
			FileParams: map[string]string{},
			GetParamsRetVal: mockedGetParametersByPathReturnPair{
				Resp: ssm.GetParametersByPathOutput{
					Parameters: []types.Parameter{
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
		// Twenty things to delete
		{
			FileParams: map[string]string{},
			GetParamsRetVal: mockedGetParametersByPathReturnPair{
				Resp: ssm.GetParametersByPathOutput{
					Parameters: []types.Parameter{
						{
							Name:  aws.String("paramOne"),
							Value: aws.String("valueOne"),
						},
						{
							Name:  aws.String("paramTwo"),
							Value: aws.String("valueTwo"),
						},
						{
							Name:  aws.String("paramThree"),
							Value: aws.String("valueThree"),
						},
						{
							Name:  aws.String("paramFour"),
							Value: aws.String("valueFour"),
						},
						{
							Name:  aws.String("paramFive"),
							Value: aws.String("valueFive"),
						},
						{
							Name:  aws.String("paramSix"),
							Value: aws.String("valueSix"),
						},
						{
							Name:  aws.String("paramSeven"),
							Value: aws.String("valueSeven"),
						},
						{
							Name:  aws.String("paramEight"),
							Value: aws.String("valueEight"),
						},
						{
							Name:  aws.String("paramNine"),
							Value: aws.String("valueNine"),
						},
						{
							Name:  aws.String("paramTen"),
							Value: aws.String("valueTen"),
						},
						{
							Name:  aws.String("paramEleven"),
							Value: aws.String("valueEleven"),
						},
						{
							Name:  aws.String("paramTwelve"),
							Value: aws.String("valueTwelve"),
						},
						{
							Name:  aws.String("paramThirteen"),
							Value: aws.String("valueThirteen"),
						},
						{
							Name:  aws.String("paramFourteen"),
							Value: aws.String("valueFourteen"),
						},
						{
							Name:  aws.String("paramFifteen"),
							Value: aws.String("valueFifteen"),
						},
						{
							Name:  aws.String("paramSixteen"),
							Value: aws.String("valueSixteen"),
						},
						{
							Name:  aws.String("paramSeventeen"),
							Value: aws.String("valueSeventeen"),
						},
						{
							Name:  aws.String("paramEighteen"),
							Value: aws.String("valueEighteen"),
						},
						{
							Name:  aws.String("paramNineteen"),
							Value: aws.String("valueNineteen"),
						},
						{
							Name:  aws.String("paramTwenty"),
							Value: aws.String("valueTwenty"),
						},
					},
				},
				Err: nil,
			},
			DeleteSuccessful: true,
			Expected: []string{
				"/path/paramOne",
				"/path/paramTwo",
				"/path/paramThree",
				"/path/paramFour",
				"/path/paramFive",
				"/path/paramSix",
				"/path/paramSeven",
				"/path/paramEight",
				"/path/paramNine",
				"/path/paramTen",
				"/path/paramEleven",
				"/path/paramTwelve",
				"/path/paramThirteen",
				"/path/paramFourteen",
				"/path/paramFifteen",
				"/path/paramSixteen",
				"/path/paramSeventeen",
				"/path/paramEighteen",
				"/path/paramNineteen",
				"/path/paramTwenty",
			},
		},
		// Twenty-two things to delete
		{
			FileParams: map[string]string{},
			GetParamsRetVal: mockedGetParametersByPathReturnPair{
				Resp: ssm.GetParametersByPathOutput{
					Parameters: []types.Parameter{
						{
							Name:  aws.String("paramOne"),
							Value: aws.String("valueOne"),
						},
						{
							Name:  aws.String("paramTwo"),
							Value: aws.String("valueTwo"),
						},
						{
							Name:  aws.String("paramThree"),
							Value: aws.String("valueThree"),
						},
						{
							Name:  aws.String("paramFour"),
							Value: aws.String("valueFour"),
						},
						{
							Name:  aws.String("paramFive"),
							Value: aws.String("valueFive"),
						},
						{
							Name:  aws.String("paramSix"),
							Value: aws.String("valueSix"),
						},
						{
							Name:  aws.String("paramSeven"),
							Value: aws.String("valueSeven"),
						},
						{
							Name:  aws.String("paramEight"),
							Value: aws.String("valueEight"),
						},
						{
							Name:  aws.String("paramNine"),
							Value: aws.String("valueNine"),
						},
						{
							Name:  aws.String("paramTen"),
							Value: aws.String("valueTen"),
						},
						{
							Name:  aws.String("paramEleven"),
							Value: aws.String("valueEleven"),
						},
						{
							Name:  aws.String("paramTwelve"),
							Value: aws.String("valueTwelve"),
						},
						{
							Name:  aws.String("paramThirteen"),
							Value: aws.String("valueThirteen"),
						},
						{
							Name:  aws.String("paramFourteen"),
							Value: aws.String("valueFourteen"),
						},
						{
							Name:  aws.String("paramFifteen"),
							Value: aws.String("valueFifteen"),
						},
						{
							Name:  aws.String("paramSixteen"),
							Value: aws.String("valueSixteen"),
						},
						{
							Name:  aws.String("paramSeventeen"),
							Value: aws.String("valueSeventeen"),
						},
						{
							Name:  aws.String("paramEighteen"),
							Value: aws.String("valueEighteen"),
						},
						{
							Name:  aws.String("paramNineteen"),
							Value: aws.String("valueNineteen"),
						},
						{
							Name:  aws.String("paramTwenty"),
							Value: aws.String("valueTwenty"),
						},
						{
							Name:  aws.String("paramTwentyOne"),
							Value: aws.String("valueTwentyOne"),
						},
						{
							Name:  aws.String("paramTwentyTwo"),
							Value: aws.String("valueTwentyTwo"),
						},
					},
				},
				Err: nil,
			},
			DeleteSuccessful: true,
			Expected: []string{
				"/path/paramOne",
				"/path/paramTwo",
				"/path/paramThree",
				"/path/paramFour",
				"/path/paramFive",
				"/path/paramSix",
				"/path/paramSeven",
				"/path/paramEight",
				"/path/paramNine",
				"/path/paramTen",
				"/path/paramEleven",
				"/path/paramTwelve",
				"/path/paramThirteen",
				"/path/paramFourteen",
				"/path/paramFifteen",
				"/path/paramSixteen",
				"/path/paramSeventeen",
				"/path/paramEighteen",
				"/path/paramNineteen",
				"/path/paramTwenty",
				"/path/paramTwentyOne",
				"/path/paramTwentyTwo",
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
			&m,
		)

		if err != nil {
			t.Fatalf("%d unexpected error %v", i, err)
		}

		for _, param := range deletedParams {
			_, present := findStringInSlice(c.Expected, param)
			if !present {
				t.Fatalf("%d expected %s in %s\n", i, c.Expected, param)
			}
		}
	}
}
