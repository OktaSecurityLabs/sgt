package dyndb

import (
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/oktasecuritylabs/sgt/logger"
	osq_types "github.com/oktasecuritylabs/sgt/osquery_types"
)

func (dyn DynDB) NewUser(u osq_types.User) (error) {
	mm, err := dynamodbattribute.MarshalMap(u)
	if err != nil {
		logger.Error(err)
		return err
	}

	_, err = dyn.DB.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("osquery_users"),
		Item:      mm,
	})
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil

}


//NewUser creates new user in DB
func NewUser(u osq_types.User, dynamoDB *dynamodb.DynamoDB) error {
	mm, err := dynamodbattribute.MarshalMap(u)
	if err != nil {
		logger.Error(err)
		return err
	}

	_, err = dynamoDB.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("osquery_users"),
		Item:      mm,
	})
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}


func (dyn DynDB) GetUser(username string) (osq_types.User, error) {
	user := osq_types.User{}
	type userquery struct {
		Username string `json:"username"`
	}

	userQuery := userquery{username}
	marshalmap, err := dynamodbattribute.MarshalMap(userQuery)
	if err != nil {
		return user, err
	}

	resp, err := dyn.DB.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("osquery_users"),
		Key:       marshalmap,
	})
	if err != nil {
		logger.Error("get item failed")
		return user, err
	}

	err = dynamodbattribute.UnmarshalMap(resp.Item, &user)
	if err != nil {
		return user, err

	}
	return user, nil

}
