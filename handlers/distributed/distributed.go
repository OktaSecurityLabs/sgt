package distributed

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/oktasecuritylabs/sgt/dyndb"
	"github.com/oktasecuritylabs/sgt/handlers/response"
	"github.com/oktasecuritylabs/sgt/logger"
	"github.com/oktasecuritylabs/sgt/osquery_types"
)

/*
func init() {
	//logger.SetFormatter(&logger.JSONFormatter{/I//})./
	f, _ := os.OpenFile("/var/log/osquery-sgt.log", os.O_WRONLY | os.O_CREATE, 0755)
	logger.SetOutput(f)
	logger.SetLevel(logger.InfoLevel)
}

var config osquery_types.ServerConfig
*/

func DistributedQueryRead(respWriter http.ResponseWriter, request *http.Request) {

	handleRequest := func() error {

		respWriter.Header().Set("Content-Type", "application/json")
		body, err := ioutil.ReadAll(request.Body)
		defer request.Body.Close()
		if err != nil {
			return fmt.Errorf("failed to read request body: %s", err)
		}

		type node struct {
			NodeKey string `json:"node_key"`
		}

		n := node{}
		err = json.Unmarshal(body, &n)
		if err != nil {
			return fmt.Errorf("unmarshal failed: %s", err)
		}

		dynSvc := dyndb.DbInstance()
		distributedQuery, err := dyndb.SearchDistributedNodeKey(n.NodeKey, dynSvc)
		if err != nil {
			return fmt.Errorf("could not find node with key '%s': %s", n.NodeKey, err)
		}

		if len(distributedQuery.Queries) == 0 {
			return errors.New("no queries in list: %s")
		}

		respWriter.Write([]byte(distributedQuery.ToJSON()))
		err = dyndb.DeleteDistributedQuery(distributedQuery, dynSvc)
		if err != nil {
			return fmt.Errorf("could not delete query: %s", err)
		}

		return nil
	}

	err := handleRequest()
	if err != nil {
		logger.Error(err)
		response.WriteError(respWriter, fmt.Sprintf("[DistributedQueryRead] %s", err))
	}
}

func ParseDistributedResults(request *http.Request) ([]osquery_types.DistributedQueryResult, error) {
	results := []osquery_types.DistributedQueryResult{}
	body, err := ioutil.ReadAll(request.Body)
	//logger.Info(string(body))
	//err = json.Unmarshal(body, &dw)
	if err != nil {
		logger.Error(err)
		return results, err
	}
	//file, err := os.Open(fn)
	type dr struct {
		NodeKey  string                         `json:"node_key"`
		Queries  map[string][]map[string]string `json:"queries"`
		Statuses map[string]string              `json:"statuses"`
	}
	//decoder := json.NewDecoder(file)
	d := dr{}
	//err = decoder.Decode(&d)
	err = json.Unmarshal(body, &d)
	if err != nil {
		logger.Error(err)
		return results, err
	}
	for k, v := range d.Queries {
		for _, v1 := range v {
			qr := osquery_types.DistributedQueryResult{
				Name:           k,
				CalendarTime:   time.Now().UTC().Format("2006-01-02 03:04:05"),
				Action:         "added",
				LogType:        "result",
				Columns:        v1,
				HostIdentifier: d.NodeKey,
			}
			results = append(results, qr)
		}
	}
	return results, nil
}

