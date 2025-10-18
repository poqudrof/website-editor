#!/bin/bash
# Quick API test script for the inline editing system

echo "🔍 Testing GET (should return empty, not edited):"
curl -s http://localhost:9000/api/content/demo:title | jq 2>/dev/null || curl -s http://localhost:9000/api/content/demo:title
echo -e "\n"

echo "📝 Testing PUT (first edit - saving original + edited content):"
curl -s -X PUT http://localhost:9000/api/content/demo:title \
  -H "Content-Type: application/json" \
  --data-raw '{"content":"My Edited Title!","original_content":"Welcome to Demo Site"}' | jq 2>/dev/null
echo -e "\n"

echo "✅ Testing GET (should return edited content):"
curl -s http://localhost:9000/api/content/demo:title | jq 2>/dev/null || curl -s http://localhost:9000/api/content/demo:title
echo -e "\n"

echo "🔄 Testing PUT again (update only edited content):"
curl -s -X PUT http://localhost:9000/api/content/demo:title \
  -H "Content-Type: application/json" \
  --data-raw '{"content":"My UPDATED Title!"}' | jq 2>/dev/null
echo -e "\n"

echo "✅ Final GET (should show updated content, original preserved):"
curl -s http://localhost:9000/api/content/demo:title | jq 2>/dev/null
echo -e "\n"

echo "✅ API tests complete!"
