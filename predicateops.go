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

// POST /dstestapi/predicates handler
func createPredicate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var predicate TestCasePredicate

	// we decode our body request params
	_ = json.NewDecoder(r.Body).Decode(&predicate)

	// insert our book model.
	result, err := predicateCollection.InsertOne(context.TODO(), predicate)

	if err != nil {
		GetError(err, w)
		return
	}

	json.NewEncoder(w).Encode(result)
}

// GET /dstestapi/predicates handler
func getPredicates(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// we created Book array
	var predicates []TestCasePredicate

	// bson.M{},  we passed empty filter. So we want to get all data.
	cur, err := predicateCollection.Find(context.TODO(), bson.M{})

	if err != nil {
		GetError(err, w)
		return
	}

	// Close the cursor once finished
	/*A defer statement defers the execution of a function until the surrounding function returns.
	simply, run cur.Close() process but after cur.Next() finished.*/
	defer cur.Close(context.TODO())

	for cur.Next(context.TODO()) {

		// create a value into which the single document can be decoded
		var predicate TestCasePredicate

		// & character returns the memory address of the following variable.
		err := cur.Decode(&predicate) // decode similar to deserialize process.
		if err != nil {
			log.Fatal(err)
		}

		// add item our array
		predicates = append(predicates, predicate)
	}

	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	json.NewEncoder(w).Encode(predicates) // encode similar to serialize process.
}

// GET /dstestapi/predicate/{id} handler
func getPredicate(w http.ResponseWriter, r *http.Request) {
	// set header.
	w.Header().Set("Content-Type", "application/json")

	var predicate TestCasePredicate

	// we get params with mux.
	var params = mux.Vars(r)

	// string to primitive.ObjectID
	id, _ := primitive.ObjectIDFromHex(params["id"])

	// We create filter. If it is unnecessary to sort data for you, you can use bson.M{}
	filter := bson.M{"_id": id}
	err := predicateCollection.FindOne(context.TODO(), filter).Decode(&predicate)

	if err != nil {
		//TODO assumption here is that the error is `not found`
		w.WriteHeader(http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(predicate)
}

// DELETE /dstestapi/predicate/{id} handler
func deletePredicate(w http.ResponseWriter, r *http.Request) {
	// Set header
	w.Header().Set("Content-Type", "application/json")

	// get params
	var params = mux.Vars(r)

	// string to primitve.ObjectID
	id, err := primitive.ObjectIDFromHex(params["id"])

	// prepare filter.
	filter := bson.M{"_id": id}

	deleteResult, err := predicateCollection.DeleteOne(context.TODO(), filter)

	if err != nil {
		GetError(err, w)
		return
	}

	json.NewEncoder(w).Encode(deleteResult)
}

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
