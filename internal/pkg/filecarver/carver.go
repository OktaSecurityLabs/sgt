package filecarver

import (
	"net/http"
	"io/ioutil"
	"fmt"
	"encoding/json"
	"github.com/oktasecuritylabs/sgt/handlers/response"
	"github.com/oktasecuritylabs/sgt/osquery_types"
	"github.com/sirupsen/logrus"
	"time"
)

var log = logrus.New()

func init() {
	log.Level = logrus.InfoLevel
	log.Formatter = &logrus.JSONFormatter{}
}

type CarverDB interface {
	CreateCarve(carveMap *osquery_types.Carve) error
	AddCarveData(data *osquery_types.CarveData) error
}


func NewSessionID() string {
	return RandString(15)
}


func StartCarve(db CarverDB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleRequest := func() (interface{}, error) {
			w.Header().Set("Content-Type", "application/json")
			body, err := ioutil.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				return nil, fmt.Errorf("failed to read request body: %s", err)
			}

			//carveMap := osquery_types.CarveMap{}
			carve := osquery_types.Carve{}

			err = json.Unmarshal(body, &carve)
			if err != nil {
				log.Errorf("invalid carve post: \n%+v", body)
				return nil, fmt.Errorf("Invalid carve post")
			}
			carve.SessionID = NewSessionID()

			log.Infof("Carve map: %+v", carve)

			err = db.CreateCarve(&carve)
			if err != nil {
				log.Errorf("failed to create carve: %s", err.Error())
				return nil, fmt.Errorf("Failed to create carve")
			}
			//create first entry in DB carve table
			type statusSuccess struct {
				Success bool `json:"success"`
				SessionID string `json:"session_id"`
			}
			ok := statusSuccess{
				Success: true,
				SessionID: carve.SessionID,
			}

			return ok, nil
		}

		ok, err := handleRequest()
		if err != nil {
			response.WriteError(w, err.Error())
		} else {
			response.WriteCustomJSON(w, ok)
		}

	})
}

func ContinueCarve(db CarverDB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleRequest := func()(interface{}, error) {
			w.Header().Set("Content-Type", "application/json")
			body, err := ioutil.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				log.Errorf("Invalid continue data: %s", err.Error())
				return nil, fmt.Errorf("Invalid continue data")
			}
			carveData := osquery_types.CarveData{}
			err = json.Unmarshal(body, &carveData)
			if err != nil {
				log.Errorf("Invalid continue data: %s", err.Error())
				return nil, fmt.Errorf("Invalid continue data")
			}
			log.Infof("block ID: %s", carveData.BlockID)
			carveData.SessionBlockID = carveData.SetSBID()
			carveData.TimeToLive = time.Now().UTC().Add(time.Hour * 4).Unix()
			err = db.AddCarveData(&carveData)
			if err != nil {
				log.Errorf("AddCarveData: %s", (err.Error()))
				return nil, fmt.Errorf("Failed to create carve")
			}
			//create first entry in DB carve table
			type statusSuccess struct {
				Success bool `json:"success"`
			}
			ok := statusSuccess{
				Success: true,
			}
			return ok, nil
		}
		status, err := handleRequest()
		if err != nil {
			log.Error(err)
		} else {
			response.WriteCustomJSON(w, status)
		}
	})
}

func DummyHandler(db CarverDB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleRequest := func() error {
			body, err := ioutil.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				return err
			}
			fmt.Println(string(body))
			return nil
		}
		err := handleRequest()
		if err != nil {
			log.Error(err)
		}
	})
}
