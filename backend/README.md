# Test build process locally

## Build Backend

```
cd backend
docker build . -f ../infra/docker/build.Dockerfile  -t foxandhound-backend
docker tag foxandhound-backend_alpha johannesrosskopp/my_private_repository:foxandhound-backend_alpha
docker push johannesrosskopp/my_private_repository:foxandhound-backend_alpha
```

## Run databse locally

```
cd database
docker-compose up
```

## Run backend container with host network

```
docker run --network="host" foxandhound-backend
```