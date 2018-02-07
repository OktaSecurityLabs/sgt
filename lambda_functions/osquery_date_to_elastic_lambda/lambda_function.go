package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func convertData(date string) (string, error) {

	oldDate := strings.TrimRight(date, " UTC")
	t, err := time.Parse(time.ANSIC, oldDate)
	if err != nil {
		return "", err
	}

	// Return the properly formatted date
	return t.Format("2006-01-02T15:04:05"), nil
}

func transform(rec events.KinesisFirehoseEventRecord) ([]byte, error) {

	recordData := make([]byte, base64.StdEncoding.DecodedLen(len(rec.Data)))
	_, err := base64.StdEncoding.Decode(recordData, rec.Data)
	if err != nil {
		return rec.Data, err
	}

	var data map[string]interface{}
	dec := json.NewDecoder(strings.NewReader(string(recordData)))
	err = dec.Decode(&data)
	if err != nil {
		return rec.Data, err
	}

	calTime := data["calendarTime"].(string)
	newCalTime, err := convertData(calTime)
	if err != nil {
		return rec.Data, err
	}

	data["calendarTime"] = newCalTime

	jsonData, err := json.Marshal(data)
	if err != nil {
		return rec.Data, err
	}

	newData := make([]byte, base64.StdEncoding.EncodedLen(len(rec.Data)))
	base64.StdEncoding.Encode(newData, jsonData)

	return newData, nil
}

// Handler is the main AWS Lambda entry point
func Handler(event events.KinesisFirehoseEvent) (events.KinesisFirehoseResponse, error) {
	var results events.KinesisFirehoseResponse

	var failed int

	for _, record := range event.Records {

		responseRec := events.KinesisFirehoseResponseRecord{
			RecordID: record.RecordID,
		}

		var err error
		responseRec.Data, err = transform(record)
		if err == nil {
			responseRec.Result = events.KinesisFirehoseTransformedStateOk
		} else {
			responseRec.Result = events.KinesisFirehoseTransformedStateProcessingFailed
			failed++
		}

		// Append the record to the list of records
		results.Records = append(results.Records, responseRec)
	}

	fmt.Printf("Processing completed. Successful records %d, Failed records %d.\n",
		len(results.Records)-failed, failed)

	return results, nil
}

func main() {
	lambda.Start(Handler)
}
