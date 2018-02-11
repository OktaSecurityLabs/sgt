package distributed

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/oktasecuritylabs/sgt/dyndb"
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

func DistributedQueryRead(respwritter http.ResponseWriter, request *http.Request) {
	respwritter.Header().Set("Content-Type", "application/json")
	body, err := ioutil.ReadAll(request.Body)
	defer request.Body.Close()
	if err != nil {
		logger.Error(err)
		respwritter.Write([]byte(`{"error": "true"}`))
		return
	}
	type node struct {
		NodeKey string `json:"node_key"`
	}
	n := node{}
	err = json.Unmarshal(body, &n)
	//js, _ := json.Marshal(n)
	if err != nil {
		logger.Error(err)

		return
	}
	dyn_svc := dyndb.DbInstance()
	distributed_q, err := dyndb.SearchDistributedNodeKey(n.NodeKey, dyn_svc)
	if err != nil {
		logger.Error(err)
		return
	}
	switch len(distributed_q.Queries) > 0 {
	case true:
		{
			respwritter.Write([]byte(distributed_q.ToJson()))
			mu := sync.Mutex{}
			err = dyndb.DeleteDistributedQuery(distributed_q, dyn_svc, &mu)
			if err != nil {
				logger.Error(err)
				return
			}
		}
	default:
		logger.Warn("returning NULLLLL")
		respwritter.Write([]byte(`{}`))
		return

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
				k,
				time.Now().UTC().Format("2006-01-02 03:04:05"),
				"added",
				"result",
				v1,
				d.NodeKey,
			}
			results = append(results, qr)
		}
	}
	return results, nil
}

func DistributedQueryWrite(respwritter http.ResponseWriter, request *http.Request) {
	fh_svc := FirehoseService()
	config, err := osquery_types.GetServerConfig("config.json")
	if err != nil {
		logger.Error(err)
		return
	}
	results, err := ParseDistributedResults(request)
	if err != nil {
		logger.Error(err)
		return
	}
	err = PutFirehoseBatch(results, config.DistributedQueryLoggerFirehoseStreamName, fh_svc)
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

func PutFirehoseBatch(dqr []osquery_types.DistributedQueryResult, streamname string, fh_svc *firehose.Firehose) error {
	records := []*firehose.Record{}
	//rec := &firehose.Record{Data: s}
	for a, i := range dqr {
		js, err := json.Marshal(i)
		if err != nil {
			logger.Error(err)
			return err
		}
		rec := &firehose.Record{Data: js}
		records = append(records, rec)
		logger.Info(a, i)
		if len(records) == 450 || a == len(dqr)-1 {
			_, err := fh_svc.PutRecordBatch(&firehose.PutRecordBatchInput{
				DeliveryStreamName: aws.String(streamname),
				Records:            records,
			})
			if err != nil {
				logger.Error(err)
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

func DistributedQueryAdd(respwritter http.ResponseWriter, request *http.Request) {
	type distributedQueryAdd struct {
		Nodes []osquery_types.DistributedQuery `json:"nodes"`
	}

	body, err := ioutil.ReadAll(request.Body)
	defer request.Body.Close()
	if err != nil {
		logger.Error(err)
		return
	}

	nodes := distributedQueryAdd{}
	err = json.Unmarshal(body, &nodes)
	if err != nil {
		logger.Error(err)
		return
	}

	dynSVC := dyndb.DbInstance()
	mu := sync.Mutex{}
	success := map[string]bool{}
	for _, j := range nodes.Nodes {
		err = dyndb.ValidNode(j.NodeKey, dynSVC)
		if err != nil {
			logger.Error(err)
			continue
		}

		err = dyndb.UpsertDistributedQuery(j, dynSVC, &mu)
		if err != nil {
			logger.Error(err)
			success[j.NodeKey] = false
		} else {
			success[j.NodeKey] = true
		}
	}

	js, err := json.Marshal(success)
	if err != nil {
		logger.Error(err)
		return
	}

	respwritter.Write(js)
}
