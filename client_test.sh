#!/bin/bash

# 🧠 Howchestrator LITE: Client Simulation Script
# This mimics an external application (like your Webapp) requesting a new resource.

BRAIN_URL="http://localhost:8080"

echo "🚀 [CLIENT] Requesting a new generic resource from Howchestrator..."
RESPONSE=$(curl -s -X POST "$BRAIN_URL/api/v1/resources")

if [ $? -eq 0 ]; then
    RESOURCE_ID=$(echo $RESPONSE | grep -oP '"id":"\K[^"]+')
    PORT=$(echo $RESPONSE | grep -oP '"port":\K\d+')
    
    echo "✅ [SUCCESS] Brain accepted request!"
    echo "   📍 Resource ID: $RESOURCE_ID"
    echo "   📍 Suggested Port: $PORT"
    echo "   📍 Status: STARTING"
    echo ""
    echo "⏳ Now, watch the 'control-plane' or 'agent' terminal logs..."
    echo "   In 3 seconds, the Agent will simulate opening the port and report back via Webhook."
else
    echo "❌ [ERROR] Could not connect to the Control Plane at $BRAIN_URL."
fi
