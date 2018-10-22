package dyndb

import (
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/aws"
	"log"
	"fmt"
	"strings"
	"github.com/oktasecuritylabs/sgt/logger"
	osq_types "github.com/oktasecuritylabs/sgt/osquery_types"
)


func (dyn DynDB) GetPackByName(packName string) (osq_types.Pack, error) {
	pack := osq_types.Pack{}
	//create query string from pack name
	type QS struct {
		PackName string `json:"pack_name"`
	}
	queryString := QS{}
	queryString.PackName = packName
	//packQuery := PackQuery{}
	//qp := QueryPack{}
	//map queryString to attribute_map
	js, err := dynamodbattribute.MarshalMap(queryString)
	if err != nil {
		logger.Error(err)
	}
	//get pack map from dynamo
	resp, err := dyn.DB.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("osquery_querypacks"),
		Key:       js,
	})
	if err != nil {
		//panic(fmt.Sprintln(err, os.Stdout))
		log.Panic(err)
		return pack, err
	}
	//create empty pack to marshal data into
	type QueryPack struct {
		PackName string   `json:"packName"`
		Queries  []string `json:"queries"`
	}
	querypack := QueryPack{}
	if len(resp.Item) > 0 {
		err = dynamodbattribute.UnmarshalMap(resp.Item, &querypack)
		if err != nil {
			fmt.Println(err)
			return pack, err
		}
		//here we actually build our osquery.Pack
		pack.PackName = querypack.PackName
		//pack.Queries = qp.Queries
		//itterate over list of queries and retrieve actual queries
		for _, query := range querypack.Queries {
			packquery, err := dyn.GetPackQuery(query)
			if err != nil {
				logger.Error(err)
			}
			pack.Queries = append(pack.Queries, packquery)
		}
		return pack, nil
	}
	return pack, nil

}

func (dyn DynDB) SearchQueryPacks(searchString string) ([]osq_types.QueryPack, error) {
	results := []osq_types.QueryPack{}
	scanItems, err := dyn.DB.Scan(&dynamodb.ScanInput{
		TableName: aws.String("osquery_querypacks"),
	})
	if err != nil {
		logger.Error(err)
		return results, err
	}
	for _, i := range scanItems.Items {
		querypack := osq_types.QueryPack{}
		err = dynamodbattribute.UnmarshalMap(i, &querypack)
		if err != nil {
			logger.Error(err)
			return results, err
		}
		if strings.Contains(querypack.PackName, searchString) {
			results = append(results, querypack)
		}
	}
	return results, nil

}


func (dyn DynDB) NewQueryPack(qp osq_types.QueryPack) (error) {
	av, err := dynamodbattribute.MarshalMap(qp)
	if err != nil {
		logger.Error(err)
		return err
	}

	_, err = dyn.DB.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("osquery_querypacks"),
		Item:      av,
	})
	if err != nil {
		logger.Error(err)
		return err
	}
	return nil

}

func (dyn DynDB) DeleteQueryPack(queryPackName string) (error) {
	type qs struct {
		PackName string `json:"pack_name"`
	}
	querystring := qs{queryPackName}
	av, err := dynamodbattribute.MarshalMap(querystring)
	if err != nil {
		logger.Error(err)
		return err
	}

	_, err = dyn.DB.DeleteItem(&dynamodb.DeleteItemInput{
		TableName: aws.String("osquery_querypacks"),
		Key:       av,
	})
	if err != nil {
		logger.Error(err)
		return err
	}
	return nil

}


func (dyn DynDB) UpsertPack(qp osq_types.QueryPack) (error) {
	//Additive upsert.
	existing, err := dyn.GetPackByName(qp.PackName)
	if err != nil {
		return err
	}

	if existing.PackName == "" {
		return dyn.NewQueryPack(qp)
	}

	existingQueries := map[string]bool{}
	for _, packQuery := range existing.Queries {
		existingQueries[packQuery.QueryName] = true
	}

	//note:  qp.Queries is a list of strings, not pack_queries
	for _, query := range qp.Queries {
		if !existingQueries[query] {
			//existing.Queries = append(existing.Queries, query)
			existingQueries[query] = true
		}
	}

	//existing queries should now be a map of both old and new
	logger.Debug(existingQueries)
	newQueryPack := osq_types.QueryPack{}
	newQueryPack.PackName = existing.PackName
	for query := range existingQueries {
		newQueryPack.Queries = append(newQueryPack.Queries, query)
	}

	err = dyn.DeleteQueryPack(qp.PackName)
	if err != nil {
		return err
	}

	return dyn.NewQueryPack(newQueryPack)

}
