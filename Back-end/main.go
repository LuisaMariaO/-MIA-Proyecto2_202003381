package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"proyecto2/analizador"
	"strings"
)

type ResponseInfo struct {
	Carnet int    `json:"carnet"`
	Nombre string `json:"nombre"`
}

type Code struct {
	Comando string `json:"comando"`
}

type ResponseResult struct {
	Result string `json:"result"`
}
type ResponseLogin struct {
	Status  byte   `json:"status"`
	Message string `json:"message"`
}
type LoginBody struct {
	Id       string `json:"id"`
	User     string `json:"user"`
	Password string `json:"password"`
}

func handler1(w http.ResponseWriter, r *http.Request) {
	reqDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("REQUEST:\n%s", string(reqDump))
	fmt.Fprintf(w, "Hello from Golang!!!")
}

func handleInfo(w http.ResponseWriter, _ *http.Request) {
	//w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	resp := ResponseInfo{202003381, "Luisa María Ortiz Romero"}
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		fmt.Printf("Error happened in JSON marshal. Err: %s", err)
	}
	w.Write(jsonResp)

}

func handleCode(w http.ResponseWriter, r *http.Request) {
	//reqDump, err := httputil.DumpRequest(r, true)
	//if err != nil {
	//	log.Fatal(err)
	//}

	//fmt.Printf("REQUEST:\n%s", string(reqDump))
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	body, _ := ioutil.ReadAll(r.Body)
	//data := string(body) //JSON de respuesta pero string
	var variable Code
	json.Unmarshal([]byte(string(body)), &variable)
	//fmt.Printf("Species: %s", variable.Comando)
	var consola string = ""
	if len(variable.Comando) >= 1 { //Porque a veces llegan peticiones vacías desde la app :c
		lineas := strings.Split(variable.Comando, "\n") //Separo por salto de línea pues cada comando es una línea diferente
		consola = analizador.Analizar(lineas)
	} else {
		consola = ""
	}

	resp := ResponseResult{consola}

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		fmt.Printf("Error happened in JSON marshal. Err: %s", err)
	}
	w.Write(jsonResp)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	body, _ := ioutil.ReadAll(r.Body)

	var variable LoginBody
	json.Unmarshal([]byte(string(body)), &variable)
	var resp ResponseLogin
	if len(variable.Id) >= 1 { //Porque a veces llegan peticiones vacías desde la app :c
		resp.Status, resp.Message = analizador.IniciarSesion(variable.User, variable.Password, variable.Id)

	} else {

	}

	//resp := ResponseResult{consola}

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		fmt.Printf("Error happened in JSON marshal. Err: %s", err)
	}
	w.Write(jsonResp)

}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	var resp ResponseLogin
	resp.Status, resp.Message = analizador.Logout()

	//resp := ResponseResult{consola}

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		fmt.Printf("Error happened in JSON marshal. Err: %s", err)
	}
	w.Write(jsonResp)
}

func Prueba() {
	fmt.Println("Que pasaaa")
}

func doNothing(w http.ResponseWriter, r *http.Request) {}

func main() {

	http.HandleFunc("/", handler1)
	http.HandleFunc("/info", handleInfo)
	http.HandleFunc("/postCode", handleCode)
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/logout", handleLogout)
	http.HandleFunc("/favicon.ico", doNothing)
	http.ListenAndServe(":8000", nil)
}
