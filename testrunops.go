package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// POST /dstestapi/testruns handler
func initTestRun(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var testrun TestRun

	// we decode our body request params
	_ = json.NewDecoder(r.Body).Decode(&testrun)

	// Set timestamp and status
	testrun.Timestamp = time.Now()
	testrun.Status = Created

	// insert our object
	result, err := testrunCollection.InsertOne(ctx, testrun)

	if err != nil {
		GetError(err, w)
		return
	}

	// Return the insertedId as the Id for this newly created testrun
	testrun.Id = result.InsertedID.(primitive.ObjectID)

	json.NewEncoder(w).Encode(testrun)
}

// GET /dstestapi/testruns handler
func getTestRuns(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// The array of TestRuns
	var testruns []TestRun

	// Check whether params were passed; set filter accordingly
	filter := bson.D{}
	apiKey := r.URL.Query().Get("apikey")
	if apiKey != "" {
		filter = bson.D{{"apikey", bson.D{{"$eq", apiKey}}}}
	}

	cur, err := testrunCollection.Find(ctx, filter)

	if err != nil {
		GetError(err, w)
		return
	}

	// Close the cursor once finished
	defer cur.Close(ctx)

	for cur.Next(ctx) {

		// create a value into which the single document can be decoded
		var testrun TestRun

		// & character returns the memory address of the following variable.
		err := cur.Decode(&testrun) // decode similar to deserialize process.
		if err != nil {
			log.Fatal(err)
		}

		// Load the Test Suite for this Test Run
		testrun, err = loadTestSuite(testrun, ctx)

		// add item our array
		testruns = append(testruns, testrun)
	}

	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	json.NewEncoder(w).Encode(testruns) // encode similar to serialize process.
}

// GET /dstestapi/testruns/{id} handler
func getTestRun(w http.ResponseWriter, r *http.Request) {
	// set header.
	w.Header().Set("Content-Type", "application/json")

	var testrun TestRun

	// we get params with mux.
	var params = mux.Vars(r)

	// string to primitive.ObjectID
	id, _ := primitive.ObjectIDFromHex(params["id"])

	// We create filter. If it is unnecessary to sort data for you, you can use bson.M{}
	filter := bson.M{"_id": id}
	err := testrunCollection.FindOne(ctx, filter).Decode(&testrun)

	if err != nil {
		//TODO assumption here is that the error is `not found`
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Load the Test Suite for this Test Run
	testrun, err = loadTestSuite(testrun, ctx)

	if err != nil {
		//TODO assumption here is that the error is `not found`
		w.WriteHeader(http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(testrun)
}

// DELETE /dstestapi/testruns/{id} handler
func deleteTestRun(w http.ResponseWriter, r *http.Request) {
	// Set header
	w.Header().Set("Content-Type", "application/json")

	// get params
	var params = mux.Vars(r)

	// string to primitve.ObjectID
	id, err := primitive.ObjectIDFromHex(params["id"])

	// prepare filter.
	filter := bson.M{"_id": id}

	deleteResult, err := testrunCollection.DeleteOne(ctx, filter)

	if err != nil {
		GetError(err, w)
		return
	}

	json.NewEncoder(w).Encode(deleteResult)
}

// POST /dstestapi/testruns/:id/stop handler
func stopTestRun(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// get params
	var params = mux.Vars(r)

	testRun, err := collectTestRun(params["id"])

	if err != nil {
		GetError(err, w)
		return
	}

	json.NewEncoder(w).Encode(testRun)
}
