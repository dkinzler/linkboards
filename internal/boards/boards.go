package boards

import (
	"github.com/d39b/linkboards/internal/auth"
	"github.com/d39b/linkboards/internal/auth/store"
	"github.com/d39b/linkboards/internal/boards/application"
	fs "github.com/d39b/linkboards/internal/boards/datastore/firestore"
	"github.com/d39b/linkboards/internal/boards/datastore/inmem"
	"github.com/d39b/linkboards/internal/boards/domain"
	"github.com/d39b/linkboards/internal/boards/transport"

	e "github.com/d39b/kit/endpoint"
	"github.com/d39b/kit/errors"
	"github.com/d39b/kit/log"

	"cloud.google.com/go/firestore"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

// Contains all the necessary elements for running the boards service, i.e. the
// boards part of the API.
type Component struct {
	ApplicationService application.BoardApplicationService
	DataStore          domain.BoardDataStore
	Endpoints          transport.EndpointSet
}

// Configures a boards component.
type Config struct {
	Logger *log.Logger

	// If true, use an in-memory data store.
	UseInmemDataStore bool
	// Used to create a firestore data store if UseInmemDataStore is set to false.
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
	var ds domain.BoardDataStore
	if config.UseInmemDataStore {
		ds = inmem.NewInmemBoardDataStore()
	} else if config.FirestoreConfig != nil {
		if config.FirestoreConfig.Client == nil {
			return nil, errors.New(nil, "boards", errors.InvalidArgument).WithInternalMessage("invalid firestore config, client is nil")
		}

		ds = fs.NewFirestoreBoardDataStore(config.FirestoreConfig.Client)
	} else {
		return nil, errors.New(nil, "boards", errors.InvalidArgument).WithInternalMessage("no datastore configured")
	}

	var as auth.AuthorizationStore
	if config.AuthorizationStore != nil {
		as = config.AuthorizationStore
	} else {
		as = store.NewDefaultAuthorizationStore(ds)
	}

	applicationService := application.NewBoardApplicationService(ds, as)

	mwBuilder := mwBuilder{config: config}
	endpoints := transport.NewEndpoints(applicationService, transport.Middlewares{
		CreateBoardEndpoint:     mwBuilder.buildMiddlewares("createBoard"),
		DeleteBoardEndpoint:     mwBuilder.buildMiddlewares("deleteBoard"),
		EditBoardEndpoint:       mwBuilder.buildMiddlewares("editBoard"),
		BoardEndpoint:           mwBuilder.buildMiddlewares("getBoard"),
		BoardsEndpoint:          mwBuilder.buildMiddlewares("getBoards"),
		CreateInviteEndpoint:    mwBuilder.buildMiddlewares("createInvite"),
		RespondToInviteEndpoint: mwBuilder.buildMiddlewares("respondToInvite"),
		DeleteInviteEndpoint:    mwBuilder.buildMiddlewares("deleteInvite"),
		InvitesEndpoint:         mwBuilder.buildMiddlewares("getInvites"),
		RemoveUserEndpoint:      mwBuilder.buildMiddlewares("removeUser"),
		EditBoardUserEndpoint:   mwBuilder.buildMiddlewares("editBoardUser"),
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
		mws = append(mws, e.ErrorLoggingMiddleware(b.config.Logger.With("component", "boards", "endpoint", endpointName)))
	}
	return mws
}
