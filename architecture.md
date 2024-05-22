# Danger ahead, read this first

To decide how to build an application, context is key.
What is it trying to accomplish? How many features? How many people building it? How many people using it? What about security?
These are just some of the many questions you might ask yourself.
Depending on your answers, a "good" application could take many different forms, as could the process of building it.
There are no one-size-fits-all solutions.
For example for a system that will have millions of users it makes sense to take some care designing it.
But if you need to write a small script, that you will use once to parse some data from a file, it would be perfectly reasonable to quickly hack something together that works.

The context for this application is to learn how to build a backend service/API in Go.
It is intended to explore the problems, patterns, approaches and solutions suited for a larger application consisting of many components, possibly a distributed microservice system.
Apart from being a learning exercise, you could say that this implementation is "over-engineered" relative to its size.
For example the use of the Go kit framework adds a lot of boilerplate code.
If you actually want to build a real application of this size, a simpler approach might be a better choice,
for some ideas see [the discussion below](#simpler-approaches).

For building this application we have made the following assumptions and requirements.
- The application should be horizontally scalable by running multiple instances while still guaranteeing consistency.
  This forces us to consider all the problems of distributed systems, e.g. lost updates if multiple API requests try to modify the same data concurrently.
  To keep the system consistent, we will use optimistic transactions.
  A more detailed discussion of these problems and solutions can be found [below](#consistency,-concurrency-and-scalability).
- We made authentication and authorization a separate component, which forces us to think about how to integrate it with other components.
- The number of users that can join a link board is limited to 32, which allows us to always load a board with all its users without having to worry about memory.
  In a NoSQL database we can e.g. store a board with all its users as a single document.
  This choice was made mostly for convenience, to not blow up the complexity and size of this application.
  To adapt the API to work with an unbounded amount of users per link board, we would have to treat board users as separate resources/entities and probably add pagination capabilities.

# Application Architecture

In the following we describe the architecture of the application and discuss some of the decisions and trade-offs made.
Of course, to really see how it all works, you might also want to look into the code :).

This is a monolithic application that consists of multiple components (modules, bounded contexts), each concerned with a different part of the application domain.
For example, while the [boards](internal/boards) component handles link boards and their invites and users,
the [links](internal/links) component provides functionality to create, rate and query links on a board.
The [auth](internal/auth) component is responsible for authentication and authorization. 

Components are mostly independent, they use their own dependencies like data stores and implement their own transport mechanism for API requests (here JSON over HTTP).
Strictly separating components makes it easy to later split off some of them into their own applications.

Components are implemented using a [hexagonal approach](https://en.wikipedia.org/wiki/Hexagonal_architecture_(software)) (also "ports and adapters"), which is described in more detail below.
This makes it easy to change and extend components, to integrate new dependencies like another database, but comes at the cost of some code complexity and boilerplate code.

## Components

Components that expose API endpoints are implemented using the hexagonal/ports and adapters approach.
They consist of a domain, application and transport layer.
The [boards](internal/boards) and [links](internal/links) components are both implemented in this way, with a Go package for each layer.

The following sections describe in more detail what each layer is responsible for and how they fit together.

### Domain Layer

The domain layer (package `domain` in code) contains all the business logic of a component.
It provides operations for working with entities/resources like boards, invites or links.
Any constraints and rules are implemented and enforced here, e.g. in this application a user can be invited to a board only if they are not already a member of it.
Note that although we call this the "domain" layer, it has nothing to do with DDD (Domain Driven Design).
Code in the domain layer could use DDD principles and practices, but doesn't have to.

### Application Layer

The application layer (package `application` in code) defines an application service that provides functionality/use cases that other systems/users can use.
In particular for an API, the application service will define the methods/operations/endpoints the API will provide.

To implement its functionality, the application layer uses the entities/resources/functions/operations of the domain layer.
It might e.g. use multiple domain operations to achieve a task or combine different pieces of data into new representations.

Furthermore, the application layer is responsible for enforcing authorization.

### Ports and adapters

The domain and application layers usually require dependencies like notification services, data stores or event publishers.
We integrate these using the ports and adapters approach.
Concretely in Go code this means that dependencies are defined using interfaces (the ports) which can then have multiple implementations (the adapters).

For example the domain layer of the [boards](internal/boards) component defines an interface [BoardDataStore](internal/boards/domain/datastore.go),
that provides methods to read and write the data that is needed to work with boards, users and invites. 
There are two implementations of the interface (i.e. adapters), one that stores the data in-memory and one that uses a Firestore NoSQL cloud database.

As another example, the application layer might e.g. define an interface (port) for obtaining authorization information about the user making a request.

Since any type implementing an interface can be used as an adapter, it is very easy to integrate new systems and run the application with different configurations.
For example for development and testing it is convenient and fast to use an in-memory data store because we won't have to setup and run an external database system.  

### Transport Layer

The transport layer makes the methods/API endpoints of an application service available to the outside using JSON over HTTP.
The [Go kit](https://github.com/go-kit/kit) framework is used to wrap the methods of the application service with HTTP handlers.
Almost all of the code of the transport layer is auto-generated using the [github.com/dkinzler/kit/codegen](https://pkg.go.dev/github.com/dkinzler/kit/codegen) package.

The transport layer is just another example of an adapter, since we could just as easily make the API endpoints available using another transport protocol like gRPC.

# Simpler approaches

The architecture described above is only one way to build a backend service/API in Go, you can find many other approaches.

Here is a simple one, instead of splitting functionality among multiple layers, you could implement an entire component in a single Go package.
Each API endpoint is implemented as an HTTP handler that for example:
* decodes the HTTP request
* checks authentication and authorization
* loads some data from a database
* performs a business logic operation or other computation
* encodes the HTTP response and sends it back to the client

For small applications this is a viable approach.
Instead of a more complex framework like Go kit you could e.g. use [Echo](https://github.com/labstack/echo).
You also don't necessarily have to hide dependencies like a database behind an interface.
Just implement your database code in the same package and use it in the HTTP handlers.

Pros and cons
* cuts down on the boilerplate code
* code can be easier to follow, everything is in one place
* you don't need to learn how a framework like Go kit works
* can quickly become a mess as application size grows
* can become harder to test code if all concerns are mixed in HTTP handlers

You could also choose a middle ground and e.g. keep transport and dependency code separate, but merge the domain layer into the application layer.

# Consistency, Concurrency and Scalability

This application is build to be concurrency safe and therefore scalable by running multiple instances.

To guarantee consistency, we need to make sure data isn't modified concurrently.
E.g. if two requests try to concurrently accept the same invite to a board, at most one of the requests should succeed.
How likely is that to happen? That really depends on the application and how many users/other systems are interacting with it.
It often seems very unlikely, but problems with concurrency and consistency can be subtle, e.g. if for some reason requests are slow (5-10s, maybe caused by a database),
the same user might send the same request multiple times.
In many domains (e.g. finance) it is absolutely necessary to make sure that data stays consistent.

We use optimistic transactions/concurrency control to maintain consistency in this application.
Optimistic transactions, in the context of processing an API request, basically work like this:
* Every time we read some data `d` e.g. from a database, we also get the last time `t` it was modified (this might not to be an actual time, but a version integer).
* Perform some task with `d`, during this time another request might also read and try to modify `d`.
* When we try to write/modify `d` in the data store, check again the last time `t'` that `d` was modified.
* If `t != t'` cancel the request, because the data was already modified by another request.

To actually implement this, you need a data store that supports optimistic transactions.
For more information on how optimistic transactions are used in this application,
refer to the code comments on [BoardDataStore](internal/boards/domain/datastore.go) and [BoardService](internal/boards/domain/service.go).

Note that there are other approaches to maintain consistency in a distributed system.
One such approach is to process requests in sequence, e.g. all requests for a particular link board are performed in order.
This could e.g. be implemented by routing all requests with a given board id to the application same instance.
As always there is a trade-off, you don't have to reason about concurrent requests anymore, but
you now need extra work on the system-level to route requests. E.g. starting a new instance is not as easy anymore, which requests does it get?
With optimistic transactions no coordination is necessary, you can start new instances and any instance can handle any request.

There are many other problems that can affect the correctness of your application.
* If you want to modify multiple separate pieces of data, you have to make sure that either all or none of the changes happen. Database transactions are the usual approach to solve this.
* Imagine, in a single request, you want to modify data using a database but also send out an event to a messaging system like Kafka. What happens if only one of these operations succeeds?
  Getting this right can be difficult.
* and probably many more ...

