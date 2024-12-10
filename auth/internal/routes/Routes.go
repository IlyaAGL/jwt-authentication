package routes

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	email "github.com/agl/auth/pkg/email"
	ip "github.com/agl/auth/pkg/ip"
	refresh "github.com/agl/auth/pkg/refresh_token"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

const PRIVATE_KEY = "MY_PRIVATE_KEY"

var savedAccessToken, savedRefreshToken string

func SetupRoutes(conn *pgx.Conn) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "go /getTokens/user/{your ID} to generate your tokens\ngo /getTokens/refresh to refresh your tokens")
	})

	http.HandleFunc("/getTokens/user/{ID}", func(w http.ResponseWriter, r *http.Request) {
		handleGetTokensUser(w, r, conn)
	})

	http.HandleFunc("/getTokens/refresh", func(w http.ResponseWriter, r *http.Request) {
		handleRefreshTokens(w, r, conn)
	})
}

func handleGetTokensUser(w http.ResponseWriter, r *http.Request, conn *pgx.Conn) {
	claims := &jwt.MapClaims{
		"iss": "Ilya",
		"exp": time.Now().Add(time.Hour).Unix(),
		"sub": r.PathValue("ID"),
		"data": map[string]string{
			"ip":    ip.GetLocalIP(),
			"email": "user@gmail.com",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	completeToken, err := token.SignedString([]byte(PRIVATE_KEY))
	if err != nil {
		fmt.Fprint(w, "Cant create token")
		return
	}

	refreshToken := refresh.GetRefreshToken()
	hashedToken, err := bcrypt.GenerateFromPassword([]byte(refreshToken), bcrypt.DefaultCost)

	id, err := strconv.Atoi(r.PathValue("ID"))
	if err != nil {
		log.Fatalf("Invalid ID: %v", err)
	}

	_, err = conn.Exec(context.Background(),
		"INSERT INTO tokens (id, token) VALUES ($1, $2)", id, string(hashedToken))

	if err != nil {
		log.Fatalf("Cant execute query: %v", err)
	}

	savedAccessToken = completeToken
	savedRefreshToken = refreshToken
	fmt.Fprint(w, "Tokens have been successfuly created")
}

func handleRefreshTokens(w http.ResponseWriter, r *http.Request, conn *pgx.Conn) {
	claims := jwt.MapClaims{}
	jwt.ParseWithClaims(savedAccessToken, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte("<YOUR VERIFICATION KEY>"), nil
	})

	userID := claims["sub"]
	current_ip := claims["data"].(map[string]interface{})["ip"].(string)
	current_email := claims["data"].(map[string]interface{})["email"].(string)

	var recievedRefreshToken string
	conn.QueryRow(context.Background(),
		"SELECT token FROM tokens WHERE id=$1", userID).Scan(&recievedRefreshToken)

	if err := bcrypt.CompareHashAndPassword([]byte(recievedRefreshToken), []byte(savedRefreshToken)); err != nil {
		fmt.Fprint(w, "Unauthorized")
		if current_ip != ip.GetLocalIP() {
			email.SendEmail(current_email)
		}
	} else {
		claims := &jwt.MapClaims{
			"iss": "Ilya",
			"exp": time.Now().Add(time.Hour).Unix(),
			"sub": r.PathValue("ID"),
			"data": map[string]string{
				"ip":    ip.GetLocalIP(),
				"email": current_email,
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		completeToken, err := token.SignedString([]byte(PRIVATE_KEY))

		if err != nil {
			fmt.Fprint(w, "Cant create token")
		}

		refreshToken := refresh.GetRefreshToken()
		hashedToken, err := bcrypt.GenerateFromPassword([]byte(refreshToken), bcrypt.DefaultCost)

		_, err = conn.Exec(context.Background(),
			"INSERT INTO tokens (id, token) VALUES ($1, $2)", userID, string(hashedToken))

		if err != nil {
			log.Fatalf("Cant execute query: %v", err)
		}

		savedAccessToken = completeToken
		savedRefreshToken = refreshToken
		fmt.Fprint(w, "Tokens have been successfuly refreshed")
	}
}
