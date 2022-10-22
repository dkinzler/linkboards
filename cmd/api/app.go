package main

import (
	"context"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"

	fbfirestore "cloud.google.com/go/firestore"
	fb "firebase.google.com/go/v4"
	fbauth "firebase.google.com/go/v4/auth"

	"linkboards/internal/auth/middleware"
	"linkboards/internal/auth/store"
	"linkboards/internal/boards"
	"linkboards/internal/links"

	lfb "github.com/d39b/kit/firebase"

	kitjwt "github.com/go-kit/kit/auth/jwt"
	"github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"

	transport "github.com/d39b/kit/transport/http"

	"github.com/d39b/kit/log"
)

const Version string = "0.1.6"

type Config struct {
	Port    int
	Address string

	UseInmemDatastores bool

	UseFirebaseEmulators       bool
	FirebaseProjectId          string
	FirebaseServiceAccountFile string

	DebugMode bool
}

func runApp(config Config) error {
	var options []log.Option
	if config.DebugMode {
		options = append(options, log.AllowDebug, log.PrettyPrint)
	}
	logger := log.DefaultJSONLogger(options...)

	logger.Info().Log("using config", config)

	// seed rng
	rand.Seed(time.Now().UnixNano())

	var fbAuthClient *fbauth.Client
	var fbFirestoreClient *fbfirestore.Client
	var err error

	if !config.UseInmemDatastores {
		_, fbAuthClient, fbFirestoreClient, err = initFirebase(config)
		if err != nil {
			logger.Log("message", "could not init firebase", "error", err)
			os.Exit(1)
		}
	}

	// endpoint authentication middleware
	var authMiddleware endpoint.Middleware
	if config.UseInmemDatastores {
		authMiddleware = middleware.NewFakeAuthEndpointMiddleware()
	} else {
		authMiddleware = middleware.NewFirebaseAuthEndpointMiddleware(fbAuthClient, false)
	}

	// create boards component
	boardsConfig := boards.Config{
		Logger:               logger,
		AuthMiddleware:       authMiddleware,
		UseLoggingMiddleware: true,
	}
	if config.UseInmemDatastores {
		boardsConfig.UseInmemDataStore = true
	} else {
		boardsConfig.FirestoreConfig = &boards.FirestoreConfig{
			Client: fbFirestoreClient,
		}
	}
	boardComponent, err := boards.NewComponent(boardsConfig)
	if err != nil {
		logger.Log("message", "could not create boards component", "error", err)
		os.Exit(1)
	}

	authorizationStore := store.NewDefaultAuthorizationStore(boardComponent.DataStore)

	// create links component
	linksConfig := links.Config{
		Logger:               logger,
		AuthMiddleware:       authMiddleware,
		UseLoggingMiddleware: true,
		AuthorizationStore:   authorizationStore,
	}
	if config.UseInmemDatastores {
		linksConfig.UseInmemDataStore = true
	} else {
		linksConfig.FirestoreConfig = &links.FirestoreConfig{
			Client: fbFirestoreClient,
		}
	}
	linksComponent, err := links.NewComponent(linksConfig)
	if err != nil {
		logger.Log("message", "could not create links component", "error", err)
		os.Exit(1)
	}

	router := mux.NewRouter()
	httpErrorEncoder := func(ctx context.Context, err error, w http.ResponseWriter) {
		transport.EncodeError(ctx, err, w)
	}

	var beforeFunc kithttp.RequestFunc
	if config.UseInmemDatastores {
		beforeFunc = kithttp.PopulateRequestContext
	} else {
		beforeFunc = kitjwt.HTTPToContext()
	}

	opts := []kithttp.ServerOption{
		kithttp.ServerBefore(beforeFunc),
		kithttp.ServerErrorEncoder(httpErrorEncoder),
		kithttp.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
	}

	boardComponent.RegisterHttpHandlers(router, opts)
	linksComponent.RegisterHttpHandlers(router, opts)

	/*
		//CORS stuff

		router.Use(mux.CORSMethodMiddleware(router))
		//TODO we might not need all these headers for every request, but is there a better way? maybe write different complete middlewares for different cases and then add them to handlers like the auth middleware
		corsCred := handlers.AllowCredentials()
		//TODO change this to appropriate domains to prevent cross site scripting
		corsOrigins := handlers.AllowedOrigins([]string{"*"})
		corsMethods := handlers.AllowedMethods([]string{"PATCH", "DELETE", "GET", "POST", "OPTIONS"})
		corsHeaders := handlers.AllowedHeaders([]string{"Content-Type", "Authorization"})
		router.Use(handlers.CORS(corsCred, corsOrigins, corsMethods, corsHeaders))
	*/

	transport.RunDefaultServer(
		router,
		nil,
		transport.NewServerConfig().
			WithAddress(config.Address).
			WithPort(config.Port).
			WithOnShutdownFunc(func(err error) {
				logger.Info().Log("msg", "shutting down...")
				if err != nil {
					logger.Info().Log("msg", "error shutting down", "error", err)
				}
			}).
			WithOnPanicFunc(func(i interface{}) {
				logger.Warn().Log("msg", "caught panic in http handler", "error", i)
			}),
	)
	return nil
}

func initFirebase(config Config) (*fb.App, *fbauth.Client, *fbfirestore.Client, error) {
	fbApp, err := lfb.NewApp(lfb.Config{
		UseEmulators:       config.UseFirebaseEmulators,
		ProjectId:          config.FirebaseProjectId,
		ServiceAccountFile: config.FirebaseServiceAccountFile,
	})
	if err != nil {
		return nil, nil, nil, err
	}

	fbAuth, err := lfb.NewAuthClient(fbApp)
	if err != nil {
		return nil, nil, nil, err
	}

	fbFirestore, err := lfb.NewFirestoreClient(fbApp)
	if err != nil {
		return nil, nil, nil, err
	}

	return fbApp, fbAuth, fbFirestore, nil
}
