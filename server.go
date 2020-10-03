package main

import (
    "encoding/json"
    "net/http"
    "time"
    "github.com/gorilla/mux"
    "strconv"
    "net"
    loggly "github.com/jamespearly/loggly"
)

type Message struct {
	Status int
	Time string
}

// Returns HTTP status codes, pushes notifications to Loggly
func status(w http.ResponseWriter, r *http.Request) {
    var tag string
    tag = "CSC482GoServer"
    
	client := loggly.New(tag)
	w.Header().Set("Content-Type", "application/json")
	
	//Get the Client IP:
	ip, port, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
        	client.Send("error", "Could not retrieve client's IP address.")
	}
    	
    if r.Method != http.MethodGet {
    	client.EchoSend("warn", "Method: " + r.Method + ", Host: " + ip + ":" + port +  ", Requested Path: " + r.URL.Path + ", HTTP Status Code: " 	+ strconv.Itoa(http.StatusMethodNotAllowed) + ".")
	w.WriteHeader(http.StatusMethodNotAllowed)
	data := "405 - Method Not Allowed."
	json.NewEncoder(w).Encode(data)
    } else {
    	if r.URL.Path != "/status" {
    		client.EchoSend("info", "Method: " + r.Method + ", Host: " + ip + ":" + port +  ", Requested Path: " + r.URL.Path + ", HTTP Status Code: " + strconv.Itoa(http.StatusNotFound) + ".")
		w.WriteHeader(http.StatusNotFound)
		data := "404 - Page Not Found."
		json.NewEncoder(w).Encode(data)
    	} else {
    	    	client.EchoSend("info", "Method: " + r.Method + ", Host: " + ip + ":" + port +  ", Requested Path: " + r.URL.Path + ", HTTP Status Code: " + strconv.Itoa(http.StatusOK) + ".")
	    	data := Message{http.StatusOK, time.Now().Format(time.RFC850)}
	    	w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(data)
	}
    }
}

// Creates an HTTP server using Mux Router.
func main() {
	r := mux.NewRouter()
	r.HandleFunc("/{[a-z]+}", status).     // Regex fails to capture "/a.../a..." but not "/a..."
	Schemes("http")

	srv := &http.Server{
		Handler: r,
	    	Addr:	":8000",
	    	
	    	WriteTimeout: 15 * time.Second,
	    	ReadTimeout: 15 * time.Second,
	}
	srv.ListenAndServe()
}
