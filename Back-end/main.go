package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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

func handler1(w http.ResponseWriter, _ *http.Request) {
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
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	body, _ := ioutil.ReadAll(r.Body)
	//data := string(body) //JSON de respuesta pero string
	var variable Code
	json.Unmarshal([]byte(string(body)), &variable)
	//fmt.Printf("Species: %s", variable.Comando)
	fmt.Print(variable.Comando)

	resp := ResponseResult{"Hola, soy el resultado"}
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		fmt.Printf("Error happened in JSON marshal. Err: %s", err)
	}
	w.Write(jsonResp)
}

func main() {
	http.HandleFunc("/", handler1)
	http.HandleFunc("/info", handleInfo)
	http.HandleFunc("/postCode", handleCode)
	http.ListenAndServe(":8000", nil)
}
