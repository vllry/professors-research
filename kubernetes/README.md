# Kubernetes Deployment

This directory contains Kubernetes manifests for deploying the Professors Research application on Google Kubernetes Engine (GKE) using Gateway API with Google Cloud L7 load balancer.

## Structure

- `namespace.yaml` - Namespace for the application
- `gateway.yaml` - Gateway API Gateway resource for Google Cloud L7 load balancer
- `httproute.yaml` - Gateway API HTTPRoute for traffic routing
- `api-deployment.yaml` - API server deployment
- `api-service.yaml` - API server service
- `web-deployment.yaml` - Web frontend deployment
- `web-service.yaml` - Web frontend service

## Prerequisites

- Google Kubernetes Engine (GKE) cluster with Gateway API enabled
- Docker installed and configured
- `kubectl` configured to access your GKE cluster
- `gcloud` CLI configured
- Access to `us-central1-docker.pkg.dev/zeitgeistlabs/us-central1-docker` registry
- DNS configured to point `professorsresearch.net` to the load balancer IP (after deployment)

## Building and Pushing Images

Build and push both images to the registry:

```bash
make docker-push
```

Or build and push individually:

```bash
make docker-push-api
make docker-push-web
```

**Note:** By default, images are tagged with the current git commit hash (e.g., `abc1234`). If there are uncommitted changes, `-dirty` is appended (e.g., `abc1234-dirty`).

To use a specific tag instead:

```bash
DOCKER_TAG=v1.0.0 make docker-push
```

## SSL Certificate Setup (Google-Managed SSL Certificate)

For Gateway API, Google manages SSL certificates through **Google-managed SSL certificates** using `gcloud compute ssl-certificates`. This is the correct approach for Gateway API.

### Create Google-Managed SSL Certificate

```bash
gcloud compute ssl-certificates create professorsresearch-net-cert \
    --domains=professorsresearch.net \
    --global
```

This creates a global SSL certificate that Google will automatically provision and renew. The Gateway is configured to use this certificate via the `networking.gke.io/pre-shared-certs` option in the TLS configuration.

