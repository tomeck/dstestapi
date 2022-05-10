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

// POST /dstestapi/testsuites handler
func createTestSuite(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var testsuite TestSuite

	// we decode our body request params
	_ = json.NewDecoder(r.Body).Decode(&testsuite)

	// Load the test case array for this test
	testsuite, err := loadTestCases(testsuite, context.TODO())

	// insert our book model.
	result, err := testsuiteCollection.InsertOne(context.TODO(), testsuite)

	if err != nil {
		GetError(err, w)
		return
	}

	json.NewEncoder(w).Encode(result)
}

// GET /dstestapi/testsuites handler
func getTestSuites(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Will store all the TestSuite elements we find
	var testSuites []TestSuite

	// TODO JTE ***** use Mongo Aggregation instead of manual lookups of child objects

	// bson.M{},  we passed empty filter. So we want to get all data.
	cur, err := testsuiteCollection.Find(context.TODO(), bson.M{})

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
		var testSuite TestSuite

		// & character returns the memory address of the following variable.
		err := cur.Decode(&testSuite) // decode similar to deserialize process.
		if err != nil {
			log.Fatal(err)
		}

		// Load the testCase array for this test suite
		testSuite, err = loadTestCases(testSuite, context.TODO())

		// add item our array
		testSuites = append(testSuites, testSuite)
	}

	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	json.NewEncoder(w).Encode(testSuites) // encode similar to serialize process.
}

// GET /dstestapi/testsuite/{id} handler
func getTestSuite(w http.ResponseWriter, r *http.Request) {
	// set header.
	w.Header().Set("Content-Type", "application/json")

	var testsuite TestSuite

	// we get params with mux.
	var params = mux.Vars(r)

	// string to primitive.ObjectID
	id, _ := primitive.ObjectIDFromHex(params["id"])

	// We create filter. If it is unnecessary to sort data for you, you can use bson.M{}
	filter := bson.M{"_id": id}
	err := testsuiteCollection.FindOne(context.TODO(), filter).Decode(&testsuite)

	if err != nil {
		//TODO assumption here is that the error is `not found`
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Load the test case array for this test
	testsuite, err = loadTestCases(testsuite, context.TODO())

	json.NewEncoder(w).Encode(testsuite)
}

// DELETE /dstestapi/testsuite/{id} handler
func deleteTestSuite(w http.ResponseWriter, r *http.Request) {
	// Set header
	w.Header().Set("Content-Type", "application/json")

	// get params
	var params = mux.Vars(r)

	// string to primitve.ObjectID
	id, err := primitive.ObjectIDFromHex(params["id"])

	// prepare filter.
	filter := bson.M{"_id": id}

	deleteResult, err := testsuiteCollection.DeleteOne(context.TODO(), filter)

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
