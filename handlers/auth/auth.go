package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/howeyc/gopass"
	"github.com/oktasecuritylabs/sgt/dyndb"
	"github.com/oktasecuritylabs/sgt/logger"
	"github.com/oktasecuritylabs/sgt/osquery_types"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/ssh/terminal"
)


//SsmClient returns an instance of ssm client with credentials provided by ec2 assumed role
func SsmClient() *ssm.SSM {
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
	ssmSVC := ssm.New(session.Must(session.NewSession(&aws.Config{
		Region:      aws.String("us-east-1"),
		Credentials: creds,
	})))
	return ssmSVC
}

//GetSsmParam returns value of a named ssm parameter
func GetSsmParam(s string) (string, error) {
	svc := SsmClient()
	ans, err := svc.GetParameter(&ssm.GetParameterInput{
		Name:           aws.String(s),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		logger.Error(err)
		return "", err
	}
	paramValue := *ans.Parameter.Value
	return paramValue, nil
}

//CrendentialedDbInstance returns an instance of dynamodb using an aws credential profile
func CrendentialedDbInstance(fn, profile string) *dynamodb.DynamoDB {
	creds := credentials.NewSharedCredentials(fn, profile)
	dynDB := dynamodb.New(session.Must(session.NewSession(&aws.Config{
		Region:      aws.String("us-east-1"),
		Credentials: creds,
	})))
	return dynDB
}

//GetPass gets password
//
// Deprecated: no longer in use
func GetPass() ([]byte, error) {
	fmt.Println("Enter Password")
	pass, err := terminal.ReadPassword(0)
	if err != nil {
		logger.Error(err)
		return []byte(""), err
	}
	return pass, nil
}

//NewUser creates new user
func NewUser(credentialsFile, profile, username, role string) error {
	u := osquery_types.User{}
	u.Username = username
	logger.Info("Enter password")
	pass1, err := gopass.GetPasswd()
	logger.Info("Enter password again")
	pass2, err := gopass.GetPasswd()
	if string(pass1) != string(pass2) {
		logger.Info("passwords do not match, please try again")
		os.Exit(0)
	}
	if err != nil {
		logger.Error(err)
	}
	hash, err := bcrypt.GenerateFromPassword(pass1, bcrypt.DefaultCost)
	if err != nil {
		logger.Error(err)
	}
	dynDB := CrendentialedDbInstance(credentialsFile, profile)
	u.Password = hash
	u.Role = role
	mu := sync.Mutex{}
	err = dyndb.NewUser(u, dynDB, mu)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

//ValidateUser checks if user is valid
func ValidateUser(request *http.Request) error {
	type userPost struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	up := userPost{}
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(body, &up)
	if err != nil {
		logger.Error(err)
		return err
	}
	dynDB := dyndb.DbInstance()
	user, err := dyndb.GetUser(up.Username, dynDB)
	if err != nil {
		logger.Error(err)
		return err
	}
	err = user.Validate(up.Password)
	if err != nil {
		return err
	}
	return nil
}

//GetTokenHandler handles requests to get-token api endpoint
func GetTokenHandler(respwritter http.ResponseWriter, request *http.Request) {
	err := ValidateUser(request)
	if err != nil {
		logger.Error(err)
		respwritter.Write([]byte(`{"Error": "Invalid Username or Password"`))
		return
	}
	logger.Info("valid user!")
	appSecret, err := GetSsmParam("sgt_app_secret")
	secret := []byte(appSecret)
	if err != nil {
		logger.Error(err)
		respwritter.Write([]byte(`{"Error": "Invalid Username or Password"`))
		return
	}
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = time.Now().Add(time.Second * 14400).Unix()
	claims["iat"] = time.Now().Unix()
	tokenString, err := token.SignedString(secret)

	if err != nil {
		logger.Error(err)
		respwritter.Write([]byte(`{"Error": "Invalid Username or Password"`))
		return
	}
	respwritter.Write([]byte(fmt.Sprintf(`{"Authorization": %q}`, tokenString)))
}

//AnotherValidation validates authorization tokens.  Is poorly named and up for refactor as time permits
func AnotherValidation(respwritter http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	appSecret, err := (GetSsmParam("sgt_app_secret"))
	secret := []byte(appSecret)
	if err != nil {
		logger.Error(err)
		logger.Info("Invalid User or Password")
		respwritter.Write([]byte(`{"Error": "Invalid Username or Password"}`))
		return
	}
	token, err := request.ParseFromRequest(req, request.AuthorizationHeaderExtractor,
		func(token *jwt.Token) (interface{}, error) {
			return secret, nil
		})
	if err != nil {
		logger.Error(err)
		respwritter.Write([]byte(`{"Error": "Invalid Username or Password"`))
		return
	}
	if token.Valid {
		next(respwritter, req)
	}
}

//GetNodeSecret gets current node secret from ssm parameter store
func GetNodeSecret() (string, error) {
	secret, err := GetSsmParam("sgt_node_secret")
	if err != nil {
		logger.Error(err)
		return "", err
	}
	return secret, nil
}

//NodeConfigurePost type for handling post requests made by node
type NodeConfigurePost struct {
	EnrollSecret   string `json:"enroll_secret"`
	NodeKey        string `json:"node_key"`
	HostIdentifier string `json:"host_identifier"`
}

//ValidNodeKey validates posted node key
func ValidNodeKey(respwritter http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	logger.Info("validating node...")

	//req_copy := req
	dynDB := dyndb.DbInstance()
	respwritter.Header().Set("Content-Type", "application/json")
	body, err := ioutil.ReadAll(req.Body)
	req.Body.Close()
	//body := ioutil.NopCloser(bytes.NewReader(buf))
	req.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	if err != nil {
		logger.Error(err)
		respwritter.Write([]byte(`{"Error": "Invalid Credentials"}`))
		return
	}
	//respwritter.Write(body)
	var data NodeConfigurePost
	// unmarshal post data into data
	err = json.Unmarshal(body, &data)
	if err != nil {
		logger.Warn("unmarshal error")
		respwritter.Write([]byte(`{"Error": "Invalid Credentials"}`))
		return
	}
	validNode, err := dyndb.ValidNode(data.NodeKey, dynDB)
	if err != nil {
		logger.Error(err)
		respwritter.Write([]byte(`{"Error": "Invalid Credentials"}`))
		return
	}
	if !validNode {
		respwritter.Write([]byte(`{"Error": "Invalid Credentials"}`))
		return
	}
	if validNode {
		next(respwritter, req)
	}
}
