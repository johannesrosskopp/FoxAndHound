# Test build process locally

## Build Backend

```
cd backend
docker build . -f ../infra/docker/build-backend.Dockerfile -t foxandhound-backend_local
docker tag foxandhound-backend_local johannesrosskopp/my_private_repository:foxandhound-backend_dev_latest
docker push johannesrosskopp/my_private_repository:foxandhound-backend_dev_latest
```

for a release version also tag and push an image with a version like foxandhound-backend_dev_1_0_0

## Run databse locally

```
cd database
docker-compose up
```

## Run backend container with host network

```
docker run --network="host" foxandhound-backend
```