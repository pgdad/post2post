// Rename this file to go.mod in a separate directory to run the test client
// This avoids conflicts with the Lambda function's go.mod

module lambda-test-client

go 1.21

require github.com/pgdad/post2post v0.0.0

// Replace with local path for development
replace github.com/pgdad/post2post => ../../..