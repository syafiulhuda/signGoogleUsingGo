package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time" // Package time sudah digunakan untuk mengatur waktu cookie

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
	http.HandleFunc("/logout", handleLogout) // Tambahkan handler logout

	// Jalankan server HTTP di port 9090
	fmt.Println("Server running on http://localhost:9090")
	if err := http.ListenAndServe(":9090", nil); err != nil {
		fmt.Printf("Error starting server: %s\n", err)
	}
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	html := `
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>Login with Google</title>
		<style>
			body {
				font-family: Arial, sans-serif;
				background-color: #f4f4f9;
				display: flex;
				justify-content: center;
				align-items: center;
				height: 100vh;
				margin: 0;
			}
			.container {
				background: white;
				padding: 2rem;
				border-radius: 8px;
				box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
				text-align: center;
			}
			.button {
				background-color: #4285F4;
				color: white;
				border: none;
				padding: 0.75rem 1.5rem;
				border-radius: 4px;
				font-size: 1rem;
				cursor: pointer;
				text-decoration: none;
				display: inline-block;
				margin-top: 1rem;
			}
			.button:hover {
				background-color: #357ABD;
			}
		</style>
	</head>
	<body>
		<div class="container">
			<h1>Welcome to My App</h1>
			<p>Please log in to continue.</p>
			<a href="/login" class="button">Login with Google</a>
			<!-- Link Logout hanya sebagai contoh, jika sudah login dan menyimpan sesi -->
			<a href="/logout" class="button">Logout</a>
		</div>
	</body>
	</html>
	`
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}
func handleLogin(w http.ResponseWriter, r *http.Request) {
	// Menambahkan parameter "prompt=select_account" agar Google selalu menampilkan halaman pemilihan akun.
	url := googleOAuthConfig.AuthCodeURL(randomState, oauth2.SetAuthURLParam("prompt", "select_account"))
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

	// Simpan token di cookie (untuk keperluan sesi)
	// Pastikan untuk mengenkripsi token dan mengamankan cookie di production.
	cookie := &http.Cookie{
		Name:     "session_token",
		Value:    token.AccessToken,
		Path:     "/",
		Expires:  time.Now().Add(1 * time.Hour),
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)

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

	// Tampilkan data pengguna dalam format HTML
	html := `
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>User Info</title>
		<style>
			body {
				font-family: Arial, sans-serif;
				background-color: #f4f4f9;
				display: flex;
				justify-content: center;
				align-items: center;
				height: 100vh;
				margin: 0;
			}
			.user-info-container {
				background: white;
				padding: 2rem;
				border-radius: 8px;
				box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
				text-align: center;
			}
			.user-info-container h1 {
				margin-bottom: 1rem;
			}
			.user-info-container pre {
				background: #f9f9f9;
				padding: 1rem;
				border-radius: 4px;
				text-align: left;
				max-width: 400px;
				overflow-x: auto;
			}
			.button {
				background-color: #4285F4;
				color: white;
				border: none;
				padding: 0.75rem 1.5rem;
				border-radius: 4px;
				font-size: 1rem;
				cursor: pointer;
				text-decoration: none;
				display: inline-block;
				margin-top: 1rem;
			}
			.button:hover {
				background-color: #357ABD;
			}
		</style>
	</head>
	<body>
		<div class="user-info-container">
			<h1>User Info</h1>
			<pre>` + string(content) + `</pre>
			<a href="/logout" class="button">Logout</a>
		</div>
	</body>
	</html>
	`
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

// handleLogout menghapus cookie session dan mengarahkan pengguna ke halaman home
func handleLogout(w http.ResponseWriter, r *http.Request) {
	// Buat cookie dengan masa berlaku sudah lewat untuk menghapusnya
	cookie := &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-time.Hour),
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)
	// Redirect ke halaman home setelah logout
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}
