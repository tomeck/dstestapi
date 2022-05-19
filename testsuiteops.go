package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

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

	// Return the insertedId as the Id for this newly created testsuite
	testsuite.Id = result.InsertedID.(primitive.ObjectID)

	json.NewEncoder(w).Encode(testsuite)
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
	defer cur.Close(context.TODO())

	for cur.Next(context.TODO()) {

		// create a value into which the single document can be decoded
		var testSuite TestSuite

		err := cur.Decode(&testSuite) // decode current document into testsuite instance
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

// GET /dstestapi/testsuite/{id}/summary handler
func getTestSuiteSummary(w http.ResponseWriter, r *http.Request) {
	// set header.
	w.Header().Set("Content-Type", "application/json")

	var testsuite TestSuite
	var sb strings.Builder

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

	fmt.Fprintf(&sb, "Test Suite Name: %s\n\n", testsuite.Name)

	// Pretty format all the test cases
	for _, testCase := range testsuite.TestCases {
		sb.WriteString(prettyFormatTestCase((*testCase)))
		sb.WriteString("\n")
	}

	json.NewEncoder(w).Encode(sb.String())
}

/*
	Name: Test Case 1
	URL: /ch/payments/v1/charges
	Expected Status Code: 201
	Criteria:
	amount.total = 300 AND
	source.sourceType = "PaymentTrack" AND
	transactionDetails.captureFlag = True
*/
func prettyFormatTestCase(testCase TestCase) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "Test Case Name: %s\n", testCase.Name)
	fmt.Fprintf(&sb, "URL: %s\n", testCase.Url)
	fmt.Fprintf(&sb, "Expected Status Code: %d\n", testCase.ExpectedStatus)
	fmt.Fprintf(&sb, "Criteria:\n")

	// Pretty format all the predicates
	for index, predicate := range testCase.Predicates {
		fmt.Fprintf(&sb, "\t%s == %s", predicate.Attribute, predicate.ExpectedValue)

		if index < len(testCase.Predicates)-1 {
			fmt.Fprintf(&sb, " AND\n")
		} else {
			fmt.Fprintf(&sb, "\n")
		}
	}

	return sb.String()
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