**Important:** The certificate will only become active after DNS is configured. See the [DNS Configuration](#dns-configuration) section below for detailed instructions on setting up DNS records.

## Configuration

### Static IP

The Gateway is configured to use the static IP `general-ingress-ip` via the `spec.addresses` field (Gateway API standard). Ensure this IP exists:

```bash
gcloud compute addresses describe general-ingress-ip --global
```

If it doesn't exist, create it:

```bash
gcloud compute addresses create general-ingress-ip --global
```

Then get the IP address:

```bash
gcloud compute addresses describe general-ingress-ip --global --format="value(address)"
```

Update your DNS to point `professorsresearch.net` to this IP address.

**Note:** For Gateway API, the static IP is specified in `spec.addresses` with `type: NamedAddress`, not as an annotation. This ensures the Gateway uses the specified IP instead of creating a new one.

### Image Tags

Update the image references in `api-deployment.yaml` and `web-deployment.yaml` if not using `latest`:

```yaml
image: us-central1-docker.pkg.dev/zeitgeistlabs/us-central1-docker/professors-research-api:your-tag
```

## Deployment

Deploy all resources:

```bash
kubectl apply -f kubernetes/
```

Or deploy individually:

```bash
kubectl apply -f kubernetes/namespace.yaml
kubectl apply -f kubernetes/api-deployment.yaml
kubectl apply -f kubernetes/api-service.yaml
kubectl apply -f kubernetes/web-deployment.yaml
kubectl apply -f kubernetes/web-service.yaml
kubectl apply -f kubernetes/gateway.yaml
kubectl apply -f kubernetes/httproute.yaml
```

## Verification

Check deployment status:

```bash
kubectl get all -n professors-research
kubectl get gateway -n professors-research
kubectl get httproute -n professors-research
```

**Important:** Ensure that your deployments are running and services have endpoints before the Gateway can create backend services:

```bash
# Check deployments are ready
kubectl get deployments -n professors-research

# Check services have endpoints
kubectl get endpoints -n professors-research

# If endpoints are missing, check pod status
kubectl get pods -n professors-research
```

If you see errors about backend services not being found, ensure:
1. Deployments are running and pods are ready
2. Services have matching endpoints
3. Services are in the same namespace as the HTTPRoute

## DNS Configuration

For the Gateway API with Google-managed SSL certificates, you **must** configure DNS to point your domain to the Gateway's IP address. This is required for the SSL certificate to be provisioned and become active.

**Deployment order:**
1. Deploy the Gateway (to get the IP address)
2. Configure DNS (point domain to Gateway IP)
3. Wait for certificate provisioning (10-60 minutes)

### Step 1: Get the Gateway IP Address

After deploying the Gateway, retrieve the load balancer IP address:

```bash
kubectl get gateway professors-research-gateway -n professors-research -o jsonpath='{.status.addresses[0].value}'
```

Or get more detailed information:

```bash
kubectl describe gateway professors-research-gateway -n professors-research
```

Look for the `Addresses` field in the output. This is the IP address you need to point your DNS to.

**Note:** If you're using a static IP (`general-ingress-ip`), you can also get it directly:

```bash
gcloud compute addresses describe general-ingress-ip --global --format="value(address)"
```

### Step 2: Configure DNS A Record

Create an A record in your DNS provider pointing `professorsresearch.net` to the Gateway IP address:

**Record Type:** `A`  
**Name/Host:** `@` or `professorsresearch.net` (depending on your DNS provider)  
**Value/IP:** `<gateway-ip-address>` (from Step 1)  
**TTL:** `3600` (or your provider's default)

**Example DNS configurations by provider:**

- **Cloudflare:** DNS → Records → Add record → Type: A, Name: @, IPv4 address: `<gateway-ip>`
- **Google Domains:** DNS → Custom resource records → Create new record → A record, Name: @, Data: `<gateway-ip>`
- **Route 53:** Hosted zones → Create record → Record type: A, Value: `<gateway-ip>`
- **Namecheap:** Advanced DNS → Add New Record → Type: A Record, Host: @, Value: `<gateway-ip>`

### Step 3: Verify DNS Propagation

Wait for DNS propagation (usually 5-60 minutes) and verify:

```bash
dig professorsresearch.net +short
# or
nslookup professorsresearch.net
```

The output should show your Gateway IP address.

### Step 4: Monitor Certificate Provisioning

Once DNS is configured and propagated, Google will begin provisioning the SSL certificate. Monitor the status:

```bash
gcloud compute ssl-certificates describe professorsresearch-net-cert --global
```

Look for the `status` field:
- `PROVISIONING` - Certificate is being provisioned (normal, wait for this to change)
- `ACTIVE` - Certificate is ready and HTTPS is enabled
- `FAILED_PROVISIONING` - Provisioning failed (check DNS configuration)

**Important:** The certificate provisioning typically takes **10-60 minutes** after DNS is correctly configured. The certificate will not become active until:
1. DNS A record points to the Gateway IP
2. DNS has propagated
3. Google validates domain ownership

## Debugging Certificate Provisioning

If your certificate is stuck in `PROVISIONING` status, use these debugging steps:

### Certificate Provisioning Requirements Checklist

Before troubleshooting, verify ALL of these requirements are met:

#### ✅ Requirement 1: DNS Points to Gateway IP (and ONLY that IP)

```bash
# Get Gateway IP
GATEWAY_IP=$(kubectl get gateway professors-research-gateway -n professors-research -o jsonpath='{.status.addresses[0].value}')
echo "Gateway IP: $GATEWAY_IP"

# Check DNS resolution
DNS_RESULT=$(dig +short professorsresearch.net | head -1)
echo "DNS resolves to: $DNS_RESULT"

# Verify they match
if [ "$GATEWAY_IP" = "$DNS_RESULT" ]; then
  echo "✓ DNS correctly points to Gateway IP"
else
  echo "✗ DNS mismatch! Update DNS A record to point to $GATEWAY_IP"
fi

# Check for multiple IPs (should only be one)
IP_COUNT=$(dig +short professorsresearch.net | wc -l)
if [ "$IP_COUNT" -eq 1 ]; then
  echo "✓ Only one IP in DNS (correct)"
else
  echo "✗ Multiple IPs found in DNS! Remove extra A records"
  dig +short professorsresearch.net
fi
```

**Critical:** DNS must point to ONLY the Gateway IP. Multiple IPs or wrong IP will prevent provisioning.

#### ✅ Requirement 2: Domain is Accessible via HTTP

Google needs to verify domain ownership by accessing it via HTTP:

```bash
# Test HTTP access
HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://professorsresearch.net)
echo "HTTP Status: $HTTP_STATUS"

if [ "$HTTP_STATUS" = "200" ] || [ "$HTTP_STATUS" = "301" ] || [ "$HTTP_STATUS" = "302" ] || [ "$HTTP_STATUS" = "404" ]; then
  echo "✓ Domain is accessible via HTTP"
else
  echo "✗ Domain not accessible! Status: $HTTP_STATUS"
  echo "  Check that Gateway is working and responding"
fi

# Test from multiple locations
echo "Testing from Google's perspective..."
curl -v http://professorsresearch.net 2>&1 | head -20
```

**Critical:** The domain must respond to HTTP requests. If it doesn't, Google cannot verify ownership.

#### ✅ Requirement 3: Gateway is Ready and Accepted

```bash
# Check Gateway status
kubectl get gateway professors-research-gateway -n professors-research -o yaml | grep -A 10 "status:"

# Look for:
# - conditions[].type: "Accepted" = "True"
# - addresses[].value should show your IP
```

**Critical:** Gateway must be `Accepted: True` and have an IP address assigned.

#### ✅ Requirement 4: Load Balancer is Operational

```bash
# Check if load balancer backend services exist
gcloud compute backend-services list --global --filter="name~gkegw"

# Check forwarding rules
gcloud compute forwarding-rules list --global --filter="IPAddress:$GATEWAY_IP"
```

**Critical:** The load balancer must be fully provisioned and operational.

#### ✅ Requirement 5: Certificate is Attached to Load Balancer

```bash
# Verify certificate is referenced in Gateway
kubectl get gateway professors-research-gateway -n professors-research -o yaml | grep -A 5 "pre-shared-certs"

# Check certificate exists
gcloud compute ssl-certificates describe professorsresearch-net-cert --global
```

**Critical:** Certificate must exist and be referenced in Gateway TLS options.

#### ✅ Requirement 6: No Proxy/CDN Interference

If using Cloudflare, CloudFront, or similar:

```bash
# Check if domain is behind a proxy
curl -I http://professorsresearch.net | grep -i "cf-\|cloudflare\|server"

# If you see Cloudflare headers, ensure:
# 1. DNS-only mode (not proxied) OR
# 2. Proxy is configured to pass through to Gateway IP
```

**Critical:** If using a proxy/CDN, it must not interfere with Google's validation requests.

#### ✅ Requirement 7: HTTP Listener is Configured

```bash
# Verify Gateway has HTTP listener on port 80
kubectl get gateway professors-research-gateway -n professors-research -o yaml | grep -A 10 "listeners:" | grep -A 5 "port: 80"
```

**Critical:** Gateway must have an HTTP (port 80) listener for Google to validate.

#### ✅ Requirement 8: No Firewall Blocking

```bash
# Test if port 80 is accessible
nc -zv professorsresearch.net 80

# Or use telnet
telnet professorsresearch.net 80
```

**Critical:** Port 80 must be accessible from the internet (Google's validation servers).

### Quick Verification Script

Run this complete check:

```bash
#!/bin/bash
echo "=== Certificate Provisioning Requirements Check ==="
echo ""

# Get Gateway IP
GATEWAY_IP=$(kubectl get gateway professors-research-gateway -n professors-research -o jsonpath='{.status.addresses[0].value}' 2>/dev/null)
if [ -z "$GATEWAY_IP" ]; then
  echo "✗ Gateway not found or has no IP"
  exit 1
fi
echo "Gateway IP: $GATEWAY_IP"
echo ""

# Check DNS
DNS_IP=$(dig +short professorsresearch.net | head -1)
if [ "$GATEWAY_IP" = "$DNS_IP" ]; then
  echo "✓ DNS points to Gateway IP"
else
  echo "✗ DNS mismatch: DNS=$DNS_IP, Gateway=$GATEWAY_IP"
fi

# Check HTTP accessibility
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" --max-time 10 http://professorsresearch.net)
if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "301" ] || [ "$HTTP_CODE" = "302" ] || [ "$HTTP_CODE" = "404" ]; then
  echo "✓ Domain accessible via HTTP (status: $HTTP_CODE)"
else
  echo "✗ Domain not accessible (status: $HTTP_CODE)"
fi

# Check Gateway status
GATEWAY_READY=$(kubectl get gateway professors-research-gateway -n professors-research -o jsonpath='{.status.conditions[?(@.type=="Accepted")].status}' 2>/dev/null)
if [ "$GATEWAY_READY" = "True" ]; then
  echo "✓ Gateway is Accepted"
else
  echo "✗ Gateway not Accepted"
fi

# Check certificate status
CERT_STATUS=$(gcloud compute ssl-certificates describe professorsresearch-net-cert --global --format="value(managed.status)" 2>/dev/null)
echo "Certificate status: $CERT_STATUS"
echo ""
echo "=== Check Complete ==="
```

### 1. Check Certificate Status in Detail

Get detailed certificate information:

```bash
gcloud compute ssl-certificates describe professorsresearch-net-cert --global --format=yaml
```

Look for:
- `managed.status` - Should show `ACTIVE` when ready
- `managed.domainStatus` - Shows status per domain with detailed error messages
- `managed.domains[].status` - Per-domain status (may show specific errors)

**Important:** Check for error messages like:
- `FAILED_NOT_VISIBLE` - Domain not accessible
- `FAILED_CAA` - CAA record issues
- `FAILED_RATE_LIMITED` - Too many certificate requests

### 1a. Check Domain-Specific Status

Google validates each domain separately. Check the detailed domain status:

```bash
gcloud compute ssl-certificates describe professorsresearch-net-cert --global \
  --format="value(managed.domains[].status)"
```

This shows the validation status for each domain in the certificate.

### 2. Verify DNS Configuration

Check that DNS is correctly pointing to your Gateway IP:

```bash
# Get the Gateway IP
GATEWAY_IP=$(kubectl get gateway professors-research-gateway -n professors-research -o jsonpath='{.status.addresses[0].value}')
echo "Gateway IP: $GATEWAY_IP"

# Check what DNS resolves to
DNS_IP=$(dig +short professorsresearch.net)
echo "DNS resolves to: $DNS_IP"

# They should match!
if [ "$GATEWAY_IP" = "$DNS_IP" ]; then
  echo "✓ DNS is correctly configured"
else
  echo "✗ DNS mismatch! Update DNS to point to $GATEWAY_IP"
fi
```

### 3. Verify DNS Propagation Globally

Check DNS from multiple locations to ensure propagation:

```bash
# Check from Google's DNS
dig @8.8.8.8 professorsresearch.net +short

# Check from Cloudflare's DNS
dig @1.1.1.1 professorsresearch.net +short

# Use online tools like:
# - https://dnschecker.org
# - https://www.whatsmydns.net
```

All should return your Gateway IP address.

### 4. Check if Domain is Accessible

Verify that Google can reach your domain:

```bash
# Test HTTP access (should work even without certificate)
curl -I http://professorsresearch.net

# Check if the Gateway is responding
curl -v http://professorsresearch.net 2>&1 | grep -i "HTTP\|server"
```

The domain should be accessible and return HTTP responses from your Gateway.

### 5. Check Certificate Domain Status

For Google-managed certificates, check the domain-specific status:

```bash
gcloud compute ssl-certificates describe professorsresearch-net-cert --global \
  --format="value(managed.domainStatus)"
```

This shows detailed status for each domain. Look for any error messages.

### 6. Common Issues and Solutions

**Issue: DNS not propagated**
- **Symptom:** DNS resolves to wrong IP or doesn't resolve
- **Solution:** Wait for DNS propagation (can take up to 48 hours, usually 5-60 minutes)
- **Check:** Use `dig` from multiple locations

**Issue: DNS points to wrong IP**
- **Symptom:** DNS resolves but not to Gateway IP
- **Solution:** Update DNS A record to point to the correct Gateway IP
- **Check:** Compare `dig professorsresearch.net` output with Gateway IP

**Issue: Gateway not accessible**
- **Symptom:** Domain doesn't respond to HTTP requests
- **Solution:** Check Gateway status, ensure it's `Accepted: True`
- **Check:** `kubectl describe gateway professors-research-gateway -n professors-research`

**Issue: Certificate stuck in PROVISIONING**
- **Symptom:** Certificate has been PROVISIONING for > 2 hours
- **Solution:** 
  1. Verify DNS is correct and propagated
  2. Ensure domain is accessible via HTTP
  3. Check for any error messages in certificate details
  4. Consider deleting and recreating the certificate if DNS was wrong initially

### 7. Force Certificate Re-validation

If DNS was incorrect initially, you may need to delete and recreate the certificate:

```bash
# Delete the certificate
gcloud compute ssl-certificates delete professorsresearch-net-cert --global

# Wait a few minutes, then recreate
gcloud compute ssl-certificates create professorsresearch-net-cert \
    --domains=professorsresearch.net \
    --global
```

**Warning:** Only do this if you're certain DNS is now correct, as it will reset the provisioning process.

### 8. Check Gateway and Load Balancer Status

Ensure the Gateway is properly configured and the load balancer is ready:

```bash
# Check Gateway status
kubectl describe gateway professors-research-gateway -n professors-research

# Look for:
# - Status: Accepted: True
# - Addresses: Should show your static IP
# - Conditions: Should show Ready: True
```

### 9. Monitor Certificate Provisioning Progress

Watch the certificate status in real-time:

```bash
watch -n 30 'gcloud compute ssl-certificates describe professorsresearch-net-cert --global --format="value(managed.status)"'
```

This updates every 30 seconds so you can see when it changes from `PROVISIONING` to `ACTIVE`.

### Step 5: Verify HTTPS

Once the certificate is `ACTIVE`, verify HTTPS is working:

```bash
curl -I https://professorsresearch.net
```

You should see a successful HTTPS response. The Gateway will automatically:
- Redirect HTTP (port 80) to HTTPS (port 443)
- Terminate TLS using the Google-managed certificate
- Route traffic based on the HTTPRoute configuration

## Routing

The HTTPRoute configures traffic routing:
- `professorsresearch.net/api/*` → API server service (port 8080)
- `professorsresearch.net/*` → Web frontend service (port 80)

## SSL Certificate

The deployment uses Google-managed SSL certificates, which automatically provisions and renews SSL certificates for `professorsresearch.net`. The certificate is referenced in the Gateway via the `networking.gke.io/pre-shared-certs` option in the TLS configuration.

**Key points:**
- Certificate provisioning requires DNS to be configured first (see DNS Configuration section above)
- Google automatically renews the certificate before expiration
- Certificate provisioning typically takes 10-60 minutes after DNS is configured and propagated
- The Gateway will automatically handle HTTP to HTTPS redirection
- This uses `gcloud compute ssl-certificates` which is the correct approach for Gateway API

## Scaling

To scale deployments:

```bash
kubectl scale deployment professors-research-api -n professors-research --replicas=3
kubectl scale deployment professors-research-web -n professors-research --replicas=3
```

## Debugging 503 Errors

If you're getting 503 errors on `/api/prize-odds`, use the debug script:

```bash
./kubernetes/debug-503.sh
```

This script checks:
1. Deployment and pod status
2. Service endpoints (most common cause of 503)
3. Pod readiness probes
4. Gateway and HTTPRoute status
5. Network connectivity

### Common Issues and Solutions

#### 1. Service Has No Endpoints

**Symptom:** `kubectl get endpoints professors-research-api` shows no addresses

**Causes:**
- Pods aren't ready (readiness probe failing)
- Pod labels don't match service selector
- Pods aren't running

**Check:**
```bash
# Check pod status
kubectl get pods -n professors-research -l app=professors-research-api

# Check if pods are ready (should show READY 1/1)
kubectl get pods -n professors-research -l app=professors-research-api -o wide

# Check pod labels match service selector
kubectl get pods -n professors-research -l app=professors-research-api --show-labels
```

**Fix:**
- If readiness probe is failing, check pod logs:
  ```bash
  kubectl logs <pod-name> -n professors-research
  ```
- Test health endpoint manually:
  ```bash
  kubectl exec -it <pod-name> -n professors-research -- curl http://localhost:8080/api/health
  ```

#### 2. Pods Running But Not Ready

**Symptom:** Pods show `Running` but `READY 0/1`

**Cause:** Readiness probe failing

**Check:**
```bash
kubectl describe pod <pod-name> -n professors-research | grep -A 5 "Readiness:"
```

**Fix:**
- Check if `/api/health` endpoint is responding
- Increase `initialDelaySeconds` if pod takes time to start
- Check pod logs for errors

#### 3. Gateway Can't Reach Backend

**Symptom:** Endpoints exist but still getting 503

**Check:**
```bash
# Check Gateway status
kubectl describe gateway professors-research-gateway -n professors-research

# Check HTTPRoute status
kubectl describe httproute professors-research-route -n professors-research

# Check for backend errors
kubectl get events -n professors-research --sort-by='.lastTimestamp' | grep -i backend
```

**Fix:**
- Ensure HTTPRoute backendRef matches service name exactly
- Check namespace matches
- Verify port matches service port

#### 4. Quick Diagnostic Commands

```bash
# Check everything at once
kubectl get all -n professors-research

# Check endpoints (most important)
kubectl get endpoints -n professors-research

# Check pod readiness
kubectl get pods -n professors-research -o custom-columns=NAME:.metadata.name,STATUS:.status.phase,READY:.status.containerStatuses[0].ready

# Check recent events
kubectl get events -n professors-research --sort-by='.lastTimestamp' | tail -20
```

## Cleanup

Remove all resources:

```bash
kubectl delete -f kubernetes/
```

Clean up SSL certificate:

```bash
gcloud compute ssl-certificates delete professorsresearch-net-cert --global
```

If you need to delete the static IP:

```bash
gcloud compute addresses delete general-ingress-ip --global
```
