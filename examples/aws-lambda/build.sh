#!/bin/bash

# Build script for AWS Lambda deployment
set -e

echo "Building AWS Lambda post2post receiver..."

# Clean up previous builds
rm -f bootstrap bootstrap.zip

# Build for Linux x86_64 (Lambda runtime)
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o bootstrap main.go

# Create deployment package
zip bootstrap.zip bootstrap

# Clean up the binary
rm bootstrap

echo "Build complete! Deployment package: bootstrap.zip"
echo ""
echo "Deploy with Terraform:"
echo "  cd terraform && terraform init && terraform apply"
echo ""
echo "Or deploy with CloudFormation:"
echo "  aws cloudformation deploy --template-file cloudformation/template.yaml --stack-name post2post-receiver --capabilities CAPABILITY_IAM"
echo ""
echo "Or deploy with AWS CLI:"
echo "  aws lambda create-function --function-name post2post-receiver --runtime provided.al2023 --role <ROLE_ARN> --handler bootstrap --zip-file fileb://bootstrap.zip"