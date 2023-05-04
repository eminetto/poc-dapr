# PoC Dapr

![APIs](resources/apis.png)

# Running locally with Docker Compose

### Building

```
docker compose up -d --build
```

## Using the services

### Auth

```
curl -X "POST" "http://localhost:8081/v1/auth" \
     -H 'Accept: application/json' \
     -H 'Content-Type: application/json' \
     -d $'{
  "email": "eminetto@email.com",
  "password": "12345"
}'

```

The result should be a token, like:

```
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImVtaW5ldHRvQGVtYWlsLmNvbSIsImV4cCI6MTY4MTM0ODQ3MSwiaWF0IjoxNjgxMzQ0ODQxLCJuYmYiOjE2ODEzNDQ4NDF9.GdUiLYqrXeUZNIgHDhGDhGIV1NpN941UiFBqgvSoS-4"
}
```

### Feedback

You need to use the token generated by the ```Auth``` service:

```
curl -X "POST" "http://localhost:8082/v1/feedback" \
     -H 'Accept: application/json' \
     -H 'Content-Type: application/json' \
	 -H 'Authorization:eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImVtaW5ldHRvQGVtYWlsLmNvbSIsImV4cCI6MTY4MTM0ODQ3MSwiaWF0IjoxNjgxMzQ0ODQxLCJuYmYiOjE2ODEzNDQ4NDF9.GdUiLYqrXeUZNIgHDhGDhGIV1NpN941UiFBqgvSoS-4' \
     -d $'{
  "title": "Feedback test",
  "body": "Feedback body"
}'
```

### Vote

You need to use the token generated by the ```Auth``` service:

```
curl -X "POST" "http://localhost:8083/v1/vote" \
     -H 'Accept: application/json' \
     -H 'Content-Type: application/json' \
	 -H 'Authorization:eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImVtaW5ldHRvQGVtYWlsLmNvbSIsImV4cCI6MTY4MTM0ODQ3MSwiaWF0IjoxNjgxMzQ0ODQxLCJuYmYiOjE2ODEzNDQ4NDF9.GdUiLYqrXeUZNIgHDhGDhGIV1NpN941UiFBqgvSoS-4' \
     -d $'{
  "talk_name": "Go e Microserviços",
  "score": "10"
}'
```


## Running on Kubernetes

### Install dapr cli

https://docs.dapr.io/getting-started/install-dapr-cli/

### Setup Dapr on your Kubernetes cluster

```
dapr init --kubernetes --wait
dapr status -k
kubectl apply --namespace dapr-system -f ops/k8s/zipkin.yaml
```

### Prometheus

```
kubectl create namespace dapr-monitoring
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
helm install dapr-prom prometheus-community/prometheus -n dapr-monitoring
```

### Grafana

```
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update
helm install grafana grafana/grafana -n dapr-monitoring
kubectl get secret --namespace dapr-monitoring grafana -o jsonpath="{.data.admin-password}" | base64 --decode ; echo
kubectl get pods -n dapr-monitoring
```

### Redis (used to pub-sub example)

```
helm repo add bitnami https://charts.bitnami.com/bitnami
helm install redis bitnami/redis --set image.tag=6.2
###password to be used in dapr/redis.yaml
kubectl get secret --namespace default redis -o jsonpath="{.data.redis-password}" | base64 --decode
```

### Configure Prometheus as data source

https://docs.dapr.io/operations/monitoring/metrics/grafana/#configure-prometheus-as-data-source

### Metrics Available

https://github.com/dapr/dapr/blob/master/docs/development/dapr-metrics.md

### Deploy apps

```
docker build -t eminetto/auth -f ./auth/Dockerfile .
docker push eminetto/auth:latest
```

```
docker build -t eminetto/feedbacks -f ./feedbacks/Dockerfile .
docker push eminetto/feedbacks:latest
```

```
docker build -t eminetto/votes -f ./votes/Dockerfile .
docker push eminetto/votes:latest
```

```
docker build -t eminetto/audit -f ./audit/Dockerfile .
docker push eminetto/audit:latest
```

```
kubectl create namespace auth
kubectl apply --namespace auth -f auth/ops/k8s/mysql.yaml
kubectl apply --namespace auth -f auth/ops/dapr/redis.yaml
kubectl apply --namespace auth -f auth/ops/dapr/ratelimit.yaml
kubectl apply --namespace auth -f auth/ops/dapr/dapr-config.yaml
kubectl apply --namespace auth -f auth/ops/k8s/auth.yaml
kubectl port-forward --namespace auth deployment/auth 3500:3500
```

```
kubectl create namespace feedbacks
kubectl apply --namespace feedbacks -f feedbacks/ops/k8s/mysql.yaml
kubectl apply --namespace feedbacks -f feedbacks/ops/dapr/dapr-config.yaml
kubectl apply --namespace feedbacks -f feedbacks/ops/k8s/feedbacks.yaml
kubectl port-forward --namespace feedbacks deployment/feedbacks 3501:3500
```

```
kubectl create namespace votes
kubectl apply --namespace votes -f votes/ops/k8s/mysql.yaml
kubectl apply --namespace votes -f votes/ops/dapr/dapr-config.yaml
kubectl apply --namespace votes -f votes/ops/k8s/votes.yaml
kubectl port-forward --namespace votes deployment/votes 3502:3500
```

```
kubectl create namespace audit
kubectl apply --namespace audit -f audit/ops/dapr/redis.yaml
kubectl apply --namespace audit -f audit/ops/dapr/dapr-config.yaml
kubectl apply --namespace audit -f audit/ops/k8s/audit.yaml
```

### Accessing services

```
curl -v -X "POST" "http://localhost:3500/v1.0/invoke/auth/method/v1/auth" \
     -H 'Accept: application/json' \
     -H 'Content-Type: application/json' \
     -d $'{
  "email": "eminetto@email.com",
  "password": "12345"
}'
```

```
curl -v -X "POST" "http://localhost:3501/v1.0/invoke/feedbacks/method/v1/feedback" \
     -H 'Accept: application/json' \
     -H 'Content-Type: application/json' \
	 -H 'Authorization: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImVtaW5ldHRvQGVtYWlsLmNvbSIsImV4cCI6MTY4MjY5MDY5MiwiaWF0IjoxNjgyNjg3MDYyLCJuYmYiOjE2ODI2ODcwNjJ9.KSZ9dW-aseGSxa6x9vZbP06wY7jWVFQ6r_kUuwsHUUk' \
     -d $'{
  "title": "Feedback test",
  "body": "Feedback body"
}'
```

```
curl -v -X "POST" "http://localhost:3502/v1.0/invoke/votes/method/v1/vote" \
     -H 'Accept: application/json' \
     -H 'Content-Type: application/json' \
	 -H 'Authorization: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImVtaW5ldHRvQGVtYWlsLmNvbSIsImV4cCI6MTY4MjY5MDY5MiwiaWF0IjoxNjgyNjg3MDYyLCJuYmYiOjE2ODI2ODcwNjJ9.KSZ9dW-aseGSxa6x9vZbP06wY7jWVFQ6r_kUuwsHUUk' \
     -d $'{
  "talk_name": "Go e Microserviços",
  "score": "10"
}'

```
### Dapr Dashboard

```
dapr dashboard -k -p 9999
```

### Zipkin

```
kubectl port-forward --namespace dapr-system deployment/zipkin 9411:9411
```
