package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/tidwall/gjson"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/ucarion/urlpath"
)

// Returns transactions for a given test run, ***in descending chronological order***
func findTransactionsForTestRun(testRun TestRun) ([]Transaction, error) {

	// JTE TODO filter also on ApiKey

	// Sort by `Timestamp` field descending
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"_id", -1}})

	filter := bson.D{{"testrunid", bson.D{{"$eq", testRun.TestRunHeaderId}}}}

	cursor, err := txCollection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}

	var transactions []Transaction
	if err = cursor.All(ctx, &transactions); err != nil {
		return nil, err
	}

	return transactions, nil
}

// Check whether all the specified predicates match the given transaction
func validatePredicatesForTransaction(predicates []*TestCasePredicate, transaction Transaction) bool {

	// Iterate through all predicates in this test case
	allMatched := true
	for _, predicate := range predicates {
		if !matchPredicate(transaction.Request, predicate.Attribute, predicate.ExpectedValue) {
			allMatched = false
			break
		}
	}

	return allMatched
}

func urlMatches(url string, urlPattern string) bool {

	// Check for exact match
	if url == urlPattern {
		return true
	}

	path := urlpath.New(urlPattern)
	_, ok := path.Match(url)

	return ok
}

// Find the most recent transaction that matches this test case
func matchTransactionToTestCase(testCase TestCase, transactions []Transaction) (Transaction, TestStatus, error) {

	var err error
	var matchingTransaction Transaction
	testStatus := NotAttempted

	for _, transaction := range transactions {

		// Check whether URL matches
		// if transaction.Url != testCase.Url {
		if !urlMatches(transaction.Url, testCase.Url) {
			continue // nope - this is not the transaction we want
		}

		// So far, so good - now attempt to match all predicates
		if validatePredicatesForTransaction(testCase.Predicates, transaction) {

			if transaction.Status == testCase.ExpectedStatus {
				testStatus = Success
			} else {
				testStatus = Failure
			}

			// This transaction matches all checks, so don't need to keep looking
			matchingTransaction = transaction

			// Actually, break only if this transaction was successful, otherwise keep looking for successful match
			if testStatus == Success {
				break
			}
		}
	}

	// If test status is still undefined, then we did not find a matching transaction
	// if testStatus == UndefinedTestStatus {
	// 	err = errors.New("Could not find transaction to match test case " + testCase.Name)
	// }

	return matchingTransaction, testStatus, err
}

// Match transactions to each test case in this test run instance
func matchTransactionsToTestRun(testRun *TestRun, transactions []Transaction) error {

	testRunStatus := Complete

	// Zero out previous matching results
	testRun.TestResults = nil

	// For each test case in the test suite associated with this run...
	for _, testCase := range testRun.TestSuite.TestCases {

		fmt.Println("Searching for transactions to match test case", testCase.Name)
		transaction, testStatus, err := matchTransactionToTestCase(*testCase, transactions)

		if err == nil {
			fmt.Println("Found matching transaction")
			testResult := TestResult{ /*TestRun: testRun, */ TestCase: testCase, Status: testStatus, Transaction: &transaction, Timestamp: time.Now()}
			testRun.TestResults = append(testRun.TestResults, &testResult)
		} else {
			// Did not find a transaction to match this test case
			fmt.Println("Did not find matching transaction")
			testResult := TestResult{ /*TestRun: testRun, */ TestCase: testCase, Status: testStatus, Transaction: &transaction, Timestamp: time.Now()}
			testRun.TestResults = append(testRun.TestResults, &testResult)
			testRunStatus = InProgress
		}
	}

	testRun.Status = testRunStatus

	return nil
}

func cleanse(input string) string {
	return strings.Trim(input, "\"")
}

