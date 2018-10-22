package dyndb

import (
	"github.com/oktasecuritylabs/sgt/logger"
	osq_types "github.com/oktasecuritylabs/sgt/osquery_types"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/aws"
	"log"
	"fmt"
	"errors"
)



func (db DynDB) ApprovePendingNode(nodeKey string) (error) {
	osqNode, err := db.SearchByNodeKey(nodeKey)
	logger.Infof("here's our node that we're approving: %+v", osqNode)
	if err != nil {
		logger.Error(err)
		return err
	}
	if osqNode.PendingRegistrationApproval {
		logger.Info("[++] Approving Node")
		logger.Info(osqNode)
		newClient := osq_types.OsqueryClient{}
		newClient.HostIdentifier = osqNode.HostIdentifier
		newClient.ConfigName = osqNode.ConfigName
		newClient.NodeKey = osqNode.NodeKey
		newClient.NodeInvalid = false
		newClient.HostDetails = osqNode.HostDetails
		newClient.ConfigurationGroup = osqNode.ConfigurationGroup
		newClient.Tags = osqNode.Tags
		err := db.UpsertClient(newClient)
		if err != nil {
			logger.Error(err)
			return err
		}
	}
	return nil

}


// ValidNode returns if a node is valid or note, specified by nodeKey
func ValidNode(nodeKey string, dyn *dynamodb.DynamoDB) error {
	db := NewDynamoDB()
	osqNode, err := db.SearchByNodeKey(nodeKey)
	if err != nil {
		return err
	}

	if len(osqNode.NodeKey) == 0 {
		return errors.New("size of osquery node key is 0")
	}

	if osqNode.PendingRegistrationApproval == true {
		return errors.New("node is pending registration approval")
	}

	return nil
}


func (db DynDB) SearchByNodeKey(nk string) (osq_types.OsqueryClient, error) {
	type QS struct {
		NodeKey string `json:"node_key"`
	}
	var qs QS
	qs.NodeKey = nk
	osqNode := osq_types.OsqueryClient{}
	if len(nk) == 0 {
		return osqNode, errors.New("invalid node key")
	}
	js, err := dynamodbattribute.MarshalMap(qs)
	if err != nil {
		logger.Error(err)
		return osqNode, err
	}
	item := dynamodb.GetItemInput{
		TableName: aws.String("osquery_clients"),
		Key:       js,
	}
	resp, err := db.DB.GetItem(&item)
	if err != nil {
		//panic(fmt.Sprintln(err, os.Stdout))
		log.Panic(err)
		return osqNode, err
	}
	if len(resp.Item) > 0 {
		err = dynamodbattribute.UnmarshalMap(resp.Item, &osqNode)
		if err != nil {
			fmt.Println(err)
			return osqNode, err
		}
		return osqNode, nil
	}
	return osqNode, nil

}

func (db DynDB) DeleteNodeByNodekey(nodeKey string) (error) {

	type NK struct {
		NodeKey string `json:"node_key"`
	}

	nk := NK{
		NodeKey: nodeKey,
	}

	av, err := dynamodbattribute.MarshalMap(nk)
	if err != nil {
		return err
	}

	_, err = db.DB.DeleteItem(&dynamodb.DeleteItemInput{
		TableName: aws.String("osquery_clients"),
		Key: av,
	})
	if err != nil {
		return err
	}
	return nil
}