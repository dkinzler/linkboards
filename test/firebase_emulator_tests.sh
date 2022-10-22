#Run all tests that require firebase emulators

export FIREBASE_PROJECT_ID="demo-project"
export FIREBASE_AUTH_EMULATOR_HOST="localhost:9099"
export FIRESTORE_EMULATOR_HOST="localhost:8080"

firebase -P "demo-project" -c "../tools/firebase_emulator/firebase.json" emulators:exec "\
	go test -v -count=1 ../internal/pkg/firebase/emulator && \
	go test -v -count=1 ../internal/pkg/firebase/auth && \
	go test -v -count=1 ../internal/pkg/firebase/firestore && \
	go test -v -count=1 ../internal/links/datastore/firestore && \
	go test -v -count=1 ../internal/boards/datastore/firestore
"
