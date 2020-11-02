package main

import (
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/gorilla/mux"
	loggly "github.com/jamespearly/loggly"
	"github.com/mrz1836/go-sanitize" // Sanitizing URLs
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Status for DynamoDB table.
type Message struct {
	Table       string
	RecordCount int64 `json:"RecordCount"`
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
		client.Send("error", "Could not retrieve client's IP address. No Echo.")
	}

	// Using a method other than "Get"
	if r.Method != http.MethodGet {
		client.EchoSend("warn", "Method: "+r.Method+", Host: "+ip+":"+port+", Requested Path: "+r.URL.Path+", HTTP Status Code: "+strconv.Itoa(http.StatusMethodNotAllowed)+".")

		w.WriteHeader(http.StatusMethodNotAllowed)
		data := "405 - Method Not Allowed."
		json.NewEncoder(w).Encode(data)
	} else {
		// Scan the entire asingh2-rates table.
		var tableName string = "asingh2-rates"
		input := &dynamodb.ScanInput{
			TableName: aws.String(tableName),
		}

		result, err := db.Scan(input)

		// Scan request error output.
		if err != nil {
			client.Send("error", "Could not scan table inside func status "+tableName+" No Echo.")
			client.EchoSend("error", "Method: "+r.Method+", Host: "+ip+":"+port+", Requested Path: "+r.URL.Path+", HTTP Status Code: "+strconv.Itoa(http.StatusInternalServerError)+".")

			w.WriteHeader(http.StatusInternalServerError)
			data := "500 - Internal Server Error."
			json.NewEncoder(w).Encode(data)
		} else {
			client.EchoSend("info", "Method: "+r.Method+", Host: "+ip+":"+port+", Requested Path: "+r.URL.Path+", HTTP Status Code: "+strconv.Itoa(http.StatusOK)+".")

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
		client.Send("error", "Could not retrieve client's IP address. No Echo.")
	}

	client.EchoSend("warn", "Method: "+r.Method+", Host: "+ip+":"+port+", Requested Path: "+r.URL.Path+", HTTP Status Code: "+strconv.Itoa(http.StatusNotFound)+".")

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
		client.Send("error", "Could not retrieve client's IP address. No Echo.")
	}

	// Validate HTTP method type.
	if r.Method != http.MethodGet {
		client.EchoSend("warn", "Method: "+r.Method+", Host: "+ip+":"+port+", Requested Path: "+r.URL.Path+", HTTP Status Code: "+strconv.Itoa(http.StatusMethodNotAllowed)+".")

		w.WriteHeader(http.StatusMethodNotAllowed)
		data := "405 - Method Not Allowed."
		json.NewEncoder(w).Encode(data)
	} else {
		// Instance of information struct for unmarshaling JSON.
		var info []Information

		// Scan the entire asingh2-rates table.
		var tableName string = "asingh2-rates"
		input := &dynamodb.ScanInput{
			TableName: aws.String(tableName),
		}

		result, err := db.Scan(input)

		// Scan is not successfull.
		if err != nil {
			// Inform client about internal server error.
			client.Send("error", "Could not perform scan on table "+tableName+". No Echo.")
			client.EchoSend("error", "Method: "+r.Method+", Host: "+ip+":"+port+", Requested Path: "+r.URL.Path+", HTTP Status Code: "+strconv.Itoa(http.StatusInternalServerError)+".")

			w.WriteHeader(http.StatusInternalServerError)
			data := "500 - Internal Server Error."
			json.NewEncoder(w).Encode(data)
		} else {
			// Unmarshal all resulting scan data into Information struct.
			err := dynamodbattribute.UnmarshalListOfMaps(result.Items, &info)
			if err != nil {
				client.EchoSend("error", "Couldn't unmarshal scan data. No echo.")
				client.EchoSend("error", "Method: "+r.Method+", Host: "+ip+":"+port+", Requested Path: "+r.URL.Path+", HTTP Status Code: "+strconv.Itoa(http.StatusInternalServerError)+".")

				w.WriteHeader(http.StatusInternalServerError)
				data := "500 - Internal Server Error."
				json.NewEncoder(w).Encode(data)
			} else {
				// Encode data as JSON and return with query.
				client.EchoSend("info", "Method: "+r.Method+", Host: "+ip+":"+port+", Requested Path: "+r.URL.Path+", HTTP Status Code: "+strconv.Itoa(http.StatusOK)+".")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(info)
			}
		}
	}
}