func matchPredicate(requestString string, predicate string, expectedValue string) bool {

	value := gjson.Get(requestString, predicate)

	if value.Raw == "" {
		return false //, errors.New("Predicate not found")
	}

	if strings.ToLower(cleanse(value.Raw)) == strings.ToLower(expectedValue) {
		return true //, nil
	} else {
		return false //, errors.New("Expected value not matched")
	}
	// Should never get here
}

// JTE WARNING: This is super inefficient
func getTestResultforTestCase(testCase TestCase, testRun TestRun) (TestResult, error) {

	// Find the test result matching the specified Test Case
	for _, testResult := range testRun.TestResults {
		if testResult.TestCase.Id == testCase.Id {
			return *testResult, nil
		}

	}

	return TestResult{}, errors.New("Could not find a test result for test case " + testCase.Name)
}

func persistTestRun(testRun TestRun) error {

	// TODO JTE is there a better way to update this document in place?
	// What's being updated is the array of TestResults and maybe the status

	_, err := testrunCollection.DeleteOne(ctx, bson.M{"_id": testRun.Id})
	if err != nil {
		return err
	}

	// TODO JTE check whether the test run is complete
	// if len(testRun.TestResults) == len(testRun.TestSuite.TestCases) {
	// 	testRun.Status = Complete
	// } else {
	// 	testRun.Status = InProgress
	// }

	// TODO JTE Warning: setting the status of the run to Complete always
	testRun.Status = Complete

	_, err = testrunCollection.InsertOne(ctx, testRun)
	if err != nil {
		return err
	}

	return nil
}

func fetchTestRun(testRunId string) (TestRun, error) {

	// The return value
	var testRun TestRun

	// Get specified test run from database

	// string to primitive.ObjectID
	id, _ := primitive.ObjectIDFromHex(testRunId)

	// We create filter. If it is unnecessary to sort data for you, you can use bson.M{}
	filter := bson.M{"_id": id}
	err := testrunCollection.FindOne(ctx, filter).Decode(&testRun)

	if err != nil {
		return TestRun{}, err
	}

	// Load the Test Suite for this Test Run
	testRun, err = loadTestSuite(testRun, ctx)

	return testRun, nil
}

func collectTestRun(testRunId string) (TestRun, error) {

	// Load the test run with the specified id from db
	testRun, err := fetchTestRun(testRunId)
	if err != nil {
		return TestRun{}, err
	}

	// Find all the transactions submitted against this test run
	transactions, err := findTransactionsForTestRun(testRun)
	if err != nil {
		return TestRun{}, err
	}

	// Match transactions to test cases; update TestRun instance
	err = matchTransactionsToTestRun(&testRun, transactions)
	if err != nil {
		return TestRun{}, err
	}

	// Persist the test run instance to db
	err = persistTestRun(testRun)

	return testRun, err
}

func compileTestRunReport(testRun TestRun) (TestRunReport, error) {

	testRunReport := TestRunReport{TestSuite: testRun.TestSuite, TestRun: &testRun, Status: testRun.Status}
	testRunReport.TestCaseReports = make([]TestCaseReport, len(testRun.TestSuite.TestCases))

	numSuccess := 0
	numAttempted := 0

	for i, testCase := range testRun.TestSuite.TestCases {
		testResult, _ := getTestResultforTestCase(*testCase, testRun)
		testStatus := NotAttempted
		if testResult.Status == Success || testResult.Status == Failure {
			numAttempted += 1
			if testResult.Status == Success {
				numSuccess += 1
				testStatus = Success
			} else {
				testStatus = Failure
			}
		} //else {
		// 	testStatus = NotAttempted
		// }

		testCaseReport := TestCaseReport{TestCase: testCase, Status: testStatus}
		//testRunReport.TestCaseReports = append(testRunReport.TestCaseReports, testCaseReport)
		testRunReport.TestCaseReports[i] = testCaseReport
	}

	testRunReport.NumTestCases = len(testRun.TestSuite.TestCases)
	testRunReport.NumTestsAttempted = numAttempted
	testRunReport.NumTestsPassed = numSuccess

	return testRunReport, nil
}
