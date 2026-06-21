package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// Structs for Ory API interactions
type KratosSession struct {
	Identity struct {
		ID     string `json:"id"`
		Traits struct {
			Email string `json:"email"`
		} `json:"traits"`
	} `json:"identity"`
}

type HydraAcceptResponse struct {
	RedirectTo string `json:"redirect_to"`
}

type ConsentRequestDetails struct {
	RequestedScope []string `json:"requested_scope"`
}

type TokenRequestBody struct {
	Code string `json:"code"`
}

type IntrospectResponse struct {
	Active bool           `json:"active"`
	Sub    string         `json:"sub"`
	Email  string         `json:"email"`
	Ext    map[string]any `json:"ext"`
}

type KratosLogoutFlow struct {
	LogoutURL string `json:"logout_url"`
}

// CORS middleware
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Cookie")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Helper to copy headers (especially cookies)
func copyHeaders(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func main() {
	mux := http.NewServeMux()

	// 1. GET /login
	mux.HandleFunc("GET /login", func(w http.ResponseWriter, r *http.Request) {
		challenge := r.URL.Query().Get("login_challenge")
		if challenge == "" {
			http.Error(w, "Missing login_challenge", http.StatusBadRequest)
			return
		}

		// Check session with Kratos
		client := &http.Client{}
		kratosReq, err := http.NewRequest("GET", "http://localhost:4433/sessions/whoami", nil)
		if err != nil {
			log.Println("Kratos req build error:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		copyHeaders(kratosReq.Header, r.Header)

		kratosRes, err := client.Do(kratosReq)
		if err != nil {
			log.Println("Kratos request failed:", err)
			http.Error(w, "Login failed", http.StatusInternalServerError)
			return
		}
		defer kratosRes.Body.Close()

		if kratosRes.StatusCode == http.StatusUnauthorized {
			// User has no session. Redirect to frontend login.
			http.Redirect(w, r, "http://localhost:3000/login?login_challenge="+challenge, http.StatusFound)
			return
		}

		if kratosRes.StatusCode != http.StatusOK {
			http.Error(w, "Login failed", http.StatusInternalServerError)
			return
		}

		var session KratosSession
		if err := json.NewDecoder(kratosRes.Body).Decode(&session); err != nil {
			log.Println("Kratos session parse error:", err)
			http.Error(w, "Login failed", http.StatusInternalServerError)
			return
		}

		// Accept login challenge at Hydra
		acceptBody := map[string]any{
			"subject":      session.Identity.ID,
			"remember":     true,
			"remember_for": 3600,
		}
		jsonBody, _ := json.Marshal(acceptBody)

		hydraReq, err := http.NewRequest("PUT", "http://localhost:4445/oauth2/auth/requests/login/accept?login_challenge="+challenge, bytes.NewBuffer(jsonBody))
		if err != nil {
			log.Println("Hydra req build error:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		hydraReq.Header.Set("Content-Type", "application/json")

		hydraRes, err := client.Do(hydraReq)
		if err != nil {
			log.Println("Hydra request failed:", err)
			http.Error(w, "Login failed", http.StatusInternalServerError)
			return
		}
		defer hydraRes.Body.Close()

		if hydraRes.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(hydraRes.Body)
			log.Printf("Hydra acceptance failed (%d): %s\n", hydraRes.StatusCode, string(bodyBytes))
			http.Error(w, "Login failed", http.StatusInternalServerError)
			return
		}

		var acceptRes HydraAcceptResponse
		if err := json.NewDecoder(hydraRes.Body).Decode(&acceptRes); err != nil {
			http.Error(w, "Login failed", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, acceptRes.RedirectTo, http.StatusFound)
	})

	// 2. GET /consent
	mux.HandleFunc("GET /consent", func(w http.ResponseWriter, r *http.Request) {
		challenge := r.URL.Query().Get("consent_challenge")
		if challenge == "" {
			http.Error(w, "Missing consent_challenge", http.StatusBadRequest)
			return
		}

		if r.Header.Get("Cookie") == "" {
			http.Redirect(w, r, "http://localhost:3000/login?consent_challenge="+challenge, http.StatusFound)
			return
		}

		client := &http.Client{}

		// Validate Kratos session
		kratosReq, err := http.NewRequest("GET", "http://localhost:4433/sessions/whoami", nil)
		if err != nil {
			http.Error(w, "Consent failed", http.StatusInternalServerError)
			return
		}
		copyHeaders(kratosReq.Header, r.Header)

		kratosRes, err := client.Do(kratosReq)
		if err != nil {
			log.Println("Kratos consent check failed:", err)
			http.Error(w, "Consent failed", http.StatusInternalServerError)
			return
		}
		defer kratosRes.Body.Close()

		if kratosRes.StatusCode == http.StatusUnauthorized {
			http.Redirect(w, r, "http://localhost:3000/login?consent_challenge="+challenge, http.StatusFound)
			return
		}

		var session KratosSession
		if err := json.NewDecoder(kratosRes.Body).Decode(&session); err != nil {
			http.Error(w, "Consent failed", http.StatusInternalServerError)
			return
		}
		email := session.Identity.Traits.Email

		// GET consent request details from Hydra
		consentReq, err := http.NewRequest("GET", "http://localhost:4445/oauth2/auth/requests/consent?consent_challenge="+challenge, nil)
		if err != nil {
			http.Error(w, "Consent failed", http.StatusInternalServerError)
			return
		}

		consentRes, err := client.Do(consentReq)
		if err != nil {
			http.Error(w, "Consent failed", http.StatusInternalServerError)
			return
		}
		defer consentRes.Body.Close()

		var consentDetails ConsentRequestDetails
		if err := json.NewDecoder(consentRes.Body).Decode(&consentDetails); err != nil {
			http.Error(w, "Consent failed", http.StatusInternalServerError)
			return
		}

		// ACCEPT consent request
		acceptBody := map[string]any{
			"grant_scope":  consentDetails.RequestedScope,
			"remember":     true,
			"remember_for": 3600,
			"session": map[string]any{
				"id_token": map[string]any{
					"email": email,
				},
				"access_token": map[string]any{
					"email": email,
				},
			},
		}
		jsonBody, _ := json.Marshal(acceptBody)

		acceptReq, err := http.NewRequest("PUT", "http://localhost:4445/oauth2/auth/requests/consent/accept?consent_challenge="+challenge, bytes.NewBuffer(jsonBody))
		if err != nil {
			http.Error(w, "Consent failed", http.StatusInternalServerError)
			return
		}
		acceptReq.Header.Set("Content-Type", "application/json")

		acceptRes, err := client.Do(acceptReq)
		if err != nil {
			http.Error(w, "Consent failed", http.StatusInternalServerError)
			return
		}
		defer acceptRes.Body.Close()

		var finalRes HydraAcceptResponse
		if err := json.NewDecoder(acceptRes.Body).Decode(&finalRes); err != nil {
			http.Error(w, "Consent failed", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, finalRes.RedirectTo, http.StatusFound)
	})

	// 3. POST /token
	mux.HandleFunc("POST /token", func(w http.ResponseWriter, r *http.Request) {
		var reqBody TokenRequestBody
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		referer := r.Header.Get("Referer")
		redirectURI := "http://localhost:3000/callback" // fallback
		if referer != "" {
			if u, err := url.Parse(referer); err == nil {
				redirectURI = fmt.Sprintf("%s://%s/callback", u.Scheme, u.Host)
			}
		}

		formValues := url.Values{}
		formValues.Set("grant_type", "authorization_code")
		formValues.Set("code", reqBody.Code)
		formValues.Set("redirect_uri", redirectURI)

		client := &http.Client{}
		hydraReq, err := http.NewRequest("POST", "http://localhost:4444/oauth2/token", strings.NewReader(formValues.Encode()))
		if err != nil {
			http.Error(w, "Token exchange failed", http.StatusInternalServerError)
			return
		}
		hydraReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		hydraReq.SetBasicAuth("36d0db37-f52e-46b6-bf1d-3923fc9cf46d", "secret")

		hydraRes, err := client.Do(hydraReq)
		if err != nil {
			http.Error(w, "Token exchange failed", http.StatusInternalServerError)
			return
		}
		defer hydraRes.Body.Close()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(hydraRes.StatusCode)
		io.Copy(w, hydraRes.Body)
	})

	// 4. GET /me
	mux.HandleFunc("GET /me", func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Missing Bearer token"})
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")

		formValues := url.Values{}
		formValues.Set("token", token)

		client := &http.Client{}
		hydraReq, err := http.NewRequest("POST", "http://localhost:4445/oauth2/introspect", strings.NewReader(formValues.Encode()))
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to validate token"})
			return
		}
		hydraReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		hydraRes, err := client.Do(hydraReq)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to validate token"})
			return
		}
		defer hydraRes.Body.Close()

		var introspect IntrospectResponse
		if err := json.NewDecoder(hydraRes.Body).Decode(&introspect); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to validate token"})
			return
		}

		if !introspect.Active {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Token is inactive or expired"})
			return
		}

		email := introspect.Email
		if email == "" {
			if extEmail, exists := introspect.Ext["email"].(string); exists {
				email = extEmail
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"email":   email,
			"subject": introspect.Sub,
		})
	})

	// 5. GET /logout
	mux.HandleFunc("GET /logout", func(w http.ResponseWriter, r *http.Request) {
		challenge := r.URL.Query().Get("logout_challenge")
		if challenge == "" {
			http.Redirect(w, r, "http://localhost:3000/", http.StatusFound)
			return
		}

		client := &http.Client{}

		// 1. Accept the logout challenge in Ory Hydra
		hydraReq, err := http.NewRequest("PUT", "http://localhost:4445/oauth2/auth/requests/logout/accept?logout_challenge="+challenge, nil)
		if err != nil {
			log.Println("Hydra logout accept build failed:", err)
			http.Error(w, "Logout failed", http.StatusInternalServerError)
			return
		}

		hydraRes, err := client.Do(hydraReq)
		if err != nil {
			log.Println("Hydra logout accept request failed:", err)
			http.Error(w, "Logout failed", http.StatusInternalServerError)
			return
		}
		defer hydraRes.Body.Close()

		var acceptRes HydraAcceptResponse
		if err := json.NewDecoder(hydraRes.Body).Decode(&acceptRes); err != nil {
			http.Error(w, "Logout failed", http.StatusInternalServerError)
			return
		}
		redirectUrl := acceptRes.RedirectTo

		// 2. Fetch browser logout URL from Ory Kratos
		kratosReq, err := http.NewRequest("GET", "http://localhost:4433/self-service/logout/browser", nil)
		if err != nil {
			http.Redirect(w, r, redirectUrl, http.StatusFound)
			return
		}
		copyHeaders(kratosReq.Header, r.Header)

		kratosRes, err := client.Do(kratosReq)
		if err != nil {
			log.Println("Kratos logout request failed:", err)
			http.Redirect(w, r, redirectUrl, http.StatusFound)
			return
		}
		defer kratosRes.Body.Close()

		if kratosRes.StatusCode != http.StatusOK {
			log.Println("Kratos logout returned status:", kratosRes.StatusCode)
			http.Redirect(w, r, redirectUrl, http.StatusFound)
			return
		}

		var logoutFlow KratosLogoutFlow
		if err := json.NewDecoder(kratosRes.Body).Decode(&logoutFlow); err != nil {
			http.Redirect(w, r, redirectUrl, http.StatusFound)
			return
		}

		finalRedirect := fmt.Sprintf("%s&return_to=%s", logoutFlow.LogoutURL, url.QueryEscape(redirectUrl))
		http.Redirect(w, r, finalRedirect, http.StatusFound)
	})

	log.Println("Backend running on http://localhost:4000")
	log.Fatal(http.ListenAndServe(":4000", enableCORS(mux)))
}
