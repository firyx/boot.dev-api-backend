package database

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"os"
	"strings"
	"time"
)

type Client struct {
	path string
}

type databaseSchema struct {
	Users map[string]User `json:"users"`
	Posts map[string]Post `json:"posts"`
}

type User struct {
	CreatedAt time.Time `json:"createdAt"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	Name      string    `json:"name"`
	Age       int       `json:"age"`
}

type Post struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UserEmail string    `json:"userEmail"`
	Text      string    `json:"text"`
}

func NewClient(path string) Client {
	return Client{
		path: path,
	}
}

func (c Client) createDB() error {
	emptyDB := databaseSchema{
		Users: map[string]User{},
		Posts: map[string]Post{},
	}
	data, err := json.Marshal(emptyDB)
	if err != nil {
		return err
	}
	return os.WriteFile(c.path, data, 0600)
}

func (c Client) EnsureDB() error {
	// check if db exists
	data, err := os.ReadFile(c.path)
	dbExists := true
	if err != nil && strings.Contains(err.Error(), "no such file or directory") {
		dbExists = false
	}
	if err != nil && dbExists {
		return err
	}

	// create db, load new data
	if !dbExists {
		err = c.createDB()
	}
	if err != nil {
		return err
	}
	if !dbExists {
		data, err = os.ReadFile(c.path)
	}
	if err != nil {
		return err
	}

	// check db data
	db := &databaseSchema{}
	err = json.Unmarshal(data, db)
	if err != nil {
		return err
	}
	return nil
}

func (c Client) updateDB(db databaseSchema) error {
	data, err := json.Marshal(db)
	if err != nil {
		return err
	}
	return os.WriteFile(c.path, data, 0600)
}

func (c Client) readDB() (databaseSchema, error) {
	data, err := os.ReadFile(c.path)
	if err != nil {
		return databaseSchema{}, err
	}
	db := &databaseSchema{}
	err = json.Unmarshal(data, db)
	if err != nil {
		return databaseSchema{}, err
	}
	return *db, nil
}

func (c Client) CreateUser(email, password, name string, age int) (User, error) {
	db, err := c.readDB()
	if err != nil {
		return User{}, err
	}
	if _, ok := db.Users[email]; ok {
		return User{}, fmt.Errorf("user with email %s already exists", email)
	}
	user := User{
		CreatedAt: time.Now().UTC(),
		Email:     email,
		Password:  password,
		Name:      name,
		Age:       age,
	}
	db.Users[email] = user
	err = c.updateDB(db)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (c Client) UpdateUser(email, password, name string, age int) (User, error) {
	db, err := c.readDB()
	if err != nil {
		return User{}, err
	}
	user, ok := db.Users[email]
	if !ok {
		return User{}, fmt.Errorf("user with email %s doesn't exist", email)
	}
	oldEmail := user.Email
	user.Email = email
	user.Password = password
	user.Name = name
	user.Age = age
	db.Users[email] = user
	if oldEmail != email {
		delete(db.Users, oldEmail)
	}
	err = c.updateDB(db)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (c Client) GetUser(email string) (User, error) {
	db, err := c.readDB()
	if err != nil {
		return User{}, err
	}
	user, ok := db.Users[email]
	if !ok {
		return User{}, fmt.Errorf("user with email %s doesn't exist", email)
	}
	return user, nil
}

func (c Client) DeleteUser(email string) error {
	db, err := c.readDB()
	if err != nil {
		return err
	}
	_, ok := db.Users[email]
	if !ok {
		return fmt.Errorf("user with email %s doesn't exist", email)
	}
	delete(db.Users, email)
	err = c.updateDB(db)
	if err != nil {
		return err
	}
	return nil
}

func (c Client) CreatePost(userEmail, text string) (Post, error) {
	db, err := c.readDB()
	if err != nil {
		return Post{}, err
	}
	_, err = c.GetUser(userEmail)
	if err != nil {
		return Post{}, err
	}
	id := uuid.NewString()
	post := Post{
		ID: id,
		CreatedAt: time.Now().UTC(),
		UserEmail: userEmail,
		Text: text,
	}
	db.Posts[id] = post
	err = c.updateDB(db)
	if err != nil {
		return Post{}, err
	}
	return post, nil
}

func (c Client) GetPost(id string) (Post, error) {
	db, err := c.readDB()
	if err != nil {
		return Post{}, err
	}
	post, ok := db.Posts[id]
	if !ok {
		return Post{}, fmt.Errorf("post with id %s doesn't exist", id)
	}
	return post, nil
}

func (c Client) GetPosts(userEmail string) ([]Post, error) {
	db, err := c.readDB()
	if err != nil {
		return nil, err
	}
	userPosts := []Post{}
	for _, post := range db.Posts {
		if post.UserEmail == userEmail {
			userPosts = append(userPosts, post)
		}
	}
	return userPosts, nil
}

func (c Client) DeletePost(id string) error {
	db, err := c.readDB()
	if err != nil {
		return err
	}
	if _, ok := db.Posts[id]; !ok {
		return fmt.Errorf("post with id %s doesn't exist", id)
	}
	delete(db.Posts, id)
	err = c.updateDB(db)
	if err != nil {
		return err
	}
	return nil
}
