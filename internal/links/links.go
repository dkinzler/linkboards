package links

import (
	"github.com/dkinzler/linkboards/internal/auth"
	"github.com/dkinzler/linkboards/internal/links/application"
	fs "github.com/dkinzler/linkboards/internal/links/datastore/firestore"
	"github.com/dkinzler/linkboards/internal/links/datastore/inmem"
	"github.com/dkinzler/linkboards/internal/links/domain"
	"github.com/dkinzler/linkboards/internal/links/transport"

	e "github.com/dkinzler/kit/endpoint"
	"github.com/dkinzler/kit/errors"
	"github.com/dkinzler/kit/log"

	"cloud.google.com/go/firestore"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

// Contains all the necessary elements for running the links service of the API.
type Component struct {
	ApplicationService application.LinkApplicationService
	DataStore          domain.LinkDataStore
	Endpoints          transport.EndpointSet
}

// Configures link component.
type Config struct {
	Logger *log.Logger

	// If true, use an in-memory data store.
	UseInmemDataStore bool
	// Used to create a firestore data store, if UseInmemDataStore is set to false.
	FirestoreConfig *FirestoreConfig
	// AuthorizationStore used to perform authorization in the application service.
	AuthorizationStore auth.AuthorizationStore

	// Middlewares that should be applied to all endpoints
	Middlewares []endpoint.Middleware
	// Authentication middleware for endpoints.
	AuthMiddleware endpoint.Middleware
	// Whether to add the logging middleware from the "internal/pkg/endpoint" package to every endpoint.
	// It logs errors from the underlying application service, not any errors produced by endpoint middlewares.
	UseLoggingMiddleware bool
}

type FirestoreConfig struct {
	Client *firestore.Client
}

func NewComponent(config Config) (*Component, error) {
	var ds domain.LinkDataStore
	if config.UseInmemDataStore {
		ds = inmem.NewInmemLinkDataStore()
	} else if config.FirestoreConfig != nil {
		if config.FirestoreConfig.Client == nil {
			return nil, errors.New(nil, "links", errors.InvalidArgument).WithInternalMessage("invalid firestore config, client is nil")
		}

		ds = fs.NewFirestoreLinkDataStore(config.FirestoreConfig.Client)
	} else {
		return nil, errors.New(nil, "links", errors.InvalidArgument).WithInternalMessage("no datastore configured")
	}

	if config.AuthorizationStore == nil {
		return nil, errors.New(nil, "links", errors.InvalidArgument).WithInternalMessage("no authorization store provided")
	}

	applicationService := application.NewLinkApplicationService(ds, config.AuthorizationStore)

	mwBuilder := mwBuilder{config: config}
	endpoints := transport.NewEndpoints(applicationService, transport.Middlewares{
		CreateLinkEndpoint: mwBuilder.buildMiddlewares("createLink"),
		DeleteLinkEndpoint: mwBuilder.buildMiddlewares("deleteLink"),
		RateLinkEndpoint:   mwBuilder.buildMiddlewares("rateLink"),
		LinkEndpoint:       mwBuilder.buildMiddlewares("getLink"),
		LinksEndpoint:      mwBuilder.buildMiddlewares("getLinks"),
	})

	return &Component{
		ApplicationService: applicationService,
		DataStore:          ds,
		Endpoints:          endpoints,
	}, nil
}

func (c *Component) RegisterHttpHandlers(router *mux.Router, httpOpts []http.ServerOption) {
	transport.RegisterHttpHandlers(c.Endpoints, router, httpOpts)
}

type mwBuilder struct {
	config Config
}

func (b mwBuilder) buildMiddlewares(endpointName string) []endpoint.Middleware {
	var mws []endpoint.Middleware
	mws = append(mws, b.config.Middlewares...)
	if b.config.AuthMiddleware != nil {
		mws = append(mws, b.config.AuthMiddleware)
	}
	if b.config.UseLoggingMiddleware && b.config.Logger != nil {
		mws = append(mws, e.ErrorLoggingMiddleware(b.config.Logger.With("component", "links", "endpoint", endpointName)))
	}
	return mws
}
