# Linkboards

![GitHub Workflow Status](https://github.com/d39b/linkboards/actions/workflows/tests.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/d39b/linkboards)](https://goreportcard.com/report/github.com/d39b/linkboards)

A sample API application for sharing links implemented in Go.
Users can create link boards and invite other users to join.
Links can be posted to boards, users can rate and discover them using queries.

The application is built using the [Go kit](https://github.com/go-kit/kit) and [github.com/d39b/kit](https://github.com/d39b/kit) frameworks.
It is inteded to showcase Go development best practices:

- Design
	- Modular monolithic application ready to be separated into different services if it becomes necessary 
    - Hexagonal architecture makes it easy to change and integrate dependencies like data stores
	- With API scalability in mind: consistent while running multiple instances and concurrent requests
- Run in the cloud with Firebase Authentication and Firestore
- Local/in-memory versions of dependencies like data stores that allows the application to be easily run locally for development and testing.
- Endpoint and http boilerplate code is fully auto-generated using package [github.com/d39b/kit/codegen](https://pkg.go.dev/github.com/d39b/kit/codegen)
- Unit, integration and end-to-end/API testing
- ...

A more in-depth description of the API is provided by the [OpenAPI documentation](https://d39b.github.io/linkboards/).

A discussion of the architecture of the app can be found [here](architecture.md).

## Running the application

To build, test and run the application more conveniently, we use the [Task](https://taskfile.dev) task runner. See the [Taskfile.yml](Taskfile.yml) file for the commands to run tasks manually.

### Locally

To run the application using local/in-memory dependencies:

```Shell
task run-inmem -- --port 9001
```

This will make the API available on `localhost:9001` in debug mode.
HTTP requests can be authenticated using basic authentication, i.e. by using the Authorization header with the format `Basic base64(userId:password)`.
Note that any userId/password combination will be accepted.

This mode of running the application is useful for development and interactive testing e.g. with Postman.

### Using Firebase emulators

To run the application using Firebase Authentication and Firestore emulators:

```Shell
task run-emulators -- --port 9001
```

Note that for this to work you need to have the [Firebase CLI](https://firebase.google.com/docs/cli) installed.

To create a Firebase Authentication user and obtain a JWT token that can be used to authenticate API requests, run e.g.: 

```Shell
task firebase-helpers -- login --email "test@test.de" --pasword "test123"
```

## Testing

Run Go unit tests:

```Shell
task test
```

Run Go tests that require Firebase emulators:

```Shell
task test-emulators
```

Run end-to-end/API tests using local Firebase emulators:

```Shell
task api-test
```

## License

[MIT](LICENSE)
