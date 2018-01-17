package dyndb

import (
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/oktasecuritylabs/sgt/logger"
	osq_types "github.com/oktasecuritylabs/sgt/osquery_types"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"sync"
	"errors"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"encoding/json"
	"strings"
	"log"
)

func DbInstance() (*dynamodb.DynamoDB) {
	sess := session.Must(session.NewSession(
		&aws.Config{
			Region:aws.String("us-east-1"),
		}))
	creds := credentials.NewChainCredentials(
		[]credentials.Provider{
			&credentials.EnvProvider{},
			&ec2rolecreds.EC2RoleProvider{
				Client: ec2metadata.New(sess),
			},
		})
	dyn_svc := dynamodb.New(session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
		Credentials: creds,
	})))
	return dyn_svc
}

func BuildOsqueryPacksAsJson(nc osq_types.OsqueryNamedConfig) json.RawMessage {
	dyn_svc := DbInstance()
	pack_json := "{"
	var pack_list []string
	for i, pack := range nc.PackList{
		logger.Debug(pack, i)
		p, err := GetNewPackByName(pack, dyn_svc)
		if err != nil {
			logger.Error(err)
		}
		pack_list = append(pack_list, p.AsString())
	}
	pack_json += strings.Join(pack_list, ", ")
	pack_json += "}"
	return json.RawMessage(pack_json)
}


func UpsertNamedConfig (dyn_svc *dynamodb.DynamoDB, onc *osq_types.OsqueryNamedConfig, mut sync.Mutex)(bool) {
	av, err := dynamodbattribute.MarshalMap(onc)
	//fmt.Println(av)
	if err != nil{
		logger.Info("Marshal Failed")
		logger.Error(err)
	}
	_, err = dyn_svc.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("osquery_configurations"),
		Item: av,
	})
	if err != nil{
		fmt.Println(err)
		return false
	}
	return true
}

func GetNamedConfig(dyn_svc *dynamodb.DynamoDB, config_name string) (osq_types.OsqueryNamedConfig, error) {
	named_config := osq_types.OsqueryNamedConfig{}
	if config_name == ""{
		return named_config, errors.New("no config name specified")
	}
	type querystring struct {
		Config_name string `json:"config_name"`
	}
	var qs querystring
	qs.Config_name = config_name
	//if config name is not "", return specified config(if exists)
	//fmt.Println(qs)
	js, err := dynamodbattribute.MarshalMap(qs)
	if err != nil {
		logger.Error(err)
		return named_config, err
	}
	//fmt.Println(js)
	resp, err := dyn_svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("osquery_configurations"),
		Key: js,
	})
	if err != nil {
		logger.Warn("[dyndb.GetNamedConfig] ", err)
		return named_config, err
	}
	if len(resp.Item) > 0 {
		err = dynamodbattribute.UnmarshalMap(resp.Item, &named_config)
		if err != nil {
			logger.Warn("UNMARSHAL ERROR")
			return named_config, err
		}
		return named_config, nil
	}
	return named_config, nil
}


func UpsertClient(oc osq_types.OsqueryClient, d *dynamodb.DynamoDB,  mut sync.Mutex)(error) {
	logger.Warn("Upserting Client: %v", oc)
	mut.Lock()
	//fmt.Println(oc)
	av, err := dynamodbattribute.MarshalMap(oc)
	if err != nil {
		logger.Warn("Marshal failed")
		logger.Warn(err)
		return err
	}
	res, err := d.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("osquery_clients"),
		Item: av,
	})
	fmt.Println("upsert res")
	fmt.Println(res)
	if err != nil {
		logger.Error(err)
		return err
	}
	mut.Unlock()
	return nil
}

