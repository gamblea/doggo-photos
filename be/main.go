package main

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"errors"
	"image"
	"io/ioutil"
	"math/rand"

	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"

	_ "image/jpeg"
)

// Photos file system storage path mapped to a volume in docker-compose
const picturesPath = "/pictures/"

const databaseName = "doggo_photos_db"
const userTable = "users"
const userSchema = "(username varchar(20), passhash binary(60), loginKey varchar(255), UNIQUE(username));"
const photoTable = "photos"
const photoSchema = "(username varchar(20), id varchar(60), date datetime, width int DEFAULT 4, height int DEFAULT 3, UNIQUE(username, id));"

const photoSizeLimit = (32 << 16)

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
		return "", err
	}

	err = os.Mkdir(path.Join(picturesPath, username), 0777)
	if err != nil {
		return "", err
	}

	sqlStatement := `INSERT INTO users (username, passhash, loginkey) VALUES (?, ?, ?)`
	_, err = db.Query(sqlStatement, username, passhash, loginKey)
	if err != nil {
		return "", err
	}

	return loginKey, nil
}

// GetUserPhotos returns the photos that a user owns
func GetUserPhotos(loginKey string) (*UserDataResponse, error) {
	db, err := getDBConn()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var dataRes UserDataResponse
	sqlFindUser := `SELECT users.username, photos.ID, photos.date, photos.width, photos.height FROM users INNER JOIN photos ON users.username = photos.username WHERE users.loginKey=? ORDER BY  photos.date DESC ;`
	res, err := db.Query(sqlFindUser, loginKey)
	if err != nil {
		return nil, err
	}

	defer res.Close()

	for res.Next() {
		var photo FEPhoto
		photoID := ""
		username := ""

		err := res.Scan(&username, &photoID, &photo.Date, &photo.Width, &photo.Height)
		if err != nil {
			return nil, err
		}

		photo.Src = PhotoIDToSrc(username, photoID)
		dataRes.Photos = append(dataRes.Photos, photo)
	}

	if len(dataRes.Photos) == 0 {
		return nil, errors.New("No photos yet :(")
	}
	return &dataRes, nil
}

// PhotoIDToSrc transforms a username and photoID to a src path the user
// can access the photo at
func PhotoIDToSrc(username string, photoID string) string {
	return path.Join(picturesPath, username, photoID)
}

