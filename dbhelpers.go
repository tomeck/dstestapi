package main

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Deep load of all test cases in the supplied test suite
func loadTestCases(testSuite TestSuite, ctx context.Context) (TestSuite, error) {

	var testCases []*TestCase

	for _, testCase := range testSuite.TestCases {

		testCase, err := getTestCaseById(testCase.Id, ctx)

		if err == nil {
			testCases = append(testCases, &testCase)
		} else {
			return testSuite, err
		}
	}

	// Replace the Predicates array on the testCase with the fully resolved array
	testSuite.TestCases = testCases
	return testSuite, nil
}

// Fetch the TestCase with the specified id
func getTestCaseById(id primitive.ObjectID, ctx context.Context) (TestCase, error) {
	// filter := bson.D{{"_id", bson.D{{"$eq", id}}}}

	var testCase TestCase
	filter := bson.M{"_id": id}
	err := testcaseCollection.FindOne(ctx, filter).Decode(&testCase)

	if err != nil {
		return testCase, err
	}

	// Now load the predicates for this testcase
	testCase, err = loadPredicates(testCase, ctx)

	if err != nil {
		return testCase, err
	} else {
		return testCase, nil
	}
}

// Deep load of all predicates in the supplied test case
func loadPredicates(testCase TestCase, ctx context.Context) (TestCase, error) {

	var predicates []*TestCasePredicate

	for _, predicate := range testCase.Predicates {

		predicate, err := getPredicateById(predicate.Id, ctx)

		if err == nil {
			predicates = append(predicates, &predicate)
		} else {
			return testCase, err
		}
	}

	// Replace the Predicates array on the testCase with the fully resolved array
	testCase.Predicates = predicates
	return testCase, nil
}

// Fetch the Predicate with the specified id
func getPredicateById(id primitive.ObjectID, ctx context.Context) (TestCasePredicate, error) {
	// filter := bson.D{{"_id", bson.D{{"$eq", id}}}}

	var predicate TestCasePredicate
	filter := bson.M{"_id": id}
	err := predicateCollection.FindOne(ctx, filter).Decode(&predicate)

	if err != nil {
		return predicate, err
	} else {
		return predicate, nil
	}
}
