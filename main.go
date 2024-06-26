package main

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/http"
	"bytes"
	"strconv"
)

func createKeyValuePairs(m map[string]string) string {
		b := new(bytes.Buffer)
		for key, value := range m {
				fmt.Fprintf(b, "%s=\"%s\"\n", key, value)
		}
		return b.String()
}

func readF(path string) string {

	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	c, err := io.ReadAll(file)
	return string(c)
}
func readDocs(collection *mongo.Collection, d *[]map[string]interface{}) {
	*d = []map[string]interface{}{}
	c, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		panic(err)
	}
	for c.Next(context.TODO()) {
		var result map[string]interface{}
		if err := c.Decode(&result); err != nil {
			panic(err)
		}
		*d = append(*d, result)
	}
}
func main() {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb+srv://trvlert:RrhE5a553UMc0LIC@turncraft.4bigr.mongodb.net/vdlg"))
	// Get a handle to the database
	collection := client.Database("vdlg").Collection("users")

	// Create a slice to store the documents
	var documents []map[string]interface{}
	level_goals := map[int]int{1: 100, 2: 200, 3: 300, 4: 400, 5: 500, 6: 600, 7: 700}
	readDocs(collection, &documents)
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
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.Method == http.MethodPost {
			println("Recieved post!")
			var s map[string]string
			err := json.NewDecoder(r.Body).Decode(&s)
			if err != nil {
				println("Uh oh", err.Error())
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			p,err := strconv.Atoi(s["points"])
			if err != nil {
				panic(err)	
			}
			filter := bson.M{"name": s["name"]}
			update := bson.M{"$inc": bson.M{"points": p}}
			_, err = collection.UpdateOne(context.TODO(), filter, update)
			if err != nil {
				panic(err)
			}
			readDocs( collection, &documents)
			for _, document := range documents {
				points := int(document["points"].(int32))
				level := int(document["level"].(int32))
				if document["name"] == s["name"] && points >= level_goals[level] {
					update := bson.M{"$inc": bson.M{"level": 1}}
					n := bson.M{"$set": bson.M{"points": points - level_goals[points]}}
					_, err = collection.UpdateOne(context.TODO(), filter, update)
					_, err = collection.UpdateOne(context.TODO(), filter, n)
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
		w.Header().Set("Access-Control-Allow-Origin", "*")
		readDocs(collection, &documents)
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
