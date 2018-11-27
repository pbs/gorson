package io

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
)

func getSSMClient(parameterStorePath *string, region *string) *ssm.SSM {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(*region),
	})
	if err != nil {
		log.Fatal(err)
	}
	client := ssm.New(sess)
	return client
}

// ReadFromParameterStore gets all parameters from a given slash-delimited parameter store path and aws region
func ReadFromParameterStore(parameterStorePath string, region string) map[string]string {
	client := getSSMClient(&parameterStorePath, &region)

	var nextToken *string
	values := make(map[string]string)

	// loop until pagination done
	for {
		decr := true
		input := ssm.GetParametersByPathInput{
			Path:           &parameterStorePath,
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
			p := outputParams[index]
			// we remove the leading path, we want the last element of the
			// slash-delimited path as the key in our key/value pair.
			s := strings.Split(*p.Name, "/")
			k := s[len(s)-1]
			values[k] = *p.Value
		}

		// we're done paginating, break out of the loop
		if output.NextToken == nil {
			break
		}
		nextToken = output.NextToken
	}
	return values
}

// WriteToParameterStore writes given parameters to a given slash-delimited parameter store path and aws region
func WriteToParameterStore(parameters map[string]string, parameterStorePath string, region string) {
	client := getSSMClient(&parameterStorePath, &region)
	overwrite := true
	valueType := "SecureString"
	keyID := "alias/aws/ssm"
	// TODO concurrency or something similarly clever
	for key, value := range parameters {
		name := parameterStorePath + key
		input := ssm.PutParameterInput{
			KeyId:     &keyID,
			Name:      &name,
			Overwrite: &overwrite,
			Type:      &valueType,
			Value:     &value,
		}
		_, err := client.PutParameter(&input)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// ReadJSONFile reads a json file of key-value pairs
func ReadJSONFile(filepath string) map[string]string {
	// TODO less cryptic error messages
	content, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Fatal(err)
	}
	var output map[string]string
	if err := json.Unmarshal(content, &output); err != nil {
		log.Fatal(err)
	}
	return output
}
