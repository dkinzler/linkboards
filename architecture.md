# Application Architecture

In the following we describe the architecture of the application and discuss some of the decisions and trade-offs made.

## Modules

The application is split into multiple modules, where each module is concerned with a different part of the application domain. In the language of DDD (Domain Driven Design) one could say that a module corresponds to one or more bounded contexts. For example while the  [boards](internal/boards) module handles link boards and their invites and users, module [links](internal/links) provides functionality to create, rate and query links on a board.

To keep modules independent, any dependency a module has on another module should generally be hidden behind an abstraction, i.e an interface in Go code.
This makes it easy to split a module off into its own application, should it ever become useful or necessary.

The structure of a module follows the [hexagonal architecture](https://en.wikipedia.org/wiki/Hexagonal_architecture_(software)) approach, which is described in the following sections.


## Hexagonal Architecture

Modules that implement and expose endpoints of the API are setup using the hexagonal architecture.
They consist of a domain, application and transport layer and use the "ports and adapters" approach to integrate dependencies like a data store or message system.
The [boards](internal/boards) and [links](internal/links) modules are both implemented using this approach.

The following sections describe the components in more detail.

### Domain Layer

The domain layer (package `domain` in code) contains all the business logic of a module. It provides operations for working with entities/resources like boards, invites or links.
Any constraints and rules are implemented and enforced here, e.g. in this application a user can be invited to a board only if they are not already a member of it.
Note that although we call this the "domain" layer, the implementation does not necessarily have to use DDD (Domain Driven Design), especially for simple applications.

### Application Layer

The application layer (package `application` in code) defines an application service that provides functionality/use cases that other systems/users can use.
In particular for an API application, the application service will define the methods/operations/endpoints the API will provide.

To implement its functionality, the application layer uses the entities/resources/functions/operations of the domain layer.
It might e.g. use multiple domain operations to achieve a task or combine different pieces of data into new representations. 

Furthermore it is responsible for enforcing authorization.

### Ports and adapters

The domain and application layers usually require dependencies like notification services, data stores or event publishers.
We integrate these using the ports and adapters approach.
Concretely in Go code this means that dependencies are defined using interfaces (the ports) which can then have multiple implementations (the adapters).

For example the domain layer of the [boards](internal/boards) module defines an interface [BoardDataStore](internal/boards/domain/datastore.go), that provides methods to read and write the data that is needed to implement the domain operations.
There are two implementations of the interface (i.e. adapters), one that stores the data in-memory and one that uses a Firestore NoSQL cloud database.

Since any type implementing an interface can be used as an adapter, it is very easy to integrate new systems and run the application with different configurations. E.g. for development and testing it is convenient and fast to use an in-memory data store because we won't have to setup and run any external database systems.  

The application layer might e.g. define an interface (port) for obtaining authorization information about the user making a request.

### Transport Layer

The transport layer makes the methods of an application service available to the outside using JSON over HTTP.
To this end the [Go kit](https://github.com/go-kit/kit) framework is used to wrap the methods of the application service using endpoints and HTTP handlers.
Almost all of the code of the transport layer is auto-generated using the [github.com/d39b/kit/codegen](https://pkg.go.dev/github.com/d39b/kit/codegen) package.

The transport layer is just another example of an adapter (see the section above), since we could just as easily make the application services/API available using another transport protocol like gRPC.


## Scaling and concurrency

This application is build to be concurrency safe and therefore scalable by running multiple instances.
To guarantee consistency we use optimistic transactions/concurrency control, which usually requires dependencies of the application, especially data stores to provide transaction capabilities.
For more information on how optimistic transactions are used in this application, refer to the code comments on [BoardDataStore](internal/boards/domain/datastore.go) and [BoardService](internal/boards/domain/service.go).

For this application we decided to limit the number of users a link board can have to 32.
This has the advantage that we can easily store and retrieve a board with all its users and invites without having to worry about memory.
The choice was made mostly for convenience to not blow up the complexity and size of this sample application. It would be straightforward to adapt the API to work with an unbounded amount of users per link board.
We would have to treat board users and invites as separate resources/entities, that need their own data store methods to create, retrieve and query them. Since the number of results is unbounded, queries need to be paginated.
