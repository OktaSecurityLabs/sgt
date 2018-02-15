package main

import (
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/oktasecuritylabs/sgt/logger"
	"fmt"
	"log"
	"github.com/oktasecuritylabs/sgt/handlers/auth"
	"strconv"
	"encoding/base64"
	"io/ioutil"
	"os"
	//"gopkg.in/h2non/filetype.v1"
	"bytes"
	"strings"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	//"gopkg.in/h2non/filetype.v1"
)

type FileCarve struct {
	BlockCount string `json:"block_count"`
	BlockSize string `json:"block_size"`
	BlocksRecieved map[int]bool `json:"blocks_recieved"`
	CarveSize string `json:"carve_size"`
	CarveGUID string `json:"carve_guid"`
	SessionID string `json:"session_id"`
}

type ContinueCarvePost struct {
	BlockID string `json:"block_id"`
	SessionID string `json:"session_id"`
	Data string `json:"data"`
	BSID string `json:"bsid,omitempty"`
}


//GetCarves
func GetCarves(db *dynamodb.DynamoDB) ([]FileCarve, error) {
	carves := []FileCarve{}
	scanItems, err := db.Scan(&dynamodb.ScanInput{
		TableName: aws.String("file_carves"),
	})
	for _, i := range scanItems.Items {
		fc := FileCarve{}
		err = dynamodbattribute.UnmarshalMap(i, &fc)
		if err != nil {
			logger.Error(err)
		}
		carves = append(carves, fc)
	}
	return carves, nil
}

func GetCarve(sessionID string, db *dynamodb.DynamoDB) (FileCarve, error) {
	type querystring struct {
		SessionID string `json:"session_id"`
	}
	qs_sid := querystring{}
	qs_sid.SessionID = sessionID
	js, err := dynamodbattribute.MarshalMap(qs_sid)
	if err != nil {
		logger.Error(err)
	}
	resp, err := db.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("file_carves"),
		Key:       js,
	})
	if err != nil {
		//panic(fmt.Sprintln(err, os.Stdout))
		log.Panic(err)
	}
	filecarve := FileCarve{}
	if len(resp.Item) > 0 {
		err = dynamodbattribute.UnmarshalMap(resp.Item, &filecarve)
		if err != nil {
			fmt.Println(err)
			return filecarve, err
		}
	}
	return filecarve, nil
}

func DeleteCarveData(fc *FileCarve, db *dynamodb.DynamoDB) (error) {
	chunkSize, err := strconv.Atoi(fc.BlockSize)
	if err != nil {
		logger.Error(err)
		return err
	}
	totalChunks, err := strconv.Atoi(fc.BlockCount)
	if err != nil {
		logger.Error(err)
		return err
	}
	for i := 0; i < totalChunks; i+= chunkSize {
		blockID := i
		fmt.Println(blockID)
		bsid := fmt.Sprintf("%d_%s", blockID, fc.SessionID)
		type querystring struct {
			BSID string `json:"bsid"`
		}
		qs := querystring{BSID:bsid}
		js, err := dynamodbattribute.MarshalMap(qs)

		_, err = db.DeleteItem(&dynamodb.DeleteItemInput{
			TableName: aws.String("file_carve_data"),
			Key: js,
		})
		if err != nil {
			logger.Error(err)
			return err
		}
	}
	return nil
}


func DeleteCarve(fc *FileCarve, db *dynamodb.DynamoDB) (error) {
	err := DeleteCarveData(fc, db)
	if err != nil {
		logger.Error(err)
		return err
	}
	err = DeleteCarveData(fc, db)
	if err != nil {
		logger.Error(err)
		return err
	}
	type querystring struct {
		SessionID string `json:"session_id"`
	}
	qs_sid := querystring{}
	qs_sid.SessionID = fc.SessionID
	js, err := dynamodbattribute.MarshalMap(qs_sid)
	if err != nil {
		logger.Error(err)
	}
	_, err = db.DeleteItem(&dynamodb.DeleteItemInput{
		TableName: aws.String("file_carves"),
		Key: js,
	})
	if err != nil {
		logger.Error(err)
		return err
	}
	return nil

}

