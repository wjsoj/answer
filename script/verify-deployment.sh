#!/bin/bash

echo "Apache Answer Deployment Verification"
echo "======================================"

ERRORS=0

# Check if container is running
echo -n "Checking container status... "
if docker ps | grep -q answer; then
    echo "✓ Running"
else
    echo "✗ Not running"
    ERRORS=$((ERRORS + 1))
fi

# Check data volume
echo -n "Checking data volume... "
if docker volume ls | grep -q answer_data; then
    echo "✓ Exists"
else
    echo "✗ Not found"
    ERRORS=$((ERRORS + 1))
fi

# Check health endpoint
echo -n "Checking health endpoint... "
if curl -sf http://localhost:9080/healthz > /dev/null 2>&1; then
    echo "✓ Responding"
else
    echo "✗ Not responding"
    ERRORS=$((ERRORS + 1))
fi

# Check UI accessibility
echo -n "Checking UI accessibility... "
if curl -sf http://localhost:9080 > /dev/null 2>&1; then
    echo "✓ Accessible"
else
    echo "✗ Not accessible"
    ERRORS=$((ERRORS + 1))
fi

# Check MCP endpoint (optional, requires API key)
if [ -n "$MCP_API_KEY" ]; then
    echo -n "Checking MCP endpoint... "
    if curl -sf -X POST http://localhost:9080/answer/api/v1/mcp \
        -H "Authorization: Bearer $MCP_API_KEY" \
        -H "Content-Type: application/json" \
        -d '{"method":"tools/list"}' > /dev/null 2>&1; then
        echo "✓ Responding"
    else
        echo "✗ Not responding (check API key)"
        ERRORS=$((ERRORS + 1))
    fi
else
    echo "Skipping MCP endpoint check (set MCP_API_KEY to test)"
fi

echo ""
if [ $ERRORS -eq 0 ]; then
    echo "✓ All checks passed!"
    exit 0
else
    echo "✗ $ERRORS check(s) failed"
    exit 1
fi
