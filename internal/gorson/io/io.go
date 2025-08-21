package io

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/pbs/gorson/internal/gorson/util"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

// SSMClient interface for mocking in tests
type SSMClient interface {
	GetParametersByPath(ctx context.Context, params *ssm.GetParametersByPathInput, optFns ...func(*ssm.Options)) (*ssm.GetParametersByPathOutput, error)
	PutParameter(ctx context.Context, params *ssm.PutParameterInput, optFns ...func(*ssm.Options)) (*ssm.PutParameterOutput, error)
	DeleteParameters(ctx context.Context, params *ssm.DeleteParametersInput, optFns ...func(*ssm.Options)) (*ssm.DeleteParametersOutput, error)
}

func getSSMClient() *ssm.Client {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	client := ssm.NewFromConfig(cfg)
	return client
}

// ReadFromParameterStore gets all parameters from a given parameter store path
func ReadFromParameterStore(path util.ParameterStorePath, client SSMClient) map[string]string {
	if client == nil {
		client = getSSMClient()
	}

	p := path.String()

	var nextToken *string
	values := make(map[string]string)

	// loop until pagination done
	for {
		decr := true
		input := ssm.GetParametersByPathInput{
			Path:           &p,
			WithDecryption: &decr,
		}
		if nextToken != nil {
			input.NextToken = nextToken
		}
		output, err := client.GetParametersByPath(context.TODO(), &input)
		if err != nil {
			log.Fatal(err)
		}
		outputParams := output.Parameters
		for index := 0; index < len(outputParams); index++ {
			o := outputParams[index]
			// we remove the leading path, we want the last element of the
			// slash-delimited path as the key in our key/value pair.
			s := strings.Split(*o.Name, "/")
			k := s[len(s)-1]
			values[k] = *o.Value
		}

		// we're done paginating, break out of the loop
		if output.NextToken == nil {
			break
		}
		nextToken = output.NextToken
	}
	return values
}

// WriteResult is the result writing a single parameter - successful if Error is nil
type WriteResult struct {
	Name  string
	Error error
}

func writeSingleParameter(c chan WriteResult, client SSMClient, name string, value string, retryCount int) {
	overwrite := true
	valueType := types.ParameterTypeSecureString
	keyID := "alias/aws/ssm"
	input := ssm.PutParameterInput{
		KeyId:     &keyID,
		Name:      &name,
		Overwrite: &overwrite,
		Type:      valueType,
		Value:     &value,
	}
	_, err := client.PutParameter(context.TODO(), &input)
	if err != nil {
		var throttlingErr *types.ThrottlingException
		if errors.As(err, &throttlingErr) {
			if retryCount < 100 {
				// Introduce exponential backoff with jitter
				r := math.Pow(2, float64(retryCount)) * (1 + rand.Float64())
				time.Sleep(time.Duration(r) * time.Millisecond)
				writeSingleParameter(c, client, name, value, retryCount+1)
			} else {
				c <- WriteResult{
					Name:  name,
					Error: errors.New("throttle retry limit reached for " + name),
				}
				return
			}
		} else {
			c <- WriteResult{
				Name:  name,
				Error: err,
			}
			return
		}
	} else {
		c <- WriteResult{
			Name:  name,
			Error: nil,
		}
		return
	}
}

// WriteToParameterStore writes given parameters to a given parameter store path
func WriteToParameterStore(parameters map[string]string, path util.ParameterStorePath, timeout time.Duration, client SSMClient) error {
	if client == nil {
		client = getSSMClient()
	}

	// the jobs channel will receive messages from successful parameter store writes
	jobs := make(chan WriteResult, len(parameters))
	for key, value := range parameters {
		name := path.String() + key
		// we pass the jobs channel into the asynchronous write function to receive
		// success messages. When throttled, parameter writes wait, then retry.
		go writeSingleParameter(jobs, client, name, value, 0)
	}

	// we keep track of the parameter store writes with results
	results := make([]WriteResult, 0)
	// the done channel will receive a message once all writes are complete
	done := make(chan bool)
	// this closure collects messages from the jobs channel: once it has enough
	// (meaning all writes are successful or one has failed), it sends a message on the done channel
	go func() {
		for result := range jobs {
			results = append(results, result)
			if len(results) == len(parameters) || result.Error != nil {
				done <- true
			}
		}
	}()

	// we let two channels race: after 1 minute, the channel from time.After wins,
	// and we error out
	select {
	case <-done:
		// if any results came back with errors, return the first error
		for _, result := range results {
			if result.Error != nil {
				return result.Error
			}
		}
		return nil
	case <-time.After(timeout):
		return errors.New("timeout")
	}
}