//GetCarveData returns all data for a given carve session id
//if data retrieved is less than the number of chunks expected for the carve, this
//function will return an InvalidData length error
func GetCarveData(fc *FileCarve, db *dynamodb.DynamoDB) ([]ContinueCarvePost, error) {
	blocks := []ContinueCarvePost{}
	logger.Infof("block size: %s", fc.BlockSize)
	chunkSize, err := strconv.Atoi(fc.BlockSize)
	if err != nil {
		logger.Error(err)
	}
	totalChunks, err := strconv.Atoi(fc.BlockCount)
	if err != nil {
		logger.Error(err)
		return []ContinueCarvePost{}, err
	}
	logger.Infof("total Chunks: %d", totalChunks)
	for i := 0; i < totalChunks; i++ {
		blockID := i*chunkSize
		bsid := fmt.Sprintf("%d_%s", blockID, fc.SessionID)
		type querystring struct {
			BSID string `json:"bsid"`
		}
		qs := querystring{BSID:bsid}
		js, err := dynamodbattribute.MarshalMap(qs)

		item, err := db.GetItem(&dynamodb.GetItemInput{
			TableName: aws.String("file_carve_data"),
			Key: js,
		})
		if err != nil {
			logger.Error(err)
			return []ContinueCarvePost{}, err
		}
		block := ContinueCarvePost{}
		err = dynamodbattribute.UnmarshalMap(item.Item, &block)
		if err != nil {
			logger.Error(err)
			return []ContinueCarvePost{}, err
		}
		blocks = append(blocks, block)
	}
	logger.Infof("got %d blocks", len(blocks))
	return blocks, nil
}

func ReconstructCarve(data []ContinueCarvePost) ([]byte, error) {
	var buffer bytes.Buffer
	outString := ""
	for i, _ := range data {
		//logger.Infof("joining %s", d.BSID)
		//buffer.WriteString(d.Data)

		logger.Infof("bsid: %s, len %d", data[i].BSID, len(data[i].Data))
		//logger.Info(d.Data[:10])
		outString = outString + data[i].Data

	}
	logger.Infof("en of buffer: %d", len(buffer.Bytes()))
	fileData, err := base64.StdEncoding.DecodeString(buffer.String())
	//fl := strings.Split(string(fileData), "\n")
	if err != nil {
		return []byte(""), err
	}
	//kind, unknown := filetype.Match(fileData)
	//head := fileData[:500]
	fn := getFilename(fileData)
	logger.Info(fn)
	logger.Infof("mode: %d", fileMode(fileData))
	//logger.Info(string(head))
	//logger.Infof("Kind: %+v", kind)
	//ogger.Infof("unknown: %+v", unknown)
	//logger.Info("byte split...")
	//ByteSplits(fileData)
	logger.Infof("File data size: %d", len(fileData))
	logger.Infof("outstring data size: %d", len(outString))
	return []byte(outString), nil
}

func ByteSplits(b []byte) {
	l := bytes.Split(b, []byte("\x00"))
	for _, i := range l {
		if len(i) > 0 {
			fmt.Println(string(i))
		}
	}

}

func getFilename(b []byte) string {
	fn := bytes.Split(b, []byte("\x00"))[0]
	return string(fn)
	/*for i := 0; i < len(b); i++ {
		b0 := b[i]
		b1 := b[i+1]
		if b0 == 0 {
			if b1 == 0 {
				return string(b[:i])
			}
		}
	}*/
	return ""
}

func fileMode(b []byte) int {
	m1 := bytes.Split(b, []byte("00"))[1]
	m2 := bytes.Split(m1, []byte("\x00"))[0]
	mode, err := strconv.Atoi(strings.Trim(string(m2), " "))
	if err != nil {
		logger.Error(err)
	}
	return mode

}

func UploadToS3(carve FileCarve, carveData []byte, bucketName string) {
	fn := "/home/mjane-a/.aws/credentials"
	profile := "playground"
	creds := credentials.NewSharedCredentials(fn, profile)
	s3Service := s3.New(session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
		Credentials: creds,
	})))
	_, err := s3Service.PutObject(&s3.PutObjectInput{
		Body: bytes.NewReader(carveData),
		Bucket: aws.String(bucketName),
		Key: aws.String(carve.CarveGUID),
	})
	if err != nil {
		logger.Error(err)
	}

}


func main() {
	fmt.Println("starting")
	db := auth.CrendentialedDbInstance("/home/mjane-a/.aws/credentials", "playground")
	fmt.Println("made db connection")
	carves, err := GetCarves(db)
	if len(carves) > 0 {
		fmt.Println("got carves")
		if err != nil {
			logger.Error(err)
		}
		for j, carve := range carves {
			data, err := GetCarveData(&carve, db)
			if err != nil {
				logger.Error(err)
			}
			fmt.Println(len(data))
			fmt.Println(j)
			fmt.Sprintf("%+v", carve)
			if err != nil {
				logger.Error(err)
			}
			fileData, err := ReconstructCarve(data)
			if err != nil {
				logger.Error(err)
			}
			//DeleteCarve(&carve, db)
			//if err != nil {
			//logger.Error(err)
			//}
			//UploadToS3(carve, fileData, "test-carve-bucket")
			ioutil.WriteFile(fmt.Sprintf("file%d", j), fileData, os.FileMode(0644))

		}
	}

}
