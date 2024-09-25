# OAUTH2 API

This is a simple OAuth2 API that allows you to authenticate users and get their access token.

## Endpoints

### GET /authorize

### POST /token

## Requirements

- Mysql database for storing clients
- Redis database (optional for storing code)
- User service api (optional for password grant type)

## Deployment

### Podman

using makefile to build and run the service

```bash
make build
make run
```

### Kubernetes
