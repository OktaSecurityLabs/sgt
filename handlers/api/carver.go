package api

import (
	"net/http"
	"log"
	"io/ioutil"
	"io"
	"github.com/oktasecuritylabs/sgt/logger"
	"github.com/oktasecuritylabs/sgt/handlers/node"
	"encoding/json"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"fmt"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/oktasecuritylabs/sgt/dyndb"
)


func readBody(b io.ReadCloser) ([]byte, error) {
	body, err := ioutil.ReadAll(b)
	defer b.Close()
	if err != nil {
		return []byte(""), err
	}
	return body, nil
}

type StartCarvePost struct {
	BlockCount string `json:"block_count"`
	BlockSize string `json:"block_size"`
	CarveSize string `json:"carve_size"`
	CarveID string `json:"carve_id"`
	NodeKey string `json:"node_key"`
}

type ContinueCarvePost struct {
	BlockID string `json:"block_id"`
	SessionID string `json:"session_id"`
	Data string `json:"data"`
	BSID string `json:"bsid,omitempty"`
}

type FileCarve struct {
	BlockCount string `json:"block_count"`
	BlockSize string `json:"block_size"`
	BlocksRecieved map[int]bool `json:"blocks_recieved"`
	CarveSize string `json:"carve_size"`
	CarveGUID string `json:"carve_guid"`
	SessionID string `json:"session_id"`
}

func StoreCarve(fc FileCarve) (error) {
	db := dyndb.DbInstance()
	av, err := dynamodbattribute.MarshalMap(fc)
	if err != nil {
		fmt.Println("marshal failed")
		fmt.Println(err)
	}
	_, err = db.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("file_carves"),
		Item:      av,
	})
	if err != nil {
		logger.Error(err)
		return err
	}
	return nil
}

func StoreCarveData(ccp ContinueCarvePost) (error) {
	db := dyndb.DbInstance()
	av, err := dynamodbattribute.MarshalMap(ccp)
	if err != nil {
		fmt.Println("marshal failed")
		fmt.Println(err)
	}
	_, err = db.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("file_carve_data"),
		Item:      av,
	})
	if err != nil {
		logger.Error(err)
		return err
	}
	return nil
}

//StartCarve start care https endpoint
func StartCarve(w http.ResponseWriter, r *http.Request) {
	sessionID := node.RandomString(30)
	//store session ID with data in dynamo
	body, err := readBody(r.Body)
	if err != nil {
		logger.Error(err)
	}
	type sessionReply struct {
		SessionID string `json:"session_id"`
	}
	sid := sessionReply{SessionID:sessionID}
	js, err := json.Marshal(sid)
	log.Printf("%+v", string(body))
	scp := StartCarvePost{}
	err = json.Unmarshal(body, &scp)
	if err != nil {
		logger.Error(err)
	}
	fc := FileCarve{}
	fc.SessionID = sessionID
	fc.BlockCount = scp.BlockCount
	fc.BlockSize = scp.BlockSize
	fc.BlocksRecieved = map[int]bool{}
	fc.CarveGUID = scp.CarveID
	fc.CarveSize = scp.CarveSize
	err = StoreCarve(fc)
	if err != nil {
		logger.Error(err)
	}
	w.Write(js)

}

func ContinueCarve(w http.ResponseWriter, r *http.Request) {
	body, err := readBody(r.Body)
	if err != nil {
		logger.Error(err)
	}
	ccp := ContinueCarvePost{}
	err = json.Unmarshal(body, &ccp)
	if err != nil {
		logger.Error(err)
	}
	ccp.BSID = fmt.Sprintf("%s_%s", ccp.BlockID, ccp.SessionID)
	err = StoreCarveData(ccp)
	if err != nil {
		w.Write([]byte(`"{}"`))
		return
	}
	w.Write([]byte(`{"success": true}`))
	return
}
