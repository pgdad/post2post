#!/bin/bash

# Deployment helper script for AWS Lambda post2post receiver
set -e

FUNCTION_NAME="${1:-post2post-receiver}"
DEPLOYMENT_METHOD="${2:-terraform}"

echo "Deploying AWS Lambda post2post receiver..."
echo "Function name: $FUNCTION_NAME"
echo "Method: $DEPLOYMENT_METHOD"
echo ""

# Build the function
echo "Building Lambda function..."
./build.sh

case $DEPLOYMENT_METHOD in
  "terraform")
    echo "Deploying with Terraform..."
    cd terraform
    terraform init
    terraform apply -var="function_name=$FUNCTION_NAME"
    FUNCTION_URL=$(terraform output -raw lambda_function_url)
    ;;
    
  "cloudformation")
    echo "Deploying with CloudFormation..."
    aws cloudformation deploy \
      --template-file cloudformation/template.yaml \
      --stack-name "$FUNCTION_NAME" \
      --capabilities CAPABILITY_IAM \
      --parameter-overrides FunctionName="$FUNCTION_NAME"
    
    FUNCTION_URL=$(aws cloudformation describe-stacks \
      --stack-name "$FUNCTION_NAME" \
      --query 'Stacks[0].Outputs[?OutputKey==`Post2PostReceiverUrl`].OutputValue' \
      --output text)
    ;;
    
  "cli")
    echo "Deploying with AWS CLI..."
    
    # Check if role exists
    ROLE_NAME="${FUNCTION_NAME}-role"
    if ! aws iam get-role --role-name "$ROLE_NAME" >/dev/null 2>&1; then
      echo "Creating IAM role..."
      aws iam create-role \
        --role-name "$ROLE_NAME" \
        --assume-role-policy-document '{
          "Version": "2012-10-17",
          "Statement": [{
            "Effect": "Allow",
            "Principal": {"Service": "lambda.amazonaws.com"},
            "Action": "sts:AssumeRole"
          }]
        }'
      
      aws iam attach-role-policy \
        --role-name "$ROLE_NAME" \
        --policy-arn "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
      
      # Create and attach STS policy
      POLICY_ARN=$(aws iam create-policy \
        --policy-name "${FUNCTION_NAME}-sts-policy" \
        --policy-document '{
          "Version": "2012-10-17",
          "Statement": [{
            "Effect": "Allow",
            "Action": ["sts:AssumeRole"],
            "Resource": "*"
          }]
        }' \
        --query 'Policy.Arn' --output text)
      
      aws iam attach-role-policy \
        --role-name "$ROLE_NAME" \
        --policy-arn "$POLICY_ARN"
      
      echo "Waiting for IAM role to propagate..."
      sleep 10
    fi
    
    ROLE_ARN=$(aws iam get-role --role-name "$ROLE_NAME" --query 'Role.Arn' --output text)
    
    # Create or update function
    if aws lambda get-function --function-name "$FUNCTION_NAME" >/dev/null 2>&1; then
      echo "Updating existing function..."
      aws lambda update-function-code \
        --function-name "$FUNCTION_NAME" \
        --zip-file fileb://bootstrap.zip
    else
      echo "Creating new function..."
      aws lambda create-function \
        --function-name "$FUNCTION_NAME" \
        --runtime provided.al2023 \
        --role "$ROLE_ARN" \
        --handler bootstrap \
        --zip-file fileb://bootstrap.zip \
        --timeout 30
    fi
    
    # Create function URL if it doesn't exist
    if ! aws lambda get-function-url-config --function-name "$FUNCTION_NAME" >/dev/null 2>&1; then
      echo "Creating function URL..."
      aws lambda create-function-url-config \
        --function-name "$FUNCTION_NAME" \
        --auth-type NONE \
        --cors '{
          "AllowOrigins": ["*"],
          "AllowMethods": ["POST"],
          "AllowHeaders": ["*"],
          "MaxAge": 86400
        }'
    fi
    
    FUNCTION_URL=$(aws lambda get-function-url-config \
      --function-name "$FUNCTION_NAME" \
      --query 'FunctionUrl' --output text)
    ;;
    
  *)
    echo "Unknown deployment method: $DEPLOYMENT_METHOD"
    echo "Supported methods: terraform, cloudformation, cli"
    exit 1
    ;;
esac

echo ""
echo "Deployment completed successfully!"
echo ""
echo "Function URL: $FUNCTION_URL"
echo ""
echo "Test your deployment:"
echo "curl -X POST $FUNCTION_URL \\"
echo "  -H 'Content-Type: application/json' \\"
echo "  -d '{\"url\":\"http://httpbin.org/post\",\"request_id\":\"test-123\",\"role_arn\":\"arn:aws:iam::YOUR_ACCOUNT:role/TestRole\"}'"
echo ""
echo "Update test-client.go with your Function URL and run:"
echo "go run test-client.go"