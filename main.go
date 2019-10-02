package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gorilla/mux"
)

type VerifyResponse struct {
	Message string `json:"message"`
	Success bool   `json:"status"`
}

type ApiResponse struct {
	Sid    string `json:"sid"`
	Status string `json:"status"`
	To     string `json:"to"`
}

func StartVerification(w http.ResponseWriter, r *http.Request) {
	via := r.FormValue("via")
	phoneNumber := r.FormValue("phone_number")
	countryCode := r.FormValue("country_code")
	fullPhone := fmt.Sprintf("+%s%s", countryCode, phoneNumber)

	reqUrl := fmt.Sprintf("https://verify.twilio.com/v2/Services/%s/Verifications", os.Getenv("VERIFY_SERVICE_SID"))

	client := &http.Client{}

	data := url.Values{}
	data.Add("Channel", via)
	data.Add("To", fullPhone)

	req, err := http.NewRequest(http.MethodPost, reqUrl, strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	ACCOUNT_SID := os.Getenv("TWILIO_ACCOUNT_SID")
	AUTH_TOKEN := os.Getenv("TWILIO_AUTH_TOKEN")
	req.SetBasicAuth(ACCOUNT_SID, AUTH_TOKEN)

	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	bodyBytes, ioErr := ioutil.ReadAll(res.Body)
	if ioErr != nil {
		log.Fatal(ioErr)
	}

	if res.StatusCode != 201 {
		response := &VerifyResponse{string(bodyBytes), false}
		json.NewEncoder(w).Encode(response)
		return
	}

	apiResponse := ApiResponse{}
	jsonErr := json.Unmarshal(bodyBytes, &apiResponse)

	if jsonErr != nil {
		log.Print(jsonErr)
	}

	if apiResponse.Status == "pending" {
		response := &VerifyResponse{fmt.Sprintf("Token sent to %s", apiResponse.To), true}
		json.NewEncoder(w).Encode(response)
	} else {
		log.Fatal("Error sending verification token")
	}
}

func CheckVerification(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	phoneNumber := r.FormValue("phone_number")
	countryCode := r.FormValue("country_code")
	fullPhone := fmt.Sprintf("+%s%s", countryCode, phoneNumber)

	reqUrl := fmt.Sprintf("https://verify.twilio.com/v2/Services/%s/VerificationCheck", os.Getenv("VERIFY_SERVICE_SID"))

	client := &http.Client{}

	data := url.Values{}
	data.Add("Code", code)
	data.Add("To", fullPhone)

	req, err := http.NewRequest(http.MethodPost, reqUrl, strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	ACCOUNT_SID := os.Getenv("TWILIO_ACCOUNT_SID")
	AUTH_TOKEN := os.Getenv("TWILIO_AUTH_TOKEN")
	req.SetBasicAuth(ACCOUNT_SID, AUTH_TOKEN)

	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	bodyBytes, ioErr := ioutil.ReadAll(res.Body)
	if ioErr != nil {
		log.Fatal(ioErr)
	}

	apiResponse := ApiResponse{}
	jsonErr := json.Unmarshal(bodyBytes, &apiResponse)

	if jsonErr != nil {
		log.Print(jsonErr)
	}

	if apiResponse.Status == "approved" {
		response := &VerifyResponse{"Correct token!", true}
		json.NewEncoder(w).Encode(response)
	} else {
		response := &VerifyResponse{"Incorrect token.", false}
		json.NewEncoder(w).Encode(response)
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if os.Getenv("VERIFY_SERVICE_SID") == "" {
		log.Fatal("$VERIFY_SERVICE_SID must be set as an environment variable.")
	}

	if os.Getenv("TWILIO_ACCOUNT_SID") == "" || os.Getenv("TWILIO_AUTH_TOKEN") == "" {
		log.Fatal("$TWILIO_ACCOUNT_SID and $TWILIO_AUTH_TOKEN must be set")
	}

	router := mux.NewRouter()

	router.HandleFunc("/start", StartVerification).Methods("POST")
	router.HandleFunc("/check", CheckVerification).Methods("POST")
	log.Fatal(http.ListenAndServe(":"+port, router))
}
