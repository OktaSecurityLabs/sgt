package main

import (
	//"github.com/aws/aws-lambda-go/lambda"
	"github.com/sirupsen/logrus"
	"github.com/oktasecuritylabs/sgt/internal/pkg/carvebuilder"
	"github.com/oktasecuritylabs/sgt/dyndb"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"os"
	lambdasvc "github.com/aws/aws-sdk-go/service/lambda"
	"github.com/gin-gonic/gin/json"
	"github.com/aws/aws-lambda-go/lambda"
)

var (
	log = logrus.New()
)



func init() {
	log.Formatter = &logrus.JSONFormatter{}
}

type ev struct {
	SessionID string `json:"session_id"`
	BlockCount string `json:"block_count"`
}

func Handler() {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	}))

	lambdaClient := lambdasvc.New(sess)

	db := dyndb.DbInstance()

	carves, err := carvebuilder.GetActiveCarves(db)
	if err != nil {
		log.Error(err)
		return
	}
	if len(carves) > 0 {
		for _, carve := range carves {
			log.Infof("invoking lambda for: %s", carve.SessionID)
			js, err := json.Marshal(carve)
			if err != nil {
				log.Error(err)
			}
			_, err = lambdaClient.Invoke(&lambdasvc.InvokeInput{
				FunctionName: aws.String(os.Getenv("CARVE_BUILDER")),
				InvocationType: aws.String("Event"),
				Payload: js,
			})
			if err != nil {
				log.Error(err)
			}
		}
	} else {
		log.Infof("No pending carves, exiting")
		return
	}



	// Perform action to transform log lines into firehose.Record type
	// these transforms and actions should be defined in the pkg for the log type.
	// see opendns pkg for example

}

func main() {
	lambda.Start(Handler)
}
