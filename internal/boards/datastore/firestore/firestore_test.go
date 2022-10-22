package firestore

import (
	"context"
	"linkboards/internal/boards/datastore"
	"os"
	"testing"
	"time"

	"github.com/d39b/kit/firebase"
	"github.com/d39b/kit/firebase/emulator"

	"cloud.google.com/go/firestore"
	"github.com/stretchr/testify/assert"
)

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

func TestMain(m *testing.M) {
	exitCode := m.Run()
	os.Exit(exitCode)
}

// This test requires a running firestore emulator.
func TestFirestoreBoardDataStore(t *testing.T) {
	// Skip this test if the environment variable is not set.
	if pid := os.Getenv("FIREBASE_PROJECT_ID"); pid == "" {
		t.Skip("set FIREBASE_PROJECT_ID to run these tests")
	}

	a := assert.New(t)

	client, err := initTest(t)
	a.Nil(err)

	ds := NewFirestoreBoardDataStore(client)
	datastore.DatastoreTest(ds, t)
}

func getContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 2*time.Second)
}
