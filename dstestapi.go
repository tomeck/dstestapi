package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Globals
// JTE TODO is there something better to do with these?
var client *mongo.Client
var ctx context.Context
var cancel context.CancelFunc
var db *mongo.Database
var txCollection *mongo.Collection
var predicateCollection *mongo.Collection
var testcaseCollection *mongo.Collection
var testsuiteCollection *mongo.Collection
var testrunCollection *mongo.Collection

// const DB_CONNECTION_STRING = "mongodb://localhost:27017"

const DB_CONNECTION_STRING = "mongodb+srv://admin:Ngokman3#@cluster0.mce8u.mongodb.net/dstest?retryWrites=true&w=majority"

// GET / handler
func healthCheck(w http.ResponseWriter, r *http.Request) {
	// set header.
	w.Header().Set("Content-Type", "application/json")
}

func main() {
	//Init Router
	r := mux.NewRouter()

	//
	// Set routes
	//
	// /destestapi/predicates
	r.HandleFunc("/dstestapi/predicates", createPredicate).Methods("POST")
	r.HandleFunc("/dstestapi/predicates", getPredicates).Methods("GET")
	r.HandleFunc("/dstestapi/predicates/{id}", getPredicate).Methods("GET")
	r.HandleFunc("/dstestapi/predicates/{id}", deletePredicate).Methods("DELETE")
	r.HandleFunc("/dstestapi/predicates/{id}", updatePredicate).Methods("PUT")

	// /destestapi/testcases
	r.HandleFunc("/dstestapi/testcases", createTestCase).Methods("POST")
	r.HandleFunc("/dstestapi/testcases", getTestCases).Methods("GET")
	r.HandleFunc("/dstestapi/testcases/{id}", getTestCase).Methods("GET")
	r.HandleFunc("/dstestapi/testcases/{id}", deleteTestCase).Methods("DELETE")

	// /destestapi/testsuites
	r.HandleFunc("/dstestapi/testsuites", createTestSuite).Methods("POST")
	r.HandleFunc("/dstestapi/testsuites", getTestSuites).Methods("GET")
	r.HandleFunc("/dstestapi/testsuites/{id}", getTestSuite).Methods("GET")
	r.HandleFunc("/dstestapi/testsuites/{id}/summary", getTestSuiteSummary).Methods("GET")
	r.HandleFunc("/dstestapi/testsuites/{id}", deleteTestSuite).Methods("DELETE")

	// /destestapi/testruns
	r.HandleFunc("/dstestapi/testruns", initTestRun).Methods("POST")
	r.HandleFunc("/dstestapi/testruns", getTestRuns).Methods("GET")
	r.HandleFunc("/dstestapi/testruns/{id}", getTestRun).Methods("GET")
	r.HandleFunc("/dstestapi/testruns/{id}/report", getTestRunReport).Methods("GET")
	r.HandleFunc("/dstestapi/testruns/{id}", deleteTestRun).Methods("DELETE")
	r.HandleFunc("/dstestapi/testruns/{id}/stop", stopTestRun).Methods("POST")

	// Health check endpoint
	r.HandleFunc("/", healthCheck).Methods("GET")

	// Initialize database connection
	// client, ctx, cancel, err := connect(DB_CONNECTION_STRING)
	// if err != nil {
	// 	panic(err)
	// }
	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().
		ApplyURI(DB_CONNECTION_STRING).
		SetServerAPIOptions(serverAPIOptions)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("x1Connected to mongo at", DB_CONNECTION_STRING)

	// Close db when the main function is returned.
	defer close(client, ctx, cancel)

	// Get target database and collection
	db = client.Database("dstest")
	txCollection = db.Collection("transactions")
	predicateCollection = db.Collection("predicates")
	testcaseCollection = db.Collection("testcases")
	testsuiteCollection = db.Collection("testsuites")
	testrunCollection = db.Collection("testruns")

	fmt.Println("Initialized db and collections")

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
		fmt.Printf("defaulting to port %s\n", port)
	}

	fmt.Println("Listening on port", port)

	// set our listen port address
	log.Fatal(http.ListenAndServe(":"+port, r))
}
