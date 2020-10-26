package main

import (
    	"encoding/json"
    	"net/http"
    	"github.com/gorilla/mux"
	"strconv"
	"net"
	"time"
	"github.com/aws/aws-sdk-go/aws"
    	"github.com/aws/aws-sdk-go/aws/session"
    	"github.com/aws/aws-sdk-go/service/dynamodb"
    	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
    	loggly "github.com/jamespearly/loggly"
)

// Status for DynamoDB table.
type Message struct {
	Table       string	
	RecordCount int64	`json:"RecordCount"`
}

// For unmarshaling DynamoDB data and encoding into JSON.
type Information struct {
	Date string `json:"date"`
	Base string `json:"base"`
	
	Currency struct {
		USD float64 `json:"USD"`
		GBP float64 `json:"GBP"`
		INR float64 `json:"INR"`
		CAD float64 `json:"CAD"`
		AUD float64 `json:"AUD"`
	} `json:"rates"`
}

// Returns HTTP status codes, pushes notifications to Loggly
func status(w http.ResponseWriter, r *http.Request) {
    var tag string
    tag = "CSC482GoServer"
    	
    	// DynamoDB session.
	sess := session.Must(session.NewSession(&aws.Config{
		Region:   aws.String("us-east-1"),
		Endpoint: aws.String("https://dynamodb.us-east-1.amazonaws.com"),
	}))
	
    	db := dynamodb.New(sess)
    	
    	// Loggly Client
	client := loggly.New(tag)
	w.Header().Set("Content-Type", "application/json")
	
	//Get the Client IP:
	ip, port, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
        	client.Send("error", "Could not retrieve client's IP address.")
	}
    	
    	// Using a method other than "Get"
    if r.Method != http.MethodGet {
    	client.EchoSend("warn", "Method: " + r.Method + ", Host: " + ip + ":" + port + ", Requested Path: " + r.URL.Path + ", HTTP Status Code: " + strconv.Itoa(http.StatusMethodNotAllowed) + ".")
    	
	w.WriteHeader(http.StatusMethodNotAllowed)
	data := "405 - Method Not Allowed."
	json.NewEncoder(w).Encode(data)
    } else { 
    	client.EchoSend("info", "Method: " + r.Method + ", Host: " + ip + ":" + port + ", Requested Path: " + r.URL.Path + ", HTTP Status Code: " + strconv.Itoa(http.StatusOK) + ".")
    	
    	// Scan the entire asingh2-rates table.
    	var tableName string = "asingh2-rates"
    	input := &dynamodb.ScanInput{
    		TableName: aws.String(tableName),
    	}
    	
    	result, err := db.Scan(input)
    	
    	// Scan request error output.
    	if err != nil {
    		client.Send("error", "Could not scan table inside func status " + tableName + " No Echo.")
    	} else {
   		// JSON response to client.
		data := Message{tableName, *result.Count}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(data)
    	}
    }
}

// Dispays error when an attempt to access non-existent path is made.
func forbidden(w http.ResponseWriter, r *http.Request) {
    	var tag string
    	tag = "CSC482GoServer"
    
	client := loggly.New(tag)
	
	// Find client's IP address and Port number.
	ip, port, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
        	client.Send("error", "Could not retrieve client's IP address.")
	}
	
    	client.EchoSend("info", "Method: " + r.Method + ", Host: " + ip + ":" + port + ", Requested Path: " + r.URL.Path + ", HTTP Status Code: " + strconv.Itoa(http.StatusNotFound) + ".")
    	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	data := "404 - Page Not Found."
	json.NewEncoder(w).Encode(data)
}

// Return everything from DynamoDB table in a JSON body to client.
func all(w http.ResponseWriter, r *http.Request) {
    var tag string
    tag = "CSC482GoServer"
    
    	// DynamoDB session.
	sess := session.Must(session.NewSession(&aws.Config{
		Region:   aws.String("us-east-1"),
		Endpoint: aws.String("https://dynamodb.us-east-1.amazonaws.com"),
	}))
	
    	db := dynamodb.New(sess)
    
	client := loggly.New(tag)
	w.Header().Set("Content-Type", "application/json")
	
	//Get the Client IP:
	ip, port, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
        	client.Send("error", "Could not retrieve client's IP address.")
	}
    	
    if r.Method != http.MethodGet {
    	client.EchoSend("warn", "Method: " + r.Method + ", Host: " + ip + ":" + port +  ", Requested Path: " + r.URL.Path + ", HTTP Status Code: " + strconv.Itoa(http.StatusMethodNotAllowed) + ".")
    	
	w.WriteHeader(http.StatusMethodNotAllowed)
	data := "405 - Method Not Allowed."
	json.NewEncoder(w).Encode(data)
    } else {
    	client.EchoSend("info", "Method: " + r.Method + ", Host: " + ip + ":" + port + ", Requested Path: " + r.URL.Path + ", HTTP Status Code: " + strconv.Itoa(http.StatusOK) + ".")
    	
    	// Instance of information struct for unmarshaling JSON.
    	var info []Information
    	
    	// Scan the entire asingh2-rates table.
    	var tableName string = "asingh2-rates"
    	input := &dynamodb.ScanInput{
    		TableName: aws.String(tableName),
    	}
    	
    	result, err := db.Scan(input)
    	
    	if err != nil {
   		client.Send("error", "Could not perform scan on table " + tableName + " in func all. No Echo.")
   	} else {
   		// Unmarshal all resulting scan data into Information struct.
   		err := dynamodbattribute.UnmarshalListOfMaps(result.Items, &info)
   		if err != nil {
   			client.EchoSend("error", "Couldn't unmarshal scan data in func all. No echo.")
   		}
   		// Encode data as JSON and return with query.
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(info)
	}
    }
} 

// Creates an HTTP server using Mux Router.
func main() {
	
	// Handle all possible routes and invalid ones.
	r := mux.NewRouter()
	r.HandleFunc("/asingh2/status", status)
	r.HandleFunc("/asingh2/status/", status)
	r.HandleFunc("/asingh2/all", all)
	r.HandleFunc("/asingh2/all/", all)
	r.HandleFunc("/", forbidden)
	r.HandleFunc("/{[a-z]+}", forbidden)
	r.HandleFunc("/{[a-z]+}/", forbidden)
	r.HandleFunc("/{[a-z]+}/{[a-z]+}", forbidden).
	Schemes("http")

	srv := &http.Server{
		Handler: r,
	    	Addr:	":8080",
	    	
	    	WriteTimeout: 15 * time.Second,
	    	ReadTimeout: 15 * time.Second,
	}
	srv.ListenAndServe()
}
