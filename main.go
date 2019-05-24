package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type VerifyResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

func StartVerification(w http.ResponseWriter, r *http.Request) {
	via := r.FormValue("via")
	phoneNumber := r.FormValue("phone_number")
	countryCode := r.FormValue("country_code")

	reqUrl := "https://api.authy.com/protected/json/phones/verification/start"

	client := http.Client{Timeout: time.Second * 2}

	data := url.Values{}
	data.Add("via", via)
	data.Add("phone_number", phoneNumber)
	data.Add("country_code", countryCode)

	req, reqErr := http.NewRequest(
		http.MethodPost,
		reqUrl,
		strings.NewReader(data.Encode()))

	if reqErr != nil {
		log.Fatal(reqErr)
	}

	api_key := os.Getenv("AUTHY_API_KEY")
	if api_key == "" {
		response := &VerifyResponse{"$AUTHY_API_KEY must be set", false}
		json.NewEncoder(w).Encode(response)
		return
	}
	req.Header.Add("X-Authy-API-Key", api_key)

	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	bodyBytes, ioErr := ioutil.ReadAll(res.Body)
	if ioErr != nil {
		log.Fatal(ioErr)
	}

	response := VerifyResponse{}
	jsonErr := json.Unmarshal(bodyBytes, &response)

	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	json.NewEncoder(w).Encode(response)
}

func CheckVerification(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	phoneNumber := r.FormValue("phone_number")
	countryCode := r.FormValue("country_code")

	reqUrl := "https://api.authy.com/protected/json/phones/verification/check"

	client := http.Client{Timeout: time.Second * 2}

	req, reqErr := http.NewRequest(
		http.MethodGet,
		reqUrl,
		nil)

	if reqErr != nil {
		log.Fatal(reqErr)
	}

	q := req.URL.Query()
	q.Add("verification_code", code)
	q.Add("phone_number", phoneNumber)
	q.Add("country_code", countryCode)

	req.URL.RawQuery = q.Encode()

	req.Header.Add("X-Authy-API-Key", os.Getenv("AUTHY_API_KEY"))

	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	bodyBytes, ioErr := ioutil.ReadAll(res.Body)
	if ioErr != nil {
		log.Fatal(ioErr)
	}

	response := VerifyResponse{}
	jsonErr := json.Unmarshal(bodyBytes, &response)

	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	json.NewEncoder(w).Encode(response)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("$PORT must be set")
	}

	router := mux.NewRouter()

	router.HandleFunc("/start", StartVerification).Methods("POST")
	router.HandleFunc("/check", CheckVerification).Methods("POST")
	log.Fatal(http.ListenAndServe(":"+port, router))
}
