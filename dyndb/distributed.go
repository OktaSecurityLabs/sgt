package dyndb

import (
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/oktasecuritylabs/sgt/logger"
	osq_types "github.com/oktasecuritylabs/sgt/osquery_types"
)

func (dyn DynDB) NewDistributedQuery(dq osq_types.DistributedQuery) (error) {
	mm, err := dynamodbattribute.MarshalMap(dq)
	if err != nil {
		logger.Error(err)
		return err
	}

	_, err = dyn.DB.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("osquery_distributed_queries"),
		Item:      mm,
	})
	if err != nil {
		logger.Error(err)
		return err
	}
	return nil

}


func (dyn DynDB) DeleteDistributedQuery(dq osq_types.DistributedQuery) (error) {
	type querykey struct {
		NodeKey string `json:"node_key"`
	}
	var qk querykey
	qk.NodeKey = dq.NodeKey
	key, err := dynamodbattribute.MarshalMap(qk)
	if err != nil {
		logger.Error(err)
	}

	_, err = dyn.DB.DeleteItem(&dynamodb.DeleteItemInput{
		TableName: aws.String("osquery_distributed_queries"),
		Key:       key,
	})
	if err != nil {
		logger.Error(err)
		return err
	}
	return nil

}


func (dyn DynDB) AppendDistributedQuery(dq osq_types.DistributedQuery) (error) {
	//
	//NOTE:  This could be optimized to take in teh results of the already made call to check if the key exists
	// This is probably worth doing at some point when its beyond POC
	//should only be called if a check has been run to verify that the node_key for the
	//distributed query already exists (EG call distributed.SearchNodeKey())
	existingDQ, err := dyn.SearchDistributedNodeKey(dq.NodeKey)
	//create copy of queries in existing distributed query
	existingQueries := map[string]bool{}
	for _, j := range existingDQ.Queries {
		existingQueries[j] = true
	}
	//delete existing distributed query
	err = dyn.DeleteDistributedQuery(existingDQ)
	if err != nil {
		logger.Error(err)
		return err
	}
	//append new queries to existing queries
	for _, j := range dq.Queries {
		if !existingQueries[j] {
			existingDQ.Queries = append(existingDQ.Queries, j)
		}
	}
	if err != nil {
		logger.Error(err)
		return err
	}
	//recreate distributed query with new + old queries
	err = dyn.NewDistributedQuery(existingDQ)
	if err != nil {
		logger.Error(err)
		return err
	}
	return nil

}


func (dyn DynDB) UpsertDistributedQuery(dq osq_types.DistributedQuery) (error) {
	//queries for node_key in dynamodb.  if found, appends queries to existing list
	//if not found, creates item and adds queries
	//Search for key
	//dynamoDB := dyndb.DbInstance()
	existing, err := dyn.SearchDistributedNodeKey(dq.NodeKey)
	if err != nil {
		return err
	}

	if existing.NodeKey != "" {
		return dyn.AppendDistributedQuery(dq)
	}

	return dyn.NewDistributedQuery(dq)

}
