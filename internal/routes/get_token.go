package routes

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/livekit/protocol/auth"
	"log"
	"net/http"
	"strconv"
	"time"
)

func GetToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid HTTP method. Only GET is allowed.", http.StatusMethodNotAllowed)
		return
	}

	paramRoomID := r.URL.Query().Get("room_id")
	roomID, err := strconv.Atoi(paramRoomID)
	if err != nil {
		http.Error(w, "room_id must be an integer", http.StatusBadRequest)
		return
	}

	cookie, err := r.Cookie(jwtName)

	if err != nil {
		http.Error(w, "Unauthenticated", http.StatusUnauthorized)
		return
	}

	token, err := jwtCheck(cookie)

	if err != nil {
		http.Error(w, "Unauthenticated", http.StatusUnauthorized)
		return
	}

	claims := token.Claims.(*jwt.RegisteredClaims)

	userExists, err := checkUserExists(claims.Issuer)
	if err != nil {
		http.Error(w, "Error while checking user authorization", http.StatusInternalServerError)
		return
	}

	if !userExists {
		log.Println("User with id ", claims.Issuer, "does not exist: ", err)
		http.Error(w, "You are not logged in", http.StatusUnauthorized)
		return
	}

	res, err := getJoinToken(roomID, claims.Subject)
	if err != nil {
		http.Error(w, "Failed to get token", http.StatusInternalServerError)
	}

	at := http.Cookie{
		Name:    "learn_live_ACCESS_TOKEN",
		Expires: time.Now().Add(24 * time.Hour),
	}

	at.Value = res
	http.SetCookie(w, &at)
}

func getJoinToken(roomID int, username string) (string, error) {
	at := auth.NewAccessToken("devkey", "secret")

	room := strconv.Itoa(roomID)

	grant := &auth.VideoGrant{
		RoomCreate: true,
		RoomJoin:   true,
		Room:       room,
	}

	log.Println("getJoinToken. roomID: ", roomID, "username: ", username)
	log.Printf("%+v\n", grant)
	at.AddGrant(grant).SetIdentity(username).SetValidFor(24 * time.Hour)
	return at.ToJWT()
}