// Extracts search parameter from URL query, sanitzes it, queries DynamoDB and returns JSON.
func search(w http.ResponseWriter, r *http.Request) {
	var tag string
	tag = "CSC482GoServer"

	client := loggly.New(tag)
	w.Header().Set("Content-Type", "application/json")

	//Get the Client IP:
	ip, port, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		client.Send("error", "Could not retrieve client's IP address. No Echo.")
	}

	// Using a method other than "Get"
	if r.Method != http.MethodGet {
		client.EchoSend("warn", "Method: "+r.Method+", Host: "+ip+":"+port+", Requested Path: "+r.URL.Path+", HTTP Status Code: "+strconv.Itoa(http.StatusMethodNotAllowed)+".")
		w.WriteHeader(http.StatusMethodNotAllowed)
		data := "405 - Method Not Allowed."
		json.NewEncoder(w).Encode(data)
	} else {
		query := r.URL.Query()

		// Extract the date parameter.
		date, present := query["date"]
		dateArr2Str := strings.Join(date, " ")

		if !present || len(date) == 0 {
			client.EchoSend("warn", "Method: "+r.Method+", Host: "+ip+":"+port+", Requested Path: "+r.URL.Path+", Requested Query: "+dateArr2Str+", HTTP Status Code: "+strconv.Itoa(http.StatusBadRequest)+".")

			client.Send("warn", "No date parameter specified in search query. No echo.")
			w.WriteHeader(http.StatusBadRequest)
			data := "400 - Bad Request."
			json.NewEncoder(w).Encode(data)
		} else {
			// Convert date string array to string for sanitization.
			sanitized := sanitize.URL(dateArr2Str)

			// Ignore any input greater than or less than 10.
			if len(sanitized) != 10 {
				client.EchoSend("warn", "Method: "+r.Method+", Host: "+ip+":"+port+", Requested Path: "+r.URL.Path+", Requested Query: "+dateArr2Str+", HTTP Status Code: "+strconv.Itoa(http.StatusBadRequest)+".")

				w.WriteHeader(http.StatusBadRequest)
				data := "400 - Bad Request. Format: YYYY-MM-DD"
				json.NewEncoder(w).Encode(data)
			} else {
				re := regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)
				regMatch := re.MatchString(sanitized)

				// Determine if the string matches the regex.
				if !regMatch {
					client.EchoSend("warn", "Method: "+r.Method+", Host: "+ip+":"+port+", Requested Path: "+r.URL.Path+", Requested Query: "+dateArr2Str+", HTTP Status Code: "+strconv.Itoa(http.StatusBadRequest)+".")

					w.WriteHeader(http.StatusBadRequest)
					data := "400 - Bad Request. Format: YYYY-MM-DD"
					json.NewEncoder(w).Encode(data)
				} else {
					// Determine if query parameter's value matches regex.
					matchedString := re.FindAllString(sanitized, -1)
					stringAry2String := strings.Join(matchedString, " ")

					// Make sure string length is only 10, which is proper input (YYYY-MM-DD). Probably redundant but why not.
					if len(stringAry2String) != 10 {
						client.EchoSend("warn", "Method: "+r.Method+", Host: "+ip+":"+port+", Requested Path: "+r.URL.Path+", Requested Query: "+dateArr2Str+", HTTP Status Code: "+strconv.Itoa(http.StatusBadRequest)+".")

						w.WriteHeader(http.StatusBadRequest)
						data := "400 - Bad Request. Format: YYYY-MM-DD"
						json.NewEncoder(w).Encode(data)
					} else {
						// Extract year, month, date from string
						yearString := stringAry2String[0:4]
						monthString := stringAry2String[5:7]
						dateString := stringAry2String[8:10]

						// Convert year, month, date to Integers for verification.
						yearInt, err1 := strconv.Atoi(yearString)
						monthInt, err2 := strconv.Atoi(monthString)
						dateInt, err3 := strconv.Atoi(dateString)

						// No success in parsing strings to integer values.
						if err1 != nil || err2 != nil || err3 != nil {
							client.EchoSend("error", "Could not convert year, month, date to string to int. Echo.")
							client.EchoSend("error", "Method: "+r.Method+", Host: "+ip+":"+port+", Requested Path: "+r.URL.Path+", HTTP Status Code: "+strconv.Itoa(http.StatusInternalServerError)+".")

							w.WriteHeader(http.StatusInternalServerError)
							data := "500 - Internal Server Error."
							json.NewEncoder(w).Encode(data)
						} else {
							// DynamoDB only contains data from 2019 and 2020.
							if yearInt < 2019 || yearInt > 2020 {
								client.EchoSend("warn", "Method: "+r.Method+", Host: "+ip+":"+port+", Requested Path: "+r.URL.Path+", Requested Query: "+dateArr2Str+", HTTP Status Code: "+strconv.Itoa(http.StatusBadRequest)+", Year not supported.")

								w.WriteHeader(http.StatusBadRequest)
								data := "400 - Bad  Request. Year range: 2019-2020."
								json.NewEncoder(w).Encode(data)
							} else if monthInt < 1 || monthInt > 12 {
								client.EchoSend("warn", "Method: "+r.Method+", Host: "+ip+":"+port+", Requested Path: "+r.URL.Path+", Requested Query: "+dateArr2Str+", HTTP Status Code: "+strconv.Itoa(http.StatusBadRequest)+", Invalid Month.")

								w.WriteHeader(http.StatusBadRequest)
								data := "400 - Bad Request. Invalid Month."
								json.NewEncoder(w).Encode(data)
							} else if dateInt < 1 || dateInt > 31 {
								client.EchoSend("warn", "Method: "+r.Method+", Host: "+ip+":"+port+", Requested Path: "+r.URL.Path+", Requested Query: "+dateArr2Str+", HTTP Status Code: "+strconv.Itoa(http.StatusBadRequest)+", Invalid Date.")

								w.WriteHeader(http.StatusBadRequest)
								data := "400 - Bad Request. Invalid Date."
								json.NewEncoder(w).Encode(data)
							} else {
								// DynamoDB session.
								sess := session.Must(session.NewSession(&aws.Config{
									Region:   aws.String("us-east-1"),
									Endpoint: aws.String("https://dynamodb.us-east-1.amazonaws.com"),
								}))

								db := dynamodb.New(sess)

								var tableName string = "asingh2-rates"

								result, err := db.GetItem(&dynamodb.GetItemInput{
									TableName: aws.String(tableName),
									Key: map[string]*dynamodb.AttributeValue{
										"date": {
											S: aws.String(stringAry2String),
										},
									},
								})

								// Could not query said item from DynamoDB table.
								if err != nil {
									client.EchoSend("error", "Could not query "+stringAry2String+" from table "+tableName+".")
									client.EchoSend("error", "Method: "+r.Method+", Host: "+ip+":"+port+", Requested Path: "+r.URL.Path+", HTTP Status Code: "+strconv.Itoa(http.StatusInternalServerError)+".")

									w.WriteHeader(http.StatusInternalServerError)
									data := "500 - Internal Server Error."
									json.NewEncoder(w).Encode(data)
								} else {
									// No data found.
									if result.Item == nil {
										client.EchoSend("info", "No such item found in DynamoDB. Echo.")
										client.EchoSend("info", "Method: "+r.Method+", Host: "+ip+":"+port+", Requested Path: "+r.URL.Path+", HTTP Status Code: "+strconv.Itoa(http.StatusNotFound)+".")

										w.WriteHeader(http.StatusNotFound)
										data := "404 - Item Not Found."
										json.NewEncoder(w).Encode(data)
									} else {
										info := Information{}

										// Unmarshal all resulting scan data into Information struct.
										err := dynamodbattribute.UnmarshalMap(result.Item, &info)
										if err != nil {
											client.EchoSend("error", "Couldn't unmarshal scan data. No echo.")
											client.EchoSend("error", "Method: "+r.Method+", Host: "+ip+":"+port+", Requested Path: "+r.URL.Path+", HTTP Status Code: "+strconv.Itoa(http.StatusInternalServerError)+".")

											w.WriteHeader(http.StatusInternalServerError)
											data := "500 - Internal Server Error."
											json.NewEncoder(w).Encode(data)
										} else {
											// Encode data as JSON and return with query.
											client.EchoSend("info", "Method: "+r.Method+", Host: "+ip+":"+port+", Requested Path: "+r.URL.Path+", HTTP Status Code: "+strconv.Itoa(http.StatusOK)+".")
											w.WriteHeader(http.StatusOK)
											json.NewEncoder(w).Encode(info)
										}
									}
								}
							}
						}
					}
				}
			}
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
	r.HandleFunc("/asingh2/search", search)
	r.HandleFunc("/", forbidden)
	r.HandleFunc("/{[a-z]+}", forbidden)
	r.HandleFunc("/{[a-z]+}/", forbidden)
	r.HandleFunc("/{[a-z]+}/{[a-z]+}", forbidden).
		Schemes("http")

	srv := &http.Server{
		Handler: r,
		Addr:    ":8080",

		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	srv.ListenAndServe()
}
