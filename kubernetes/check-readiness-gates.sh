#!/bin/bash
# Check readiness gates and conditions on API pods

NAMESPACE="professors-research"
LABEL="app=professors-research-api"

echo "=== Checking Readiness Gates and Conditions ==="
echo ""

PODS=$(kubectl get pods -n $NAMESPACE -l $LABEL -o jsonpath='{.items[*].metadata.name}')

for pod in $PODS; do
  echo "Pod: $pod"
  echo "---"
  
  echo "Readiness Gates:"
  GATES=$(kubectl get pod $pod -n $NAMESPACE -o jsonpath='{.spec.readinessGates[*].conditionType}' 2>/dev/null)
  if [ -z "$GATES" ]; then
    echo "  (none configured)"
  else
    echo "$GATES" | tr ' ' '\n' | sed 's/^/  - /'
  fi
  echo ""
  
  echo "Pod Conditions:"
  kubectl get pod $pod -n $NAMESPACE -o jsonpath='{range .status.conditions[*]}{.type}={.status} (Reason: {.reason}){if .message} - {.message}{end}{"\n"}{end}' 2>/dev/null | sed 's/^/  /'
  echo ""
  
  echo "Readiness Status:"
  READY=$(kubectl get pod $pod -n $NAMESPACE -o jsonpath='{.status.containerStatuses[0].ready}' 2>/dev/null)
  echo "  Container Ready: $READY"
  
  READINESS_GATES_READY=$(kubectl get pod $pod -n $NAMESPACE -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}' 2>/dev/null)
  echo "  Overall Ready: $READINESS_GATES_READY"
  echo ""
  
  echo "Pod Events (last 5):"
  kubectl get events -n $NAMESPACE --field-selector involvedObject.name=$pod --sort-by='.lastTimestamp' | tail -5 | sed 's/^/  /'
  echo ""
  echo "=========================================="
  echo ""
done

echo "Service Endpoints:"
kubectl get endpoints professors-research-api -n $NAMESPACE -o yaml | grep -A 10 "subsets:" || echo "  No endpoints found"





