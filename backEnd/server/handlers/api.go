package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/dgrijalva/jwt-go"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB

func InitDB() error {
	var err error
	db, err = sql.Open("mysql", "root:1234@tcp(localhost:3306)/tugas_pt_jpg")
	if err != nil {
		return err
	}

	err = db.Ping()
	if err != nil {
		return err
	}

	fmt.Println("Database connected")
	return nil
}

// register
type User struct {
	Email    string `json:"email"`
	Nama     string `json:"nama"`
	Password string `json:"password"`
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !isValidEmail(user.Email) {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	if isEmailRegistered(user.Email) {
		http.Error(w, "Email is already registered", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to encrypt password", http.StatusInternalServerError)
		return
	}
	user.Password = string(hashedPassword)

	err = SaveUser(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func SaveUser(user User) error {
	_, err := db.Exec("INSERT INTO user (email, nama, password) VALUES (?, ?, ?)", user.Email, user.Nama, user.Password)
	return err
}

func isValidEmail(email string) bool {
	// Regex for email validation
	regex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	match, _ := regexp.MatchString(regex, email)
	return match
}

func isEmailRegistered(email string) bool {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM user WHERE email = ?", email).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}

// login
type LoginResponse struct {
	Token   string `json:"token,omitempty"`
	Message string `json:"message,omitempty"`
}

type UserLogin struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Nama     string `json:"nama"`
	Password string `json:"password"`
}

type RefreshToken struct {
	UserID       int    `json:"user_id"`
	RefreshToken string `json:"refresh_token"`
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var user UserLogin
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Cari user berdasarkan email
	storedUser, err := getUserByEmail(user.Email)
	if err != nil {
		http.Error(w, "Email atau password Anda salah", http.StatusUnauthorized)
		return
	}

	// Bandingkan password yang dienkripsi
	err = bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(user.Password))
	if err != nil {
		http.Error(w, "Email atau password Anda salah", http.StatusUnauthorized)
		return
	}

	// Generate token akses
	accessToken, err := generateAccessToken(storedUser.ID)
	if err != nil {
		http.Error(w, "Gagal membuat token akses", http.StatusInternalServerError)
		return
	}

	// Generate refresh token
	refreshToken, err := generateRefreshToken(storedUser.ID)
	if err != nil {
		http.Error(w, "Gagal membuat refresh token", http.StatusInternalServerError)
		return
	}

	// Cek apakah data user sudah ada di dalam tabel auth
	_, err = getUserIDFromAuth(storedUser.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			// Jika belum ada, gunakan saveAuth
			err = saveAuth(storedUser.ID, accessToken, refreshToken)
			if err != nil {
				http.Error(w, "Gagal menyimpan token akses", http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, "Gagal memeriksa keberadaan data user", http.StatusInternalServerError)
			return
		}
	} else {
		// Jika sudah ada, gunakan updateAuth
		err = updateAuth(storedUser.ID, accessToken, refreshToken)
		if err != nil {
			http.Error(w, "Gagal memperbarui token akses", http.StatusInternalServerError)
			return
		}
	}

	// Kirimkan token akses sebagai respons
	response := LoginResponse{Token: accessToken}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getUserByEmail(email string) (UserLogin, error) {
	var user UserLogin
	err := db.QueryRow("SELECT * FROM user WHERE email = ?", email).Scan(&user.ID, &user.Email, &user.Nama, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return UserLogin{}, errors.New("User not found")
		}
		return UserLogin{}, err
	}
	return user, nil
}

func generateAccessToken(userID int) (string, error) {
	// Set waktu kedaluwarsa token (contoh: 1 jam)
	expirationTime := time.Now().Add(1 * time.Hour)

	// Buat token dengan claims
	claims := jwt.MapClaims{
		"id":  userID,
		"exp": expirationTime.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Tandatangani token dan kembalikan sebagai string
	accessToken, err := token.SignedString([]byte("access_secret"))
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

func generateRefreshToken(userID int) (string, error) {
	// Generate refresh token unik
	refreshToken := uuid.New().String()

	return refreshToken, nil
}

func saveAuth(userID int, accessToken string, refreshToken string) error {
	// Simpan access token dan refresh token ke dalam tabel auth
	_, err := db.Exec("INSERT INTO auth (id, access_token, refresh_token) VALUES (?, ?, ?)", userID, accessToken, refreshToken)
	if err != nil {
		return err
	}
	return nil
}

func updateAuth(userID int, accessToken string, refreshToken string) error {
	// Deklarasikan variabel err
	var err error

	// Update access token dan refresh token di tabel auth
	_, err = db.Exec("UPDATE auth SET access_token = ?, refresh_token = ? WHERE id = ?", accessToken, refreshToken, userID)
	if err != nil {
		return err
	}
	return nil
}

func getUserIDFromAuth(userID int) (bool, error) {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM auth WHERE id = ?)", userID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// refresh token

func RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	// Parse refresh token dari permintaan
	var requestData struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Gagal memproses permintaan", http.StatusBadRequest)
		return
	}

	// Validasi refresh token
	userID, err := validateRefreshToken(requestData.RefreshToken)
	if err != nil {
		http.Error(w, "Refresh token tidak valid", http.StatusUnauthorized)
		return
	}

	// Generate token akses baru
	accessToken, err := generateAccessToken(userID)
	if err != nil {
		http.Error(w, "Gagal membuat token akses", http.StatusInternalServerError)
		return
	}

	// Kirimkan token akses baru sebagai respons
	response := LoginResponse{Token: accessToken}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func validateRefreshToken(refreshToken string) (int, error) {
	// Cari userID berdasarkan refreshToken dari tabel auth
	var userID int
	err := db.QueryRow("SELECT id FROM auth WHERE refresh_token = ?", refreshToken).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, errors.New("Refresh token tidak valid")
		}
		return 0, err
	}
	return userID, nil
}

type Sales struct {
	Tanggal time.Time `json:"tanggal"`
	Jenis   string    `json:"jenis"`
	Nominal int       `json:"nominal"`
}

func InputSalesHandler(w http.ResponseWriter, r *http.Request) {
	// Mendapatkan token akses dari header permintaan
	accessToken := r.Header.Get("Authorization")
	if accessToken == "" {
		http.Error(w, "Token akses diperlukan", http.StatusUnauthorized)
		return
	}

	// Validasi token akses dan ambil ID pengguna dari tabel auth
	userID, err := getUserIDByAccessToken(accessToken)
	if err != nil {
		http.Error(w, "Token akses tidak valid", http.StatusUnauthorized)
		return
	}

	// Ambil nama pengguna dari tabel user
	nama, email, err := getUserDataByID(userID)
	if err != nil {
		http.Error(w, "Gagal mendapatkan nama pengguna", http.StatusInternalServerError)
		return
	}

	// Parsing data input sales dari body permintaan
	var sales Sales
	err = json.NewDecoder(r.Body).Decode(&sales)
	if err != nil {
		http.Error(w, "Gagal memproses permintaan", http.StatusBadRequest)
		return
	}

	// Simpan data input sales ke dalam database
	err = saveSales(nama, email, sales)
	if err != nil {
		http.Error(w, "Gagal menyimpan data sales", http.StatusInternalServerError)
		return
	}

	// Kirimkan respons sukses
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Data sales berhasil disimpan"))
}

func getUserIDByAccessToken(accessToken string) (int, error) {
	// Get the user ID from the auth table based on the access token
	var userID int
	err := db.QueryRow("SELECT id FROM auth WHERE access_token = ?", accessToken).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, errors.New("Invalid access token")
		}
		return 0, err
	}
	return userID, nil
}

