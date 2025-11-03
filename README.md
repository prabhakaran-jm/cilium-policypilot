# Cilium PolicyPilot

Turn real traffic into safe **CiliumNetworkPolicies** in minutes.

PolicyPilot learns from Hubble flows, proposes **least-privilege** policies, verifies them safely in kind, and explains results with diagrams.

## Quickstart
```bash
go build -o cpp ./cmd/cpp
./cpp learn
./cpp propose
./cpp verify
./cpp explain
```

## Why
Writing CiliumNetworkPolicies by hand is error-prone - too tight and you break workloads, too loose and you open security holes. PolicyPilot helps engineers find the sweet spot.

