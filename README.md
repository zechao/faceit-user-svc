# faceit-user-svc

The faceit-user-svc is a user management service that provides functionalities to add, modify, remove, and list users with pagination and filtering capabilities. The service is designed to be scalable, maintainable, and easy to integrate with other services.

## Quickstart

### Prerequisites

- Docker
- Docker Compose

### Running the Service

To run the service, use the following command, this will run the application in production mode:

```sh
docker-compose --env-file .env.production up -d --build
```
or

```sh
make run
```

