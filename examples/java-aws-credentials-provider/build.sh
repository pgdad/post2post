#!/bin/bash

# Build script for Post2Post AWS Credentials Provider Java library
# This script compiles, tests, and packages the Java library

set -e

echo "🚀 Building Post2Post AWS Credentials Provider Java Library"
echo "============================================================"

# Check if Maven is installed
if ! command -v mvn &> /dev/null; then
    echo "❌ Error: Maven is not installed or not in PATH"
    echo "Please install Maven 3.6+ to build this library"
    exit 1
fi

# Check Maven version
echo "📋 Checking Maven version..."
mvn_version=$(mvn -version | head -1)
echo "Using: $mvn_version"

# Check Java version
echo "📋 Checking Java version..."
java_version=$(java -version 2>&1 | head -1)
echo "Using: $java_version"

# Ensure we're in the right directory
if [ ! -f "pom.xml" ]; then
    echo "❌ Error: pom.xml not found. Please run this script from the java-aws-credentials-provider directory"
    exit 1
fi

# Clean previous builds
echo "🧹 Cleaning previous builds..."
mvn clean

# Compile the code
echo "🔨 Compiling source code..."
mvn compile

# Run tests
echo "🧪 Running unit tests..."
mvn test

# Package the library
echo "📦 Packaging JAR files..."
mvn package

# Display build results
echo ""
echo "✅ Build completed successfully!"
echo ""
echo "📊 Build artifacts created:"
echo "   - JAR file: target/aws-credentials-provider-1.0.0.jar"
echo "   - Sources JAR: target/aws-credentials-provider-1.0.0-sources.jar"
echo "   - Javadoc JAR: target/aws-credentials-provider-1.0.0-javadoc.jar"
echo ""

# Show JAR file details
if [ -f "target/aws-credentials-provider-1.0.0.jar" ]; then
    jar_size=$(du -h target/aws-credentials-provider-1.0.0.jar | cut -f1)
    echo "📈 JAR file size: $jar_size"
    echo ""
    
    echo "📋 JAR contents:"
    jar tf target/aws-credentials-provider-1.0.0.jar | grep -E "\.class$" | head -10
    class_count=$(jar tf target/aws-credentials-provider-1.0.0.jar | grep -E "\.class$" | wc -l)
    if [ $class_count -gt 10 ]; then
        echo "   ... and $((class_count - 10)) more classes"
    fi
    echo ""
fi

# Installation instructions
echo "🔧 To install to local Maven repository:"
echo "   mvn install"
echo ""
echo "🔧 To use in your project, add to pom.xml:"
echo "   <dependency>"
echo "     <groupId>org.pgdad.post2post</groupId>"
echo "     <artifactId>aws-credentials-provider</artifactId>"
echo "     <version>1.0.0</version>"
echo "   </dependency>"
echo ""

# Example usage
echo "🎯 To run the example application:"
echo "   export AWS_LAMBDA_URL=\"https://your-lambda-url.amazonaws.com/\""
echo "   export AWS_ROLE_ARN=\"arn:aws:iam::123456789012:role/remote/MyRole\""
echo "   export TAILSCALE_AUTH_KEY=\"tskey-auth-your-key\""
echo "   export CALLBACK_URL=\"http://your-callback.tailnet.ts.net:8080/callback\""
echo "   mvn exec:java -Dexec.mainClass=\"org.pgdad.post2post.AwsCredentialsExample\""
echo ""

echo "🎉 Java library build complete!"