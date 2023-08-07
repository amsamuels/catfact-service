package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Server struct {
	client *mongo.Client
}

func NewServer(client *mongo.Client) *Server {
	return &Server{
		client: client,
	}
}

func (s *Server) handlergetFact(w http.ResponseWriter, r *http.Request) {
	coll := s.client.Database("catfact").Collection("fact")

	query := bson.M{}
	cursor, err := coll.Find(context.TODO(), query)
	if err != nil {
		log.Fatal(err)
	}

	results := []bson.M{}
	if err = cursor.All(context.TODO(), &results); err != nil {
		log.Fatal(err)
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)

}

func (s *Server) handlerPostFact(w http.ResponseWriter, r *http.Request) {
	coll := s.client.Database("catfact").Collection("fact")

	fmt.Println("Request Body: ", r.Body)
	var catFact bson.M
	err := json.NewDecoder(r.Body).Decode(&catFact)

	if err != nil {
		log.Fatal(err)
	}

	_, err = coll.InsertOne(context.TODO(), catFact)
	if err != nil {
		log.Fatal(err)
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(catFact)

}

type CatFactWorker struct {
	client *mongo.Client
}

func NewCatFactWorker(c *mongo.Client) *CatFactWorker {
	return &CatFactWorker{
		client: c,
	}
}

func (cfw *CatFactWorker) start() error {
	coll := cfw.client.Database("catfact").Collection("fact")
	ticker := time.NewTicker(2 * time.Second)

	i := 0
	for i < 10 {
		resp, err := http.Get("https://catfact.ninja/fact")
		if err != nil {
			return fmt.Errorf("error getting cat fact: %v", err)
		}
		defer resp.Body.Close()

		var catFact bson.M
		err = json.NewDecoder(resp.Body).Decode(&catFact)
		fmt.Println("Cat Fact: ", catFact)
		if err != nil {
			return fmt.Errorf("error decoding cat fact: %v", err)
		}
		_, err = coll.InsertOne(context.TODO(), catFact)
		if err != nil {
			return fmt.Errorf("error inserting cat fact: %v", err)
		}
		fmt.Println("Inserted document: ", catFact)
		fmt.Println("Iteration: ", i)
		<-ticker.C
		i++
	}

	return nil
}

func main() {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:2717"))
	if err != nil {
		panic(err)
	}

	worker := NewCatFactWorker(client)
	fmt.Println("Starting worker", worker)
	go worker.start()

	server := NewServer(client)
	http.HandleFunc("/facts", server.handlergetFact)
	http.HandleFunc("/add-fact", server.handlerPostFact)
	http.ListenAndServe(":3000", nil)
}
