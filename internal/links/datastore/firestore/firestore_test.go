package firestore

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/dkinzler/linkboards/internal/links/datastore"

	"github.com/dkinzler/kit/firebase"
	"github.com/dkinzler/kit/firebase/emulator"

	"cloud.google.com/go/firestore"
	"github.com/stretchr/testify/assert"
)

// These tests require a running firestore emulator.
func TestMain(m *testing.M) {
	// Skip these tests if the environment variable is not set.
	if pid := os.Getenv("FIREBASE_PROJECT_ID"); pid == "" {
		log.Println("set FIREBASE_PROJECT_ID to run these tests")
		os.Exit(0)
	}

	exitCode := m.Run()
	os.Exit(exitCode)
}

// Creates a firestore emulator client that can be used to reset the emulator after each test and an instance of firestore.Client to implement the tests.
func initTest(t *testing.T) (*firestore.Client, error) {
	firestoreEmulatorClient, err := emulator.NewFirestoreEmulatorClient()
	if err != nil {
		t.Fatalf("could not create firestore emulator client: %v", err)
	}

	app, err := firebase.NewApp(firebase.Config{UseEmulators: true})
	if err != nil {
		t.Fatalf("could not create firebase app: %v", err)
	}
	ctx, cancel := getContext()
	defer cancel()
	firestore, err := app.Firestore(ctx)
	if err != nil {
		t.Fatalf("could not create firestore client: %v", err)
	}

	t.Cleanup(func() {
		err := firestoreEmulatorClient.ResetEmulator()
		if err != nil {
			t.Logf("could not reset firestore emulator: %v", err)
		}
	})

	return firestore, nil
}

func TestFirestoreBoardDataStore(t *testing.T) {
	a := assert.New(t)

	client, err := initTest(t)
	a.Nil(err)

	ds := NewFirestoreLinkDataStore(client)
	datastore.DatastoreTest(ds, t)
}

func getContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 2*time.Second)
}
