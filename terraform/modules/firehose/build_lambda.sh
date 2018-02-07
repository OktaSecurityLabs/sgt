#!/bin/bash
GOOS=linux go build -o ../../../lambda_functions/osquery_date_to_elastic_lambda/lambda_function ../../../lambda_functions/osquery_date_to_elastic_lambda/lambda_function.go
cwd=$(pwd)
cd ../../../lambda_functions/osquery_date_to_elastic_lambda
zip -r9 $cwd/lambda.zip lambda_function
rm lambda_function
cd $cwd
