package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// POST /dstestapi/testcases handler
func createTestCase(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var testcase TestCase

	// we decode our body request params
	_ = json.NewDecoder(r.Body).Decode(&testcase)

	// insert our book model.
	result, err := testcaseCollection.InsertOne(context.TODO(), testcase)

	if err != nil {
		GetError(err, w)
		return
	}

	json.NewEncoder(w).Encode(result)
}

// GET /dstestapi/testcases handler
func getTestCases(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Will store all the TestCase elements we find
	var testCases []TestCase

	// TODO JTE ***** use Mongo Aggretation instead of manual lookups of child objects

	// bson.M{},  we passed empty filter. So we want to get all data.
	cur, err := testcaseCollection.Find(context.TODO(), bson.M{})

	if err != nil {
		GetError(err, w)
		return
	}

	// Close the cursor once finished
	// A defer statement defers the execution of a function until the surrounding function returns.
	// simply, run cur.Close() process but after cur.Next() finished.
	defer cur.Close(context.TODO())

	for cur.Next(context.TODO()) {

		// create a value into which the single document can be decoded
		var testCase TestCase

		// & character returns the memory address of the following variable.
		err := cur.Decode(&testCase) // decode similar to deserialize process.
		if err != nil {
			log.Fatal(err)
		}

		// Load the predicates array for this test case
		testCase, err = loadPredicates(testCase, context.TODO())

		// add item our array
		testCases = append(testCases, testCase)
	}

	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	json.NewEncoder(w).Encode(testCases) // encode similar to serialize process.
}

// GET /dstestapi/testcase/{id} handler
func getTestCase(w http.ResponseWriter, r *http.Request) {
	// set header.
	w.Header().Set("Content-Type", "application/json")

	var testcase TestCase

	// we get params with mux.
	var params = mux.Vars(r)

	// string to primitive.ObjectID
	id, _ := primitive.ObjectIDFromHex(params["id"])

	// We create filter. If it is unnecessary to sort data for you, you can use bson.M{}
	filter := bson.M{"_id": id}
	err := testcaseCollection.FindOne(context.TODO(), filter).Decode(&testcase)

	if err != nil {
		//TODO assumption here is that the error is `not found`
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Load the predicates array for this test case
	testcase, err = loadPredicates(testcase, context.TODO())

	json.NewEncoder(w).Encode(testcase)
}

// DELETE /dstestapi/testcase/{id} handler
func deleteTestCase(w http.ResponseWriter, r *http.Request) {
	// Set header
	w.Header().Set("Content-Type", "application/json")

	// get params
	var params = mux.Vars(r)

	// string to primitve.ObjectID
	id, err := primitive.ObjectIDFromHex(params["id"])

	// prepare filter.
	filter := bson.M{"_id": id}

	deleteResult, err := testcaseCollection.DeleteOne(context.TODO(), filter)

	if err != nil {
		GetError(err, w)
		return
	}

	json.NewEncoder(w).Encode(deleteResult)
}

/*
// PUT /dstestapi/predicate/{id} handler
func updatePredicate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var params = mux.Vars(r)

	//Get id from parameters
	id, _ := primitive.ObjectIDFromHex(params["id"])

	var predicate TestCasePredicate

	// Create filter
	filter := bson.M{"_id": id}

	// Read update model from body request
	_ = json.NewDecoder(r.Body).Decode(&predicate)

	// prepare update model.
	update := bson.D{
		{"$set", bson.D{
			{"attribute", predicate.Attribute},
			{"expected_value", predicate.ExpectedValue},
		}},
	}

	// Set options so that the updated object is returned
	// upsert := true
	// after := options.After
	// opt := options.FindOneAndUpdateOptions{
	// 	ReturnDocument: &after,
	// 	Upsert:         &upsert,
	// }

	predicateCollection.FindOneAndUpdate(context.TODO(), filter, update) //.Decode(&predicate)

	// TODO JTE handle errors
	// if err != nil {
	// 	GetError(err, w)
	// 	return
	// }

	predicate.Id = id

	json.NewEncoder(w).Encode(predicate)
}
*/