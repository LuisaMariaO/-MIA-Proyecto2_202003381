package analizador

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// This func must be Exported, Capitalized, and comment added.
func Holis() {
	fmt.Println("Holis :3")
}

var consola string

func espacioCadena(comando string) string {
	var cadena bool = false
	a := []byte(comando)
	for i := 0; i < len(a); i++ {
		if a[i] == '"' {
			if cadena {
				cadena = false //Fin de cadena
			} else {
				cadena = true

			}
		}
		if cadena && a[i] == ' ' {
			a[i] = '$' //Si me encuentro dentro de una cadena y encuentro un espacio, lo sustituyo temporalmente por un $

		}
	}

	return string(a)
}

func getTipoValor(parametro string) (string, string) {
	//fmt.Println(parametro)
	if parametro[0] == '#' {
		return parametro, "" //Si es un comentario, retorno el mismo comando
	}

	par := strings.Split(parametro, "=")
	par[0] = strings.ToLower(par[0]) //Paso el tipo de parameto a minúsculas
	return par[0], par[1]

}

func regresarEspacio(ruta string) string {
	ruta = strings.Replace(ruta, "$", " ", -1)
	ruta = strings.Replace(ruta, "\"", "", -1) //Quito las comillas
	return ruta
}
func verifyDirectory(ruta string) {
	dir := filepath.Dir(ruta)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		//Si el directorio no existe, se crea
		if err2 := os.MkdirAll(dir, os.ModePerm); err2 != nil {
			consola += err2.Error() + "\n"
		}
	}
}

func crearArchivoB(ruta string, tamano int, unit byte) bool {

	var buffer [1024]byte
	//fmt.Println(ruta)
	file, err := os.Create(ruta)
	if err != nil {
		consola += "Error al crear el archivo binario\n"
		return false
	}
	defer file.Close()

	if unit == 'k' {
		for i := 0; i < tamano; i++ {
			err = binary.Write(file, binary.LittleEndian, &buffer)
			if err != nil {
				consola += "Error al escribir en el archivo binario\n"
				return false
			}
		}
	} else {
		for i := 0; i < (tamano * 1024); i++ {
			err = binary.Write(file, binary.LittleEndian, &buffer)
			if err != nil {
				consola += "Error al escribir en el archivo binario\n"
				return false
			}
		}
	}

	return true
}

func mkdisk(parametros []string) {
	var fsize, funit, fpath, ffit bool = false, false, false, false

	mbr := MBR{}
	var size int = 0 //TODO: Revisar que no de problema solo int
	var path string
	var unit, fit byte
	for len(parametros) > 0 {

		tmp := parametros[0]
		tipo, valor := getTipoValor(tmp)
		if tipo == ">size" {
			size, _ = strconv.Atoi(valor)
			if size > 0 {
				fsize = true
			} else {
				consola += "Error: Tamaño de disco <" + valor + "> inválido\n"
				break
			}
		} else if tipo == ">fit" {
			if strings.EqualFold(valor, "bf") {
				fit = 'B'
				ffit = true
			} else if strings.EqualFold(valor, "ff") {
				fit = 'F'
				ffit = true
			} else if strings.EqualFold(valor, "wf") {
				fit = 'W'
				ffit = true
			} else {
				consola += "Error: Valor de ajuste <" + valor + "> inválido\n"
			}
		} else if tipo == ">unit" {
			if strings.EqualFold(valor, "k") {
				unit = 'k'
				funit = true
			} else if strings.EqualFold(valor, "M") {
				unit = 'm'
				funit = true
			} else {
				consola += "Error: Valor de unidad <" + valor + "> inválido\n"
			}
		} else if tipo == ">path" {
			valor = regresarEspacio(valor)
			verifyDirectory(valor) //Verifico si el directorio no existe para crearlo
			path = valor
			fpath = true
		} else if tipo[0] == '#' {
			break
		} else {
			consola += "Error: Parámetro <" + tipo + "> no válido\n"
		}

		parametros = parametros[1:] //Elimino el parámetro que ya se analizó
	}

	if fsize && fpath {
		/*
			if !funit {
				size = size * 1024 * 1024 //Megabytes
			} else {
				if unit == 'k' {
					size = size * 1024
				} else {
					size = size * 1024 * 1024
				}
			}
		*/
		if !funit {
			unit = 'm'
		}
		if !ffit {
			fit = 'f'
		}

		mbr.Mbr_tamano = int32(size)
		mbr.Dsk_fit = fit
		t := time.Now()
		fecha := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d",
			t.Year(), t.Month(), t.Day(),
			t.Hour(), t.Minute(), t.Second())
		copy(mbr.Mbr_fecha_creacion[:], []byte(fecha))
		seed := rand.New(rand.NewSource(time.Now().UnixNano()))
		mbr.Mbr_dsk_signature = seed.Int31n(100000)
		//fmt.Println(unsafe.Sizeof(mbr))

		if crearArchivoB(path, size, unit) {
			file, err := os.OpenFile(path, os.O_RDWR, 0777)
			if err != nil {
				consola += "Error al abrir el archivo binario\n"
				return

			}
			defer file.Close()
			file.Seek(0, 0)
			binary.Write(file, binary.LittleEndian, &mbr)
			consola += "¡Disco creado con éxito!\n"
		}

	} else {
		consola += "Error: No es posible crear el disco duro, faltan parámetros obligatorios\n"
	}
}

func removeFile(path string) {
	e := os.Remove(path)
	if e != nil {
		consola += "Error: No se encontró el archivo a eliminar\n"
		return
	}
	consola += "¡Disco eliminado con éxito!\n"
}

func rmdisk(parametros []string) {
	var path string

	for len(parametros) > 0 {
		tmp := parametros[0]
		tipo, valor := getTipoValor(tmp)

		if tipo == ">path" {
			valor = regresarEspacio(valor)
			path = valor
		} else if tipo[0] == '#' {
			break
		} else {
			consola += "Error: Parámetro <" + valor + "> no válido\n"
		}
		parametros = parametros[1:]
	}
	removeFile(path)
}

func Analizar(lineas []string) string {
	consola = "" //Reestableciendo la consola cada vez que se llama a analizar
	for _, linea := range lineas {
		linea = espacioCadena(linea) //Cambio temporalmente los espacios dentro de cadenas por $

		params := strings.Split(linea, " ")         //Separo por espacio
		if strings.EqualFold(params[0], "mkdisk") { //Comparación case insensitve
			//Elimino el primer elemento (el nombre del comando)
			params = params[1:]
			mkdisk(params)
		} else if strings.EqualFold(params[0], "rmdisk") {
			params = params[1:]
			rmdisk(params)
		} else if params[0][0] == '#' {
			//Si es un comentario, no pasa nada
		} else {
			consola += "Error: Comando <" + params[0] + "> no reconocido\n"
		}
	}
	return consola
}
