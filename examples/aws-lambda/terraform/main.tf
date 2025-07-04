terraform {
  required_version = ">= 1.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "function_name" {
  description = "Lambda function name"
  type        = string
  default     = "post2post-receiver"
}

variable "tailscale_auth_key" {
  description = "Tailscale auth key for secure networking (optional)"
  type        = string
  default     = ""
  sensitive   = true
}

# IAM role for the Lambda function
resource "aws_iam_role" "lambda_role" {
  name = "${var.function_name}-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })
}

# IAM policy for basic Lambda execution
resource "aws_iam_role_policy_attachment" "lambda_basic" {
  role       = aws_iam_role.lambda_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

# IAM policy for STS AssumeRole - allows assuming any role
# In production, you should restrict this to specific roles
resource "aws_iam_policy" "sts_assume_role" {
  name        = "${var.function_name}-sts-policy"
  description = "Allow Lambda to assume roles via STS"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "sts:AssumeRole"
        ]
        Resource = "*"  # In production, restrict to specific role ARNs
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "lambda_sts" {
  role       = aws_iam_role.lambda_role.name
  policy_arn = aws_iam_policy.sts_assume_role.arn
}

# Lambda function
resource "aws_lambda_function" "post2post_receiver" {
  filename         = "../bootstrap.zip"
  function_name    = var.function_name
  role            = aws_iam_role.lambda_role.arn
  handler         = "bootstrap"
  runtime         = "provided.al2023"
  timeout         = 30

  environment {
    variables = {
      TAILSCALE_AUTH_KEY = var.tailscale_auth_key
    }
  }

  depends_on = [
    aws_iam_role_policy_attachment.lambda_basic,
    aws_iam_role_policy_attachment.lambda_sts,
  ]
}

# Lambda Function URL for HTTP access
resource "aws_lambda_function_url" "post2post_url" {
  function_name      = aws_lambda_function.post2post_receiver.function_name
  authorization_type = "NONE"  # Public access - adjust as needed

  cors {
    allow_credentials = false
    allow_origins     = ["*"]
    allow_methods     = ["POST"]
    allow_headers     = ["date", "keep-alive", "content-type"]
    expose_headers    = ["date", "keep-alive"]
    max_age          = 86400
  }
}

# Outputs
output "lambda_function_url" {
  description = "URL of the Lambda function"
  value       = aws_lambda_function_url.post2post_url.function_url
}

output "lambda_function_name" {
  description = "Name of the Lambda function"
  value       = aws_lambda_function.post2post_receiver.function_name
}

output "lambda_function_arn" {
  description = "ARN of the Lambda function"
  value       = aws_lambda_function.post2post_receiver.arn
}