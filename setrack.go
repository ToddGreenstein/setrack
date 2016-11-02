package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	sRand "math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/couchbase/gocb"
	"github.com/gorilla/mux"
)

// Test Key for AES Cryptographic hashing of user password
const key string = "Really_Unsecure_Key"

// bucket reference
var bucket *gocb.Bucket

func CreateActivityHandler(w http.ResponseWriter, req *http.Request) {
	var sessionActivity SessionActivity
	_ = json.NewDecoder(req.Body).Decode(&sessionActivity.Activity)
	curActivity, err := sessionActivity.Create()
	if err != nil {
		w.WriteHeader(401)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(curActivity)
}

func RetrieveActivityHandler(w http.ResponseWriter, req *http.Request) {
	var sessionActivity SessionActivity
	vars := mux.Vars(req)
	curActivity, err := sessionActivity.Retrieve(vars["activityId"])
	if err != nil {
		w.WriteHeader(401)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(curActivity)
}

func AddActivityHandler(w http.ResponseWriter, req *http.Request) {
	_, err := AddActivity()
	if err != nil {
		w.WriteHeader(401)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	//json.NewEncoder(w).Encode(curActivity)
}

func LoginHandler(w http.ResponseWriter, req *http.Request) {
	var session Session
	vars := mux.Vars(req)
	curUser, err := session.Login(vars["email"], vars["pass"])
	if err != nil {
		w.WriteHeader(401)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(curUser)
}

func CreateLoginHandler(w http.ResponseWriter, req *http.Request) {
	var session Session
	_ = json.NewDecoder(req.Body).Decode(&session.User)
	curUser, err := session.Create()
	if err != nil {
		w.WriteHeader(401)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(curUser)
}

func RetrieveUserHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	var session Session
	curUser, err := session.Retrieve(vars["userId"])
	if err != nil {
		w.WriteHeader(401)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(curUser)
}

func RetrieveAllUserHandler(w http.ResponseWriter, req *http.Request) {
	var session Session
	curUser, err := session.RetrieveAll()
	if err != nil {
		w.WriteHeader(401)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(curUser)
}

func GenLoadHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	threads, _ := strconv.Atoi(vars["threads"])
	threshold, _ := strconv.Atoi(vars["threshold"])
	go GenLoad(threads, threshold)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode("Data Loader:" + vars["threads"] + " threads, " + vars["threshold"] + " items")
}

func GenLoadAddUsers(w http.ResponseWriter, req *http.Request) {
	AddUsers()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode("Add Users For Testing")
}

func GenLoad(threads int, threshold int) {
	c := make(chan bool, 128)
	for i := 0; i < threads; i++ {
		go func() {
			for {
				select {
				case <-c:
					_, err := AddActivity()
					if err != nil {
						fmt.Printf("Sleep Error\n")
						time.Sleep(1000 * time.Millisecond)
					}
				default:
					return
				}
			}
		}()
	}
	for i := 0; i < threshold; i++ {
		c <- true
	}
	return
}

func GenUUID() (uuid string) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	uuid = fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	return
}

func GenINT(min int, max int) string {
	var bytes int
	bytes = min + sRand.Intn(max)
	return strconv.Itoa(bytes)
}

func main() {
	//runtime.GOMAXPROCS(4)
	// Cluster connection and bucket for couchbase

	cluster, _ := gocb.Connect("couchbase://localhost")
	bucket, _ = cluster.OpenBucket("travel-sample", "")

	// Http Router
	r := mux.NewRouter()

	// Activity Routes
	r.HandleFunc("/api/activity/get/{activityId}", RetrieveActivityHandler).Methods("GET")
	r.HandleFunc("/api/activity/create", CreateActivityHandler).Methods("POST")

	// User Routes
	r.HandleFunc("/api/user/login/{email}/{pass}", LoginHandler).Methods("GET")
	r.HandleFunc("/api/user/getAll", RetrieveAllUserHandler).Methods("GET")
	r.HandleFunc("/api/user/get/{userId}", RetrieveUserHandler).Methods("GET")
	r.HandleFunc("/api/user/create", CreateLoginHandler).Methods("POST")

	// Unit Test Routes
	r.HandleFunc("/api/activity/add", AddActivityHandler).Methods("POST")

	// LoadGen Routes
	r.HandleFunc("/api/activity/gen/{threads}/{threshold}", GenLoadHandler).Methods("POST")
	r.HandleFunc("/api/users/gen", GenLoadAddUsers).Methods("POST")

	fmt.Printf("Starting server on :3000\n")
	http.ListenAndServe(":3000", r)
}
