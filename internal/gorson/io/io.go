package io

import (
	"bufio"
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

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
)

func getSSMClient() *ssm.SSM {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	client := ssm.New(sess)
	return client
}

// ReadFromParameterStore gets all parameters from a given parameter store path
func ReadFromParameterStore(path util.ParameterStorePath, client ssmiface.SSMAPI) map[string]string {
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
		output, err := client.GetParametersByPath(&input)
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

// WriteResult is the result writing a single parameter - successful is Error is nil
type WriteResult struct {
	Name  string
	Error error
}

func writeSingleParameter(c chan WriteResult, client ssmiface.SSMAPI, name string, value string, retryCount int) {
	overwrite := true
	valueType := "SecureString"
	keyID := "alias/aws/ssm"
	input := ssm.PutParameterInput{
		KeyId:     &keyID,
		Name:      &name,
		Overwrite: &overwrite,
		Type:      &valueType,
		Value:     &value,
	}
	_, err := client.PutParameter(&input)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() != "ThrottlingException" {
				c <- WriteResult{
					Name:  name,
					Error: awsErr,
				}
				return
			}
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
		}
	} else {
		c <- WriteResult{
			Name:  name,
			Error: err,
		}
		return
	}
}

// WriteToParameterStore writes given parameters to a given parameter store path
func WriteToParameterStore(parameters map[string]string, path util.ParameterStorePath, timeout time.Duration, client ssmiface.SSMAPI) error {
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
		fmt.Printf(red(fullParameterPath))
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
func find(slice []*string, val *string) (int, bool) {
	for i, item := range slice {
		if *item == *val {
			return i, true
		}
	}
	return -1, false
}

// deleteFromParameterStore deletes parameters at a given path from parameter store
func deleteFromParameterStore(parameters []string, path util.ParameterStorePath, client ssmiface.SSMAPI) ([]string, error) {
	fullPathParameters := make([]*string, len(parameters))
	for idx, parameter := range parameters {
		fullPathParameter := fmt.Sprintf("%s%s", path.String(), parameter)
		fullPathParameters[idx] = &fullPathParameter
	}

	deleteParametersInput := ssm.DeleteParametersInput{
		Names: fullPathParameters,
	}

	output, err := client.DeleteParameters(&deleteParametersInput)

	if len(output.DeletedParameters) != len(parameters) {
		fmt.Println("Some parameters failed to delete:")
		for _, parameter := range fullPathParameters {
			idx, found := find(output.DeletedParameters, parameter)
			if !found {
				fmt.Println(*output.DeletedParameters[idx])
			}
		}
	}
	if len(output.InvalidParameters) != 0 {
		fmt.Println("Some parameters failed to delete due to being invalid:")
		for _, invalidParameter := range output.InvalidParameters {
			fmt.Println(*invalidParameter)
		}
	}

	deletedParams := []string{}
	for _, deletedParam := range output.DeletedParameters {
		deletedParams = append(deletedParams, *deletedParam)
	}

	return deletedParams, err
}

// DeleteDeltaFromParameterStore deletes the parameters that exist in parameter store, but not in the parameters variable
func DeleteDeltaFromParameterStore(parameters map[string]string, path util.ParameterStorePath, autoApprove bool, client ssmiface.SSMAPI) ([]string, error) {
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
