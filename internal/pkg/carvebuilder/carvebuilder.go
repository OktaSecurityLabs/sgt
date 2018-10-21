package carvebuilder

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/oktasecuritylabs/sgt/osquery_types"
	"github.com/sirupsen/logrus"
	"strconv"
)

var (
	log = logrus.New()
)

func init() {
	log.Level = logrus.InfoLevel
	log.Formatter = &logrus.JSONFormatter{}
}

func GetActiveCarves(db *dynamodb.DynamoDB) ([]*osquery_types.Carve, error) {
	carves := []*osquery_types.Carve{}
	resp, err := db.Scan(&dynamodb.ScanInput{
		TableName: aws.String("filecarves"),
	})
	if err != nil {
		return nil, err
	}
	for _, i := range resp.Items {
		c := &osquery_types.Carve{}
		err := dynamodbattribute.UnmarshalMap(i, &c)
		if err != nil {
			return nil, err
		}
		carves = append(carves, c)
	}
	return carves, nil
}

func DeleteCarve(db *dynamodb.DynamoDB, carve *osquery_types.Carve) error {
	type deleteQuery struct {
		SessionID string `json:"session_id"`
	}
	q := deleteQuery{
		SessionID: carve.SessionID,
	}
	mm, err := dynamodbattribute.MarshalMap(q)
	if err != nil {
		return err
	}
	params := &dynamodb.DeleteItemInput{
		TableName: aws.String("filecarves"),
		Key:       mm,
	}
	_, err = db.DeleteItem(params)
	if err != nil {
		return err
	}
	return nil
}

func GetCarveDataBySBID(db *dynamodb.DynamoDB, sbid string) (*osquery_types.CarveData, error) {
	keyExpr := expression.KeyEqual(expression.Key("session_block_id"), expression.Value(sbid))
	expr, err := expression.NewBuilder().WithKeyCondition(keyExpr).Build()
	if err != nil {
		return nil, err
	}
	params := &dynamodb.QueryInput{
		TableName:                 aws.String("carve_data"),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}
	result, err := db.Query(params)
	if err != nil {
		return nil, err
	}
	if len(result.Items) > 0 {
		data := &osquery_types.CarveData{}
		err = dynamodbattribute.UnmarshalMap(result.Items[0], &data)
		if err != nil {
			return nil, err
		}
		return data, nil
	}
	return nil, nil
}

func CarveFinished(db *dynamodb.DynamoDB, carve *osquery_types.Carve) (bool, []*osquery_types.CarveData, error) {
	finished := false
	bc, err := strconv.Atoi(carve.BlockCount)
	if err != nil {
		return finished, nil, err
	}

	data := make([]*osquery_types.CarveData, bc)
	for i := bc - 1; i >= 0; i-- {
		sbid := fmt.Sprintf("%s-%d", carve.SessionID, i)
		cd, err := GetCarveDataBySBID(db, sbid)
		if err != nil {
			return finished, nil, err
		}
		if cd == nil {
			log.Infof("Carve: %s is not ready yet, missing: %s", carve.SessionID, sbid)
			return finished, nil, nil
		}
		data[i] = cd
	}
	finished = true
	return finished, data, nil
}
