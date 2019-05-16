package dyndb

import (
	"errors"
	"fmt"
			"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/oktasecuritylabs/sgt/logger"
	osq_types "github.com/oktasecuritylabs/sgt/osquery_types"
)

type DynDB struct {
	DB *dynamodb.DynamoDB
}

// DbInstance creates a new pointer to dynamodb from assumed role by ec2 instance
func DbInstance() *dynamodb.DynamoDB {
	sess := session.Must(session.NewSession(
		&aws.Config{
			Region: aws.String("us-east-1"),
		}))
	creds := credentials.NewChainCredentials(
		[]credentials.Provider{
			&credentials.EnvProvider{},
			&ec2rolecreds.EC2RoleProvider{
				Client: ec2metadata.New(sess),
			},
		})
	dynamoDB := dynamodb.New(session.Must(session.NewSession(&aws.Config{
		Region:      aws.String("us-east-1"),
		Credentials: creds,
	})))
	return dynamoDB
}

func NewDynamoDB() DynDB {
	dynDB := DynDB{}
	dynDB.DB = DbInstance()
	return dynDB
}


// BuildNamedConfig returns the fully built Named config, minus the credentials which are supplied during node config


func (dyn DynDB) SearchDistributedNodeKey(nk string) (osq_types.DistributedQuery, error) {
	type nodequery struct {
		NodeKey string `json:"node_key"`
	}
	querystring := nodequery{nk}
	marshalmap, err := dynamodbattribute.MarshalMap(querystring)
	dq := osq_types.DistributedQuery{}
	if err != nil {
		logger.Error(err)
		return dq, err
	}
	item := dynamodb.GetItemInput{
		TableName: aws.String("osquery_distributed_queries"),
		Key:       marshalmap,
	}
	resp, err := dyn.DB.GetItem(&item)
	if err != nil {
		logger.Error(err)
		return dq, err
	}
	if len(resp.Item) > 0 {
		err = dynamodbattribute.UnmarshalMap(resp.Item, &dq)
		if err != nil {
			logger.Error(err)
			return dq, err
		}

	}
	return dq, nil

}



// UpsertClient upsers an osqueryClient
func (db DynDB) UpsertClient(oc osq_types.OsqueryClient) (error) {
	logger.Debugf("Upserting Client: %v", oc)

	av, err := dynamodbattribute.MarshalMap(oc)
	if err != nil {
		logger.Warn("Marshal failed")
		logger.Warn(err)
		return err
	}
	_, err = db.DB.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("osquery_clients"),
		Item:      av,
	})
	if err != nil {
		logger.Error(err)
		return err
	}
	return nil

}


func (db DynDB) SearchByHostIdentifier(hid string) ([]osq_types.OsqueryClient, error) {
	Results := []osq_types.OsqueryClient{}
	type BS struct {
		HostIdentifier string `json:"host_identifier"`
	}
	var bs BS
	bs.HostIdentifier = hid
	scanParams := &dynamodb.ScanInput{
		TableName: aws.String("osquery_clients"),
	}

        pageNum := 0
        err := db.DB.ScanPages(scanParams,
            func(page *dynamodb.ScanOutput, lastPage bool) bool {
                pageNum++
                if hid != "" {
                    for _, i := range page.Items {
                        o := osq_types.OsqueryClient{}
                        err2 := dynamodbattribute.UnmarshalMap(i, &o)
                        if err2 != nil {
                              logger.Error(err2)
                        }
                        if hid == string(o.HostIdentifier) {
                              Results = append(Results, o)
                              fmt.Println(o)
                        }
                    }
                } else {
                      for _, i := range page.Items {
                          client := osq_types.OsqueryClient{}
                          err3 := dynamodbattribute.UnmarshalMap(i, &client)
                          if err3 != nil {
                              logger.Error(err3)
                          }
                          Results = append(Results, client)
                      }
                  }

                return pageNum <= 10
            })

	if err != nil {
		logger.Error(err)
		return Results, err
	}
	return Results, nil

}




func (db DynDB) ValidNode(nodeKey string) (error) {
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







func (db DynDB) APIGetPackQueries() ([]osq_types.PackQuery, error) {
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
		results = append(results, packquery)
	}
	return results, nil


}





// GetPackByName returns pack specified by name
func GetPackByName(s string, db *dynamodb.DynamoDB) (string, error) {
	type qs struct {
		PackName string `json:"packName"`
		PackOS   string `json:"pack_os"`
	}
	query := qs{}
	type pack struct {
		PackName string `json:"packName"`
		PackOS   string `json:"pack_os"`
		Queries  string `json:"queries"`
	}
	p := pack{}
	query.PackName = s
	query.PackOS = "Linux"
	js, err := dynamodbattribute.MarshalMap(query)
	item := dynamodb.GetItemInput{
		TableName: aws.String("osquery_packs"),
		Key:       js,
	}
	resp, err := db.GetItem(&item)
	if err != nil {
		logger.Info(err)
	}
	if len(resp.Item) > 0 {
		err = dynamodbattribute.UnmarshalMap(resp.Item, &p)
		if err != nil {
			logger.Info(err)
		}
		return p.Queries, nil
	}
	return p.Queries, nil
}