func getUserDataByID(userID int) (string, string, error) {
	// Ambil nama dan email pengguna dari tabel user berdasarkan ID
	var nama, email string
	err := db.QueryRow("SELECT nama, email FROM user WHERE id = ?", userID).Scan(&nama, &email)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", "", errors.New("Pengguna tidak ditemukan")
		}
		return "", "", err
	}
	return nama, email, nil
}

func saveSales(userName string, email string, sales Sales) error {
	// Simpan data input sales ke dalam database
	_, err := db.Exec("INSERT INTO sales (tanggal, jenis, nominal, nama, email) VALUES (?, ?, ?, ?, ?)",
		sales.Tanggal, sales.Jenis, sales.Nominal, userName, email)
	if err != nil {
		return err
	}
	return nil
}

func ReportSalesHandler(w http.ResponseWriter, r *http.Request) {
	// Mendapatkan token akses dari header permintaan
	accessToken := r.Header.Get("Authorization")
	if accessToken == "" {
		http.Error(w, "Token akses diperlukan", http.StatusUnauthorized)
		return
	}

	// Validasi token akses dan ambil ID pengguna dari tabel auth
	userID, err := getUserIDByAccessToken(accessToken)
	if err != nil {
		http.Error(w, "Token akses tidak valid", http.StatusUnauthorized)
		return
	}

	// Mendapatkan nama pengguna dari tabel user berdasarkan ID
	userName, email, err := getUserDataByID(userID)
	if err != nil {
		http.Error(w, "Gagal mendapatkan nama pengguna", http.StatusInternalServerError)
		return
	}
	// Mendapatkan parameter tanggal awal dan tanggal akhir dari query string
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	// Validasi tanggal awal dan tanggal akhir
	if startDate == "" || endDate == "" {
		http.Error(w, "Parameter tanggal awal dan tanggal akhir diperlukan", http.StatusBadRequest)
		return
	}

	// Parsing tanggal awal dan tanggal akhir
	startDateTime, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		http.Error(w, "Format tanggal awal tidak valid", http.StatusBadRequest)
		return
	}
	endDateTime, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		http.Error(w, "Format tanggal akhir tidak valid", http.StatusBadRequest)
		return
	}

	// Query database untuk mendapatkan laporan penjualan
	rows, err := db.Query("SELECT user.nama, SUM(CASE WHEN sales.jenis = 'Barang' THEN 1 ELSE 0 END) AS jumlah_transaksi_barang, SUM(CASE WHEN sales.jenis = 'Jasa' THEN 1 ELSE 0 END) AS jumlah_transaksi_jasa, SUM(CASE WHEN sales.jenis = 'Barang' THEN sales.nominal ELSE 0 END) AS nominal_transaksi_barang, SUM(CASE WHEN sales.jenis = 'Jasa' THEN sales.nominal ELSE 0 END) AS nominal_transaksi_jasa FROM sales JOIN user ON sales.email = user.email WHERE sales.tanggal BETWEEN ? AND ? AND sales.email = ? GROUP BY user.nama", startDateTime.Format("2006-01-02"), endDateTime.Format("2006-01-02"), email)
	if err != nil {
		http.Error(w, "Gagal mendapatkan laporan penjualan", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Membuat file Excel
	file := excelize.NewFile()
	sheetName := "Report"
	file.NewSheet(sheetName)

	file.DeleteSheet("Sheet1")
	// Menulis header
	headers := []string{"User", "Jumlah Hari Kerja", "Jumlah Transaksi Barang", "Jumlah Transaksi Jasa", "Nominal Transaksi Barang", "Nominal Transaksi Jasa"}
	for col, header := range headers {
		cell := excelize.ToAlphaString(col) + "7"
		file.SetCellValue(sheetName, cell, header)
	}

	// Menulis data laporan penjualan
	row := 8
	for rows.Next() {
		var userName string
		var numTransactionsBarang, numTransactionsJasa int
		var nominalTransactionsBarang, nominalTransactionsJasa float64
		err := rows.Scan(&userName, &numTransactionsBarang, &numTransactionsJasa, &nominalTransactionsBarang, &nominalTransactionsJasa)
		if err != nil {
			http.Error(w, "Gagal membaca hasil laporan penjualan", http.StatusInternalServerError)
			return
		}

		// Menghitung jumlah hari kerja
		numWorkingDays := calculateWorkingDays(startDateTime, endDateTime, email)

		// Menulis data ke setiap kolom
		file.SetCellValue(sheetName, fmt.Sprintf("A%d", row), userName)
		file.SetCellValue(sheetName, fmt.Sprintf("B%d", row), numWorkingDays)
		strNominalTransactionsBarang := FormatNominal(nominalTransactionsBarang)
		strNominalTransactionsJasa := FormatNominal(nominalTransactionsJasa)
		file.SetCellValue(sheetName, fmt.Sprintf("C%d", row), numTransactionsBarang)
		file.SetCellValue(sheetName, fmt.Sprintf("D%d", row), numTransactionsJasa)
		file.SetCellValue(sheetName, fmt.Sprintf("E%d", row), strNominalTransactionsBarang)
		file.SetCellValue(sheetName, fmt.Sprintf("F%d", row), strNominalTransactionsJasa)

		row++
	}

	boldStyle, _ := file.NewStyle(`{"font":{"bold":true}}`)
	// Menulis requestor
	file.SetCellValue(sheetName, "A1", "Requestor")
	file.SetCellStyle(sheetName, "A1", "A1", boldStyle)
	file.SetCellValue(sheetName, "B1", fmt.Sprintf("%s(%s)", userName, email))

	// Menulis parameter
	file.SetCellValue(sheetName, "A3", "Parameter")
	file.SetCellStyle(sheetName, "A3", "A3", boldStyle)
	file.SetCellValue(sheetName, "A4", "Start Date")
	file.SetCellStyle(sheetName, "A4", "A4", boldStyle)
	file.SetCellValue(sheetName, "B4", startDateTime.Format("02 January 2006"))
	file.SetCellValue(sheetName, "A5", "End Date")
	file.SetCellStyle(sheetName, "A5", "A5", boldStyle)
	file.SetCellValue(sheetName, "B5", endDateTime.Format("02 January 2006"))
	file.SetCellStyle(sheetName, "A6", "A6", boldStyle)
	file.SetCellStyle(sheetName, "B6", "B6", boldStyle)
	file.SetCellStyle(sheetName, "C6", "C6", boldStyle)
	file.SetCellStyle(sheetName, "D6", "D6", boldStyle)
	file.SetCellStyle(sheetName, "E6", "E6", boldStyle)
	file.SetCellStyle(sheetName, "F6", "F6", boldStyle)
	// Mengatur dimensi kolom
	file.SetColWidth(sheetName, "A", "E", 15)

	// Mengatur style
	style, _ := file.NewStyle(`{"alignment":{"horizontal":"center"}}`)
	file.SetCellStyle(sheetName, "A1", fmt.Sprintf("E%d", row), style)

	// Menyimpan file Excel ke response
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=report.xlsx")
	err = file.Write(w)
	if err != nil {
		http.Error(w, "Gagal menyimpan file Excel", http.StatusInternalServerError)
		return
	}
}

func calculateWorkingDays(startDate, endDate time.Time, email string) int {
	// Query untuk menghitung jumlah hari kerja berdasarkan data input penjualan
	query := `
        SELECT COUNT(DISTINCT DATE(tanggal)) AS workingDays
        FROM sales
        WHERE tanggal BETWEEN ? AND ? AND email = ?
    `

	var workingDays int
	err := db.QueryRow(query, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"), email).Scan(&workingDays)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0 // Tidak ada data, tidak ada hari kerja
		}
		// Tangani kesalahan lainnya jika perlu
		return 0
	}

	return workingDays
}

func FormatNominal(value float64) string {
	// Konversi nilai float64 ke string tanpa digit desimal
	strValue := strconv.FormatFloat(value, 'f', 0, 64)

	// Pisahkan angka menjadi tiga digit terakhir dan sisanya
	length := len(strValue)
	remainder := length % 3
	result := strValue[:remainder]

	// Tambahkan pemisah ribuan setiap tiga digit dan gabungkan dengan angka sisanya
	for i := remainder; i < length; i += 3 {
		if result != "" {
			result += ","
		}
		result += strValue[i : i+3]
	}

	return result
}
