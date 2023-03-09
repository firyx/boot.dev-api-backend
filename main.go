package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/firyx/boot.dev-api-backend/internal/database"
)

type errorBody struct {
	Error string `json:"error"`
}

type apiConfig struct {
	dbClient database.Client
	usersPrefix string
	postsprefix string
}



func main() {
	c := database.NewClient("./db.json")
	c.EnsureDB()

	apiCfg := apiConfig{
		dbClient: c,
		usersPrefix: "/users",
		postsprefix: "/posts",
	}

	serveMux := http.NewServeMux()

	serveMux.HandleFunc(apiCfg.usersPrefix, apiCfg.endpointUsersHandler)
    serveMux.HandleFunc(apiCfg.usersPrefix + "/", apiCfg.endpointUsersHandler)
    serveMux.HandleFunc(apiCfg.postsprefix, apiCfg.endpointPostsHandler)
    serveMux.HandleFunc(apiCfg.postsprefix + "/", apiCfg.endpointPostsHandler)

	const addr = "localhost:8080"
	srv := http.Server{
		Handler:      serveMux,
		Addr:         addr,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
	}
	srv.ListenAndServe()
}

func (apiCfg apiConfig) endpointPostsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// call GET handler
		apiCfg.handlerRetrievePosts(w,r)
	case http.MethodPost:
		// call POST handler
		apiCfg.handlerCreatePost(w, r)
	case http.MethodPut:
		// call PUT handler
	case http.MethodDelete:
		// call DELETE handler
		apiCfg.handlerDeletePost(w, r)
	default:
		respondWithError(w, 404, errors.New("method not supported"))
	}
}

func (apiCfg apiConfig) endpointUsersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// call GET handler
		apiCfg.handlerGetUser(w,r)
	case http.MethodPost:
		// call POST handler
		apiCfg.handlerCreateUser(w, r)
	case http.MethodPut:
		// call PUT handler
		apiCfg.handlerUpdateUser(w, r)
	case http.MethodDelete:
		// call DELETE handler
		apiCfg.handlerDeleteUser(w, r)
	default:
		respondWithError(w, 404, errors.New("method not supported"))
	}
}

func (apiCfg apiConfig) handlerCreatePost(w http.ResponseWriter, r *http.Request) {
	// get params
	type parameters struct {
		UserEmail string `json:"userEmail"`
		Text      string `json:"text"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err)
		return
	}
	
	// check user exists
	if !userExists(apiCfg, params.UserEmail) {
		respondWithError(w, http.StatusBadRequest, errors.New("user with that email doesn't exist"))
		return
	}

	// create post
	post, err := apiCfg.dbClient.CreatePost(params.UserEmail, params.Text)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}
	respondWithJSON(w, http.StatusCreated, post)
}

func (apiCfg apiConfig) handlerRetrievePosts(w http.ResponseWriter, r *http.Request) {
	// get params
	type parameters struct {
		UserEmail string `json:"userEmail"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err)
		return
	}
	
	// check user exists
	if !userExists(apiCfg, params.UserEmail) {
		respondWithError(w, http.StatusBadRequest, errors.New("user with that email doesn't exist"))
		return
	}

	// return posts
	posts, err := apiCfg.dbClient.GetPosts(params.UserEmail)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}
	respondWithJSON(w, http.StatusOK, posts)
}

func (apiCfg apiConfig) handlerDeletePost(w http.ResponseWriter, r *http.Request) {
	// check path
	id, err := getPostUuid(apiCfg, r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, errors.New("bad request, correct format is: /users/{post-id}"))
		return
	}

	// check post exists
	if !postExists(apiCfg, id) {
		respondWithError(w, http.StatusBadRequest, errors.New("post with that id doesn't exist"))
		return
	}

	// delete post
	err = apiCfg.dbClient.DeletePost(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}
	respondWithJSON(w, http.StatusOK, struct{}{})
}

func (apiCfg apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	// get params
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
		Age      int    `json:"age"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err)
		return
	}

	// check user doesn't exist
	if userExists(apiCfg, params.Email) {
		respondWithError(w, http.StatusBadRequest, errors.New("user with that email already exists"))
		return
	}

	// create user
	user, err := apiCfg.dbClient.CreateUser(params.Email, params.Password, params.Name, params.Age)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}
	respondWithJSON(w, http.StatusCreated, user)
}

func (apiCfg apiConfig) handlerGetUser(w http.ResponseWriter, r *http.Request) {
	// check path
	email, err := getUserEmail(apiCfg, r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, errors.New("bad request, correct format is: /users/{email}"))
		return
	}

	// check user exists
	if !userExists(apiCfg, email) {
		respondWithError(w, http.StatusBadRequest, errors.New("user with that email doesn't exist"))
		return
	}

	// return user
	user, err := apiCfg.dbClient.GetUser(email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
	}
	respondWithJSON(w, http.StatusOK, user)
}

func (apiCfg apiConfig) handlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	// get params
	type parameters struct {
		Password string `json:"password"`
		Name     string `json:"name"`
		Age      int    `json:"age"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err)
		return
	}
	// check path
	email, err := getUserEmail(apiCfg, r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, errors.New("bad request, correct format is: /users/{email}"))
		return
	}

	// check user exists
	if !userExists(apiCfg, email) {
		respondWithError(w, http.StatusBadRequest, errors.New("user with that email doesn't exist"))
		return
	}

	// update user
	user, err := apiCfg.dbClient.UpdateUser(email, params.Password, params.Name, params.Age)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}
	respondWithJSON(w, http.StatusOK, user)
}

func (apiCfg apiConfig) handlerDeleteUser(w http.ResponseWriter, r *http.Request) {
	// check path
	email, err := getUserEmail(apiCfg, r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, errors.New("bad request, correct format is: /users/{email}"))
		return
	}

	// check user exists
	if !userExists(apiCfg, email) {
		respondWithError(w, http.StatusBadRequest, errors.New("user with that email doesn't exist"))
		return
	}

	// delete user
	err = apiCfg.dbClient.DeleteUser(email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}
	respondWithJSON(w, http.StatusOK, struct{}{})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	response, err := json.Marshal(payload)
	if err != nil {
		code = http.StatusInternalServerError
		response = []byte(fmt.Sprintf("{\"error\":\"%s\"}", "error marshalling to JSON" + err.Error()))
	}
	w.WriteHeader(code)
    w.Write(response)
}

func respondWithError(w http.ResponseWriter, code int, err error) {
	errorBody := errorBody{
		Error: err.Error(),
	}
	respondWithJSON(w, code, errorBody)
}

func getUserEmail(apiCfg apiConfig, r *http.Request) (string, error) {
	prefix := apiCfg.usersPrefix + "/"
	return trimPrefix(r.URL.Path, prefix, "not a valid URL: %s{email}")
}

func getPostUuid(apiConfig apiConfig, r *http.Request) (string, error) {
	prefix := apiConfig.postsprefix + "/"
	return trimPrefix(r.URL.Path, prefix, "not a valid URL: %s{post-id}")
}

func trimPrefix(str, prefix, errMsg string) (string, error) {
	res := strings.TrimPrefix(str, prefix)
	if !strings.HasPrefix(str, prefix) || res == "" {
		return "", fmt.Errorf(errMsg, prefix)
	}
	return res, nil
}

func postExists(apiCfg apiConfig, id string) bool {
	_, err := apiCfg.dbClient.GetPost(id)
	return err == nil
}

func userExists(apiCfg apiConfig, email string) bool {
	_, err := apiCfg.dbClient.GetUser(email)
	return err == nil
}
