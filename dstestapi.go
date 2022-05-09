package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
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
	// r.HandleFunc("/dstestapi/predicates/{id}", updatePredicate).Methods("PUT")

	// /destestapi/testsuites
	r.HandleFunc("/dstestapi/testsuites", createTestSuite).Methods("POST")
	r.HandleFunc("/dstestapi/testsuites", getTestSuites).Methods("GET")
	r.HandleFunc("/dstestapi/testsuites/{id}", getTestSuite).Methods("GET")
	r.HandleFunc("/dstestapi/testsuites/{id}", deleteTestSuite).Methods("DELETE")
	// r.HandleFunc("/dstestapi/predicates/{id}", updatePredicate).Methods("PUT")

	// Initialize database (hardcoded for local machine)
	client, ctx, cancel, err := connect("mongodb://localhost:27017")
	if err != nil {
		panic(err)
	}

	// Close db when the main function is returned.
	defer close(client, ctx, cancel)
	fmt.Println("Connected to local mongodb")

	// Get target database and collection
	db = client.Database("dstest")
	txCollection = db.Collection("transactions")
	predicateCollection = db.Collection("predicates")
	testcaseCollection = db.Collection("testcases")
	testsuiteCollection = db.Collection("testsuites")

	fmt.Println("Initialized db and collections")

	// set our port address
	log.Fatal(http.ListenAndServe(":8000", r))
}
