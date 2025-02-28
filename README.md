# faceit-user-svc

The [faceit-user-svc](https://github.com/zechao/faceit-user-svc) is a user management service that provides functionalities to add, modify, remove, and list users with pagination and filtering capabilities. The service is designed to be scalable, maintainable, and easy to integrate with other services.

## Quickstart
### Prerequisites

- Docker
- Docker Compose

### Running the Service
The project is containerized with all dependencies configured. 
To run the service, use the following command, this will run the application in production mode, this will set environment variables for the docker containers:

```sh
docker-compose --env-file .env.production up -d --build
```
or

```sh
make run
```

To run the service in development mode simply, it will load `.env` by default
```sh
docker-compose up -d --build
```
or

```sh
make run-dev
```
### Postman collection

You can import the provided Postman collection [`faceit-user.postman_collection.json`](faceit-user.postman_collection.json) to test the HTTP endpoints of the service. The collection includes all the necessary requests and examples to interact with the API. Once imported, you can run the requests and test the service endpoints.
  <p align="center">
  <img src="images/postman.png" align="center" alt="drawing" width="600"/>
  <p/>



## Project architecture and
The faceit-user-svc follows a modular, layered architecture designed with clean architecture principles to ensure separation of concerns, testability, and maintainability, all code is well tested.


### Architecture and design overview
The service is structured into several key components, all layer are interacted with interfaces, so dependencies are injected rather than created directly, facilitating testing and flexibility. Each component has a single responsibility, making the codebase easier to maintain.
- **Domain Model** [user](user/): Defines core domain entities and business rules
- **API Layer** [http](http/): Handles HTTP requests, input validation, and response formatting
- **Service Layer**  [service](service/): Contains business logic and orchestrates operations by calling repository for DB actions and send event to event handlers
- **Repository Layer** [repository](postgres): Manages data persistence and retrieval
- **Event System** [event](event/): Handles asynchronous communication
- **Query Handling** [query](query/): Manages filtering and pagination
- **Error Handling**: Centralized error handling [errors/errors.go](errors/errors.go) for consistent error responses.
- **Observability**: Built-in logging [log](log/) and tracing [tracing](tracing/) for monitoring and troubleshooting.
- **Database Migrations**: Structured migration system [migrations/migration.go](migrations/migration.go) for database schema evolution.
- **Configuration**: All configuration is read from system environment variables (ENV) [config](config/config.go). 


## Design details

### API Design
### User Model Design

The user model in the faceit-user-svc service is designed to encapsulate all the necessary attributes and behaviors of a user entity. It follows the principles of domain-driven design to ensure that the model is both expressive and encapsulated.




### Observability
I defined `X-Trace-ID`, which will be sent by the client as a unique trace ID in the request Header. This ID will be propagated across the service using context. The purpose of this trace ID is to track the call throughout the service. If the client does not provide the `X-Trace-ID`, it will be generated in the middleware from [`tracing.go`](tracing/tracing.go). As you can see in the following image trace ID is set in the request and response header  
  <p align="center">
  <img src="images/tracing.png" align="center" alt="drawing" width="600"/>
  <p/>
  
The log messages are structured to include key information such as timestamps, log levels, trace IDs, and contextual data to facilitate easy searching and filtering. Here is an example of a log message generated from a previous request that returned a "record already exists" error. **Note: User information is not logged to protect sensitive data.**

```json
{"time":"2025-02-28T16:12:02.643639136Z","level":"INFO","msg":"creating new user","user_id":"d297d011-84bc-4a01-a4b0-2b1e412c59e2","trace_id":"da79667b-3b8e-4f2a-95c4-4d54c846499c"}
{"time":"2025-02-28T16:12:02.738414397Z","level":"WARN","msg":"service error","error":"record already exists","trace_id":"da79667b-3b8e-4f2a-95c4-4d54c846499c"}
{"time":"2025-02-28T16:12:02.73850743Z","level":"INFO","msg":"HTTP REQUEST","status_code":409,"duration":96939071,"client_ip":"192.168.65.1","method":"POST","path":"/users","raw":"","response_size":46,"trace_id":"da79667b-3b8e-4f2a-95c4-4d54c846499c"}
```

### Configuration
 All configurations are read from system environment variables (ENV), here for simplicity we save those variables in `.env*` file, which will be loaded by our app or docker compose.
 However, in real-life microservices, it is recommended to use a configuration management system or service such as Consul, etcd, or AWS Systems Manager Parameter Store. These tools provide centralized management, versioning, and secure storage of configuration data, which enhances the scalability, maintainability, and security of your microservices.