// determineParameterDelta determines the parameters that are present in parameter store, but missing locally
func determineParameterDelta(parameters map[string]string, ssmParams map[string]string) []string {
	parameterDelta := make([]string, 0)
	for ssmParam := range ssmParams {
		if _, ok := parameters[ssmParam]; !ok {
			parameterDelta = append(parameterDelta, ssmParam)
		}
	}
	return parameterDelta
}

// promptUserDeltaWarning prompt the user with a warning based on the delta of parameters in file vs ssm, and return approval
func promptUserDeltaWarning(parameters []string, path util.ParameterStorePath) (bool, error) {
	red := color.New(color.FgRed).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	fmt.Println("The following are not present in the file, but are in parameter store:")
	for _, parameter := range parameters {
		fullParameterPath := fmt.Sprintf("%s%s\n", path.String(), parameter)
		fmt.Print(red(fullParameterPath))
	}
	fmt.Printf("Are you sure you'd like to %s these parameters?\nType %s to proceed:\n", red("delete"), green("yes"))
	reader := bufio.NewReader(os.Stdin)
	text, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}
	approval := strings.TrimSpace(text) == "yes"
	return approval, nil
}

// find returns the index and presence of a value in a slice and returns -1, false if it's not there.
func find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

// findStringInSlice is a helper function for finding strings in slices
func findStringInSlice(slice []string, val string) (int, bool) {
	return find(slice, val)
}

// deleteFromParameterStore deletes parameters at a given path from parameter store
func deleteFromParameterStore(parameters []string, path util.ParameterStorePath, client SSMClient) (deletedParams []string, err error) {
	deletedParams = []string{}

	fullPathParameters := make([]string, len(parameters))
	for idx, parameter := range parameters {
		fullPathParameter := fmt.Sprintf("%s%s", path.String(), parameter)
		fullPathParameters[idx] = fullPathParameter
	}

	var chunkSize int
	var chunkedParameters [][]string

	for {
		if len(fullPathParameters) == 0 {
			break
		}

		if len(fullPathParameters) < 10 {
			chunkSize = len(fullPathParameters)
		} else {
			chunkSize = 10
		}

		chunkedParameters = append(chunkedParameters, fullPathParameters[0:chunkSize])
		fullPathParameters = fullPathParameters[chunkSize:]
	}

	for _, params := range chunkedParameters {
		deleteParametersInput := ssm.DeleteParametersInput{
			Names: params,
		}

		output, err := client.DeleteParameters(context.TODO(), &deleteParametersInput)

		if err != nil {
			fmt.Println(err)
		}

		if len(output.DeletedParameters) != len(params) {
			fmt.Println("Some parameters failed to delete:")
			for _, parameter := range params {
				_, found := findStringInSlice(output.DeletedParameters, parameter)
				if !found {
					fmt.Println(parameter)
				}
			}
		}

		if len(output.InvalidParameters) != 0 {
			fmt.Println("Some parameters failed to delete due to being invalid:")
			for _, invalidParameter := range output.InvalidParameters {
				fmt.Println(invalidParameter)
			}
		}

		for _, deletedParam := range output.DeletedParameters {
			deletedParams = append(deletedParams, deletedParam)
		}
	}

	return deletedParams, err
}

// DeleteDeltaFromParameterStore deletes the parameters that exist in parameter store, but not in the parameters variable
func DeleteDeltaFromParameterStore(parameters map[string]string, path util.ParameterStorePath, autoApprove bool, client SSMClient) ([]string, error) {
	if client == nil {
		client = getSSMClient()
	}
	ssmParams := ReadFromParameterStore(path, client)
	parameterDelta := determineParameterDelta(parameters, ssmParams)
	if len(parameterDelta) == 0 {
		return []string{}, nil
	}
	if !autoApprove {
		approved, err := promptUserDeltaWarning(parameterDelta, path)
		if err != nil {
			return []string{}, err
		}
		if !approved {
			return []string{}, nil
		}
	}
	return deleteFromParameterStore(parameterDelta, path, client)
}

// ReadJSONFile reads a json file of key-value pairs
func ReadJSONFile(filepath string) map[string]string {
	content, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Fatal(err)
	}
	var output map[string]string
	if err := json.Unmarshal(content, &output); err != nil {
		switch err := err.(type) {
		case *json.SyntaxError:
			log.Fatal("error reading " + filepath + ": check that it's valid json")
		case *json.UnmarshalTypeError:
			log.Fatal("error reading " + filepath + ": it should contain only string key/value pairs")

		default:
			fmt.Println(reflect.TypeOf(err))
			log.Fatal(err)
		}
	}
	return output
}
