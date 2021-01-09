package main

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"io/ioutil"
	"math/rand"

	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

const picturesPath = "/pictures/"

var db sql.DB

const databaseName = "doggo_photos_db"
const userTable = "users"

// add id uniqueness
const userSchema = "(username varchar(20), passhash binary(60), loginKey varchar(255));"
const photoTable = "photos"
const photoSchema = "(username varchar(20), id varchar(20), date datetime);"

func createDB() {
	db, err := sql.Open("mysql", "root:password@tcp(db:3306)/")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	db.Exec("DROP DATABASE " + databaseName + ";")
	_, err = db.Exec("CREATE DATABASE " + databaseName + ";")
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("USE " + databaseName)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("CREATE TABLE " + userTable + " " + userSchema)
	if err != nil {
		panic(err)
	}
	_, err = db.Exec("CREATE TABLE " + photoTable + " " + photoSchema)
	if err != nil {
		panic(err)
	}

	//  Insert admin users

	// _, err = createNewAccount("admin", "adminpassword")
	// if err != nil {
	// 	panic(err)
	// }

}

func createNewAccount(username string, password string) (string, error) {
	db, err := getDBConn()

	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	usernameDB := ""
	sqlFindUser := `SELECT username FROM users WHERE username=?`
	db.QueryRow(sqlFindUser, username).Scan(&usernameDB)
	if usernameDB != "" {
		return "", errors.New("User already exists")
	}

	loginKey := generateToken(username)
	passBytes := []byte(password)
	passhash, err := bcrypt.GenerateFromPassword(passBytes, bcrypt.MinCost)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(len(passhash))

	fmt.Print(username, password)
	sqlStatement := `INSERT INTO users (username, passhash, loginkey) VALUES (?, ?, ?)`
	_, err = db.Query(sqlStatement, username, passhash, loginKey)
	if err != nil {
		return "", err
	}
	return loginKey, nil
}

func GetUserName(loginKey string) (string, error) {
	db, err := getDBConn()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	println(loginKey)
	username := ""
	sqlFindUser := `SELECT username FROM users WHERE loginKey=?`
	db.QueryRow(sqlFindUser, loginKey).Scan(&username)
	println(username)
	if username == "" {
		return "", errors.New("Please login again.")
	}
	return username, nil
}

func accountLogin(username string, password string) (string, error) {
	db, err := getDBConn()

	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var passwordHash []byte
	sqlStatement := `SELECT passhash FROM users WHERE username=?`
	db.QueryRow(sqlStatement, username).Scan(&passwordHash)
	err = bcrypt.CompareHashAndPassword(passwordHash, []byte(password))
	if err != nil {
		return "", err
	}
	newLoginToken := generateToken(username)

	sqlUpdateToken := `UPDATE users SET loginKey=? WHERE username=?`
	db.QueryRow(sqlUpdateToken, newLoginToken, username)

	return newLoginToken, nil
}

func getDBConn() (*sql.DB, error) {
	db, err := sql.Open("mysql", "root:password@tcp(db:3306)/"+databaseName)
	if err != nil {
		return nil, err
	}
	// See "Important settings" section.
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	return db, nil
}

func main() {
	http.HandleFunc("/", HelloServer)
	http.HandleFunc("/admin/createdb", CreateDBServe)
	http.HandleFunc("/pictures/", PicturesServer)
	http.HandleFunc("/api/account/create", CreateAccountServe)
	http.HandleFunc("/api/account/login", LoginServe)
	http.HandleFunc("/api/account/user", TokenLoginServe)
	http.HandleFunc("/api/photos/upload", UploadPhotosService)
	http.ListenAndServe(":3000", nil)
}

// Creates the DB to start the service
func CreateDBServe(w http.ResponseWriter, r *http.Request) {
	createDB()
}

const photoSizeLimit = (32 << 12)

func UploadPhotosService(w http.ResponseWriter, r *http.Request) {
	fmt.Print("starting photo upload\n")
	db, err := getDBConn()
	defer db.Close()
	if err != nil {
		panic(err)
	}

	r.ParseMultipartForm(32 << 20)
	var errors []string
	fileHeaders := r.MultipartForm.File["photos"]
	loginKey := r.Form.Get("loginKey")
	fmt.Printf("loginKey: %s\n", loginKey)
	username, err := GetUserName(loginKey)
	if err != nil {
		fmt.Printf("Error: cant find user")
		return
	}

	for _, fh := range fileHeaders {
		fmt.Printf("%s\n", fh.Filename)
	}

	for _, fileHeader := range fileHeaders {
		filename := fileHeader.Filename

		if fileHeader.Size > photoSizeLimit {
			errors = append(errors, filename+" cannot be uploaded it is too big")
			continue
		}

		if !strings.HasSuffix(filename, ".jpg") {
			errors = append(errors, filename+" cannot be uploaded it is not a jpg")
			continue
		}

		f, err := fileHeader.Open()
		if err != nil {
			errors = append(errors, filename+" cannot be uploaded an error occured 1")
			continue
		}
		defer f.Close()

		fileBytes, err := ioutil.ReadAll(f)
		if err != nil {
			errors = append(errors, filename+" cannot be uploaded an error occured 2")
			continue
		}

		photo := Photo{Data: fileBytes, User: username, ID: filename}

		err = photo.UploadImage(db)
		if err != nil {
			errors = append(errors, filename+" cannot be uploaded an error: "+err.Error())
			continue
		}
	}
	fmt.Printf("%+v\n", errors)
}

