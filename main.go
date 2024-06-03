package main

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/http"
	"strconv"
)

func readF(path string) string {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	c, err := io.ReadAll(file)
	return string(c)
}
func readDocs(c *mongo.Cursor, d *[]map[string]interface{}) {
	for c.Next(context.TODO()) {
		var result map[string]interface{}
		if err := c.Decode(&result); err != nil {
			panic(err)
		}
		*d = append(*d, result)
	}
}
func main() {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(os.Getenv("tok")))
	// Get a handle to the database
	collection := client.Database("vdlg").Collection("users")
	// Find all documents in the collection
	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		panic(err)
	}

	// Create a slice to store the documents
	var documents []map[string]interface{}
	level_goals := map[int]int{1: 100, 2: 200, 3: 300, 4: 400, 5: 500, 6: 600, 7: 700}
	readDocs(cursor, &documents)
	// Iterate through the results and append them to the slice

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write([]byte(readF("index.html")))
	})

	http.HandleFunc("/set-points", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var s map[string]interface{}
			err := json.NewDecoder(r.Body).Decode(&s)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			p, err := strconv.Atoi(s["points"].(string))
			if err != nil {
				panic(err)
			}
			println(s)
			filter := bson.M{"name": s["name"]}
			update := bson.M{"$set": bson.M{"points": p}}
			_, err = collection.UpdateOne(context.TODO(), filter, update)

			readDocs(cursor, &documents)

			for _, document := range documents {
				println(document)
				if document["name"] == s["name"] && int(document["points"].(float64)) >= level_goals[int(document["points"].(float64))] {
					update := bson.M{"$inc": bson.M{"level": 1}}
					_, err = collection.UpdateOne(context.TODO(), filter, update)
					if err != nil {
						panic(err)
					} else {
						println("level up!")
					}
					
				}
			}
			println("Success!!!")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("{\"status\": \"success\", \"message\": \"Points updated successfully.\"}"))
		
		} else {
			w.Write([]byte(r.Method + ": " + http.StatusText(http.StatusMethodNotAllowed)))
		}
	})

	http.HandleFunc("/points", func(w http.ResponseWriter, r *http.Request) {
		v, err := json.Marshal(documents)
		if err != nil {
			panic(err)
		}
		w.Write(v)
	})
	port, env := os.LookupEnv("PORT")
	if !env {
		port = "3000"
	}
	http.ListenAndServe(":"+port, nil)
	println("Server runs successfully on port", port, "!")
}
