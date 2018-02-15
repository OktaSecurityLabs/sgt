package main

import (
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
)

func TestConvertDate01(t *testing.T) {
	dateString := "Mon Oct 30 03:00:41 2017 UTC"
	d, err := convertData(dateString)
	if err != nil {
		t.Error(err)
	}

	if d != "2017-10-30 03:00:41" {
		t.Error("Invalid date conversion")
	}
}

func TestConvertDate02(t *testing.T) {
	dateString := "Tue Sep 30 17:37:30 2014"
	d, err := convertData(dateString)
	if err != nil {
		t.Error(err)
	}

	if d != "2014-09-30 17:37:30" {
		t.Error("Invalid date conversion")
	}
}

func TestConvertDateInvalid(t *testing.T) {
	dateString := "2014-09-30T17:37:30"
	_, err := convertData(dateString)
	if err == nil {
		t.Error("ANSIC date conversion should fail")
	}
}

func TestHandlerValid(t *testing.T) {
	data := `{
  "action": "removed",
  "columns": {
    "name": "osqueryd",
    "path": "/usr/local/bin/osqueryd",
    "pid": "97650"
  },
  "name": "processes",
  "hostname": "hostname.local",
  "calendarTime": "Tue Sep 30 17:37:30 2014 UTC",
  "unixTime": "1412123850"
}`

	event := events.KinesisFirehoseEvent{
		InvocationID:      "01960bf1-64cb-4e41-bbc0-92fa5b14665e",
		DeliveryStreamArn: "arn:aws:firehose:us-east-1:123456789012:deliverystream/delivery-stream-name",
		Region:            "us-east-1",
		Records: []events.KinesisFirehoseEventRecord{
			events.KinesisFirehoseEventRecord{
				RecordID:                    "001",
				ApproximateArrivalTimestamp: events.MilliSecondsEpochTime{time.Now()},
				Data: []byte(data),
			},
		},
	}

	results, _ := Handler(event)

	if len(results.Records) != 1 {
		t.Errorf("Expected 1 record, got %d records", len(results.Records))
	}

	if results.Records[0].Result != events.KinesisFirehoseTransformedStateOk {
		t.Error("Expected successful record")
	}
}

func TestHandlerInvalid(t *testing.T) {
	data := `{
  "action": "removed",
  "columns": {
    "name": "osqueryd",
    "path": "/usr/local/bin/osqueryd",
    "pid": "97650"
  },
  "name": "processes",
  "hostname": "hostname.local",
  "calendarTime": "2014-09-30T17:37:30",
  "unixTime": "1412123850"
}`

	event := events.KinesisFirehoseEvent{
		InvocationID:      "01960bf1-64cb-4e41-bbc0-92fa5b14665e",
		DeliveryStreamArn: "arn:aws:firehose:us-east-1:123456789012:deliverystream/delivery-stream-name",
		Region:            "us-east-1",
		Records: []events.KinesisFirehoseEventRecord{
			events.KinesisFirehoseEventRecord{
				RecordID:                    "001",
				ApproximateArrivalTimestamp: events.MilliSecondsEpochTime{time.Now()},
				Data: []byte(data),
			},
		},
	}

	results, _ := Handler(event)

	if len(results.Records) != 1 {
		t.Errorf("Expected 1 record, got %d records", len(results.Records))
	}

	if results.Records[0].Result != events.KinesisFirehoseTransformedStateProcessingFailed {
		t.Error("Expected failed record")
	}
}
