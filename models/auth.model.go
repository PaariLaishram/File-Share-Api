package models

import (
	"FileShare/database"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type LoginBody struct {
	Email    *string `json:"email,omitempty"`
	Password *string `json:"password,omitempty"`
}

type LoginResult struct {
	User         *int    `json:"user,omitempty"`
	Email        *string `json:"email,omitempty"`
	Password     *string `json:"password,omitempty"`
	AccessToken  *string `json:"accessToken"`
	RefreshToken *string `json:"refreshToken,omitempty"`
}

type TokenPayload struct {
	Iss   string  `json:"iss"`
	Sub   int     `json:"sub"`
	Email *string `json:"email,omitempty"`
	Exp   int     `json:"exp"`
	Iat   int     `json:"iat"`
	Jti   *string `json:"jti,omitempty"`
}

type TokenHeader struct {
	Typ string `json:"typ"`
	Alg string `json:"alg"`
}

type RefreshToken struct {
	User         *int    `json:"user,omitempty"`
	RefreshToken *string `json:"refreshToken,omitempty"`
	AccessToken  *string `json:"accessToken,omitempty"`
	Jti          *string `json:"jti,omitempty"`
}

// This function is used to login
func (login *LoginBody) Login() (bool, string, LoginResult) {
	message := "Invalid Credentials"
	var user LoginResult
	user_fields := []any{&user.User, &user.Email, &user.Password}
	user_filter := []any{SanitizeData(login.Email)}
	get_user_by_email := get_user_query + user_email_filter
	user_result, err := database.RunQuery(get_user_by_email, user_fields, user_filter, &user)
	if err != nil {
		log.Error("Error getting user:", err.Error())
		return false, message, user
	}
	if len(user_result) == 0 {
		return false, message, user
	}
	encrypted_password := SanitizeData(user_result[0].(LoginResult).Password)
	err = bcrypt.CompareHashAndPassword([]byte(encrypted_password), []byte(SanitizeData(login.Password)))
	if err != nil {
		log.Info("Invalid Password")
		return false, message, user
	}
	tokenHeader := getTokenHeader()
	iss := os.Getenv("ISS")
	accessTokenPayload := TokenPayload{
		Iss:   iss,
		Sub:   SanitizeData(user.User),
		Email: user.Email,
		Exp:   int(time.Now().Add(5 * time.Minute).Unix()),
		Iat:   int(time.Now().Unix()),
	}

	secret := os.Getenv("ACCESS_SECRET")
	access_token, err := getJwtToken(*tokenHeader, accessTokenPayload, secret)
	if err != nil {
		return false, message, user
	}
	jti := getJti()
	refresh_token_expires_at := int(time.Now().Add(30 * time.Minute).Unix())
	refreshTokenPayload := TokenPayload{
		Iss: iss,
		Sub: SanitizeData(user.User),
		Exp: refresh_token_expires_at,
		Iat: int(time.Now().Unix()),
		Jti: &jti,
	}

	refresh_token_secret := os.Getenv("REFRESH_SECRET")

	refresh_token, err := getJwtToken(*tokenHeader, refreshTokenPayload, refresh_token_secret)
	if err != nil {
		return false, message, user
	}

	insert_fields := []any{SanitizeData(user.User), refresh_token, jti, refresh_token_expires_at}
	_, err = database.RunInsert(insert_refresh_token_query, insert_fields)
	if err != nil {
		log.Error("Error inserting refresh token: ", err.Error())
		return false, message, user
	}

	user.AccessToken = &access_token
	user.RefreshToken = &refresh_token
	user.Password = nil
	return true, "", user
}

func Logout(user string) (bool, string) {
	delete_refresh_token_query := `DELETE from refresh_tokens WHERE user = ?`
	update, err := database.RunUpdate(delete_refresh_token_query, []any{user})
	if err != nil {
		return false, "Error logging out user"
	}
	return update, ""
}

func decodeJwtToken(token string) (*TokenHeader, *TokenPayload, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, nil, errors.New("error decoding jwt token: should contain 3 parts")
	}
	base64_header, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		log.Error("Error decoding:", err.Error())
		return nil, nil, err
	}
	base64_payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		log.Error("Error decoding:", err.Error())
		return nil, nil, err
	}
	var header TokenHeader
	var payload TokenPayload

	if err := json.Unmarshal(base64_header, &header); err != nil {
		log.Error("Error unmarshalling: ", err.Error())
		return nil, nil, err
	}

	if err := json.Unmarshal(base64_payload, &payload); err != nil {
		log.Error("Error unmarshalling:", err.Error())
		return nil, nil, err
	}
	return &header, &payload, nil
}

