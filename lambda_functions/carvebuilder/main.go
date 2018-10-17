package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sirupsen/logrus"
	"github.com/oktasecuritylabs/sgt/osquery_types"
	"github.com/oktasecuritylabs/sgt/internal/pkg/carvebuilder"
	"github.com/oktasecuritylabs/sgt/dyndb"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"os"
	"bytes"
	"io/ioutil"
)

var (
	log = logrus.New()
	carveBucket string
)



func init() {
	log.Formatter = &logrus.JSONFormatter{}
	carveBucket = os.Getenv("CARVE_BUCKET")
}

type ev struct {
	SessionID string `json:"session_id"`
	BlockCount string `json:"block_count"`
}

func Handler(event ev) {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	}))
	s3uploader := s3manager.NewUploader(sess)

	c := &osquery_types.Carve{
		SessionID: event.SessionID,
		BlockCount: event.BlockCount,
	}

	db := dyndb.DbInstance()

	ready, data, err := carvebuilder.CarveFinished(db, c)
	if err != nil {
		log.Error(err)
		return
	}
	if ready {
		fc := osquery_types.FileCarve{
			SessionID: c.SessionID,
			Chunks: data,
		}
		path := fmt.Sprintf("/tmp/%s.tar", fc.SessionID)
		err = fc.SaveToFile(path)
		if err != nil {
			log.Fatal(err)
		}

		file, err := os.Open(path)
		defer file.Close()
		if err != nil {
			log.Error(err)
		}

		body, err := ioutil.ReadAll(file)
		if err != nil {
			log.Error(err)
		}


		_, err = s3uploader.Upload(&s3manager.UploadInput{
			Bucket: aws.String(carveBucket),
			Key: aws.String(fmt.Sprintf("filecarves/%s", fc.SessionID)),
			Body: bytes.NewReader(body),
		})
		// copy to s3 here, if successfull, delete

		err = carvebuilder.DeleteCarve(db, c)
		if err != nil {
			log.Fatal(err)
		}
	}
	return

	// Perform action to transform log lines into firehose.Record type
	// these transforms and actions should be defined in the pkg for the log type.
	// see opendns pkg for example

}

func main() {
	lambda.Start(Handler)
}
