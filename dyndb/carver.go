package dyndb

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/oktasecuritylabs/sgt/osquery_types"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func init() {
	log.Level = logrus.InfoLevel
	log.Formatter = &logrus.JSONFormatter{}
}

func (dyn DynDB) CreateCarve(carveMap *osquery_types.Carve) error {
	mm, err := dynamodbattribute.MarshalMap(carveMap)
	if err != nil {
		return err
	}
	_, err = dyn.DB.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("filecarves"),
		Item:      mm,
	})
	if err != nil {
		return err
	}
	return nil
}

func (dyn DynDB) CarveDataExists(data *osquery_types.CarveData) (bool, error) {
	type query struct {
		SessionBlockID string `json:"session_block_id"`
	}
	q := query{
		SessionBlockID: data.SessionBlockID,
	}

	mm, err := dynamodbattribute.MarshalMap(q)
	if err != nil {
		return false, err
	}

	resp, err := dyn.DB.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("carve_data"),
		Key:       mm,
	})
	if err != nil {
		log.Errorf("GetItemFailed: %s", err.Error())
		return false, err
	}
	log.Infof("RESP: %+v", resp)
	if len(resp.Item) > 0 {
		return true, nil
	}
	return false, nil
}

func (dyn DynDB) AddCarveData(data *osquery_types.CarveData) error {
	exists, err := dyn.CarveDataExists(data)
	if err != nil {
		log.Error(err)
		return err
	}
	if !exists {
		mm, err := dynamodbattribute.MarshalMap(data)
		if err != nil {
			return err
		}
		_, err = dyn.DB.PutItem(&dynamodb.PutItemInput{
			TableName: aws.String("carve_data"),
			Item:      mm,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
