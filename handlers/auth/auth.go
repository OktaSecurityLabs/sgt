package auth

import (
	"bytes"
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
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/howeyc/gopass"
	"github.com/oktasecuritylabs/sgt/dyndb"
	"github.com/oktasecuritylabs/sgt/handlers/response"
	"github.com/oktasecuritylabs/sgt/logger"
	"github.com/oktasecuritylabs/sgt/osquery_types"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/ssh/terminal"
)

// NodeConfigurePost type for handling post requests made by node
type NodeConfigurePost struct {
	EnrollSecret   string `json:"enroll_secret"`
	NodeKey        string `json:"node_key"`
	HostIdentifier string `json:"host_identifier"`
}

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

// NewUser creates new user
func NewUser(credentialsFile, profile, username, role string) error {
	fmt.Print("Enter password: ")
	pass1, err := gopass.GetPasswd()
	if err != nil {
		return err
	}

	fmt.Print("Enter password again: ")
	pass2, err := gopass.GetPasswd()
	if err != nil {
		return err
	}

	if string(pass1) != string(pass2) {
		return errors.New("passwords do not match, please try again")
	}

	hash, err := bcrypt.GenerateFromPassword(pass1, bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := osquery_types.User{
		Username: username,
		Password: hash,
		Role:     role,
	}

	dynDB := CrendentialedDbInstance(credentialsFile, profile)

	return dyndb.NewUser(user, dynDB)
}

// ValidateUser checks if user is valid
func ValidateUser(request *http.Request) error {
	type userPost struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		panic(err)
	}

	up := userPost{}
	err = json.Unmarshal(body, &up)
	if err != nil {
		return err
	}

	dynDB := dyndb.DbInstance()
	user, err := dyndb.GetUser(up.Username, dynDB)
	if err != nil {
		return err
	}

	err = user.Validate(up.Password)
	if err != nil {
		return err
	}

	return nil
}

// GetTokenHandler handles requests to get-token api endpoint
func GetTokenHandler(respWriter http.ResponseWriter, request *http.Request) {

	handleRequest := func() (string, error) {

		err := ValidateUser(request)
		if err != nil {
			return "", err
		}

		logger.Info("valid user!")

		appSecret, err := GetSsmParam("sgt_app_secret")
		if err != nil {
			return "", err
		}

		token := jwt.New(jwt.SigningMethodHS256)
		claims := token.Claims.(jwt.MapClaims)
		claims["exp"] = time.Now().Add(time.Second * 14400).Unix()
		claims["iat"] = time.Now().Unix()
		tokenString, err := token.SignedString(appSecret)

		if err != nil {
			return "", err
		}

		return tokenString, nil
	}

	tokenValue, err := handleRequest()
	if err != nil {
		logger.Error(err)
		errString := fmt.Sprintf("[GetTokenHandler] invalid username or password: %s", err)
		response.WriteError(respWriter, errString)
	} else {
		response.WriteCustomJSON(respWriter, response.SGTCustomResponse{"Authorization": tokenValue})
	}
}

// AnotherValidation validates authorization tokens.  Is poorly named and up for refactor as time permits
func AnotherValidation(respWriter http.ResponseWriter, req *http.Request, next http.HandlerFunc) {

	handleRequest := func() (*jwt.Token, error) {

		appSecret, err := GetSsmParam("sgt_app_secret")
		secret := []byte(appSecret)
		if err != nil {
			return nil, err
		}

		keyFunc := func(token *jwt.Token) (interface{}, error) {
			return secret, nil
		}
		return request.ParseFromRequest(req, request.AuthorizationHeaderExtractor, keyFunc)
	}

	token, err := handleRequest()
	if err != nil {
		logger.Error(err)
		errString := fmt.Sprintf("[AnotherValidation] invalid username or password: %s", err)
		response.WriteError(respWriter, errString)
	} else if token.Valid {
		next(respWriter, req)
	}
}

// GetNodeSecret gets current node secret from ssm parameter store
func GetNodeSecret() (string, error) {
	secret, err := GetSsmParam("sgt_node_secret")
	if err != nil {
		logger.Error(err)
		return "", err
	}
	return secret, nil
}

// ValidNodeKey validates posted node key
func ValidNodeKey(respWriter http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	logger.Info("validating node...")

	handleRequest := func() error {

		dynDB := dyndb.DbInstance()
		respWriter.Header().Set("Content-Type", "application/json")
		body, err := ioutil.ReadAll(req.Body)
		req.Body.Close()

		req.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		if err != nil {
			return err
		}

		var data NodeConfigurePost
		// unmarshal post data into data
		err = json.Unmarshal(body, &data)
		if err != nil {
			return fmt.Errorf("unmarshal failed: %s", err)
		}

		return dyndb.ValidNode(data.NodeKey, dynDB)
	}

	err := handleRequest()
	if err != nil {
		logger.Error(err)
		errString := fmt.Sprintf("[ValidNodeKey] invalid node key: %s", err)
		response.WriteError(respWriter, errString)
	} else {
		next(respWriter, req)
	}
}
