package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/oktasecuritylabs/sgt/logger"
)

// convertData reformats an osquery date/timestamp to an ElasticSearch readable date/timestamp
func convertData(date string) (string, error) {

	oldDate := strings.TrimRight(date, " UTC")
	t, err := time.Parse(time.ANSIC, oldDate)
	if err != nil {
		return "", err
	}

	// Return the properly formatted date
	return t.Format("2006-01-02 15:04:05"), nil
}

// transform recieves a KinesisFirehoseEvent Record
// converts the timestamp to ElasticSearch readable
// and returns the new json representation of the record data
func transform(rec events.KinesisFirehoseEventRecord) ([]byte, error) {

	var data map[string]interface{}
	err := json.Unmarshal(rec.Data, &data)
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

	return jsonData, nil
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
		if err != nil {
			logger.Error("could not transform record:", err)
			responseRec.Result = events.KinesisFirehoseTransformedStateProcessingFailed
			failed++
		} else {
			responseRec.Result = events.KinesisFirehoseTransformedStateOk
		}

		results.Records = append(results.Records, responseRec)
	}

	fmt.Printf("Processing completed. Successful records %d, Failed records %d.\n",
		len(results.Records)-failed, failed)

	return results, nil
}

func main() {
	lambda.Start(Handler)
}
