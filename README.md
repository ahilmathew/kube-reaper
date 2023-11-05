# kube-reaper

A tool to automatically delete namespaces after a certain TTL. 

To enable namespace deletion tag the namespace with - 

```
kube-reaper/expires: 2024-05-03T13:32
```

## Installation

Install the chart directly into your kubernetes cluster using - 

```
helm repo add kube-reaper https://raw.githubusercontent.com/ahilmathew/kube-reaper/main/chart

helm upgrade --install kube-reaper kube-reaper/kube-reaper 
```

## Running locally

1. `go build .`
2. `LOCAL=true ./kube-reaper`