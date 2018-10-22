package dyndb

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/aws"
	osq_types "github.com/oktasecuritylabs/sgt/osquery_types"
	"github.com/oktasecuritylabs/sgt/logger"
	"errors"
)

func (db DynDB) BuildNamedConfig(configName string) (osq_types.OsqueryNamedConfig, error) {
	storedNC := osq_types.OsqueryNamedConfig{}
	oc := osq_types.OsqueryConfig{}
	storedNC, err := db.GetNamedConfig(configName)
	if err != nil {
		return storedNC, err
	}
	storedNC.OsqueryConfig.Packs = make(map[string]map[string]map[string]map[string]string)
	//oc = storedNC.OsqueryConfig
	for _, packName := range storedNC.PackList {
		fmt.Printf("adding %s to config", packName)
		fmt.Printf("config now: %+v", oc)
		p, err := db.GetPackByName(packName)
		if err != nil {
			return storedNC, err
		}
		storedNC.OsqueryConfig.Packs[packName] = p.AsMap()
	}

	return storedNC, nil
}


// UpsertNamedConfig upserts named config to dynamo db.  Returns true if successful, else false
func (db DynDB) UpsertNamedConfig(onc *osq_types.OsqueryNamedConfig) error {

	av, err := dynamodbattribute.MarshalMap(onc)
	if err != nil {
		logger.Error("Marshal Failed")
		return err
	}

	_, err = db.DB.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("osquery_configurations"),
		Item:      av,
	})

	return err
}

func UpsertNamedConfig(dynamoDB *dynamodb.DynamoDB, onc *osq_types.OsqueryNamedConfig) error {

	av, err := dynamodbattribute.MarshalMap(onc)
	if err != nil {
		logger.Error("Marshal Failed")
		return err
	}

	_, err = dynamoDB.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("osquery_configurations"),
		Item:      av,
	})

	return err
}

// GetNamedConfigs returns all named configs
func (db DynDB) GetNamedConfigs() ([]osq_types.OsqueryNamedConfig, error) {
	results := []osq_types.OsqueryNamedConfig{}
	scanItems, err := db.DB.Scan(&dynamodb.ScanInput{
		TableName: aws.String("osquery_configurations"),
	})
	if err != nil {
		logger.Error(err)
		return results, err
	}
	for _, i := range scanItems.Items {
		config := osq_types.OsqueryNamedConfig{}
		err = dynamodbattribute.UnmarshalMap(i, &config)
		if err != nil {
			logger.Error(err)
			return results, err
		}
		results = append(results, config)
	}
	return results, nil
}

func GetNamedConfigs(dynamoDB *dynamodb.DynamoDB) ([]osq_types.OsqueryNamedConfig, error) {
	results := []osq_types.OsqueryNamedConfig{}
	scanItems, err := dynamoDB.Scan(&dynamodb.ScanInput{
		TableName: aws.String("osquery_configurations"),
	})
	if err != nil {
		logger.Error(err)
		return results, err
	}
	for _, i := range scanItems.Items {
		config := osq_types.OsqueryNamedConfig{}
		err = dynamodbattribute.UnmarshalMap(i, &config)
		if err != nil {
			logger.Error(err)
			return results, err
		}
		results = append(results, config)
	}
	return results, nil
}

func (db DynDB) GetNamedConfig(configName string) (osq_types.OsqueryNamedConfig, error) {
	namedConfig := osq_types.OsqueryNamedConfig{}
	if configName == "" {
		return namedConfig, errors.New("no config name specified")
	}
	type querystring struct {
		ConfigName string `json:"config_name"`
	}
	var qs querystring
	qs.ConfigName = configName
	//if config name is not "", return specified config(if exists)
	//fmt.Println(qs)
	js, err := dynamodbattribute.MarshalMap(qs)
	if err != nil {
		logger.Error(err)
		return namedConfig, err
	}
	//fmt.Println(js)
	resp, err := db.DB.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("osquery_configurations"),
		Key:       js,
	})
	if err != nil {
		logger.Warn("[dyndb.GetNamedConfig] ", err)
		return namedConfig, err
	}
	if len(resp.Item) > 0 {
		err = dynamodbattribute.UnmarshalMap(resp.Item, &namedConfig)
		if err != nil {
			logger.Warn("UNMARSHAL ERROR")
			return namedConfig, err
		}
		return namedConfig, nil
	}
	return namedConfig, nil

}
