#!/bin/bash
# Generate traffic between services in demo namespace

NAMESPACE=${1:-demo}
DURATION=${2:-60}

echo "Generating traffic in namespace: $NAMESPACE for $DURATION seconds..."

# Get pod names
FRONTEND_POD=$(kubectl get pods -n $NAMESPACE -l app=frontend -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
BACKEND_POD=$(kubectl get pods -n $NAMESPACE -l app=backend -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
DATABASE_POD=$(kubectl get pods -n $NAMESPACE -l app=database -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)

if [ -z "$FRONTEND_POD" ] || [ -z "$BACKEND_POD" ] || [ -z "$DATABASE_POD" ]; then
    echo "Error: Not all pods found. Make sure demo app is deployed."
    exit 1
fi

echo "Found pods:"
echo "  Frontend: $FRONTEND_POD"
echo "  Backend: $BACKEND_POD"
echo "  Database: $DATABASE_POD"
echo ""
echo "Generating traffic..."

# Generate traffic from frontend to backend
kubectl exec -n $NAMESPACE $FRONTEND_POD -- sh -c "
    for i in \$(seq 1 $DURATION); do
        wget -q -O- http://backend > /dev/null 2>&1
        sleep 1
    done
" &

# Generate traffic from backend to database
kubectl exec -n $NAMESPACE $BACKEND_POD -- sh -c "
    for i in \$(seq 1 $DURATION); do
        wget -q -O- http://database > /dev/null 2>&1
        sleep 1
    done
" &

echo "Traffic generation started. Waiting $DURATION seconds..."
sleep $DURATION

echo "Traffic generation complete!"
echo "Capture flows with: hubble observe --namespace $NAMESPACE --since ${DURATION}s -o json > flows.json"