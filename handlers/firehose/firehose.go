package firehose

//	"github.com/aws/aws-sdk-go/service/firehose"
//	"io"
//	"encoding/json"
//	log "github.com/sirupsen/logrus"

type FirehoseRecord struct {
	Data string `json:"data"`
}

type DistributedWritePost struct {
	NodeKey  string            `json:"node_key"`
	Statuses map[string]string `json:"statuses"`
	Queries  map[string][]map[string]string
}

/*
func BodyToRecordBatch(body *io.ByteReader){
	decoder := json.NewDecoder(body)
	err := decoder.Decode(&config)
	if err != nil {
		log.Error(err)
		return config, err
	}
}
*/
