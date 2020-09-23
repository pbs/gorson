package io

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"reflect"
	"strings"
	"time"

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
