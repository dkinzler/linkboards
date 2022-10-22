package links

import (
	"go-sample/internal/auth"
	"go-sample/internal/links/application"
	fs "go-sample/internal/links/datastore/firestore"
	"go-sample/internal/links/datastore/inmem"
	"go-sample/internal/links/domain"
	"go-sample/internal/links/transport"

	e "github.com/d39b/kit/endpoint"
	"github.com/d39b/kit/errors"
	"github.com/d39b/kit/log"

	"cloud.google.com/go/firestore"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

type Component struct {
	ApplicationService application.LinkApplicationService
	DataStore          domain.LinkDataStore
	Endpoints          transport.EndpointSet
}

type Config struct {
	Logger *log.Logger

	UseInmemDataStore  bool
	FirestoreConfig    *FirestoreConfig
	AuthorizationStore auth.AuthorizationStore

	// Middlewares that should be applied to all endpoints
	Middlewares    []endpoint.Middleware
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
