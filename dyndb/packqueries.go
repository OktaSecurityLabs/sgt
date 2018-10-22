package dyndb

import (
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"strings"
	osq_types "github.com/oktasecuritylabs/sgt/osquery_types"
	"github.com/oktasecuritylabs/sgt/logger"
	"log"
	"fmt"
)

func (db DynDB) APISearchPackQueries(searchString string) ([]osq_types.PackQuery, error) {
	results := []osq_types.PackQuery{}
	scanItems, err := db.DB.Scan(&dynamodb.ScanInput{
		TableName: aws.String("osquery_packqueries"),
	})
	if err != nil {
		logger.Error(err)
		return results, err
	}
	for _, i := range scanItems.Items {
		packquery := osq_types.PackQuery{}
		err = dynamodbattribute.UnmarshalMap(i, &packquery)
		if err != nil {
			logger.Error(err)
			return results, err
		}
		if strings.Contains(packquery.QueryName, searchString) {
			results = append(results, packquery)
		}
	}
	return results, nil
}


func (db DynDB) GetPackQuery(queryName string) (osq_types.PackQuery, error) {
	type QS struct {
		QueryName string `json:"query_name"`
	}
	qs := QS{}
	qs.QueryName = queryName
	packQuery := osq_types.PackQuery{}
	js, err := dynamodbattribute.MarshalMap(qs)
	if err != nil {
		logger.Error(err)
	}
	item := dynamodb.GetItemInput{
		TableName: aws.String("osquery_packqueries"),
		Key:       js,
	}
	resp, err := db.DB.GetItem(&item)
	if err != nil {
		//panic(fmt.Sprintln(err, os.Stdout))
		log.Panic(err)
		return packQuery, err
	}
	if len(resp.Item) > 0 {
		err = dynamodbattribute.UnmarshalMap(resp.Item, &packQuery)
		if err != nil {
			fmt.Println(err)
			return packQuery, err
		}
		return packQuery, nil
	}
	return packQuery, nil

}



func (dyn DynDB) UpsertPackQuery(pq osq_types.PackQuery) (error) {
	av, err := dynamodbattribute.MarshalMap(pq)
	if err != nil {
		logger.Warn("Marshal failed")
		logger.Error(err)
		return err
	}

	_, err = dyn.DB.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("osquery_packqueries"),
		Item:      av,
	})
	if err != nil {
		logger.Error(err)
		return err
	}

	return err

}
