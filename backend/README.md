# Test build process locally

## Build Backend

```
cd backend
docker build . -f ../infra/docker/build.Dockerfile  -t foxandhound-backend
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