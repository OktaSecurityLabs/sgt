#!/bin/bash
GOOS=linux go build -o ../../../lambda_functions/osquery_date_to_elastic_lambda/main ../../../lambda_functions/osquery_date_to_elastic_lambda/lambda_function.go
cwd=$(pwd)
cd ../../../lambda_functions/osquery_date_to_elastic_lambda
zip -r9 $cwd/lambda.zip main
rm main
cd $cwd