func SearchByHostIdentifier(hid string, s*dynamodb.DynamoDB)([]osq_types.OsqueryClient, error){
	Results := []osq_types.OsqueryClient{}
	type BS struct {
		Host_identifier string `json:"host_identifier"`
	}
	var bs BS
	bs.Host_identifier = hid
	scan_item := dynamodb.ScanInput{
		TableName: aws.String("osquery_clients"),
	}
	a, err := s.Scan(&scan_item)
	if err != nil {
		logger.Error(err)
		return Results, err
	}
	if len(hid) > 0 {
		for _, i := range (a.Items) {
			//fmt.Println(i)
			o := osq_types.OsqueryClient{}
			err := dynamodbattribute.UnmarshalMap(i, &o)
			if err != nil {
				logger.Error(err)
				return Results, err
			}
			if hid == string(o.Host_identifier) {
				Results = append(Results, o)
				fmt.Println(o)
			}

		}
	} else {
		for _, i := range(a.Items) {
			client := osq_types.OsqueryClient{}
			err = dynamodbattribute.UnmarshalMap(i, &client)
			if err != nil {
				logger.Error(err)
				return Results, err
			}
			Results = append(Results, client)
		}
	}
	//resp, err := s.GetItem(&item)
	if err != nil {
		logger.Error(err)
		return Results, err
	}
	return Results, nil
}


func ApprovePendingNode(nodekey string, dyn *dynamodb.DynamoDB, mut sync.Mutex)(error) {
	//approve a pending node validation.  Returns true if successfull
	osq_node, err := SearchByNodeKey(nodekey, dyn)
	logger.Warn("here's our node that we're approving: %+v", osq_node)
	if err != nil {
		logger.Error(err)
		return err
	}
	if osq_node.Pending_registration_approval {
		logger.Info("[++] Approving Node")
		logger.Info(osq_node)
		new_client := osq_types.OsqueryClient{}
		new_client.Host_identifier = osq_node.Host_identifier
		new_client.Config_name = osq_node.Config_name
		new_client.Node_key = osq_node.Node_key
		new_client.Node_invalid = false
		new_client.HostDetails = osq_node.HostDetails
		new_client.Configuration_group = osq_node.Configuration_group
		new_client.Tags = osq_node.Tags
		err := UpsertClient(new_client, dyn, mut)
		if err != nil {
			logger.Error(err)
			return err
		}
	}
	return nil
}

func ValidNode(nodekey string, dyn *dynamodb.DynamoDB)(bool, error) {
	osq_node, err := SearchByNodeKey(nodekey, dyn)
	if err != nil {
		logger.Error(err)
		return false, err
	}
	if len(osq_node.Node_key) > 0 {
		if osq_node.Pending_registration_approval == false {
			return true, nil
		}
	}
	return false, nil
}

func SearchByNodeKey(nk string, s *dynamodb.DynamoDB)(osq_types.OsqueryClient, error){

	type QS struct {
		Node_key string `json:"node_key"`
	}
	var qs QS
	qs.Node_key = nk
	osq_node := osq_types.OsqueryClient{}
	if len(nk) == 0 {
		return osq_node, errors.New("invalid node key")
	}
	js, err := dynamodbattribute.MarshalMap(qs)
	if err != nil {
		logger.Error(err)
		return osq_node, err
	}
	item := dynamodb.GetItemInput{
		TableName: aws.String("osquery_clients"),
		Key: js,
	}
	resp, err := s.GetItem(&item)
	if err != nil {
		//panic(fmt.Sprintln(err, os.Stdout))
		log.Panic(err)
		return osq_node, err
	}
	if len(resp.Item) > 0 {
		err = dynamodbattribute.UnmarshalMap(resp.Item, &osq_node)
		if err != nil {
			fmt.Println(err)
			return osq_node, err
		}
		return osq_node, nil
	}
	return osq_node, nil
}

