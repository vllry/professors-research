#!/bin/bash
# Debug script for 503 errors on /api/prize-odds
# Run this script to diagnose why the API service is returning 503

set -e

NAMESPACE="professors-research"
SERVICE_NAME="professors-research-api"
DEPLOYMENT_NAME="professors-research-api"

echo "=== Debugging 503 Error on /api/prize-odds ==="
echo ""

echo "1. Checking Deployment Status..."
kubectl get deployment $DEPLOYMENT_NAME -n $NAMESPACE
echo ""

echo "2. Checking Pod Status..."
kubectl get pods -n $NAMESPACE -l app=$SERVICE_NAME -o wide
echo ""

echo "3. Checking Service..."
kubectl get service $SERVICE_NAME -n $NAMESPACE
echo ""

echo "4. Checking Endpoints (CRITICAL - should have addresses)..."
kubectl get endpoints $SERVICE_NAME -n $NAMESPACE
echo ""

echo "5. Detailed Endpoint Information..."
kubectl get endpoints $SERVICE_NAME -n $NAMESPACE -o yaml | grep -A 10 "subsets:" || echo "  ⚠️  No endpoints found - this is likely the problem!"
echo ""

echo "6. Checking Pod Labels (should match service selector)..."
kubectl get pods -n $NAMESPACE -l app=$SERVICE_NAME --show-labels
echo ""

echo "7. Checking Readiness Probe Status..."
PODS=$(kubectl get pods -n $NAMESPACE -l app=$SERVICE_NAME -o jsonpath='{.items[*].metadata.name}')
for pod in $PODS; do
  echo "  Pod: $pod"
  kubectl get pod $pod -n $NAMESPACE -o jsonpath='{.status.conditions[?(@.type=="Ready")]}' | jq -r '.status, .reason, .message' 2>/dev/null || echo "    (jq not available, checking manually...)"
  echo ""
done

echo "8. Testing Health Endpoint from Pod (if pod exists)..."
POD=$(kubectl get pods -n $NAMESPACE -l app=$SERVICE_NAME -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
if [ -n "$POD" ]; then
  echo "  Testing /api/health on pod: $POD"
  kubectl exec -n $NAMESPACE $POD -- curl -s -o /dev/null -w "HTTP Status: %{http_code}\n" http://localhost:8080/api/health || echo "  ⚠️  Failed to connect to pod"
else
  echo "  ⚠️  No pods found"
fi
echo ""

echo "9. Checking Gateway Status..."
kubectl get gateway professors-research-gateway -n $NAMESPACE
echo ""

echo "10. Checking HTTPRoute Status..."
kubectl get httproute professors-research-route -n $NAMESPACE
echo ""

echo "11. Detailed HTTPRoute Status..."
kubectl describe httproute professors-research-route -n $NAMESPACE | grep -A 20 "Status:" || echo "  (No status section found)"
echo ""

echo "12. Checking for Recent Pod Events..."
kubectl get events -n $NAMESPACE --field-selector involvedObject.kind=Pod --sort-by='.lastTimestamp' | tail -10
echo ""

echo "=== Summary ==="
ENDPOINTS=$(kubectl get endpoints $SERVICE_NAME -n $NAMESPACE -o jsonpath='{.subsets[*].addresses[*].ip}' 2>/dev/null)
if [ -z "$ENDPOINTS" ]; then
  echo "❌ PROBLEM: Service has no endpoints!"
  echo "   This is why you're getting 503. Check:"
  echo "   - Are pods running? (step 2)"
  echo "   - Are pods ready? (step 7)"
  echo "   - Do pod labels match service selector? (step 6)"
else
  echo "✓ Service has endpoints: $ENDPOINTS"
  echo "  If still getting 503, check:"
  echo "   - Gateway/HTTPRoute status (steps 9-11)"
  echo "   - Network policies blocking traffic"
  echo "   - Backend health checks"
fi