func getTokenHeader() *TokenHeader {
	return &TokenHeader{
		Typ: "JWT",
		Alg: "HS256",
	}
}

func ValidateRefreshToken(token string) (bool, RefreshToken) {
	var db = database.DB
	var refresh_token RefreshToken
	header, payload, err := decodeJwtToken(token)
	if err != nil {
		log.Error("Error decoding jwt token: ", err.Error())
		return false, refresh_token
	}

	if payload.Iss != os.Getenv("ISS") {
		return false, refresh_token
	}
	if payload.Exp < int(time.Now().Unix()) {
		log.Warn("Refresh token has expired. User must login again.")
		return false, refresh_token
	}
	user := payload.Sub
	var token_exist_query = ` DELETE FROM refresh_tokens WHERE user = ? AND jti = ? `
	var token_filter = []interface{}{user, payload.Jti}

	tx, err := db.Begin()
	if err != nil {
		log.Error("Error beginning transaction: ", err.Error())
		return false, refresh_token
	}
	defer tx.Rollback()
	delete_refresh_token, err := tx.Exec(token_exist_query, token_filter...)
	if err != nil {
		log.Error("Error in deleting refresh token: ", err.Error())
		return false, refresh_token
	}
	deleted_rows, err := delete_refresh_token.RowsAffected()
	if err != nil {
		log.Error("Error in getting deleted rows: ", deleted_rows)
		return false, refresh_token
	}
	if deleted_rows == 0 {
		log.Warn("No refresh token found, unauthorized user")
		return false, refresh_token
	}
	access_token_payload := TokenPayload{
		Iss: os.Getenv("ISS"),
		Sub: user,
		Exp: int(time.Now().Add(5 * time.Minute).Unix()),
		Iat: int(time.Now().Unix()),
	}

	jti := getJti()
	refresh_token_expires_at := int(time.Now().Add(30 * time.Minute).Unix())
	refresh_token_payload := TokenPayload{
		Iss: os.Getenv("ISS"),
		Sub: user,
		Exp: refresh_token_expires_at,
		Iat: int(time.Now().Unix()),
		Jti: &jti,
	}

	access_jwt_token, err := getJwtToken(*header, access_token_payload, os.Getenv("ACCESS_SECRET"))
	if err != nil {
		log.Error("Error getting access jwt token: ", err.Error())
		return false, refresh_token
	}

	refresh_jwt_token, err := getJwtToken(*header, refresh_token_payload, os.Getenv("REFRESH_SECRET"))
	if err != nil {
		log.Error("Error getting refresh jwt token: ", err.Error())
		return false, refresh_token
	}

	insert_fields := []any{user, refresh_jwt_token, jti, refresh_token_expires_at}
	_, err = tx.Exec(insert_refresh_token_query, insert_fields...)
	if err != nil {
		log.Error("Error inserting refresh token: ", err.Error())
		return false, refresh_token
	}

	if err := tx.Commit(); err != nil {
		log.Error("Error commiting validate refresh token: ", err.Error())
		return false, refresh_token
	}

	refresh_token.AccessToken = &access_jwt_token
	refresh_token.RefreshToken = &refresh_jwt_token
	refresh_token.User = &user
	return true, refresh_token
}

func getJwtToken(header TokenHeader, payload TokenPayload, secret string) (string, error) {
	message := "Failed to create jwt token"
	json_header, err := json.Marshal(header)
	if err != nil {
		log.Error("Error converting header to JSON: ", err.Error())
		return message, err
	}
	json_access_token_payload, err := json.Marshal(payload)
	if err != nil {
		log.Error("Error converting payload to JSON: ", err.Error())
		return message, err
	}
	encoded_header := base64UrlEncode(json_header)
	encoded_payload := base64UrlEncode(json_access_token_payload)
	signature := getJwtSignature(encoded_header, encoded_payload, secret)
	return encoded_header + "." + encoded_payload + "." + signature, nil
}

func getJwtSignature(encodedHeader, encodedPayload, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	input := encodedHeader + "." + encodedPayload
	h.Write([]byte(input))
	signature := h.Sum(nil)
	return base64UrlEncode(signature)
}

func base64UrlEncode(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

func getJti() string {
	return uuid.NewString()
}