//PackQueries
func APIGetPackQueries(dyn_svc *dynamodb.DynamoDB) ([]osq_types.PackQuery, error) {
	results := []osq_types.PackQuery{}
	scan_items, err := dyn_svc.Scan(&dynamodb.ScanInput{
		TableName: aws.String("osquery_packqueries"),
	})
	if err != nil {
		logger.Error(err)
		return results, err
	}
	for _, i := range(scan_items.Items) {
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

func APISearchPackQueries(search_string string, dyn_svc *dynamodb.DynamoDB) ([]osq_types.PackQuery, error) {
	results := []osq_types.PackQuery{}
	scan_items, err := dyn_svc.Scan(&dynamodb.ScanInput{
		TableName: aws.String("osquery_packqueries"),
	})
	if err != nil {
		logger.Error(err)
		return results, err
	}
	for _, i := range(scan_items.Items) {
		packquery := osq_types.PackQuery{}
		err = dynamodbattribute.UnmarshalMap(i, &packquery)
		if err != nil {
			logger.Error(err)
			return results, err
		}
		if strings.Contains(packquery.QueryName, search_string){
			results = append(results, packquery)
		}
	}
	return results, nil

}

func GetPackQuery(query_name string, db *dynamodb.DynamoDB)(osq_types.PackQuery, error) {
	type QS struct {
		QueryName string `json:"query_name"`
	}
	qs := QS{}
	qs.QueryName = query_name
	pack_query := osq_types.PackQuery{}
	js, err := dynamodbattribute.MarshalMap(qs)
	if err != nil {
		logger.Error(err)
	}
	item := dynamodb.GetItemInput{
		TableName: aws.String("osquery_packqueries"),
		Key: js,
	}
	resp, err := db.GetItem(&item)
	if err != nil {
		//panic(fmt.Sprintln(err, os.Stdout))
		log.Panic(err)
		return pack_query, err
	}
	if len(resp.Item) > 0 {
		err = dynamodbattribute.UnmarshalMap(resp.Item, &pack_query)
		if err != nil {
			fmt.Println(err)
			return pack_query, err
		}
		return pack_query, nil
	}
	return pack_query, nil
}
func UpsertPackQuery(pq osq_types.PackQuery, db *dynamodb.DynamoDB, mut sync.Mutex)(bool, error){
	mut.Lock()
	//fmt.Println(oc)
	av, err := dynamodbattribute.MarshalMap(pq)
	if err != nil {
		fmt.Println("marshal failed")
		fmt.Println(err)
	}
	_, err = db.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("osquery_packqueries"),
		Item: av,
	})
	if err != nil {
		fmt.Println(err)
		return false, err
	}
	mut.Unlock()
	return true, nil
}

//Packs
func GetPackByName(s string, db *dynamodb.DynamoDB) (string, error){
	type qs struct {
		Pack_name string `json:"pack_name"`
		Pack_os string `json:"pack_os"`
	}
	query := qs{}
	type pack struct {
		Pack_name string `json:"pack_name"`
		Pack_os string `json:"pack_os"`
		Queries string `json:"queries"`

	}
	p := pack{}
	query.Pack_name = s
	query.Pack_os = "Linux"
	js, err := dynamodbattribute.MarshalMap(query)
	item := dynamodb.GetItemInput{
		TableName: aws.String("osquery_packs"),
		Key: js,
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

func GetNewPackByName(pack_name string, dyn_svc *dynamodb.DynamoDB)(osq_types.Pack, error) {
	pack := osq_types.Pack{}
	//create query string from pack name
	type QS struct {
		PackName string `json:"pack_name"`
	}
	query_string := QS{}
	query_string.PackName = pack_name
	//pack_query := PackQuery{}
	//qp := QueryPack{}
	//map query_string to attribute_map
	js, err := dynamodbattribute.MarshalMap(query_string)
	if err != nil {
	}
	//get pack map from dynamo
	resp, err := dyn_svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("osquery_querypacks"),
		Key: js,
	})
	//create empty pack to marshal data into
	type QueryPack struct {
		PackName string `json:"pack_name"`
		Queries []string `json:"queries"`
	}
	querypack := QueryPack{}
	if err != nil {
		//panic(fmt.Sprintln(err, os.Stdout))
		log.Panic(err)
		return pack, err
	}
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
			packquery, err := GetPackQuery(query, dyn_svc)
			if err != nil {
				logger.Error(err)
			}
			pack.Queries = append(pack.Queries, packquery)
		}
		return pack, nil
	}
	return pack, nil
}