func TokenLoginServe(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var data tokenLoginBody
	json.NewDecoder(r.Body).Decode(&data)
	fmt.Printf("%+v", data)
	username, err := GetUserName(data.LoginKey)
	if err != nil {
		json.NewEncoder(w).Encode(errorBody{Error: "Unauthorized"})
		return
	}

	json.NewEncoder(w).Encode(userRequest{Username: username})
}

// Login API
func LoginServe(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var data createAccountBody
	json.NewDecoder(r.Body).Decode(&data)

	loginKey, err := accountLogin(data.Username, data.Password)
	if err != nil {
		json.NewEncoder(w).Encode(errorBody{Error: "Incorrect Password"})
		return
	}

	json.NewEncoder(w).Encode(loginResponse{LoginKey: loginKey})
}

func generateToken(username string) string {
	usernameBytes, err := bcrypt.GenerateFromPassword([]byte(username), bcrypt.MinCost)
	if err != nil {
		log.Fatal(err)
	}
	tokenBytes := []byte(string(rand.Uint32()))
	usernameBytes = append(usernameBytes, tokenBytes...)
	tokenStr := base64.StdEncoding.EncodeToString(usernameBytes)
	return tokenStr
}

func insertUser(username string, passhash []byte, loginKey string) error {
	sqlStatement := `INSERT INTO users (username, passhash, loginkey) VALUES (?, ?, '?)`
	_, err := db.Query(sqlStatement, username, passhash, loginKey)
	return err
}

// CreateAccountServe API
func CreateAccountServe(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var data createAccountBody
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	loginKey, err := createNewAccount(data.Username, data.Password)
	if err != nil {
		json.NewEncoder(w).Encode(errorBody{Error: err.Error()})
		return
	}

	json.NewEncoder(w).Encode(loginResponse{LoginKey: loginKey})
}

// GetPicture API
func GetPicture(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	token := query.Get("key")
	if token != "" {

	} else {
		fmt.Fprint(w, "Cannot access")
	}

}

// PicturesServer serves images
func PicturesServer(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	user := parts[2]
	imageID := parts[3]

	_ = user
	_ = imageID

	log.Print(user)
	log.Print(imageID)

	query := r.URL.Query()
	loginKey := query.Get("key")
	if loginKey == "" {
		fmt.Fprintf(w, "Access Restricted")
		return
	}

	username := ""
	sqlStatement := `SELECT username FROM users WHERE loginKey=?`
	db.QueryRow(sqlStatement, loginKey).Scan(&username)
	if username != user {
		fmt.Fprint(w, "Access Restricted")
		return
	}

	err := writePicture(w, username, imageID)
	if err != nil {
		fmt.Fprint(w, "Access Restricted")
	}
}

func (photo *Photo) UploadImage(db *sql.DB) error {
	// "/picture/{user}/{id}"
	// add uniqueness

	// Check if the picture is saved
	filePath := path.Join(picturesPath, photo.User, photo.ID)
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		return errors.New("File already exists")
	}
	err := ioutil.WriteFile(filePath, photo.Data, 0644)
	if err != nil {
		return err
	}

	const insertSql = `INSERT INTO photos (username, id, date) VALUES (?, ?, now())`
	_, err = db.Query(insertSql, photo.User, photo.ID)

	if err != nil {
		removeErr := os.Remove(filePath)
		if removeErr != nil {
			panic(removeErr)
		}
		return err
	}

	return nil
}

type Photo struct {
	User string
	ID   string
	Data []byte
}

func HelloServer(w http.ResponseWriter, r *http.Request) {
	db, err := getDBConn()
	defer db.Close()
	if err != nil {
		panic(err)
	}
	res, err := db.Query("SELECT * FROM users")
	if err != nil {
		panic(err)
	}
	for res.Next() {
		var username string
		var password string

		//err := res.Scan(&username, &password)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%v %v\n", username, password)
	}
	fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
}

func writePicture(w http.ResponseWriter, user string, fileName string) error {
	f, err := os.Open(path.Join(picturesPath, user, fileName))
	if err != nil {
		return err
	}
	defer f.Close()
	w.Header().Set("Content-Type", "image/jpg")
	io.Copy(w, f)
	return nil
}

/*
Useres:
username, password (hashed), hash_key

Groups:

Photo Access:
picture_id, username, access_ability


*/