func DistributedQueryWrite(respWriter http.ResponseWriter, request *http.Request) {

	handleRequest := func() error {

		fhSvc := FirehoseService()
		config, err := osquery_types.GetServerConfig("config.json")
		if err != nil {
			return fmt.Errorf("could not get server config: %s", err)
		}
		results, err := ParseDistributedResults(request)
		if err != nil {
			return fmt.Errorf("could not parsed results: %s", err)
		}
		return PutFirehoseBatch(results, config.DistributedQueryLoggerFirehoseStreamName, fhSvc)
	}

	err := handleRequest()
	if err != nil {
		logger.Error(err)
		response.WriteError(respWriter, fmt.Sprintf("[DistributedQueryWrite] %s", err))
	}

	/*type distributed_write struct {
		NodeKey  string `json:"node_key"`
		Queries  map[string][]map[string]string `json:"queries"`
		Statuses map[string]string `json:"statuses"`
	}
	dw := distributed_write{}
	*/
	//logger.Infof("Here's our dqa object %+v", dqa)
	//logger.Info(dqa.Statuses)
	//logger.Info(dqa.Queries)
	//err = PutFirehoseBatch(body, config.DistributedQueryLoggerFirehoseStreamName, FirehoseService())
	//if err != nil {
	//logger.Error(err)
	//return
	//}
	//return
}

func FirehoseService() *firehose.Firehose {
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
	fh_svc := firehose.New(session.Must(session.NewSession(&aws.Config{
		Region:      aws.String("us-east-1"),
		Credentials: creds,
	})))
	return fh_svc
}

func PutFirehoseBatch(dqr []osquery_types.DistributedQueryResult, streamname string, fhSvc *firehose.Firehose) error {
	records := []*firehose.Record{}
	//rec := &firehose.Record{Data: s}
	for a, i := range dqr {
		js, err := json.Marshal(i)
		if err != nil {
			return err
		}
		rec := &firehose.Record{Data: js}
		records = append(records, rec)
		logger.Info(a, i)
		if len(records) == 450 || a == len(dqr)-1 {
			_, err := fhSvc.PutRecordBatch(&firehose.PutRecordBatchInput{
				DeliveryStreamName: aws.String(streamname),
				Records:            records,
			})
			if err != nil {
				return err
			}
			records = records[:0]
		}
	}
	//records = append(records, rec)
	//_, err := fh_svc.PutRecordBatch(&firehose.PutRecordBatchInput{
	//DeliveryStreamName: aws.String(streamname),
	//Records: records,
	//})
	//if err != nil {
	//logger.Error(err)
	//return err
	//}
	return nil
}

// code for bulk firehose puts if needed for later
/*for a, i := range s {
rec := &firehose.Record{Data: []byte(fmt.Sprintf("%s \n", i))}
records = append(records, rec)
logger.Info(a, i)
if len(records) == 450 || a == len(s) -1 {
	_, err := fh_svc.PutRecordBatch(&firehose.PutRecordBatchInput{
		DeliveryStreamName: aws.String(streamname),
		Records: records,
	})
	if err != nil {
		logger.Error(err)
		return err
	}
	records = records[:0]
}*/

func DistributedQueryAdd(respWriter http.ResponseWriter, request *http.Request) {

	handleRequest := func() (interface{}, error) {

		type distributedQueryAdd struct {
			Nodes []osquery_types.DistributedQuery `json:"nodes"`
		}

		body, err := ioutil.ReadAll(request.Body)
		defer request.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %s", err)
		}

		nodes := distributedQueryAdd{}
		err = json.Unmarshal(body, &nodes)
		if err != nil {
			return nil, fmt.Errorf("unmarshal failed: %s", err)
		}

		dynSVC := dyndb.DbInstance()
		success := map[string]bool{}
		for _, j := range nodes.Nodes {
			err = dyndb.ValidNode(j.NodeKey, dynSVC)
			if err != nil {
				logger.Error(err)
				response.WriteError(respWriter, fmt.Sprintf("node is not valid: %s", err))
				continue
			}

			err = dyndb.UpsertDistributedQuery(j, dynSVC)
			if err != nil {
				logger.Error(err)
				success[j.NodeKey] = false
			} else {
				success[j.NodeKey] = true
			}
		}

		return success, nil
	}

	result, err := handleRequest()
	if err != nil {
		logger.Error(err)
		response.WriteError(respWriter, fmt.Sprintf("[DistributedQueryAdd] %s", err))
	} else {
		response.WriteCustomJSON(respWriter, result)
	}
}