func SearchQueryPacks(search_string string, dyn_svc *dynamodb.DynamoDB) ([]osq_types.QueryPack, error) {
	results := []osq_types.QueryPack{}
	scan_items, err := dyn_svc.Scan(&dynamodb.ScanInput{
		TableName: aws.String("osquery_querypacks"),
	})
	if err != nil {
		logger.Error(err)
		return results, err
	}
	for _, i := range(scan_items.Items) {
		querypack := osq_types.QueryPack{}
		err = dynamodbattribute.UnmarshalMap(i, &querypack)
		if err != nil {
			logger.Error(err)
			return results, err
		}
		if strings.Contains(querypack.PackName, search_string){
			results = append(results, querypack)
		}
	}
	return results, nil
}

func NewQueryPack(qp osq_types.QueryPack, dyn_svc *dynamodb.DynamoDB, mu sync.Mutex) (error) {
	av, err := dynamodbattribute.MarshalMap(qp)
	if err != nil {
		logger.Error(err)
		return err
	}
	mu.Lock()
	_, err = dyn_svc.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("osquery_querypacks"),
		Item: av,
	})
	mu.Unlock()
	if err != nil {
		logger.Error(err)
		return err
	}
	return nil
}

func DeleteQueryPack(qp_name string, dyn_svc *dynamodb.DynamoDB, mu sync.Mutex) (error) {
	type qs struct {
		PackName string `json:"pack_name"`
	}
	querystring := qs{qp_name}
	av, err := dynamodbattribute.MarshalMap(querystring)
	if err != nil {
		logger.Error(err)
		return err
	}
	mu.Lock()
	_, err = dyn_svc.DeleteItem(&dynamodb.DeleteItemInput{
		TableName: aws.String("osquery_querypacks"),
		Key: av,
	})
	mu.Unlock()
	if err != nil {
		logger.Error(err)
		return err
	}
	return nil
}

func UpsertPack(qp osq_types.QueryPack, dyn_svc *dynamodb.DynamoDB, mu sync.Mutex) (error) {
	//Additive upsert.
	existing, err := GetNewPackByName(qp.PackName, dyn_svc)
	if err != nil {
		logger.Error(err)
		return err
	}
	switch len(existing.PackName) > 0 {
	case true : {
		existing_queries := map[string]bool{}
		//
		for _, pack_query := range existing.Queries {
			existing_queries[pack_query.QueryName] = true
		}
		if err != nil {
			logger.Error(err)
			return err
		}
		//note:  qp.Queries is a list of strings, not pack_queries
		for _, query := range qp.Queries {
			if !existing_queries[query] {
				//existing.Queries = append(existing.Queries, query)
				existing_queries[query] = true
			}
		}
		//existing queries should now be a map of both old and new
		logger.Debug(existing_queries)
		new_querypack := osq_types.QueryPack{}
		new_querypack.PackName = existing.PackName
		for query, _ := range existing_queries {
			new_querypack.Queries = append(new_querypack.Queries, query)
		}
		err = DeleteQueryPack(qp.PackName, dyn_svc, mu)
		if err != nil {
			logger.Error(err)
			return err
		}
		err = NewQueryPack(new_querypack, dyn_svc, mu)
		if err != nil {
			logger.Error(err)
			return err
		}
	}
	default: {
		err = NewQueryPack(qp, dyn_svc, mu)
		if err != nil {
			logger.Error(err)
			return err
		}
	}

	}

	return nil
}


func SearchDistributedNodeKey(nk string, dyn_svc *dynamodb.DynamoDB) (osq_types.DistributedQuery, error) {
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
		Key: marshalmap,
	}
	resp, err := dyn_svc.GetItem(&item)
	if err != nil {
		logger.Error(err)
		return dq, err
	}
	if len(resp.Item) > 0 {
		err = dynamodbattribute.UnmarshalMap(resp.Item, &dq)
		if err != nil {
			logger.Warn("Error unmarshalling distributed query")
			logger.Error(err)
			return dq, err
		}

	}
	return dq, nil
}

func NewDistributedQuery(dq osq_types.DistributedQuery, dyn_svc *dynamodb.DynamoDB, mu sync.Mutex) (error) {
	mm, err := dynamodbattribute.MarshalMap(dq)
	if err != nil {
		logger.Error(err)
		return err
	}
	mu.Lock()
	_, err = dyn_svc.PutItem(&dynamodb.PutItemInput{
	TableName: aws.String("osquery_distributed_queries"),
	Item: mm,
	})
	if err != nil {
		logger.Error(err)
		return err
	}
	mu.Unlock()
	return nil
}

