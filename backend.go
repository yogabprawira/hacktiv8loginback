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

func respSend(w http.ResponseWriter, errStr string, data UserData) {
	resp := &UserDataList{}
	resp.Resp.ErrStr = errStr
	resp.UserList = append(resp.UserList, data)
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
		respSend(w, "Login error!", UserData{})
		return
	}
	if len(userData.Username) <= 0 {
		log.Println("Username must not empty")
		respSend(w, "Login error!", UserData{})
		return
	}
	rows, err := stmtLogin.Query(userData.Username)
	if err != nil {
		log.Println(err)
		respSend(w, "Login error!", UserData{})
		return
	}
	isSuccess := false
	var u, n, p, e, role string
	for rows.Next() {
		err = rows.Scan(&u, &n, &p, &e, &role)
		if err != nil {
			log.Println(err)
			respSend(w, "Login error!", UserData{})
			return
		}
		log.Println("user:", u, "password hash:", p, "password from user:", userData.Password)
		if reflect.DeepEqual(userData.Password, p) {
			isSuccess = true
			break
		}
	}
	if !isSuccess {
		respSend(w, "Password or username is not match!", UserData{})
	}
	userDataResp := UserData{
		Name:     n,
		Username: u,
		Email:    e,
		Password: "",
		Role:     role,
	}
	log.Println(userDataResp)
	respSend(w, "success", userDataResp)
}

func register(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	userData := &UserData{}
	err := json.NewDecoder(r.Body).Decode(userData)
	if err != nil {
		log.Println(err)
		respSend(w, "Registration error!", UserData{})
		return
	}
	rows, err := stmtCheckRegister.Query(userData.Username)
	if err != nil {
		log.Println(err)
		respSend(w, "Registration error!", UserData{})
		return
	}
	for rows.Next() {
		respSend(w, "Username is exist!", UserData{})
		return
	}
	_, err = stmtRegister.Exec(userData.Username, userData.Name, userData.Email, userData.Password, userData.Role)
	if err != nil {
		log.Println(err)
		respSend(w, "Registration error!", UserData{})
		return
	}
	userDataResp := UserData{
		Name:     userData.Name,
		Username: userData.Username,
		Email:    userData.Email,
		Password: "",
		Role:     userData.Role,
	}
	respSend(w, "success", userDataResp)
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

func getUserDetail(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	usernameReq := ps.ByName("username")
	rows, err := stmtUserDetail.Query(usernameReq)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var userData UserData
	for rows.Next() {
		err = rows.Scan(&userData.Username, &userData.Name, &userData.Email, &userData.Role)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	log.Println(userData)
	if len(userData.Username) <= 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(&userData)
}

func editUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var userData UserData
	err := json.NewDecoder(r.Body).Decode(&userData)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(userData.Password) > 0 {
		_, err = stmtEditUser.Exec(userData.Name, userData.Email, userData.Password, userData.Role, userData.Username)
	} else {
		_, err = stmtEditUserWithoutPass.Exec(userData.Name, userData.Email, userData.Role, userData.Username)
	}
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var responseData ResponseData
	responseData.ErrStr = "success"
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(&responseData)
}

var db *sql.DB
var stmtRegister *sql.Stmt
var stmtLogin *sql.Stmt
var stmtGetList *sql.Stmt
var stmtCheckRegister *sql.Stmt
var stmtUserDetail *sql.Stmt
var stmtEditUser *sql.Stmt
var stmtEditUserWithoutPass *sql.Stmt

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
	stmtLogin, err = db.Prepare("select username, name, password, email, role from userinfo where username = ?")
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
	stmtUserDetail, err = db.Prepare("select username, name, email, role from `userinfo` where username = ?;")
	if err != nil {
		log.Fatalln(err)
	}
	stmtEditUser, err = db.Prepare("update `userinfo` set name = ?, email = ?, password = ?, role = ? where username = ?;")
	if err != nil {
		log.Fatalln(err)
	}
	stmtEditUserWithoutPass, err = db.Prepare("update `userinfo` set name = ?, email = ?, role = ? where username = ?;")
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
	router.GET("/user/:username", getUserDetail)
	router.POST("/edit/:username", editUser)
	log.Fatal(http.ListenAndServe(":9090", router))
}
