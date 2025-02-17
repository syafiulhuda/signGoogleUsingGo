package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var googleOAuthConfig *oauth2.Config
var randomState = "random"

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	// Ambil nilai Client ID dan Client Secret dari environment variables
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")

	// Print untuk memastikan variabel terbaca
	fmt.Println("Client ID:", clientID)
	fmt.Println("Client Secret:", clientSecret)

	// Pastikan Client ID dan Secret tidak kosong
	if clientID == "" || clientSecret == "" {
		fmt.Println("Client ID atau Client Secret tidak ditemukan. Pastikan sudah diatur di .env!")
		return
	}

	// Inisialisasi konfigurasi OAuth2
	googleOAuthConfig = &oauth2.Config{
		RedirectURL:  "http://localhost:9090/callback",
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}

	// Daftarkan handler untuk route
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/callback", handleCallback)

	// Jalankan server HTTP di port 9090
	fmt.Println("Server running on http://localhost:9090")
	if err := http.ListenAndServe(":9090", nil); err != nil {
		fmt.Printf("Error starting server: %s\n", err)
	}
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	halaman := "/login"
	html := "<html><body><a href='" + halaman + "'>Login using Google</a></body></html>"
	fmt.Fprint(w, html)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	url := googleOAuthConfig.AuthCodeURL(randomState)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("state") != randomState {
		fmt.Println("State is not valid")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	token, err := googleOAuthConfig.Exchange(context.Background(), r.FormValue("code"))
	if err != nil {
		fmt.Printf("Could not get token: %s\n", err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Ambil data pengguna dari Google API
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		fmt.Printf("Could not create request: %s\n", err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Could not read response body: %s\n", err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Tampilkan data yang didapat
	fmt.Fprintf(w, "Content: %s", content)
}
