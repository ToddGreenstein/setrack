package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/couchbase/gocb"
	gen "github.com/couchbaselabs/setrack/datagen"
)

type Session struct {
	User User
}

// User Struct Type
type User struct {
	Type      string `json:"_type,omitempty"`
	ID        string `json:"_id"`
	CreatedOn string `json:"createdON"`
	Name      string `json:"name"`
	Address   struct {
		Street  string `json:"street"`
		City    string `json:"city"`
		State   string `json:"state"`
		Zip     string `json:"zip"`
		Country string `json:"country"`
	} `json:"address"`
	Region   string `json:"region"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Phone    string `json:"phone"`
	Company  string `json:"company"`
	Active   bool   `json:"active"`
}

func (u *Session) Login(email string, password string) (*User, error) {
	// Login a single user.   Uses a N1QL query to retrieve a single instance
	// and return the user object back to the front end application.
	myQuery := gocb.NewN1qlQuery("SELECT _id,active,address,company,createdON," +
		"email,name,`password`,phone FROM `comply` " +
		"WHERE _type = 'User' and email='" + email + "' ").Consistency(gocb.RequestPlus)

	rows, err := bucket.ExecuteN1qlQuery(myQuery, nil)
	if err != nil {
		return nil, err
	}

	// Interfaces for handling streaming return values
	var user User

	// Stream the values returned from the query into an untyped and unstructred
	// array of interfaces
	err = rows.One(&user)
	if err != nil {
		return nil, errors.New("User Not Found")
	}

	// Check the password, compare.  If correct return the user object.
	// NOTE: for demonstration purposes, this approach is NOT SECURE
	if user.Password == password {
		return &user, nil
	}
	return nil, errors.New("Password is invalid")
}

func (u *Session) Create() (*User, error) {
	// Create a new user instance.   This method uses the User struct within the
	// Session struct when it's passed in.   It adds the specific item fields
	// Not set by the rest endpoint in the JSON body, and then stores within
	// the appropriate bucket.
	u.User.Type = "User"
	u.User.ID = GenUUID()
	u.User.CreatedOn = time.Now().Format(time.RFC3339)
	u.User.Active = true

	// Store in couchbase, check for error.   If no errors, return the user object
	// back to the front end application.
	_, err := bucket.Upsert(u.User.Name, u.User, 0)
	if err != nil {
		return nil, err
	}
	return &u.User, nil
}

func (u *Session) Retrieve(id string) (*User, error) {
	// Retrieve a single user instance from the id.  Uses a get operation against
	// the database and returns the User object to the front end application.
	_, err := bucket.Get(id, &u.User)
	if err != nil {
		return nil, err
	}
	return &u.User, nil
}

func (u *Session) RetrieveAll() ([]User, error) {
	// Retrieves all users from the database.   Does not implement filtering
	// and returns the array of users back to the front end application.
	myQuery := gocb.NewN1qlQuery("SELECT * FROM `comply` " +
		"WHERE _type = 'User'")
	rows, err := bucket.ExecuteN1qlQuery(myQuery, nil)
	if err != nil {
		return nil, err
	}

	// Wrapper struct needed for parsing results from N1QL
	// The results will always come wrapped in the "bucket" name
	type wrapUser struct {
		User User `json:"comply"`
	}

	// Temporary variables to parse the results.
	var row wrapUser
	var curUsers []User

	// Parse the n1ql results and build the array of Users to return to the
	// front end application
	for rows.Next(&row) {
		curUsers = append(curUsers, row.User)
	}
	return curUsers, nil
}

func AddUsers() (bool, error) {

	eng := []string{
		"Addison Smith",
		"Shirley Martinez",
		"Chloe Jones",
		"Charlotte Smith",
		"Sofia Martin",
		"Liam Miller",
		"Jacob Thompson",
		"Paul Moore",
		"Matthew Harris",
		"Elijah Robninson"}

	// Local Activity Struct
	var user Session

	for _, se := range eng {

		// Externally Visible Fields
		fullname := strings.Split(se, " ")
		email := fullname[0] + "@couchbase.com"
		user.User.Active = true
		user.User.Name = se
		user.User.Address.Street = gen.Street()
		user.User.Address.State = gen.State(gen.Large)
		user.User.Address.City = gen.City()
		user.User.Address.Country = "USA"
		user.User.Address.Zip = gen.PostalCode("us")
		user.User.Email = email

		user.User.Password = encrypt([]byte(key), gen.SillyName())
		user.User.Region = gen.Region()
		user.User.Company = "Couchbase Inc"

		_, err := user.Create()
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

func addBase64Padding(value string) string {
	m := len(value) % 4
	if m != 0 {
		value += strings.Repeat("=", 4-m)
	}

	return value
}

func removeBase64Padding(value string) string {
	return strings.Replace(value, "=", "", -1)
}

func Pad(src []byte) []byte {
	padding := aes.BlockSize - len(src)%aes.BlockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

func Unpad(src []byte) ([]byte, error) {
	length := len(src)
	unpadding := int(src[length-1])

	if unpadding > length {
		return nil, errors.New("unpad error. This could happen when incorrect encryption key is used")
	}

	return src[:(length - unpadding)], nil
}

func encrypt(key []byte, text string) string {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "LOAD GENERATOR DATA, NO ENCRYPTION"
	}

	msg := Pad([]byte(text))
	ciphertext := make([]byte, aes.BlockSize+len(msg))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "LOAD GENERATOR DATA, NO ENCRYPTION"
	}

	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(msg))
	finalMsg := removeBase64Padding(base64.URLEncoding.EncodeToString(ciphertext))
	return finalMsg
}

func decrypt(key []byte, text string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	decodedMsg, err := base64.URLEncoding.DecodeString(addBase64Padding(text))
	if err != nil {
		return "", err
	}

	if (len(decodedMsg) % aes.BlockSize) != 0 {
		return "", errors.New("blocksize must be multipe of decoded message length")
	}

	iv := decodedMsg[:aes.BlockSize]
	msg := decodedMsg[aes.BlockSize:]

	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(msg, msg)

	unpadMsg, err := Unpad(msg)
	if err != nil {
		return "", err
	}

	return string(unpadMsg), nil
}
