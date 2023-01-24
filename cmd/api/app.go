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

	"github.com/d39b/linkboards/internal/auth/middleware"
	"github.com/d39b/linkboards/internal/auth/store"
	"github.com/d39b/linkboards/internal/boards"
	"github.com/d39b/linkboards/internal/links"

	lfb "github.com/d39b/kit/firebase"

	kitjwt "github.com/go-kit/kit/auth/jwt"
	"github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"

	dhttp "github.com/d39b/kit/transport/http"

	"github.com/d39b/kit/log"
)

// Configures the application.
//
// Which dependencies are used and how they are created is determined in-order as follows:
//   - If UseInmemDependencies is true, use local in-memory dependencies.
//   - If UseFirebaseEmulators is true, attempt to use firebase emulators.
//   - If FirebaseServiceAccountFile is not empty, attempt to use it to connect to firebase services,
//     Firebase Authentication for authentication and Firestore for any data stores.
//   - Attempt to use application default credentials to connect to firebase services.
//     These are usually set automatically if the application is run in a google cloud product like Cloud Run or App Engine.
type Config struct {
	// Port the API will be reachable at.
	Port int
	// Address the API will be reachable at, usually empty to be reachable
	// on all network interfaces of the system.
	Address string

	// If true, use in memory dependencies like data stores and authentication mechanism.
	UseInmemDependencies bool

	// If true, attempt to connect to firebase emulators to use for authentication and data stores.
	// Note that the emulators will have to be running already.
	UseFirebaseEmulators bool
	FirebaseProjectId    string
	// If not empty, try to connect to firebase services using the given service account file.
	FirebaseServiceAccountFile string

	// In debug mode:
	//   - log messages with level Debug will be output
	//   - log messages will be pretty printed, i.e. as JSON with multiple indented lines
	DebugMode bool
}

// Runs the application using the given config.
// Will create a boards and links component and expose their endpoints using a http server.
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

	if !config.UseInmemDependencies {
		_, fbAuthClient, fbFirestoreClient, err = initFirebase(config)
		if err != nil {
			logger.Log("message", "could not init firebase", "error", err)
			os.Exit(1)
		}
	}

	// endpoint authentication middleware
	var authMiddleware endpoint.Middleware
	if config.UseInmemDependencies {
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
	if config.UseInmemDependencies {
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
	if config.UseInmemDependencies {
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
		dhttp.EncodeError(ctx, err, w)
	}

	var beforeFunc kithttp.RequestFunc
	if config.UseInmemDependencies {
		beforeFunc = kithttp.PopulateRequestContext
	} else {
		beforeFunc = kitjwt.HTTPToContext()
	}

	opts := []kithttp.ServerOption{
		kithttp.ServerBefore(beforeFunc),
		kithttp.ServerErrorEncoder(httpErrorEncoder),
		kithttp.ServerErrorHandler(dhttp.NewLogErrorHandler(logger)),
	}

	boardComponent.RegisterHttpHandlers(router, opts)
	linksComponent.RegisterHttpHandlers(router, opts)

	err = dhttp.RunDefaultServer(
		router,
		nil,
		dhttp.NewServerConfig().
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

	if err != nil {
		logger.Error().Log("msg", "error running http server", "error", err)
	}

	return err
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
