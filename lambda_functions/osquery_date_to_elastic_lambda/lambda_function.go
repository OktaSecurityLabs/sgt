package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"strings"
	"time"
)

//convertData reformats an osquery date stamp to an ElasticSearch readable Timestamp
func convertData(date string) (string, error) {

	oldDate := strings.TrimRight(date, " UTC")
	t, err := time.Parse(time.ANSIC, oldDate)
	if err != nil {
		return "", err
	}

	// Return the properly formatted date
	return t.Format("2006-01-02 15:04:05"), nil
}

//transform recieves a KinesisFirehoseEvent Record
// converts the timestamp to ElasticSearch readable
// and returns a []byte representation of the record data
func transform(rec events.KinesisFirehoseEventRecord) ([]byte, error) {
	recString := string(rec.Data)

	var data map[string]interface{}
	dec := json.NewDecoder(strings.NewReader(recString))
	err := dec.Decode(&data)
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

		responseRec := events.KinesisFirehoseResponseRecord{}
		tranformedData, err := transform(record)

		if err != nil {
			responseRec.RecordID = record.RecordID
			//responseRec.Result = events.KinesisFirehoseTransformedStateProcessingFailed
			responseRec.Result = "ProcessingFailed"
			failed++

		} else {

			responseRec.RecordID = record.RecordID
			//responseRec.Result = events.KinesisFirehoseTransformedStateOk
			responseRec.Result = "Ok"
			responseRec.Data = tranformedData
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