// accountLogin logs a user with their credentials and returns a login token
func accountLogin(username string, password string) (string, error) {
	db, err := getDBConn()

	if err != nil {
		return "", err
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

// All the handlers for the backend API
func main() {
	http.HandleFunc("/admin/createdb", CreateDBServe)
	http.HandleFunc("/pictures/", PicturesServer)
	http.HandleFunc("/api/account/photos", ServeUserPhotos)
	http.HandleFunc("/api/account/create", CreateAccountServe)
	http.HandleFunc("/api/account/login", LoginServe)
	http.HandleFunc("/api/account/user", TokenLoginServe)
	http.HandleFunc("/api/photos/upload", UploadPhotosService)
	http.ListenAndServe(":5000", nil)
}

// CreateDBServe creates the DB to start the service
// Add authentication
func CreateDBServe(w http.ResponseWriter, r *http.Request) {
	createDB()
}

// Builds and resets the database
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
	os.RemoveAll(picturesPath)
}

// PicturesServer serves images
func PicturesServer(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	user := parts[2]
	imageID := parts[3]

	query := r.URL.Query()
	loginKey := query.Get("key")
	if loginKey == "" {
		fmt.Fprintf(w, "Access Restricted no key")
		return
	}

	username, _ := GetUserName(loginKey)
	if username != user {
		fmt.Fprint(w, "Access Restricted")
		return
	}

	err := writePicture(w, username, imageID)
	if err != nil {
		fmt.Fprint(w, "Access Restricted")
	}
}

// writePicture writes a photo to be sent to the user
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

// ServeUserPhotos serves requests to get the photo metadata of a user
func ServeUserPhotos(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var data tokenLoginBody
	json.NewDecoder(r.Body).Decode(&data)

	res, err := GetUserPhotos(data.LoginKey)
	if err != nil {
		json.NewEncoder(w).Encode(ErrorBody{Error: err.Error()})
		return
	}

	json.NewEncoder(w).Encode(res)
}

// GetUserName returns the name of the user with the respective login key
func GetUserName(loginKey string) (string, error) {
	db, err := getDBConn()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	username := ""
	sqlFindUser := `SELECT username FROM users WHERE loginKey=?`
	db.QueryRow(sqlFindUser, loginKey).Scan(&username)
	if username == "" {
		return "", errors.New("Please login again")
	}
	return username, nil
}

// UploadPhotosService handles uploading photos form submittions
func UploadPhotosService(w http.ResponseWriter, r *http.Request) {
	db, err := getDBConn()
	defer db.Close()
	if err != nil {
		panic(err)
	}

	r.ParseMultipartForm(32 << 20)
	var errors []string
	fileHeaders := r.MultipartForm.File["photos"]
	loginKey := r.Form.Get("loginKey")
	username, err := GetUserName(loginKey)

	// The user does not exists do not upload files
	if err != nil {
		return
	}

	for _, fileHeader := range fileHeaders {
		filename := fileHeader.Filename

		if fileHeader.Size > photoSizeLimit {
			errors = append(errors, filename+" cannot be uploaded it is too big")
			continue
		}

		if !(strings.HasSuffix(filename, ".jpg") || strings.HasSuffix(filename, ".jpeg")) {
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

		image, _, err := image.DecodeConfig(bytes.NewReader(fileBytes))
		if err != nil {
			errors = append(errors, err.Error())
			continue
		}

		width := 0
		height := 0
		if image.Width > image.Height {
			width = 4
			height = 3
		} else if image.Width < image.Height {
			width = 3
			height = 4
		} else {
			width = 1
			height = 1
		}

		photo := Photo{Data: fileBytes, User: username, ID: filename, Width: width, Height: height}
		err = photo.UploadImage(db)
		if err != nil {
			errors = append(errors, filename+" cannot be uploaded an error: "+err.Error())
			continue
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ErrorBody{Error: strings.Join(errors, ",")})
}

// UploadImage writes a photo to database and to file system
func (photo *Photo) UploadImage(db *sql.DB) error {
	filePath := path.Join(picturesPath, photo.User, photo.ID)

	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		return errors.New("File already exists")
	}
	err := ioutil.WriteFile(filePath, photo.Data, 0666)
	if err != nil {
		return err
	}

	const insertSQL = `INSERT INTO photos (username, id, date, width, height) VALUES (?, ?, now(), ?, ?)`
	_, err = db.Query(insertSQL, photo.User, photo.ID, photo.Width, photo.Height)

	if err != nil {
		removeErr := os.Remove(filePath)
		if removeErr != nil {
			panic(removeErr)
		}
		return err
	}

	return nil
}

// TokenLoginServe validates a token and returns the username of the respective user
func TokenLoginServe(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var data tokenLoginBody
	json.NewDecoder(r.Body).Decode(&data)

	username, err := GetUserName(data.LoginKey)
	if err != nil {
		json.NewEncoder(w).Encode(ErrorBody{Error: "Unauthorized"})
		return
	}

	json.NewEncoder(w).Encode(UserRequest{Username: username})
}

// LoginServe handles login requests
func LoginServe(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var data createAccountBody
	json.NewDecoder(r.Body).Decode(&data)

	loginKey, err := accountLogin(data.Username, data.Password)
	if err != nil {
		json.NewEncoder(w).Encode(ErrorBody{Error: "Incorrect Password"})
		return
	}

	json.NewEncoder(w).Encode(LoginResponse{LoginKey: loginKey})
}

// generateToken generates login token to be cached by the browser to request further resouces
func generateToken(username string) string {
	usernameBytes, err := bcrypt.GenerateFromPassword([]byte(username), bcrypt.MinCost)
	if err != nil {
		log.Fatal(err)
	}

	tokenBytes := []byte(string(rand.Uint32()))
	usernameBytes = append(usernameBytes, tokenBytes...)
	tokenStr := base64.StdEncoding.EncodeToString(usernameBytes)
	// Remove '+' as they conflict with query string syntax
	tokenStr = strings.Replace(tokenStr, "+", "", -1)

	return tokenStr
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
		json.NewEncoder(w).Encode(ErrorBody{Error: err.Error()})
		return
	}

	json.NewEncoder(w).Encode(LoginResponse{LoginKey: loginKey})
}
