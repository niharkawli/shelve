package main

import (
    "context"
    "encoding/json"
    "fmt"
    "github.com/gorilla/mux"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "log"
    "net/http"
    "time"
)

type Shelve struct {
    ID             primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
    Value          string             `json:"value,omitempty" bson:"value,omitempty"`
    //ExpirationDate time.Time          `json:"expiration_date,omitempty" bson:"expiration_date,omitempty"`
}

var client *mongo.Client

func CreateItemEndpoint(response http.ResponseWriter, request *http.Request) {
    response.Header().Set("content-type", "application/json")
    var item Shelve
    _ = json.NewDecoder(request.Body).Decode(&item)
    collection := client.Database("shelve").Collection("items")
    ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
    result, _ := collection.InsertOne(ctx, item)
    _ = json.NewEncoder(response).Encode(result)
}

func GetItemEndpoint(response http.ResponseWriter, request *http.Request) {
    response.Header().Set("content-type", "application/json")
    params := mux.Vars(request)
    id, _ := primitive.ObjectIDFromHex(params["id"])
    var item Shelve
    collection := client.Database("shelve").Collection("items")
    ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
    err := collection.FindOne(ctx, Shelve{ID: id}).Decode(&item)
    if err != nil {
        response.WriteHeader(http.StatusInternalServerError)
        _, _ = response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
        return
    }
    _ = json.NewEncoder(response).Encode(item)
}

func GetItemsEndpoint(response http.ResponseWriter, request *http.Request) {
    response.Header().Set("content-type", "application/json")
    var items []Shelve
    collection := client.Database("shelve").Collection("items")
    ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
    cursor, err := collection.Find(ctx, bson.M{})
    if err != nil {
        response.WriteHeader(http.StatusInternalServerError)
        _, _ = response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
        return
    }
    defer cursor.Close(ctx)
    for cursor.Next(ctx) {
        var item Shelve
        _ = cursor.Decode(&item)
        items = append(items, item)
    }
    if err := cursor.Err(); err != nil {
        response.WriteHeader(http.StatusInternalServerError)
        _, _ = response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
        return
    }
    _ = json.NewEncoder(response).Encode(items)
}

func main() {
    fmt.Println("Starting the application...")
    ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
    clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
    client, _ = mongo.Connect(ctx, clientOptions)
    router := mux.NewRouter()
    router.HandleFunc("/item", CreateItemEndpoint).Methods("POST")
    router.HandleFunc("/items", GetItemsEndpoint).Methods("GET")
    router.HandleFunc("/item/{id}", GetItemEndpoint).Methods("GET")
    defer fmt.Println("Closing...")
    log.Fatal(http.ListenAndServe(":8080", router))

}
