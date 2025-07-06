#!/bin/bash

# GitHub Repository Setup Script
# Run this after creating pgdad/post2post repository on GitHub

set -e

echo "🚀 Setting up GitHub remote for pgdad/post2post"
echo "================================================"

# Check if we're in a git repository
if [ ! -d ".git" ]; then
    echo "❌ Error: Not in a git repository"
    exit 1
fi

# Check current status
echo "📋 Current repository status:"
git status --short
echo

# Add GitHub remote
echo "🔗 Adding GitHub remote..."
git remote add origin https://github.com/pgdad/post2post.git

# Verify remote
echo "📍 Remote configuration:"
git remote -v
echo

# Push master branch
echo "📤 Pushing master branch to GitHub..."
git push -u origin master

# Push any existing tags
echo "🏷️ Pushing tags..."
git push --tags

echo
echo "✅ Repository successfully pushed to GitHub!"
echo "🌐 Repository URL: https://github.com/pgdad/post2post"
echo
echo "📊 Repository contents:"
echo "   - Core library: post2post.go, processors.go"
echo "   - Examples: Complete example suite with Tailscale integration"
echo "   - Architecture: Comprehensive documentation suite (11 files)"
echo "   - Tests: Complete test suite"
echo
echo "📋 Next steps:"
echo "   1. Visit https://github.com/pgdad/post2post"
echo "   2. Configure repository settings (branch protection, topics)"
echo "   3. Review uploaded documentation"
echo "   4. Consider creating a v1.0.0 release"