func DeleteDistributedQuery(dq osq_types.DistributedQuery, dyn_svc *dynamodb.DynamoDB, mu sync.Mutex) (error) {
	type querykey struct {
		NodeKey string `json:"node_key"`
	}
	var qk querykey
	qk.NodeKey = dq.NodeKey
	key, err := dynamodbattribute.MarshalMap(qk)
	if err != nil {
		logger.Error(err)
	}
	mu.Lock()
	_ , err = dyn_svc.DeleteItem(&dynamodb.DeleteItemInput{
		TableName: aws.String("osquery_distributed_queries"),
		Key: key,
	})
	if err != nil {
		logger.Error(err)
		return err
	}
	mu.Unlock()
	return nil
}


func AppendDistributedQuery(dq osq_types.DistributedQuery, dyn_svc *dynamodb.DynamoDB, mu sync.Mutex) (error) {
	//
	//NOTE:  This could be optimized to take in teh results of the already made call to check if the key exists
	// This is probably worth doing at some point when its beyond POC
	//should only be called if a check has been run to verify that the node_key for the
	//distributed query already exists (EG call distributed.SearchNodeKey())
	existing_dq, err := SearchDistributedNodeKey(dq.NodeKey, dyn_svc)
	//create copy of queries in existing distributed query
	existing_queries := map[string]bool{}
	for _, j := range existing_dq.Queries {
		existing_queries[j] = true
	}
	//delete existing distributed query
	err = DeleteDistributedQuery(existing_dq, dyn_svc, mu)
	if err != nil {
		logger.Error(err)
		return err
	}
	//append new queries to existing queries
	for _, j := range dq.Queries {
		if !existing_queries[j] {
			existing_dq.Queries = append(existing_dq.Queries, j)
		}
	}
	if err != nil {
		logger.Error(err)
		return err
	}
	//recreate distributed query with new + old queries
	err = NewDistributedQuery(existing_dq, dyn_svc, mu)
	if err != nil {
		logger.Error(err)
		return err
	}
	return nil
}


func UpsertDistributedQuery(dq osq_types.DistributedQuery, dyn_svc *dynamodb.DynamoDB, mu sync.Mutex) (error) {
	//queries for node_key in dynamodb.  if found, appends queries to existing list
	//if not found, creates item and adds queries
	//Search for key
	//dyn_svc := dyndb.DbInstance()
	existing, err := SearchDistributedNodeKey(dq.NodeKey, dyn_svc)
	if err != nil {
		logger.Error(err)
		return err
	}
	switch len(existing.NodeKey) > 0 {
	case true: {
		err = AppendDistributedQuery(dq, dyn_svc, mu)
	}
	default:
		err = NewDistributedQuery(dq, dyn_svc, mu)
	}
	if err != nil {
		logger.Error(err)
		return err
	}
	return nil
}

func NewUser(u osq_types.User, dyn_svc *dynamodb.DynamoDB, mu sync.Mutex) (error){
	mm, err := dynamodbattribute.MarshalMap(u)
	if err != nil {
		logger.Error(err)
		return err
	}
	mu.Lock()
	_, err = dyn_svc.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("osquery_users"),
		Item: mm,
	})
	if err != nil {
		logger.Error(err)
		return err
	}
	mu.Unlock()
	return nil
}

func GetUser(username string, dyn_svc *dynamodb.DynamoDB) (osq_types.User, error) {
	u := osq_types.User{}
	type userquery struct {
		Username string `json:"username"`
	}

	user_q := userquery{username}
	marshalmap, err := dynamodbattribute.MarshalMap(user_q)
	if err != nil {
		logger.Error(err)
		return u, err
	}
	resp, err := dyn_svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("osquery_users"),
		Key: marshalmap,
	})
	if err != nil {
		logger.Info("get item failed")
		logger.Error(err)
		return u, err
	}
	err = dynamodbattribute.UnmarshalMap(resp.Item, &u)
	if err != nil {
		logger.Error(err)
		return u, err

	}
	return u, nil
}


