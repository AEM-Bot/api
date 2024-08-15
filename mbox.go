package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func getToken() (string, error) {
	// Prepare the URL-encoded form data
	data := url.Values{}
	data.Set("client_id", "f30cec70e0f54d859a6821b6e9c4fe53")
	data.Set("client_secret", "f3352e48-f51c-48d0-91bd-d72125149bee")
	data.Set("grant_type", "client_credentials")
	data.Set("scope", "openid,AdobeID,target_sdk,additional_info.roles,read_organizations,additional_info.projectedProductContext")

	// Create the HTTP request
	req, err := http.NewRequest("POST", "https://ims-na1.adobelogin.com/ims/token/v3", strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Execute the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read and parse the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	// Extract the access token from the response
	token, ok := result["access_token"].(string)
	if !ok {
		return "", fmt.Errorf("could not extract access_token from response")
	}

	return token, nil
}

func getMboxes(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

	// Handle preflight OPTIONS request
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Get the access token
	token, err := getToken()
	if err != nil {
		http.Error(w, "Failed to get token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Create the HTTP request to fetch mboxes
	req, err := http.NewRequest("GET", "https://mc.adobe.io/lbrands/target/mboxes?limit=1000&sortBy=name&offset=0", nil)
	if err != nil {
		http.Error(w, "Failed to create request: "+err.Error(), http.StatusInternalServerError)
		return
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", "f30cec70e0f54d859a6821b6e9c4fe53")

	// Execute the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to fetch mboxes: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Read and forward the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read mboxes response: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func main() {
	http.HandleFunc("/getmboxes", getMboxes)
	fmt.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
