# Step 1 - Setup Dapr on your Kubernetes cluster
dapr init --kubernetes --wait
dapr status -k
kubectl apply --namespace dapr-system -f zipkin.yaml
kubectl apply --namespace dapr-system -f ratelimit.yaml

# Install dependencies on cluster
#o11y

# Deploy apps
docker build -t eminetto/auth -f ./auth/Dockerfile .
docker push eminetto/auth:latest

docker build -t eminetto/feedbacks -f ./feedbacks/Dockerfile .
docker push eminetto/feedbacks:latest

docker build -t eminetto/votes -f ./votes/Dockerfile .
docker push eminetto/votes:latest

kubectl create namespace auth
kubectl apply --namespace auth -f mysql_auth.yaml
kubectl apply --namespace auth -f dapr-config-auth.yaml
kubectl apply --namespace auth -f auth.yaml
kubectl port-forward --namespace auth deployment/auth 3500:3500

kubectl create namespace feedbacks
kubectl apply --namespace feedbacks -f mysql_feedbacks.yaml
kubectl apply --namespace feedbacks -f dapr-config-feedbacks.yaml
kubectl apply --namespace feedbacks -f feedbacks.yaml
kubectl port-forward --namespace feedbacks deployment/feedbacks 3501:3500

kubectl create namespace votes
kubectl apply --namespace votes -f mysql_votes.yaml
kubectl apply --namespace votes -f dapr-config-votes.yaml
kubectl apply --namespace votes -f votes.yaml
kubectl port-forward --namespace votes deployment/votes 3502:3500


## Accessing services

curl -v -X "POST" "http://localhost:3500/v1.0/invoke/auth/method/v1/auth" \
     -H 'Accept: application/json' \
     -H 'Content-Type: application/json' \
     -d $'{
  "email": "eminetto@email.com",
  "password": "12345"
}'

curl -v -X "POST" "http://localhost:3501/v1.0/invoke/feedbacks/method/v1/feedback" \
     -H 'Accept: application/json' \
     -H 'Content-Type: application/json' \
	 -H 'Authorization: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImVtaW5ldHRvQGVtYWlsLmNvbSIsImV4cCI6MTY4MjY5MDY5MiwiaWF0IjoxNjgyNjg3MDYyLCJuYmYiOjE2ODI2ODcwNjJ9.KSZ9dW-aseGSxa6x9vZbP06wY7jWVFQ6r_kUuwsHUUk' \
     -d $'{
  "title": "Feedback test",
  "body": "Feedback body"
}'

curl -v -X "POST" "http://localhost:3502/v1.0/invoke/votes/method/v1/vote" \
     -H 'Accept: application/json' \
     -H 'Content-Type: application/json' \
	 -H 'Authorization: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImVtaW5ldHRvQGVtYWlsLmNvbSIsImV4cCI6MTY4MjY5MDY5MiwiaWF0IjoxNjgyNjg3MDYyLCJuYmYiOjE2ODI2ODcwNjJ9.KSZ9dW-aseGSxa6x9vZbP06wY7jWVFQ6r_kUuwsHUUk' \
     -d $'{
  "talk_name": "Go e Microserviços",
  "score": "10"
}'

## dashboard
dapr dashboard -k -p 9999

##zipkin
kubectl port-forward --namespace dapr-system deployment/zipkin 9411:9411

## para acessar entre clusters precisa uma "gambiarra"
https://carlos.mendible.com/2020/04/05/kubernetes-nginx-ingress-controller-with-dapr/