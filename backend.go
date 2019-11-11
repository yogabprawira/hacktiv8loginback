package main

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"reflect"

	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

type UserData struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type UserDataList struct {
	Resp     ResponseData `json:"resp"`
	UserList []UserData   `json:"user_list"`
}

type ResponseData struct {
	ErrStr string `json:"err_str"`
}

func respSend(w http.ResponseWriter, errStr string) {
	resp := &ResponseData{}
	resp.ErrStr = errStr
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func listSend(w http.ResponseWriter, u UserDataList) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(u)
}

func login(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	userData := &UserData{}
	err := json.NewDecoder(r.Body).Decode(userData)
	if err != nil {
		log.Println(err)
		respSend(w, "error")
		return
	}
	log.Println(userData)
	if len(userData.Username) <= 0 {
		log.Println("Username must not empty")
		respSend(w, "error")
		return
	}
	rows, err := stmtLogin.Query(userData.Username)
	if err != nil {
		log.Println(err)
		respSend(w, "error")
		return
	}
	isSuccess := false
	for rows.Next() {
		var u, p string
		err = rows.Scan(&u, &p)
		if err != nil {
			log.Println(err)
			respSend(w, "error")
			return
		}
		log.Println("user:", u, "password hash:", p, "password from user:", userData.Password)
		if reflect.DeepEqual(userData.Password, p) {
			isSuccess = true
		}
	}
	if !isSuccess {
		respSend(w, "error")
	}
	respSend(w, "success")
}

func register(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	userData := &UserData{}
	err := json.NewDecoder(r.Body).Decode(userData)
	if err != nil {
		log.Println(err)
		respSend(w, "error")
		return
	}
	log.Println(userData)
	rows, err := stmtCheckRegister.Query(userData.Username)
	if err != nil {
		log.Println(err)
		respSend(w, "error")
		return
	}
	for rows.Next() {
		log.Println("Username is exist!")
		respSend(w, "error")
		return
	}
	_, err = stmtRegister.Exec(userData.Username, userData.Name, userData.Email, userData.Password, userData.Role)
	if err != nil {
		log.Println(err)
		respSend(w, "error")
		return
	}
	respSend(w, "success")
}

func getUserList(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	userDataList := UserDataList{}
	userDataList.Resp.ErrStr = "error"
	rows, err := stmtGetList.Query()
	if err != nil {
		log.Println(err)
		listSend(w, userDataList)
		return
	}
	listTemp := make([]UserData, 0)
	for rows.Next() {
		var u, n, e, p, r string
		err = rows.Scan(&u, &n, &e, &p, &r)
		if err != nil {
			log.Println(err)
			listSend(w, userDataList)
			return
		}
		listTemp = append(listTemp, UserData{
			Name:     n,
			Username: u,
			Email:    e,
			Password: p,
			Role:     r,
		})
	}
	userDataList.Resp.ErrStr = "success"
	userDataList.UserList = listTemp
	listSend(w, userDataList)
}

var db *sql.DB
var stmtRegister *sql.Stmt
var stmtLogin *sql.Stmt
var stmtGetList *sql.Stmt
var stmtCheckRegister *sql.Stmt

const UserDb = "root"
const PwdDb = "yoga123"
const DbName = "login"

func dbInit() {
	var err error
	dataSourceName := fmt.Sprintf("%s:%s@/%s?charset=utf8", UserDb, PwdDb, DbName)
	db, err = sql.Open("mysql", dataSourceName)
	if err != nil {
		log.Fatalln(err)
	}
	stmtRegister, err = db.Prepare("insert into `userinfo` values (null, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatalln(err)
	}
	stmtLogin, err = db.Prepare("select username, password from userinfo where username = ?")
	if err != nil {
		log.Fatalln(err)
	}
	stmtGetList, err = db.Prepare("select username, name, email, password, role from userinfo")
	if err != nil {
		log.Fatalln(err)
	}
	stmtCheckRegister, err = db.Prepare("select username from userinfo where username = ?")
	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	dbInit()
	router := httprouter.New()
	router.POST("/list", getUserList)
	router.POST("/login", login)
	router.POST("/register", register)
	log.Fatal(http.ListenAndServe(":9090", router))
}
