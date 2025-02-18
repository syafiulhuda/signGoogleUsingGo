package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

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

	// Tambahkan route baru untuk Dashboard (protected route)
	http.HandleFunc("/dashboard", authMiddleware(handleDashboard))

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
			<a href="/dashboard" class="button">Dashboard</a>
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

// authMiddleware adalah middleware untuk memeriksa otentikasi pengguna
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil || cookie.Value == "" {
			// Jika tidak ada token, arahkan ke halaman login
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}
		// Token ditemukan, lanjutkan ke handler berikutnya
		next(w, r)
	}
}

// handleDashboard adalah endpoint yang dilindungi, hanya dapat diakses jika pengguna telah login.
// Di sini, bagian data JSON dipecah per bagian (ID, Name, Given Name, Family Name, Picture)
func handleDashboard(w http.ResponseWriter, r *http.Request) {
	// Ambil token dari cookie
	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	// Ambil data pengguna dari Google API menggunakan token dari cookie
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + cookie.Value)
	if err != nil {
		http.Error(w, "Gagal mengambil data pengguna", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Gagal membaca data pengguna", http.StatusInternalServerError)
		return
	}

	// Definisikan struct untuk memetakan data JSON
	type UserInfo struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		GivenName  string `json:"given_name"`
		FamilyName string `json:"family_name"`
		Picture    string `json:"picture"`
	}

	var user UserInfo
	if err := json.Unmarshal(content, &user); err != nil {
		http.Error(w, "Gagal memparsing data pengguna", http.StatusInternalServerError)
		return
	}

	// Tampilkan dashboard dengan informasi pengguna yang dipecah per bagian
	html := `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Dashboard</title>
	<style>
	   body {
		   font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
		   background: linear-gradient(135deg, #74ABE2, #5563DE);
		   margin: 0;
		   padding: 0;
		   color: #333;
	   }
	   .container {
		   max-width: 960px;
		   margin: 0 auto;
		   padding: 20px;
	   }
	   header {
		   display: flex;
		   justify-content: space-between;
		   align-items: center;
		   background: rgba(255, 255, 255, 0.9);
		   padding: 10px 20px;
		   border-radius: 8px;
		   box-shadow: 0 2px 4px rgba(0,0,0,0.1);
		   margin-bottom: 20px;
	   }
	   header h1 {
		   margin: 0;
		   font-size: 24px;
	   }
	   header a.button {
		   background-color: #4285F4;
		   color: #fff;
		   text-decoration: none;
		   padding: 8px 16px;
		   border-radius: 4px;
		   transition: background-color 0.3s ease;
	   }
	   header a.button:hover {
		   background-color: #357ABD;
	   }
	   .card {
		   background: white;
		   border-radius: 8px;
		   padding: 20px;
		   box-shadow: 0 4px 6px rgba(0,0,0,0.1);
		   margin-bottom: 20px;
	   }
	   .card ul {
		   list-style: none;
		   padding: 0;
	   }
	   .card li {
		   padding: 10px 0;
		   border-bottom: 1px solid #f4f4f4;
	   }
	   .card li:last-child {
		   border-bottom: none;
	   }
	   .card li strong {
		   display: inline-block;
		   width: 120px;
	   }
	   .profile-pic {
		   width: 50px;
		   height: 50px;
		   border-radius: 50%;
	   }
	</style>
</head>
<body>
   <div class="container">
	   <header>
	   	   <img src="` + user.Picture + `" alt="Profile Picture" class="profile-pic">
		   <h1>Dashboard</h1>
		   <a href="/logout" class="button">Logout</a>
	   </header>
	   <div class="card">
		   <h3>Informasi Pengguna</h3>
		   <ul>
			   <li><strong>ID:</strong> ` + user.ID + `</li>
			   <li><strong>Name:</strong> ` + user.Name + `</li>
			   <li><strong>Given Name:</strong> ` + user.GivenName + `</li>
			   <li><strong>Family Name:</strong> ` + user.FamilyName + `</li>
		   </ul>
	   </div>
   </div>
</body>
</html>
	`
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}
