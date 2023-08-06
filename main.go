package main

import (
	"context"
	"encoding/json"
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

type CatFactWorker struct {
	client *mongo.Client
}

func NewCatFactWorker(client *mongo.Client) *CatFactWorker {
	return &CatFactWorker{
		client: client,
	}
}

func (cfw *CatFactWorker) start() error {
	coll := cfw.client.Database("catfact").Collection("facts")
	ticker := time.NewTicker(2 * time.Second)

	i := 0
	for i < 10 {
		resp, err := http.Get("https://catfact.ninja/fact")
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		var catFact bson.M
		err = json.NewDecoder(resp.Body).Decode(&catFact)
		if err != nil {
			return err
		}
		_, err = coll.InsertOne(context.TODO(), catFact)
		if err != nil {
			return err
		}

		<-ticker.C
		i++
	}

	return nil
}

func main() {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		panic(err)
	}

	worker := NewCatFactWorker(client)
	go worker.start()

}
