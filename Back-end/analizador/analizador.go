package analizador

import (
	"fmt"
	"strings"
)

// This func must be Exported, Capitalized, and comment added.
func Holis() {
	fmt.Println("Holis :3")
}

var consola string

func espacioCadena(comando string) string {
	var cadena bool = false

	for _, char := range comando {
		if char == '"' {
			if cadena {
				cadena = false //Fin de cadena
			} else {
				cadena = true
			}
		}
		if cadena && char == ' ' {
			char = '$' //Si me encuentro dentro de una cadena y encuentro un espacio, lo sustituyo temporalmente por un $
		}
	}
	return comando
}

func Analizar(lineas []string) string {
	consola = "" //Reestableciendo la consola cada vez que se llama a analizar
	for _, linea := range lineas {
		linea = espacioCadena(linea)                //Cambio temporalmente los espacios dentro de cadenas por $
		params := strings.Split(linea, " ")         //Separo por espacio
		if strings.EqualFold(params[0], "mkdisk") { //Comparación case insensitve
			//Elimino el primer elemento (el nombre del comando)
			params = params[1:]
			consola += "Creo mkdisk\n"
		}
	}
	return consola
}
