# Linkboards

![GitHub Workflow Status](https://github.com/dkinzler/linkboards/actions/workflows/tests.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/dkinzler/linkboards)](https://goreportcard.com/report/github.com/dkinzler/linkboards)

A sample HTTP API/backend service for sharing links implemented in Go.
Users can create link boards and invite other users to join.
Links can be posted to boards, users can rate links and discover them using queries.

The application is intended to learn about backend development with Go, to explore common problems and solutions.
- Modular application design
- Run in the cloud with Firebase Authentication and Firestore
- Support development and testing with local/in-memory versions of dependencies like data stores
- Built using the [Go kit](https://github.com/go-kit/kit) framework
- Auto-generate endpoint and HTTP boilerplate code using [github.com/dkinzler/kit/codegen](https://pkg.go.dev/github.com/dkinzler/kit/codegen)
- Unit, integration and end-to-end/API testing
- ...

A more in-depth description of the API is provided by the [OpenAPI documentation](https://dkinzler.github.io/linkboards/).

A discussion of the architecture of the app can be found [here](architecture.md).
Among other things, it explains some of the decisions and trade-offs made, how to scale and deal with issues of concurrency and consistency.

Note that the goal of this project is to learn.
There are many ways to build an application and whether or not a particular solution is a good choice depends on the context.
The implementation presented here was designed with a larger application in mind, with many components/modules/services and possibly distributed.
If you want to build a real application of this size, you might want to consider a simpler approach (see also the [discussion here](architecture.md)).

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

This mode of running the application is useful for development and interactive testing.

### Using Firebase emulators

To run the application using Firebase Authentication and Firestore emulators:

```Shell
task run-emulators -- --port 9001
```

Note that for this to work you need to have the [Firebase CLI](https://firebase.google.com/docs/cli) installed.

To create a Firebase Authentication user and obtain a JWT token that can be used to authenticate API requests, run e.g.: 

```Shell
task firebase-helpers -- login --email "test@test.com" --pasword "test123"
```

### Run with Docker

To build a Docker image tagged as `linkboardsapi`, run:

```Shell
TAG=linkboardsapi task docker-build
```

Run the application using the image:

```Shell
docker run -it -p 9004:9004 linkboardsapi --inmem --debug --port 9004
```

## Testing

Run Go unit tests:

```Shell
task test
```

Run integration tests that require Firebase emulators:

```Shell
task test-emulators
```

Run end-to-end/API tests using local Firebase emulators:

```Shell
task api-test
```

## Go Links and Resources

* [awesome-go](https://github.com/avelino/awesome-go) - curated list of frameworks, packages, projects, tools
* [Effective Go](https://go.dev/doc/effective_go), [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
* [GopherCon 2017: Mitchell Hashimoto - Advanced Testing with Go](https://www.youtube.com/watch?v=8hQG7QlcLBk)
* [How To Use Go Interfaces](https://blog.chewxy.com/2018/03/18/golang-interfaces/)
* [Go for Industrial Programming](https://peter.bourgon.org/go-for-industrial-programming/)
* [Book: 100 Go Mistakes and How to Avoid Them](https://100go.co/)
* [How I write HTTP services in Go after 13 years - Mat Ryer](https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years/)

## License

[MIT](LICENSE)
