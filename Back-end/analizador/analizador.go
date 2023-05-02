package analizador

import (
	"bufio"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

// This func must be Exported, Capitalized, and comment added.
func Holis() {
	fmt.Println("Holis :3")
}

type Atributos struct {
	ruta     string
	nombre   [16]byte
	inicio   int
	tamano   int
	tipo     byte
	numDisco int
}

/*Variables globales*/
var consola string

/*<id,atributos>*/
var montadas = make(map[string]Atributos)
var ultDisco = 0

var arbol, conexiones = "", ""

var userLog, uidLog, idLog, gidLog string
var logged bool = false

type reportes struct {
	Reportes []Reporte `json:"reportes"`
}
type Reporte struct {
	Name    string `json:"nombre"`
	Reporte string `json:"reporte"`
}

var Reportes reportes

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
	if len(parametro) < 2 {
		return "#vacio", ""
	}
	if parametro[0] == '#' {
		return parametro, "" //Si es un comentario, retorno el mismo comando
	}

	par := strings.Split(parametro, "=")
	par[0] = strings.ToLower(par[0]) //Paso el tipo de parameto a minúsculas
	if len(par) == 2 {
		return par[0], par[1]
	} else {
		return par[0], ""
	}

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
			err = binary.Write(file, binary.BigEndian, &buffer)
			if err != nil {
				consola += "Error al escribir en el archivo binario\n"
				return false
			}
		}
	} else {
		for i := 0; i < (tamano * 1024); i++ {
			err = binary.Write(file, binary.BigEndian, &buffer)
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
	var size, fixsize int = 0, 0 //TODO: Revisar que no de problema solo int
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

		if !funit {
			fixsize = size * 1024 * 1024 //Megabytes
		} else {
			if unit == 'k' {
				fixsize = size * 1024
			} else {
				fixsize = size * 1024 * 1024
			}
		}

		if !funit {
			unit = 'm'
		}
		if !ffit {
			fit = 'f'
		}

		mbr.Mbr_tamano = int32(fixsize)
		mbr.Dsk_fit = fit
		t := time.Now()
		fecha := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d",
			t.Year(), t.Month(), t.Day(),
			t.Hour(), t.Minute(), t.Second())
		copy(mbr.Mbr_fecha_creacion[:], []byte(fecha))
		seed := rand.New(rand.NewSource(time.Now().UnixNano()))
		mbr.Mbr_dsk_signature = seed.Int31n(100000)
		for i := 0; i < 4; i++ {
			mbr.Mbr_partition[i].Part_status = '0'
		}
		//fmt.Println(unsafe.Sizeof(mbr))

		if crearArchivoB(path, size, unit) {
			file, err := os.OpenFile(path, os.O_RDWR, 0777)
			if err != nil {
				consola += "Error al abrir el archivo binario\n"
				return

			}
			defer file.Close()
			file.Seek(0, 0)
			binary.Write(file, binary.BigEndian, &mbr)
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

func existsFile(name string) bool {
	if _, err := os.Stat(name); err == nil {
		return true
	} else {
		return false
	}
}

func ajustarP(ruta string, size int, name [16]byte, fit byte) {

	file, err := os.OpenFile(ruta, os.O_RDWR, 0777)
	if err != nil {
		consola += "Error: No se pudo abrir el disco duro\n"
	}
	defer file.Close()

	file.Seek(0, 0)
	var mbr MBR
	binary.Read(file, binary.BigEndian, &mbr)

	//Inicializando parámetros
	var particion Partition
	particion.Part_status = '1'
	particion.Part_type = 'P'
	particion.Part_fit = fit
	particion.Part_start = 0
	particion.Part_size = int32(size)
	copy(particion.Part_name[:], name[:])

	ocupado := unsafe.Sizeof(mbr)
	var crear bool = true
	pos := -1

	if mbr.Mbr_partition[0].Part_status == '1' && mbr.Mbr_partition[1].Part_status == '1' && mbr.Mbr_partition[2].Part_status == '1' && mbr.Mbr_partition[3].Part_status == '1' {
		consola += "Error: Límite de 4 particiones primarias y extendidas alcanzado\n"
		crear = false
	} else {
		for i := 0; i < 4; i++ { //Encontrando el espacio ocupado actualmente
			if mbr.Mbr_partition[i].Part_status == '1' {
				ocupado += uintptr(mbr.Mbr_partition[i].Part_size)

			}
		}
		//Todo será ajustado según el FirstFit pues no se eliminarán particiones

		for i := 0; i < 4; i++ {
			if mbr.Mbr_partition[i].Part_status == '0' {
				//Si la partición está libre
				if ocupado+uintptr(particion.Part_size) <= uintptr(mbr.Mbr_tamano) {
					//Si el nuevo tamaño ocupado cabe en el disco
					pos = i
					break
				} else {
					consola += "Error: Espacio insuficiente para la partición\n"
					crear = false
					break
				}
			}
		}
	}

	if crear {
		particion.Part_start = int32(ocupado)
		mbr.Mbr_partition[pos] = particion
		consola += "¡Partición Primaria <" + string(name[:]) + "> creada!\n"

		file.Seek(0, 0)
		binary.Write(file, binary.BigEndian, &mbr)
	}
}
func ajustarE(ruta string, size int, name [16]byte, fit byte) {
	extendida := false
	file, err := os.OpenFile(ruta, os.O_RDWR, 0777)
	if err != nil {
		consola += "Error: No se pudo abrir el disco duro\n"
	}
	defer file.Close()

	file.Seek(0, 0)
	var mbr MBR
	binary.Read(file, binary.BigEndian, &mbr)

	//Inicializando parámetros
	var particion Partition
	particion.Part_status = '1'
	particion.Part_type = 'E'
	particion.Part_fit = fit
	particion.Part_start = 0
	particion.Part_size = int32(size)
	copy(particion.Part_name[:], name[:])

	ocupado := unsafe.Sizeof(mbr)
	var crear bool = true
	pos := -1
	//Verifico si ya existe una partición extendida
	for i := 0; i < 4; i++ {
		if mbr.Mbr_partition[i].Part_type == 'E' {
			extendida = true
			crear = false
			consola += "Error: Ya existe una partición extendida\n"
		}
	}
	if mbr.Mbr_partition[0].Part_status == '1' && mbr.Mbr_partition[1].Part_status == '1' && mbr.Mbr_partition[2].Part_status == '1' && mbr.Mbr_partition[3].Part_status == '1' {
		consola += "Error: Límite de 4 particiones primarias y extendidas alcanzado\n"
		crear = false
	} else if !extendida {
		for i := 0; i < 4; i++ { //Encontrando el espacio ocupado actualmente
			if mbr.Mbr_partition[i].Part_status == '1' {
				ocupado += uintptr(mbr.Mbr_partition[i].Part_size)

			}
		}
		//Todo será ajustado según el FirstFit pues no se eliminarán particiones
		for i := 0; i < 4; i++ {
			if mbr.Mbr_partition[i].Part_status == '0' {
				//Si la partición está libre
				if ocupado+uintptr(particion.Part_size) <= uintptr(mbr.Mbr_tamano) {
					//Si el nuevo tamaño ocupado cabe en el disco
					pos = i
					break
				} else {
					consola += "Error: Espacio insuficiente para la partición\n"
					crear = false
					break
				}
			}
		}
	}

	if crear {
		particion.Part_start = int32(ocupado)
		mbr.Mbr_partition[pos] = particion
		file.Seek(0, 0)
		binary.Write(file, binary.BigEndian, &mbr) //Escribiendo el mbr actualizado

		var ebr EBR
		ebr.Part_status = '0'
		ebr.Part_fit = fit
		ebr.Part_start = 0
		ebr.Part_size = 0
		ebr.Part_next = -1

		file.Seek(int64(particion.Part_start), 0)
		binary.Write(file, binary.BigEndian, &ebr)

		consola += "¡Partición extendida <" + string(name[:]) + "> creada!\n"
	}
}
func ajustarL(ruta string, size int, name [16]byte, fit byte) {
	extendida := false
	ocupado := 0

	file, err := os.OpenFile(ruta, os.O_RDWR, 0777)
	if err != nil {
		consola += "Error: No se pudo abrir el disco duro\n"
	}
	defer file.Close()

	file.Seek(0, 0)
	var mbr MBR
	binary.Read(file, binary.BigEndian, &mbr)

	var ext Partition

	var logica EBR
	logica.Part_fit = fit
	copy(logica.Part_name[:], name[:])
	logica.Part_status = '1'
	logica.Part_size = int32(size)
	logica.Part_start = 0
	logica.Part_next = -1

	//Verifico si ya existe una partición extendida
	for i := 0; i < 4; i++ {
		if mbr.Mbr_partition[i].Part_type == 'E' {
			extendida = true
			ext = mbr.Mbr_partition[i]
			break
		}
	}

	if extendida {
		//Si existe una partición extendida, se pueden crear lógica
		var tmp, ultimo EBR

		file.Seek(int64(ext.Part_start), 0)
		binary.Read(file, binary.BigEndian, &tmp)

		file.Seek(int64(ext.Part_start), 0)
		binary.Read(file, binary.BigEndian, &ultimo)

		ocupado += int(tmp.Part_size)

		if tmp.Part_status == '0' {
			//Si el primer EBR no está siendo usado, se inserta la partición
			if (ocupado + int(logica.Part_size)) <= int(ext.Part_size) {
				//Si cabe en la partición extendida
				logica.Part_start = ext.Part_start
				file.Seek(int64(ext.Part_start), 0) //Para actualizar el ebr
				binary.Write(file, binary.BigEndian, &logica)
				consola += "¡Partición lógica <" + string(logica.Part_name[:]) + "> creada!\n"
			} else {
				consola += "Error: Espacio insuficiente en la partición extendida\n"
			}
		} else {
			for tmp.Part_next != -1 {
				file.Seek(int64(tmp.Part_next), 0)
				binary.Read(file, binary.BigEndian, &tmp)
				ocupado += int(tmp.Part_size)
				if tmp.Part_next == -1 {
					//Se encontró otro ebr
					ultimo = tmp
				}
			}

			if (ocupado + int(logica.Part_size)) <= int(ext.Part_size) {
				//Ya estoy en el final de la lista enlazada, ya puedo colocar la partición
				logica.Part_start = ultimo.Part_start + ultimo.Part_size //Inicio de la nueva partición
				ultimo.Part_next = logica.Part_start                     //Apuntador del anterior a la nueva
				//Actualizo el anterior
				file.Seek(int64(ultimo.Part_start), 0)
				binary.Write(file, binary.BigEndian, &ultimo)
				//Escribo el nuevo
				file.Seek(int64(logica.Part_start), 0)
				binary.Write(file, binary.BigEndian, &logica)
				consola += "¡Partición lógica <" + string(name[:]) + "> creada!\n"
			} else {
				consola += "Error: Espacio insuficiente en la partición extendida\n"
			}
		}
	} else {
		consola += "Error: No existe una partición extendida\n"
	}
}

func ajustar(ruta string, size int, unit byte, name [16]byte, typep byte, fit byte) {
	var colocar bool = true
	file, err := os.Open(ruta)
	if err != nil {
		consola += "Error: No se puede abrir el disco duro\n"
	}
	defer file.Close()
	file.Seek(0, 0) //Coloco el puntero al inicio para obtener el mbr
	var mbr MBR
	binary.Read(file, binary.BigEndian, &mbr)

	//Primero verifico que no se repita el nombre de la partición en el disco
	for i := 0; i < 4; i++ {
		if name == mbr.Mbr_partition[i].Part_name {
			consola += "Error: Ya existe una partición llamada <" + string(name[:]) + ">\n"
			colocar = false
			break
		}
		if mbr.Mbr_partition[i].Part_type == 'E' { //Si es extendida, reviso en las particiones lógicas
			var tmp EBR
			file.Seek(int64(mbr.Mbr_partition[i].Part_start), 0) //Inicio de la partición extendida
			binary.Read(file, binary.BigEndian, &tmp)

			if name == tmp.Part_name {
				consola += "Error: Ya existe una partición llamada <" + string(name[:]) + ">\n"
				colocar = false
				break
			}
			for tmp.Part_next != -1 {
				if name == tmp.Part_name {
					consola += "Error: Ya existe una partición llamada <" + string(name[:]) + ">\n"
					colocar = false
					break
				}
				file.Seek(int64(tmp.Part_next), 0)
				binary.Read(file, binary.BigEndian, &tmp)

				if tmp.Part_next == -1 {
					if name == tmp.Part_name {
						consola += "Error: Ya existe una partición llamada <" + string(name[:]) + ">\n"
						colocar = false
						break
					}
				}
			}

		}
	}
	if colocar {
		if typep == 'P' {
			ajustarP(ruta, size, name, fit)
		} else if typep == 'E' {
			ajustarE(ruta, size, name, fit)
		} else if typep == 'L' {
			ajustarL(ruta, size, name, fit)
		}
	}
}

func fdisk(parametros []string) {
	var fsize, funit, fpath, ftype, ffit, fname bool = false, false, false, false, false, false
	var size int
	var unit, typep, fit byte
	var path string
	var name [16]byte

	for len(parametros) > 0 {
		tmp := parametros[0]

		tipo, valor := getTipoValor(tmp)

		if tipo == ">size" {
			size, _ = strconv.Atoi(valor)
			fsize = true
		} else if tipo == ">unit" {
			if strings.EqualFold(valor, "b") {
				unit = 'B'
				funit = true
			} else if strings.EqualFold(valor, "k") {
				unit = 'K'
				funit = true
			} else if strings.EqualFold(valor, "m") {
				unit = 'M'
				funit = true
			} else {
				consola += "Error: Valor de >unit <" + valor + "> inválido\n"
				break
			}
		} else if strings.EqualFold(tipo, ">path") {
			valor = regresarEspacio(valor)
			if existsFile(valor) {
				path = valor
				fpath = true
			} else {
				consola += "Error: No se encontró un disco duro en <" + valor + ">\n"
				break
			}
		} else if strings.EqualFold(tipo, ">name") {
			valor = regresarEspacio(valor)
			copy(name[:], []byte(valor))
			fname = true
		} else if strings.EqualFold(tipo, ">type") {
			if strings.EqualFold(valor, "p") {
				typep = 'P'
				ftype = true
			} else if strings.EqualFold(valor, "e") {
				typep = 'E'
				ftype = true
			} else if strings.EqualFold(valor, "l") {
				typep = 'L'
				ftype = true
			} else {
				consola += "Tipo de partición <" + valor + "> no válido\n"
			}
		} else if strings.EqualFold(tipo, ">fit") {
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
				consola += "Ajuste <" + valor + "> inválido\n"
			}
		} else if tipo[0] == '#' {
			break
		} else {
			consola += "Parámetro <" + valor + "> no válido\n"
		}

		parametros = parametros[1:] //Elimino el parámetro que ya se analizó
	}

	if fsize && fpath && fname {
		if !funit {
			size = size * 1024 //Si no se ingresó, se toma kb por defecto
		} else {
			if unit == 'K' {
				size = size * 1024
			} else if unit == 'M' {
				size = size * 1024 * 1024
			}
		}
		if !ftype {
			typep = 'P' //Si no se ingresó, se toma Primaria por defecto
		}
		if !ffit {
			fit = 'W' //Si no se ingresó, se toma Worst Fit por defecto
		}
		ajustar(path, size, unit, name, typep, fit)

	} else {
		consola += "Error: Faltan parámetros obligatorios\n"
	}
}

func montarParticion(ruta string, name [16]byte) {
	var atributos Atributos
	atributos.nombre = name
	atributos.ruta = ruta
	encontrada := false
	file, err := os.Open(ruta)
	if err != nil {
		consola += "Error: No se puede abrir el disco duro\n"
	}
	defer file.Close()
	file.Seek(0, 0) //Coloco el puntero al inicio para obtener el mbr
	var mbr MBR
	binary.Read(file, binary.BigEndian, &mbr)

	//Primero verifico que exista la partición
	for i := 0; i < 4; i++ {
		if name == mbr.Mbr_partition[i].Part_name {
			encontrada = true
			atributos.inicio = int(mbr.Mbr_partition[i].Part_start)
			atributos.tamano = int(mbr.Mbr_partition[i].Part_size)
			atributos.tipo = mbr.Mbr_partition[i].Part_type
			break
		}
		if mbr.Mbr_partition[i].Part_type == 'E' { //Si es extendida, reviso en las particiones lógicas
			var tmp EBR
			file.Seek(int64(mbr.Mbr_partition[i].Part_start), 0) //Inicio de la partición extendida
			binary.Read(file, binary.BigEndian, &tmp)

			if name == tmp.Part_name {
				encontrada = true
				atributos.inicio = int(tmp.Part_start)
				atributos.tamano = int(tmp.Part_size) - int(unsafe.Sizeof(tmp)) //Le resto el ebr porque ese espacio no se usa para guardar archivos y carpetas
				atributos.tipo = 'L'
				break
			}

			for tmp.Part_next != -1 {

				if name == tmp.Part_name {
					encontrada = true
					atributos.inicio = int(tmp.Part_start)
					atributos.tamano = int(tmp.Part_size) - int(unsafe.Sizeof(tmp))
					atributos.tipo = 'L'
					break
				}
				file.Seek(int64(tmp.Part_next), 0)
				binary.Read(file, binary.BigEndian, &tmp)

				if tmp.Part_next == -1 {
					if name == tmp.Part_name {
						encontrada = true
						atributos.inicio = int(tmp.Part_start)
						atributos.tamano = int(tmp.Part_size) - int(unsafe.Sizeof(tmp))
						atributos.tipo = 'L'
						break
					}
				}
			}

		}
	}
	if encontrada {
		//Carner 202003381
		id := "81"
		numero := 0
		asci := 65

		for _, part := range montadas {
			if part.ruta == ruta {
				numero = part.numDisco
				atributos.numDisco = part.numDisco
				break
			}
		}
		if numero == 0 {
			//Si el número sigue siendo 0, entonces no se ha montado ninguna partición del disco, se crea un nuevo número
			ultDisco++
			numero, atributos.numDisco = ultDisco, ultDisco

		}
		id += strconv.Itoa(int(numero))
		for _, part := range montadas {
			if part.ruta == ruta {
				asci++
			}
		}
		id += string(rune(asci)) //Convierto a string según el número ascii
		/*TODO: Agregar código para última vez de montaje eb particiones formateadas*/
		montadas[id] = atributos
		/*Si la partición ya fue formateada, actualizo la última fecha de montaje*/
		file.Seek(int64(atributos.inicio), 0) //Me muevo al inicio de la partición
		var superbloque SuperBloque
		binary.Read(file, binary.BigEndian, &superbloque)
		if superbloque.S_filesystem_type != 0 {
			t := time.Now()
			fecha := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d",
				t.Year(), t.Month(), t.Day(),
				t.Hour(), t.Minute(), t.Second())
			copy(superbloque.S_mtime[:], []byte(fecha))
			superbloque.S_mnt_count++
			file.Seek(int64(atributos.inicio), 0)
			binary.Write(file, binary.BigEndian, &superbloque) //Escribo el superbloque con la información actualizada
		}
		consola += "¡Partición <" + string(name[:]) + "> montada! ID: " + id + "\n"

	} else {
		consola += "Error: No se encontró la partición\n"
	}

}
func mount(parametros []string) {
	fpath, fname := false, false
	var path string
	var name [16]byte

	for len(parametros) > 0 {
		tmp := parametros[0]
		tipo, valor := getTipoValor(tmp)

		if tipo == ">path" {
			valor = regresarEspacio(valor)
			if existsFile(valor) {
				path = valor
				fpath = true
			} else {
				consola += "Error: No se encontró el disco duro\n"
				break
			}
		} else if tipo == ">name" {
			valor = regresarEspacio(valor)
			copy(name[:], []byte(valor))
			fname = true

		} else if tipo[0] == '#' {
			break
		} else {
			consola += "Parámetro <" + valor + "> no válido\n"
		}

		parametros = parametros[1:] //Elimino el parámetro analizado
	}

	if fname && fpath {
		montarParticion(path, name)
	} else {
		consola += "Error: Faltan parámetros obligatorios\n"
	}
}

func getPathWName(ruta string) string {
	dir := filepath.Dir(ruta)
	return dir
}
func getFileName(ruta string) string {
	name := filepath.Base(ruta)
	return strings.TrimSuffix(name, filepath.Ext(name))
}

func dotToPng(path string, name string) {
	res := exec.Command("dot", "-Tpng", name+".dot", "-o", name+".jpg")
	res.Dir = path

	stdout, err := res.Output()

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// Print the output
	fmt.Println(string(stdout))

}

func repDisk(ruta string, id string) {
	colorMbr := "\"#39F91A\""
	colorParticion := "\"#7E1DF2\""
	colorEbr := "\"#F927A9\""
	colorEbrInfo := "\"#FA96D4\""
	colorLibre := "\"#D3D3D3\""
	graficar := false

	var mbr MBR
	var vacio Atributos
	//Primero reviso si la partición está montada
	if montadas[id] != vacio {
		graficar = true
	} else {
		consola += "Error: No se encontró el id <" + id + ">\n"
	}

	if graficar {
		rutaDot := getPathWName(ruta)
		rutaDot += "/" + getFileName(ruta) + ".dot"

		file, err := os.Create(rutaDot)
		if err != nil {
			consola += "Error: No se pudo crear el archivo\n"
			return
		}

		fileb, err := os.Open(montadas[id].ruta)
		if err != nil {
			consola += "Error: No se puede leer el disco duro\n"
		}

		fileb.Seek(0, 0) //Coloco el puntero al inicio para obtener el mbr
		binary.Read(fileb, binary.BigEndian, &mbr)

		dot := ""
		dot += "digraph G {\n"
		dot += "a0[shape=none label=<\n"
		dot += "<TABLE cellspacing=\"1\" cellpadding=\"0\">\n"
		dot += "<TR>\n"
		dot += "<TD bgcolor="
		dot += colorMbr
		dot += "> MBR </TD>\n"
		porcentaje, libre := 0, 0
		var calc float64
		ocupado := 0

		for i := 0; i < 4; i++ {
			ocupado += int(mbr.Mbr_partition[i].Part_size)
			if mbr.Mbr_partition[i].Part_type == 'E' && mbr.Mbr_partition[i].Part_status == '1' {
				//Si es extendida y está activa, reviso las lógicas

				dot += "<TD>\n"
				dot += "\n"
				dot += "<TABLE cellspacing=\"1\" cellpadding=\"0\">\n"
				dot += "<TR>\n"
				dot += "<TD color=\"#FFFFFF\">Extendida</TD>\n"
				dot += "</TR>\n"

				var tmp EBR
				fileb.Seek(int64(mbr.Mbr_partition[i].Part_start), 0)
				binary.Read(fileb, binary.BigEndian, &tmp)

				if tmp.Part_next == -1 {
					dot += "<TR>\n"
					dot += "<TD bgcolor=" + colorEbr + ">EBR</TD>\n"

					libre = int(mbr.Mbr_partition[i].Part_size)

					calc = 1 * float64(libre)
					calc = calc / float64(mbr.Mbr_tamano)
					calc *= 100
					porcentaje = int(math.Round(calc))
					if libre > 0 {
						dot += "<TD bgcolor=" + colorLibre + ">Libre <BR></BR> " + strconv.Itoa(int(porcentaje)) + "% del disco</TD>\n"
					}

				} else {
					dot += "<TR>\n"
				}
				/*
					scanner := bufio.NewScanner(os.Stdin)
				*/
				for tmp.Part_next != -1 {

					//scanner.Scan()
					dot += "<TD bgcolor=" + colorEbr + ">EBR</TD>\n"
					calc = float64(tmp.Part_size) * 1
					calc = calc / float64(mbr.Mbr_tamano)
					calc *= 100
					porcentaje = int(math.Round(calc))
					dot += "<TD bgcolor=" + colorEbrInfo + ">Lógica <BR></BR> " + strconv.Itoa(int(porcentaje)) + "% del disco</TD>\n"

					fileb.Seek(int64(tmp.Part_next), 0)
					binary.Read(fileb, binary.BigEndian, &tmp)

					if tmp.Part_next == -1 {
						//Código para graficar la última partición lógica
						dot += "<TD bgcolor=" + colorEbr + ">EBR</TD>\n"
						calc = float64(tmp.Part_size) * 1
						calc = calc / float64(mbr.Mbr_tamano)
						calc *= 100
						porcentaje = int(math.Round(calc))
						dot += "<TD bgcolor=" + colorEbrInfo + ">Lógica <BR></BR> " + strconv.Itoa(int(porcentaje)) + "% del disco</TD>\n"

						//Ahora reviso si hay espacio libre al final de la extendida
						libre = (int(mbr.Mbr_partition[i].Part_start) + int(mbr.Mbr_partition[i].Part_size)) - (int(tmp.Part_start) + int(tmp.Part_size))

						calc = float64(libre) * 1
						calc = calc / float64(mbr.Mbr_tamano)
						calc *= 100
						porcentaje = int(math.Round(calc))
						if libre > 0 {
							dot += "<TD bgcolor=" + colorLibre + ">Libre <BR></BR> " + strconv.Itoa(int(porcentaje)) + "% del disco</TD>\n"
						}
					}
				}
				dot += "</TR>\n"
				dot += "</TABLE>\n"
				dot += "</TD>\n"
			} else {
				if mbr.Mbr_partition[i].Part_status == '0' {
					//Partición que no está siendo usada
					libre = int(mbr.Mbr_partition[i].Part_size)
					calc = float64(libre) * 1
					calc = calc / float64(mbr.Mbr_tamano)
					calc *= 100
					porcentaje = int(math.Round(calc))

					dot += "<TD bgcolor=" + colorLibre + ">Libre <BR></BR> " + strconv.Itoa(int(porcentaje)) + "% del disco</TD>\n"
				} else {
					calc = float64(mbr.Mbr_partition[i].Part_size) * 1
					calc = calc / float64(mbr.Mbr_tamano)
					calc *= 100
					porcentaje = int(math.Round(calc))

					dot += "<TD bgcolor=" + colorParticion + ">Primaria <BR></BR> " + strconv.Itoa(int(porcentaje)) + "% del disco</TD>\n"
				}
			}
		}
		//Espacio libre al final del disco
		libre = int(mbr.Mbr_tamano) - (ocupado)
		if libre > 0 {
			calc = float64(libre) * 1
			calc = calc / float64(mbr.Mbr_tamano)
			calc *= 100
			porcentaje = int(math.Round(calc))

			dot += "<TD bgcolor=" + colorLibre + ">Libre <BR></BR> " + strconv.Itoa(int(porcentaje)) + "% del disco</TD>\n"
		}

		dot += "</TR>\n"
		dot += "</TABLE>\n"
		dot += ">]\n"
		dot += "label=\"" + getFileName(montadas[id].ruta) + ".dsk\""
		dot += "}\n"

		file.WriteString(dot)
		file.Close()
		fileb.Close()
		nombreA := getFileName(ruta)
		rutaA := getPathWName(ruta)
		/*
			comando := "dot -Tjpg "
			comando += rutaDot
			comando += " -o "
			comando += rutaA
			comando += "/"
			comando += nombreA
			comando += ".jpg"*/
		dotToPng(rutaA, nombreA)

		consola += "¡Reporte generado con éxito!\n"
		encodeBase64(rutaA + "/" + nombreA + ".jpg")
	}

}

func repSb(ruta string, id string) {
	colorSb, colorSbInfo := "\"#FF8B00\"", "\"#FEB358\""
	graficar := false

	var superbloque SuperBloque
	var vacio Atributos
	//Primero reviso si la partición está montada
	if montadas[id] != vacio {
		graficar = true
	} else {
		consola += "Error: No se encontró el id <" + id + ">\n"
	}

	if graficar {
		rutaDot := getPathWName(ruta)
		rutaDot += "/" + getFileName(ruta) + ".dot"

		file, err := os.Create(rutaDot)
		if err != nil {
			consola += "Error: No se pudo crear el archivo\n"
			return
		}

		fileb, err := os.Open(montadas[id].ruta)
		if err != nil {
			consola += "Error: No se puede leer el disco duro\n"
		}

		fileb.Seek(int64(montadas[id].inicio), 0) //Coloco el puntero al inicio de la partición para leer el superbloque
		binary.Read(fileb, binary.BigEndian, &superbloque)

		dot := ""
		dot += "digraph G {\n"
		dot += "a0[shape=none label=<\n"
		dot += "<TABLE cellspacing=\"0\" cellpadding=\"0\">\n"
		dot += "<TR>\n"
		dot += "<TD bgcolor="
		dot += colorSb
		dot += "> REPORTE DE SUPERBLOQUE</TD>\n"
		dot += "<TD bgcolor="
		dot += colorSb
		dot += "></TD>\n"
		dot += "</TR>\n"
		//Comienzo con la información del superbloque
		dot += "<TR>\n"
		dot += "<TD bgcolor=" + colorSbInfo + ">s_filesystem_type</TD>\n"
		dot += "<TD bgcolor=" + colorSbInfo + ">" + strconv.Itoa(int(superbloque.S_filesystem_type)) + "</TD>\n"
		dot += "</TR>\n"

		dot += "<TR>\n"
		dot += "<TD bgcolor=" + colorSbInfo + ">s_inodes_count</TD>\n"
		dot += "<TD bgcolor=" + colorSbInfo + ">" + strconv.Itoa(int(superbloque.S_inodes_count)) + "</TD>\n"
		dot += "</TR>\n"

		dot += "<TR>\n"
		dot += "<TD bgcolor=" + colorSbInfo + ">s_blocks_count</TD>\n"
		dot += "<TD bgcolor=" + colorSbInfo + ">" + strconv.Itoa(int(superbloque.S_blocks_count)) + "</TD>\n"
		dot += "</TR>\n"

		dot += "<TR>\n"
		dot += "<TD bgcolor=" + colorSbInfo + ">s_free_blocks_count</TD>\n"
		dot += "<TD bgcolor=" + colorSbInfo + ">" + strconv.Itoa(int(superbloque.S_free_blocks_count)) + "</TD>\n"
		dot += "</TR>\n"

		dot += "<TR>\n"
		dot += "<TD bgcolor=" + colorSbInfo + ">s_free_inodes_count</TD>\n"
		dot += "<TD bgcolor=" + colorSbInfo + ">" + strconv.Itoa(int(superbloque.S_free_inodes_count)) + "</TD>\n"
		dot += "</TR>\n"

		var s_mtime []byte

		for _, char := range superbloque.S_mtime {
			if char != 0 {
				s_mtime = append(s_mtime, char)
			}
		}
		dot += "<TR>\n"
		dot += "<TD bgcolor=" + colorSbInfo + ">s_mtime</TD>\n"
		dot += "<TD bgcolor=" + colorSbInfo + ">" + string(s_mtime[:]) + "</TD>\n"
		dot += "</TR>\n"

		dot += "<TR>\n"
		dot += "<TD bgcolor=" + colorSbInfo + ">s_mnt_count</TD>\n"
		dot += "<TD bgcolor=" + colorSbInfo + ">" + strconv.Itoa(int(superbloque.S_mnt_count)) + "</TD>\n"
		dot += "</TR>\n"

		dot += "<TR>\n"
		dot += "<TD bgcolor=" + colorSbInfo + ">s_magic</TD>\n"
		dot += "<TD bgcolor=" + colorSbInfo + ">" + strconv.Itoa(int(superbloque.S_magic)) + "</TD>\n"
		dot += "</TR>\n"

		dot += "<TR>\n"
		dot += "<TD bgcolor=" + colorSbInfo + ">s_inode_s</TD>\n"
		dot += "<TD bgcolor=" + colorSbInfo + ">" + strconv.Itoa(int(superbloque.S_inode_size)) + "</TD>\n"
		dot += "</TR>\n"

		dot += "<TR>\n"
		dot += "<TD bgcolor=" + colorSbInfo + ">s_block_s</TD>\n"
		dot += "<TD bgcolor=" + colorSbInfo + ">" + strconv.Itoa(int(superbloque.S_block_size)) + "</TD>\n"
		dot += "</TR>\n"

		dot += "<TR>\n"
		dot += "<TD bgcolor=" + colorSbInfo + ">s_first_ino</TD>\n"
		dot += "<TD bgcolor=" + colorSbInfo + ">" + strconv.Itoa(int(superbloque.S_first_ino)) + "</TD>\n"
		dot += "</TR>\n"

		dot += "<TR>\n"
		dot += "<TD bgcolor=" + colorSbInfo + ">s_first_block</TD>\n"
		dot += "<TD bgcolor=" + colorSbInfo + ">" + strconv.Itoa(int(superbloque.S_first_blo)) + "</TD>\n"
		dot += "</TR>\n"

		dot += "<TR>\n"
		dot += "<TD bgcolor=" + colorSbInfo + ">s_bm_inode_start</TD>\n"
		dot += "<TD bgcolor=" + colorSbInfo + ">" + strconv.Itoa(int(superbloque.S_bm_inode_start)) + "</TD>\n"
		dot += "</TR>\n"

		dot += "<TR>\n"
		dot += "<TD bgcolor=" + colorSbInfo + ">s_bm_block_start</TD>\n"
		dot += "<TD bgcolor=" + colorSbInfo + ">" + strconv.Itoa(int(superbloque.S_bm_block_start)) + "</TD>\n"
		dot += "</TR>\n"

		dot += "<TR>\n"
		dot += "<TD bgcolor=" + colorSbInfo + ">s_inode_start</TD>\n"
		dot += "<TD bgcolor=" + colorSbInfo + ">" + strconv.Itoa(int(superbloque.S_inode_start)) + "</TD>\n"
		dot += "</TR>\n"

		dot += "<TR>\n"
		dot += "<TD bgcolor=" + colorSbInfo + ">s_block_start</TD>\n"
		dot += "<TD bgcolor=" + colorSbInfo + ">" + strconv.Itoa(int(superbloque.S_block_start)) + "</TD>\n"
		dot += "</TR>\n"

		dot += "</TABLE>\n"
		dot += ">]\n"
		dot += "}\n"

		file.WriteString(dot)
		file.Close()
		fileb.Close()
		nombreA := getFileName(ruta)
		rutaA := getPathWName(ruta)
		dotToPng(rutaA, nombreA)

		consola += "¡Reporte generado con éxito!\n"
		encodeBase64(rutaA + "/" + nombreA + ".jpg")
	}
}

func repTree(ruta string, id string) {
	graficar := false

	var superbloque SuperBloque
	var vacio Atributos
	//Primero reviso si la partición está montada
	if montadas[id] != vacio {
		graficar = true
	} else {
		consola += "Error: No se encontró el id <" + id + ">\n"
	}

	if graficar {
		rutaDot := getPathWName(ruta)
		rutaDot += "/" + getFileName(ruta) + ".dot"

		file, err := os.Create(rutaDot)
		if err != nil {
			consola += "Error: No se pudo crear el archivo\n"
			return
		}

		fileb, err := os.OpenFile(montadas[id].ruta, os.O_RDWR, 0777)
		if err != nil {
			consola += "Error: No se puede leer el disco duro\n"
		}

		arbol = ""
		conexiones = ""
		dot := ""

		fileb.Seek(int64(montadas[id].inicio), 0)
		binary.Read(fileb, binary.BigEndian, &superbloque)
		var raiz Inodo
		nombreNodo := strconv.Itoa(int(superbloque.S_inode_start))

		fileb.Seek(int64(superbloque.S_inode_start), 0)
		//fmt.Println(superbloque.S_inode_start)
		//fmt.Println(superbloque.S_inode_start + int32(unsafe.Sizeof(Inodo{})))
		binary.Read(fileb, binary.BigEndian, &raiz)
		recorrer(raiz, fileb, nombreNodo)

		//fileb.Seek(int64(superbloque.S_block_start), 0)
		//fmt.Println(superbloque.S_block_start)
		//var bloquecarpeta BloqueCarpetas
		//binary.Read(fileb, binary.BigEndian, &bloquecarpeta)

		//fmt.Println(bloquecarpeta.B_content[2].B_inodo)

		//file.Seek(int64(bloquecarpeta.B_content[2].B_inodo), 0)
		//binary.Read(fileb, binary.BigEndian, &raiz)
		//fmt.Println(raiz.I_uid)
		dot += "digraph G {\n"
		dot += "rankdir=\"LR\"\n"
		dot += arbol
		dot += conexiones
		dot += "}"

		file.WriteString(dot)
		file.Close()
		fileb.Close()

		nombreA := getFileName(ruta)
		rutaA := getPathWName(ruta)
		dotToPng(rutaA, nombreA)

		consola += "¡Reporte generado con éxito!\n"
		encodeBase64(rutaA + "/" + nombreA + ".jpg")
	}
}

func recorrer(raiz Inodo, archivo *os.File, nombreNodo string) string {
	var carpeta BloqueCarpetas
	var barchivo BloqueArchivos
	var nombreNodo2 string

	var inodo Inodo

	arbol += "\n"
	for i := 0; i < 16; i++ {
		//Grafico el bloque correspondiente y busco el inodo siguiente
		if raiz.I_block[i] != 0 { //Si está ocupado
			//Leo el bloque
			if raiz.I_type == '0' { //Si es carpeta
				archivo.Seek(int64(raiz.I_block[i]), 0)
				binary.Read(archivo, binary.BigEndian, &carpeta)

				nombreNodo2 = strconv.Itoa(int(int(raiz.I_block[i])))

				arbol += nombreNodo2
				arbol += "[shape=none label=<\n"
				arbol += "<TABLE cellspacing=\"0\" cellpadding=\"0\">\n"
				arbol += "<TR>\n"
				arbol += "<TD bgcolor="
				arbol += "\"#FF0093 \""
				arbol += ">Bloque de carpeta</TD>\n"
				arbol += "<TD bgcolor="
				arbol += "\"#FF0093 \""
				arbol += "></TD>\n"
				arbol += "</TR>\n"

				var name1 []byte

				for _, char := range carpeta.B_content[0].B_name {
					if char != 0 {
						name1 = append(name1, char)
					}
				}
				arbol += "<TR>\n"
				arbol += "<TD bgcolor="
				arbol += "\"#FA8CCB\""
				arbol += ">b_name</TD>\n"
				arbol += "<TD bgcolor="
				arbol += "\"#FA8CCB \""
				arbol += ">"
				arbol += string(name1[:])
				arbol += "</TD>\n"
				arbol += "</TR>\n"

				var name2 []byte

				for _, char := range carpeta.B_content[1].B_name {
					if char != 0 {
						name2 = append(name2, char)
					}
				}
				arbol += "<TR>\n"
				arbol += "<TD bgcolor="
				arbol += "\"#FA8CCB\""
				arbol += ">b_name</TD>\n"
				arbol += "<TD bgcolor="
				arbol += "\"#FA8CCB \""
				arbol += ">"
				arbol += string(name2[:])
				arbol += "</TD>\n"
				arbol += "</TR>\n"

				var name3 []byte

				for _, char := range carpeta.B_content[2].B_name {
					if char != 0 {
						name3 = append(name3, char)
					}
				}
				arbol += "<TR>\n"
				arbol += "<TD bgcolor="
				arbol += "\"#FA8CCB\""
				arbol += ">b_name</TD>\n"
				arbol += "<TD bgcolor="
				arbol += "\"#FA8CCB \""
				arbol += ">"
				arbol += string(name3[:])
				arbol += "</TD>\n"
				arbol += "</TR>\n"

				var name4 []byte

				for _, char := range carpeta.B_content[3].B_name {
					if char != 0 {
						name4 = append(name4, char)
					}
				}
				arbol += "<TR>\n"
				arbol += "<TD bgcolor="
				arbol += "\"#FA8CCB\""
				arbol += ">b_name</TD>\n"
				arbol += "<TD bgcolor="
				arbol += "\"#FA8CCB \""
				arbol += ">"
				arbol += string(name4[:])
				arbol += "</TD>\n"
				arbol += "</TR>\n"

				arbol += "</TABLE>\n"
				arbol += ">]\n"
				if i == 0 { //Para no regresar a la raiz
					for j := 2; j < 4; j++ {
						//fmt.Println(j)
						if carpeta.B_content[j].B_inodo != 0 {
							conexiones += nombreNodo2
							conexiones += "->"
							conexiones += strconv.Itoa(int(carpeta.B_content[j].B_inodo))
							conexiones += "\n"
							archivo.Seek(int64(carpeta.B_content[j].B_inodo), 0)
							binary.Read(archivo, binary.BigEndian, &inodo)
							//fmt.Println(inodo.I_type)
							recorrer(inodo, archivo, strconv.Itoa(int(carpeta.B_content[j].B_inodo)))
						}
					}
				} else {
					for j := 0; j < 4; j++ {
						if carpeta.B_content[j].B_inodo != 0 {
							conexiones += nombreNodo2
							conexiones += "->"
							conexiones += strconv.Itoa(int(carpeta.B_content[j].B_inodo))
							conexiones += "\n"
							archivo.Seek(int64(carpeta.B_content[j].B_inodo), 0)
							binary.Read(archivo, binary.BigEndian, &inodo)

							recorrer(inodo, archivo, strconv.Itoa(int(carpeta.B_content[j].B_inodo)))
						}
					}
				}

			} else { //Si es archivo

				archivo.Seek(int64(raiz.I_block[i]), 0)
				binary.Read(archivo, binary.BigEndian, &barchivo)

				nombreNodo2 = strconv.Itoa(int(raiz.I_block[i]))

				arbol += nombreNodo2
				arbol += "[shape=none label=<\n"
				arbol += "<TABLE cellspacing=\"0\" cellpadding=\"0\">\n"
				arbol += "<TR>\n"
				arbol += "<TD bgcolor="
				arbol += "\"#FFF000\""
				arbol += ">Bloque de archivo</TD>\n"
				arbol += "<TD bgcolor="
				arbol += "\"#FFF000\""
				arbol += "></TD>\n"
				arbol += "</TR>\n"

				arbol += "<TR>\n"
				arbol += "<TD bgcolor="
				arbol += "\"#FFF66C\""
				arbol += ">"

				for k := 0; k < 64; k++ {
					if barchivo.B_content[k] == '\n' {
						arbol += "<BR></BR>"
					} else if barchivo.B_content[k] < 32 || barchivo.B_content[k] > 126 {
						//No pasa nada
						fmt.Print("")
					} else {
						arbol += string(barchivo.B_content[k])
					}
				}
				//arbol+=barchivo.b_content;
				arbol += "</TD>\n"
				arbol += "</TR>\n"

				arbol += "</TABLE>\n"
				arbol += ">]\n"

			}

		}

	}

	//Grafico el inodo actual
	arbol += nombreNodo
	arbol += "[shape=none label=<\n"
	arbol += "<TABLE cellspacing=\"0\" cellpadding=\"0\">\n"
	arbol += "<TR>\n"
	arbol += "<TD bgcolor="
	arbol += "\"#00C9FF\""
	arbol += "> Inodo</TD>\n"
	arbol += "<TD bgcolor="
	arbol += "\"#00C9FF\""
	arbol += "></TD>\n"
	arbol += "</TR>\n"

	arbol += "<TR>\n"
	arbol += "<TD bgcolor="
	arbol += "\"#64DEFF\""
	arbol += ">I_uid</TD>\n"
	arbol += "<TD bgcolor="
	arbol += "\"#64DEFF \""
	arbol += ">" + strconv.Itoa(int(raiz.I_uid))
	arbol += "</TD>\n"
	arbol += "</TR>\n"

	arbol += "<TR>\n"
	arbol += "<TD bgcolor="
	arbol += "\"#64DEFF\""
	arbol += ">I_gid</TD>\n"
	arbol += "<TD bgcolor="
	arbol += "\"#64DEFF \""
	arbol += ">" + strconv.Itoa(int(raiz.I_gid))
	arbol += "</TD>\n"
	arbol += "</TR>\n"

	arbol += "<TR>\n"
	arbol += "<TD bgcolor="
	arbol += "\"#64DEFF\""
	arbol += ">I_size</TD>\n"
	arbol += "<TD bgcolor="
	arbol += "\"#64DEFF \""
	arbol += ">" + strconv.Itoa(int(raiz.I_size))
	arbol += "</TD>\n"
	arbol += "</TR>\n"

	var fechahora1 []byte

	for _, char := range raiz.I_atime {
		if char != 0 {
			fechahora1 = append(fechahora1, char)
		}
	}

	arbol += "<TR>\n"
	arbol += "<TD bgcolor="
	arbol += "\"#64DEFF\""
	arbol += ">I_atime</TD>\n"
	arbol += "<TD bgcolor="
	arbol += "\"#64DEFF\""
	arbol += ">" + string(fechahora1[:])
	arbol += "</TD>\n"
	arbol += "</TR>\n"

	var fechahora2 []byte

	for _, char := range raiz.I_ctime {
		if char != 0 {
			fechahora2 = append(fechahora2, char)
		}
	}
	arbol += "<TR>\n"
	arbol += "<TD bgcolor="
	arbol += "\"#64DEFF\""
	arbol += ">I_ctime</TD>\n"
	arbol += "<TD bgcolor="
	arbol += "\"#64DEFF\""
	arbol += ">" + string(fechahora2[:])
	arbol += "</TD>\n"
	arbol += "</TR>\n"

	var fechahora3 []byte

	for _, char := range raiz.I_mtime {
		if char != 0 {
			fechahora3 = append(fechahora3, char)
		}
	}

	arbol += "<TR>\n"
	arbol += "<TD bgcolor="
	arbol += "\"#64DEFF\""
	arbol += ">I_mtime</TD>\n"
	arbol += "<TD bgcolor="
	arbol += "\"#64DEFF\""
	arbol += ">" + string(fechahora3[:])
	arbol += "</TD>\n"
	arbol += "</TR>\n"

	arbol += "<TR>\n"
	arbol += "<TD bgcolor="
	arbol += "\"#64DEFF\""
	arbol += ">I_type</TD>\n"
	arbol += "<TD bgcolor="
	arbol += "\"#64DEFF\""
	arbol += ">"
	if raiz.I_type == '0' {
		arbol += "0"
	} else {
		arbol += "1"
	}

	arbol += "</TD>\n"
	arbol += "</TR>\n"

	arbol += "<TR>\n"
	arbol += "<TD bgcolor="
	arbol += "\"#64DEFF\""
	arbol += ">I_perm</TD>\n"
	arbol += "<TD bgcolor="
	arbol += "\"#64DEFF \""
	arbol += ">" + strconv.Itoa(int(raiz.I_perm))
	arbol += "</TD>\n"
	arbol += "</TR>\n"

	for i := 0; i <= 15; i++ {
		if raiz.I_block[i] != 0 {
			conexiones += nombreNodo //Los nombres de los nodos será su dirección en disco
			conexiones += "->"
			conexiones += strconv.Itoa(int(raiz.I_block[i]))
			conexiones += "\n"
		}
	}

	arbol += "</TABLE>\n"
	arbol += ">]\n"
	return arbol
}

func repFile(path string, id string, rutaF string) {
	graficar := false

	var superbloque SuperBloque
	var vacio Atributos
	//Primero reviso si la partición está montada
	if montadas[id] != vacio {
		graficar = true
	} else {
		consola += "Error: No se encontró el id <" + id + ">\n"
	}

	if graficar {

		file, err := os.Create(path)
		if err != nil {
			consola += "Error: No se pudo crear el archivo\n"
			return
		}

		fileb, err := os.OpenFile(montadas[id].ruta, os.O_RDWR, 0777)
		if err != nil {
			consola += "Error: No se puede leer el disco duro\n"
		}
		defer file.Close()
		defer fileb.Close()
		fileb.Seek(int64(montadas[id].inicio), 0)
		binary.Read(fileb, binary.BigEndian, &superbloque)

		//Muevo el puntero al inicio de los inodos, para buscar el inodo raiz
		fileb.Seek(int64(superbloque.S_inode_start), 0)

		var inodo Inodo
		binary.Read(fileb, binary.BigEndian, &inodo)

		//Hago una lista con los dferentes directorios
		lista_ruta := strings.Split(rutaF, "/")
		lista_ruta = lista_ruta[1:] //Si comienza con / la primera posición es un espacio vacío
		var carpeta BloqueCarpetas
		pos := 0
		finded := false //Para determinar si se encontró o no
		/*Primero voy a encontrar la carpeta que contiene al archivo*/
		if rutaF[0] == '/' { //Si no empieza con la raiz, la ruta está mal
			for len(lista_ruta) > 0 {
				if inodo.I_type == '0' { //Si es carpeta
					for i := 0; i <= 15; i++ { //recorro los apuntadores
						if inodo.I_block[i] != 0 { //Si está ocupado
							fileb.Seek(int64(inodo.I_block[i]), 0)
							binary.Read(fileb, binary.BigEndian, &carpeta)

							if i == 0 {
								//fmt.Println(string(carpeta.B_content[2].B_name[:]))
								//fmt.Println(lista_ruta[0])

								var name []byte
								for _, char := range carpeta.B_content[2].B_name {
									if char != 0 {
										name = append(name, char)
									}
								}

								var name2 []byte
								for _, char := range carpeta.B_content[3].B_name {
									if char != 0 {
										name2 = append(name2, char)
									}
								}
								//name2 := lista_ruta[0]
								if string(name[:]) == lista_ruta[0] {
									lista_ruta = lista_ruta[1:]
									if len(lista_ruta) == 0 {
										//Si ya se encontraron todos los directorios, doy por encontrada la carpeta
										finded = true
										pos = 2
										break
									} else {
										fileb.Seek(int64(carpeta.B_content[2].B_inodo), 0)
										binary.Read(fileb, binary.BigEndian, &inodo)
										break
									}

								} else if string(name2[:]) == lista_ruta[0] {
									lista_ruta = lista_ruta[1:]
									if len(lista_ruta) == 0 {
										//Si ya se encontraron todos los directorios, doy por encontrada la carpeta
										finded = true
										pos = 3
										break
									} else {
										fileb.Seek(int64(carpeta.B_content[3].B_inodo), 0)
										binary.Read(fileb, binary.BigEndian, &inodo)
										break
									}
								}

							} else {
								var name []byte
								for _, char := range carpeta.B_content[0].B_name {
									if char != 0 {
										name = append(name, char)
									}
								}

								var name1 []byte
								for _, char := range carpeta.B_content[1].B_name {
									if char != 0 {
										name1 = append(name1, char)
									}
								}

								var name2 []byte
								for _, char := range carpeta.B_content[2].B_name {
									if char != 0 {
										name2 = append(name2, char)
									}
								}

								var name3 []byte
								for _, char := range carpeta.B_content[3].B_name {
									if char != 0 {
										name3 = append(name3, char)
									}
								}
								if string(name[:]) == lista_ruta[0] {
									lista_ruta = lista_ruta[1:]
									if len(lista_ruta) == 0 {
										//Si ya se encontraron todos los directorios, doy por encontrada la carpeta
										finded = true
										pos = 0
										break
									} else {
										fileb.Seek(int64(carpeta.B_content[0].B_inodo), 0)
										//Me muevo al inodo del apuntador
										binary.Read(fileb, binary.BigEndian, &inodo)
										break
									}
								} else if string(name1[:]) == lista_ruta[0] {
									lista_ruta = lista_ruta[1:]
									if len(lista_ruta) == 0 {
										//Si ya se encontraron todos los directorios, doy por encontrada la carpeta
										finded = true
										pos = 1
										break
									} else {
										fileb.Seek(int64(carpeta.B_content[1].B_inodo), 0)
										//Me muevo al inodo del apuntador
										binary.Read(fileb, binary.BigEndian, &inodo)
										break
									}
								} else if string(name2[:]) == lista_ruta[0] {
									lista_ruta = lista_ruta[1:]
									if len(lista_ruta) == 0 {
										//Si ya se encontraron todos los directorios, doy por encontrada la carpeta
										finded = true
										pos = 2
										break
									} else {
										fileb.Seek(int64(carpeta.B_content[2].B_inodo), 0)
										binary.Read(fileb, binary.BigEndian, &inodo)
										break
									}
								} else if string(name3[:]) == lista_ruta[0] {
									lista_ruta = lista_ruta[1:]
									if len(lista_ruta) == 0 {
										//Si ya se encontraron todos los directorios, doy por encontrada la carpeta
										finded = true
										pos = 3
										break
									} else {
										fileb.Seek(int64(carpeta.B_content[3].B_inodo), 0)
										binary.Read(fileb, binary.BigEndian, &inodo)
										break
									}
								}
							}

						}

						if i == 15 && !finded {
							//Si ya se llegó al final y no se ha encontrado la dirección, no será encontrada
							//fmt.Println("FINAL")
							consola += "Error: No se encontró la ruta del archivo a reportar\n"
							//Limpio la lista para que se salga del ciclo
							lista_ruta = nil
						}
					}
				}
			}
		}

		if finded {
			//Ahora que encontré la carpeta, me muevo al archivo correspondiente

			fileb.Seek(int64(carpeta.B_content[pos].B_inodo), 0)
			binary.Read(fileb, binary.BigEndian, &inodo)
			var bloquea BloqueArchivos
			for i := 0; i <= 15; i++ {
				if inodo.I_block[i] != 0 {
					fileb.Seek(int64(inodo.I_block[i]), 0)
					binary.Read(fileb, binary.BigEndian, &bloquea)
					for _, char := range bloquea.B_content {
						if char != 0 {
							file.WriteString(string(char))
						}
					}
					//file.WriteString(string(bloquea.B_content[:]))

				}
			}

			consola += "¡Reporte generado correctamente!\n"
			encodeBase64(path)

		}
	}
}

func rep(parametros []string) {
	fname, fpath, fid, fruta := false, false, false, false

	var name, path, id, ruta string
	for len(parametros) > 0 {
		tmp := parametros[0]
		tipo, valor := getTipoValor(tmp)
		if tipo == ">name" {
			name = valor
			fname = true
		} else if tipo == ">path" {
			valor = regresarEspacio(valor)
			verifyDirectory(valor)
			path = valor
			fpath = true
		} else if tipo == ">id" {
			id = valor
			fid = true
		} else if tipo == ">ruta" {
			valor = regresarEspacio(valor)
			ruta = valor
			fruta = true
		} else if tipo[0] == '#' {
			break
		} else {
			consola += "Error: Parámetro <" + valor + "> no válido\n"
		}
		parametros = parametros[1:]
	}
	if fname && fpath && fid {
		if strings.EqualFold(name, "disk") {
			repDisk(path, id)
		} else if strings.EqualFold(name, "sb") {
			repSb(path, id)
		} else if strings.EqualFold(name, "tree") {
			repTree(path, id)
		} else if strings.EqualFold(name, "file") {
			if fruta {
				repFile(path, id, ruta)
			} else {
				consola += "Error: Falta el parámetro obligatorio >ruta"
			}
		} else {
			consola += "Error: reporte <" + name + "> no disponible\n"
		}
	} else {

		consola += "Error: Faltan parámetros obligatorios\n"
	}
}

func formatear(id string) {
	encontrada := false
	var vacio, particion Atributos
	if montadas[id] != vacio {
		encontrada = true
		particion = montadas[id]
	} else {
		consola += "Error: Id de partición <" + id + "> no encontrado\n"
	}

	if encontrada {
		sizeInodo := unsafe.Sizeof(Inodo{})
		sizeBloque := unsafe.Sizeof(BloqueArchivos{})
		sizeSuperBloque := unsafe.Sizeof(SuperBloque{})
		file, err := os.OpenFile(particion.ruta, os.O_RDWR, 0777)
		if err != nil {
			consola += "Error: No se puede abrir el disco duro\n"
		}
		defer file.Close()

		var n, bloques, inodos int
		var n_numerador, n_denominador, inodos_parcial float64
		var superbloque SuperBloque

		n_numerador = float64(particion.tamano) - float64(sizeSuperBloque)
		n_denominador = 4 + float64(sizeInodo) + (3 * float64(sizeBloque))
		n_numerador /= n_denominador
		n = int(n_numerador)

		inodos_parcial = float64(n / 4)
		inodos = int(inodos_parcial)

		bloques = 3 * inodos

		//Creo la carpeta raiz y el archivo users.txt
		//Inodo carpeta
		var raiz Inodo
		raiz.I_uid = 1
		raiz.I_gid = 1
		raiz.I_size = 27
		t := time.Now()
		fecha := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d",
			t.Year(), t.Month(), t.Day(),
			t.Hour(), t.Minute(), t.Second())
		copy(raiz.I_ctime[:], []byte(fecha))
		copy(raiz.I_mtime[:], []byte(fecha))

		raiz.I_type = '0' //Tipo carpeta
		raiz.I_perm = 664

		//Bloque carpeta
		var bloquecarpeta BloqueCarpetas
		copy(bloquecarpeta.B_content[0].B_name[:], []byte("/"))
		copy(bloquecarpeta.B_content[1].B_name[:], []byte("/"))
		copy(bloquecarpeta.B_content[2].B_name[:], []byte("users.txt"))

		//Inodo de archivo
		var archivo Inodo
		archivo.I_uid = 1
		archivo.I_gid = 1
		archivo.I_size = 27
		t = time.Now()
		fecha = fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d",
			t.Year(), t.Month(), t.Day(),
			t.Hour(), t.Minute(), t.Second())
		copy(archivo.I_ctime[:], []byte(fecha))
		copy(archivo.I_mtime[:], []byte(fecha))
		archivo.I_type = '1' //Tipo archivo
		archivo.I_perm = 664

		//Asocio el inodo raiz con el bloque de carpeta
		raiz.I_block[0] = int32(particion.inicio) + int32(sizeSuperBloque) + int32(inodos) + int32(bloques) + int32((int32(inodos) * int32(sizeInodo))) //Posición del primer bloque
		//Asocio el bloque de carpeta con el inodo de archivo
		bloquecarpeta.B_content[2].B_inodo = int32(particion.inicio) + int32(sizeSuperBloque) + int32(inodos) + int32(bloques) + int32(sizeInodo)

		//Escribo el contenido de users.txt
		var barchivo BloqueArchivos
		copy(barchivo.B_content[:], []byte("1,G,root\n1,U,root,root,123\n"))

		//Asocio el inodo de archivo con el bloque de archivo
		archivo.I_block[0] = int32(particion.inicio) + int32(sizeSuperBloque) + int32(inodos) + int32(bloques) + int32((int32(inodos) * int32(sizeInodo))) + int32(sizeBloque)

		//Agrego los atributos del superbloque
		superbloque.S_filesystem_type = 2
		superbloque.S_inodes_count = int32(inodos)
		superbloque.S_blocks_count = int32(bloques)
		superbloque.S_free_inodes_count = int32(inodos) - 2
		superbloque.S_free_blocks_count = int32(bloques) - 2
		copy(superbloque.S_mtime[:], []byte("0000-00-00T00:00:00"))
		superbloque.S_magic = 0xEF53
		superbloque.S_block_size = int32(sizeBloque)
		superbloque.S_inode_size = int32(sizeInodo)
		superbloque.S_first_ino = int32(particion.inicio) + int32(sizeSuperBloque) + int32(inodos) + int32(bloques) + int32(int32(2)*int32(sizeInodo))
		superbloque.S_first_blo = int32(particion.inicio) + int32(sizeSuperBloque) + int32(inodos) + int32(bloques) + int32(int32(inodos)*int32(sizeInodo)) + int32(int32(2)*int32(sizeBloque))
		superbloque.S_bm_inode_start = int32(particion.inicio) + int32(sizeSuperBloque)
		superbloque.S_bm_block_start = int32(particion.inicio) + int32(sizeSuperBloque) + int32(inodos)
		superbloque.S_inode_start = int32(particion.inicio) + int32(sizeSuperBloque) + int32(inodos) + int32(bloques)
		superbloque.S_block_start = int32(particion.inicio) + int32(sizeSuperBloque) + int32(inodos) + int32(bloques) + int32(int32(inodos)*int32(sizeInodo))

		/*Escribiendo las estructuras en el archivo*/
		file.Seek(int64(particion.inicio), 0)
		//Al inicio de la partición va el superbloque
		binary.Write(file, binary.BigEndian, &superbloque)
		var cero, uno byte = '0', '1'
		//Bitmap de inodos
		file.Seek(int64(superbloque.S_bm_inode_start), 0)
		binary.Write(file, binary.BigEndian, &uno) //Ya se escribieron dos inodos
		binary.Write(file, binary.BigEndian, &uno)
		//Ahora los 0s
		for i := 2; i < inodos; i++ {
			binary.Write(file, binary.BigEndian, &cero)
		}

		//Bitmap de bloques
		file.Seek(int64(superbloque.S_bm_block_start), 0)
		binary.Write(file, binary.BigEndian, &uno) //Ya se escribieron dos bloques
		binary.Write(file, binary.BigEndian, &uno)
		//Ahora los 0s
		for i := 2; i < bloques; i++ {
			binary.Write(file, binary.BigEndian, &cero)
		}

		//Escribo los inodos creados
		file.Seek(int64(superbloque.S_inode_start), 0)
		binary.Write(file, binary.BigEndian, &raiz)
		file.Seek(int64(bloquecarpeta.B_content[2].B_inodo), 0)
		binary.Write(file, binary.BigEndian, &archivo)

		//Escribo los bloques creados
		file.Seek(int64(superbloque.S_block_start), 0)
		binary.Write(file, binary.BigEndian, &bloquecarpeta)
		file.Seek(int64(archivo.I_block[0]), 0)
		binary.Write(file, binary.BigEndian, &barchivo)

		file.Seek(int64(superbloque.S_inode_start), 0)
		binary.Read(file, binary.BigEndian, raiz)
		binary.Read(file, binary.BigEndian, raiz)

		consola += "¡Partición formateada con el sistema de archivos EXT2!\n"

	}

}
func mkfs(parametros []string) {
	fid := false
	var id string
	for len(parametros) > 0 {
		tmp := parametros[0]
		tipo, valor := getTipoValor(tmp)
		if tipo == ">id" {
			valor = regresarEspacio(valor)
			id = valor
			fid = true
		} else if tipo == ">type" {
			if strings.EqualFold(valor, "full") {
				//ftype=true
			} else {
				consola += "Error: Valor para >type <" + valor + "> inválido\n"
			}
		} else if tipo[0] == '#' {
			break
		} else {
			consola += "Error: Parámetro <" + valor + "> no válido\n"
		}
		parametros = parametros[1:]
	}

	if fid {
		formatear(id)

	} else {
		consola += "Error: Parámetros insuficientes para realizar una acción\n"
	}
}

func IniciarSesion(user string, pass string, id string) (byte, string) {
	if !logged {
		encontrado := false //usuario
		var agrupo, auser, apass, auid string

		encontrada := false //Partición
		var vacio, particion Atributos
		if montadas[id] != vacio {
			encontrada = true
			particion = montadas[id]
		} else {
			consola += "Error: Id de partición <" + id + "> no encontrado\n"
			return '0', "Id de partición <" + id + "> no encontrado"
		}

		if encontrada {
			file, err := os.OpenFile(particion.ruta, os.O_RDWR, 0777)
			if err != nil {
				consola += "Error: No se puede abrir el disco duro\n"
			}
			defer file.Close()

			var inodo Inodo
			var barchivo BloqueArchivos
			var superbloque SuperBloque
			//Me muevo al inicio de la partición
			file.Seek(int64(particion.inicio), 0)
			binary.Read(file, binary.BigEndian, &superbloque)

			//Busco el inodo de users.txt, el cuál es el segundo inodo del bitmap
			file.Seek((int64(superbloque.S_inode_start) + int64(unsafe.Sizeof(Inodo{}))), 0)
			binary.Read(file, binary.BigEndian, &inodo)
			var lineas []string
			var contenido []byte

			for i := 0; i <= 15; i++ {
				if inodo.I_block[i] != 0 { //Si está ocupado
					file.Seek(int64(inodo.I_block[i]), 0)
					binary.Read(file, binary.BigEndian, &barchivo)
					/*Guardo los usuarios que puedan haber en todos los bloques de archivo del inodo*/
					for _, char := range barchivo.B_content {
						if char != 0 {
							contenido = append(contenido, char)
						}
					}
					//Separando por líneas

				}
			}

			lineas = strings.Split(string(contenido[:]), "\n")
			lineas = lineas[:len(lineas)-1] //Por el salto de línea al final, elimino el último elemento
			for len(lineas) > 0 {
				linea := strings.Split(lineas[0], ",")
				for len(linea) > 0 {
					auid = linea[0]
					if auid == "0" { //Si está eliminiado, deja de analizar
						break
					}
					linea = linea[1:]

					if linea[0] == "G" {
						//Estoy leyendo un grupo, así que me salto a otra línea
						linea = nil
					} else {
						linea = linea[1:]
						//Leo el grupo
						agrupo = linea[0]
						linea = linea[1:]
						//Leo el usuario
						auser = linea[0]
						linea = linea[1:]
						//Leo la contraseña
						apass = linea[0]

						if auser == user && apass == pass {
							//Se inicia sesión
							encontrado = true
							uidLog = auid
							idLog = id
							userLog = user
							gidLog = agrupo

						}
						linea = nil

					}

				}
				if encontrado {
					break
				}
				lineas = lineas[1:]
			}

			if encontrado {
				logged = true
				consola += "¡Sesión iniciada con éxito!\n"
				return '1', ""
			} else {
				consola += "Error: Usuario o contraseña incorrectos\n"
				return '0', "Error: Usuario o contraseña incorrectos"
			}
		}

	} else {
		consola += "Error: Ya existe una sesión iniciada\n"
		return '0', "Error: Ya existe una sesión iniciada"
	}
	return '0', ""
}

func login(parametros []string) {
	fuser, fpass, fid := false, false, false
	var user, pass, id string
	for len(parametros) > 0 {
		tmp := parametros[0]
		tipo, valor := getTipoValor(tmp)
		if tipo == ">id" {
			valor = regresarEspacio(valor)
			id = valor
			fid = true
		} else if tipo == ">pwd" {
			valor = regresarEspacio(valor)
			pass = valor
			fpass = true
		} else if tipo == ">user" {
			valor = regresarEspacio(valor)
			user = valor
			fuser = true
		} else if tipo[0] == '#' {
			break
		} else {
			consola += "Error: Parámetro <" + valor + "> no válido\n"
		}
		parametros = parametros[1:]
	}

	if fid && fuser && fpass {
		IniciarSesion(user, pass, id)
	} else {
		consola += "Error: Faltan parámetros obligatorios\n"
	}
}

func Logout() (byte, string) {
	if logged {
		userLog = ""
		idLog = ""
		gidLog = ""
		uidLog = ""
		logged = false
		consola += "¡Sesión Cerrada!"
		return '1', ""
	} else {
		consola += "Error: No existe una sesión iniciada\n"
		return '0', "Error: No existe una sesión iniciada\n"
	}
}

func encodeBase64(ruta string) {
	//fmt.Println(ruta)
	file, _ := os.Open(ruta)

	defer file.Close()
	reader := bufio.NewReader(file)
	content, _ := ioutil.ReadAll(reader)
	// Encode as base64.
	encoded := base64.StdEncoding.EncodeToString(content)
	//fmt.Println(encoded)
	var reporte Reporte

	reporte.Reporte = encoded
	reporte.Name = filepath.Base(ruta)

	Reportes.Reportes = append(Reportes.Reportes, reporte)
}

func crearGrupo(name string) {
	if logged {
		if userLog == "root" {
			file, err := os.OpenFile(montadas[idLog].ruta, os.O_RDWR, 0777)
			if err != nil {
				consola += "Error: No se puede abrir el disco duro\n"
			}
			defer file.Close()

			//Busco el inodo de users.txt, es decir, el segundo inodo de la tabla de inodos
			var superbloque SuperBloque
			file.Seek(int64(montadas[idLog].inicio), 0)
			binary.Read(file, binary.BigEndian, &superbloque)

			var inodo Inodo
			file.Seek(int64(superbloque.S_inode_start)+int64(unsafe.Sizeof(Inodo{})), 0)
			binary.Read(file, binary.BigEndian, &inodo)

			var contenido []byte
			var block, pos int //block -> Ultimo bloque de archivo ocupado, pos->posición en el bloque del último caracter usado
			var barchivo BloqueArchivos

			for i := 0; i <= 15; i++ {
				if inodo.I_block[i] != 0 {
					//Si el bloque está ocupado
					block = i
					file.Seek(int64(inodo.I_block[i]), 0)
					binary.Read(file, binary.BigEndian, &barchivo)

					for j := 0; j < 64; j++ {
						if barchivo.B_content[j] != 0 {
							contenido = append(contenido, barchivo.B_content[j])
							pos = j
						}
					}

				}
			}

			gid := 0 //Ultimo gid encontrado
			gidaux := 0
			encontrado := false
			lineas := strings.Split(string(contenido[:]), "\n")
			lineas = lineas[:len(lineas)-1] //Por el salto de línea al final, elimino el último elemento
			for len(lineas) > 0 {
				linea := strings.Split(lineas[0], ",")
				for len(linea) > 0 {
					gidaux, _ = strconv.Atoi(string(linea[0]))
					linea = linea[1:]
					if linea[0] == "G" {
						//Estoy leyendo un grupo
						//	fmt.Println("Grupito")
						if gidaux != 0 { //Si el grupo no ha sido eliminado, sigo analizando
							gid = gidaux
							linea = linea[1:]
							//Leo el grupo
							if linea[0] == name {
								encontrado = true
							}
							linea = linea[1:]
						} else {
							linea = nil
						}

					} else {
						//Si estoy leyendo un usuario, salto la linea
						linea = nil

					}
					linea = nil

				}
				if encontrado {
					break
				}
				lineas = lineas[1:]
			}

			if !encontrado {

				gid++
				pos++
				lineaNueva := ""
				var nuevoBloque bool = false
				var barchivo2 BloqueArchivos
				var bm byte
				var cont int
				file.Seek(int64(superbloque.S_bm_block_start), 0)
				for i := 0; i < int(superbloque.S_blocks_count); i++ {
					binary.Read(file, binary.BigEndian, &bm)
					if bm == '1' {
						cont++
					} else {
						break
						//Si es un cero, me detengo pues ya encontré el espacio
					}
				}
				if len(name) <= 10 {
					lineaNueva = strconv.Itoa(gid) + ",G," + name + "\n"
					inodo.I_size += int32(len(lineaNueva))
					lineaNuevaByte := []byte(lineaNueva)
					for _, char := range lineaNuevaByte {
						if pos == 64 {
							//Escribo los cambios en el bloque actual
							file.Seek(int64(inodo.I_block[block]), 0)
							binary.Write(file, binary.BigEndian, &barchivo)

							barchivo = barchivo2
							nuevoBloque = true
							pos = 0
							block++
							inodo.I_block[block] = superbloque.S_block_start + (int32(int32(cont+1) * int32(unsafe.Sizeof(BloqueArchivos{}))))

							//Actualizando el bitmap de bloques
							var uno byte = '1'
							file.Seek(int64(superbloque.S_bm_block_start), 0)
							for i := 0; i < int(superbloque.S_blocks_count); i++ {
								binary.Read(file, binary.BigEndian, &bm)
								if bm == '1' {
								} else {
									file.Seek(int64(superbloque.S_bm_block_start)+int64(i), 0)
									break
									//Si es un cero, me detengo pues ya encontré el espacio
								}
							}
							binary.Write(file, binary.BigEndian, uno)
						}
						barchivo.B_content[pos] = char
						pos++
					}
					/*TODO SABADO : Escribir el bloque dependiendo del bitmap de bloques y actualizarlo
					  Actualizar el bloque según el correspondiente block*/
					if nuevoBloque {
						superbloque.S_first_blo += int32(unsafe.Sizeof(BloqueArchivos{}))
						superbloque.S_free_blocks_count -= 1
						//Escribo el nuevo bloque
						file.Seek(int64(inodo.I_block[block]), 0)
						binary.Write(file, binary.BigEndian, &barchivo)

						//Actualizo el superbloque
						file.Seek(int64(montadas[idLog].inicio), 0)
						binary.Write(file, binary.BigEndian, &superbloque)

					} else {
						//Actualizo el bloque de archivos
						file.Seek(int64(inodo.I_block[block]), 0)
						binary.Write(file, binary.BigEndian, &barchivo)
					}
					//Actualizo el inodo
					file.Seek(int64(superbloque.S_inode_start)+int64(unsafe.Sizeof(Inodo{})), 0)
					binary.Write(file, binary.BigEndian, &inodo)
					consola += "¡Grupo <" + name + "> creado con éxito!\n"

				} else {
					consola += "Error: El nombre del grupo debe tener un máximo de 10 caracteres\n"
				}
			} else {
				consola += "Error: Ya existe un grupo llamado <" + name + ">\n"
			}

		} else {
			consola += "Error: Acción no disponible para el usuario actual\n"
		}
	} else {
		consola += "Error: No hay una sesión iniciada\n"
	}
}

func mkgrp(parametros []string) {
	fname := false
	var name string
	for len(parametros) > 0 {
		tmp := parametros[0]
		tipo, valor := getTipoValor(tmp)
		if tipo == ">name" {
			valor = regresarEspacio(valor)
			name = valor
			fname = true
		} else if tipo[0] == '#' {
			break
		} else {
			consola += "Error: Parámetro <" + valor + "> no válido\n"
		}
		parametros = parametros[1:]
	}

	if fname {
		crearGrupo(name)
	} else {
		consola += "Error: Faltan parámetros obligatorios\n"
	}
}
func eliminarGrupo(name string) {
	if logged {
		if userLog == "root" {
			file, err := os.OpenFile(montadas[idLog].ruta, os.O_RDWR, 0777)
			if err != nil {
				consola += "Error: No se puede abrir el disco duro\n"
			}
			defer file.Close()

			//Busco el inodo de users.txt, es decir, el segundo inodo de la tabla de inodos
			var superbloque SuperBloque
			file.Seek(int64(montadas[idLog].inicio), 0)
			binary.Read(file, binary.BigEndian, &superbloque)

			var inodo Inodo
			file.Seek(int64(superbloque.S_inode_start)+int64(unsafe.Sizeof(Inodo{})), 0)
			binary.Read(file, binary.BigEndian, &inodo)

			var contenido []byte

			var barchivo BloqueArchivos

			for i := 0; i <= 15; i++ {
				if inodo.I_block[i] != 0 {
					//Si el bloque está ocupado

					file.Seek(int64(inodo.I_block[i]), 0)
					binary.Read(file, binary.BigEndian, &barchivo)

					for j := 0; j < 64; j++ {
						if barchivo.B_content[j] != 0 {
							contenido = append(contenido, barchivo.B_content[j])

						}
					}

				}
			}

			gidaux := 0
			encontrado := false
			punteroL := 0 //Actualiza la posición en la que inicia la línea del grupo a eliminar
			lineas := strings.Split(string(contenido[:]), "\n")
			lineas = lineas[:len(lineas)-1] //Por el salto de línea al final, elimino el último elemento
			for len(lineas) > 0 {

				linea := strings.Split(lineas[0], ",")

				for len(linea) > 0 {
					gidaux, _ = strconv.Atoi(string(linea[0]))
					linea = linea[1:]
					if linea[0] == "G" {
						//Estoy leyendo un grupo
						//	fmt.Println("Grupito")
						if gidaux != 0 { //Si el grupo no ha sido eliminado, sigo analizando

							linea = linea[1:]
							//Leo el grupo
							if linea[0] == name {
								encontrado = true
							}
							linea = linea[1:]
						} else {
							linea = nil
						}

					} else {
						//Si estoy leyendo un usuario, salto la linea
						linea = nil

					}
					linea = nil

				}
				if encontrado {
					break
				}
				punteroL += len(lineas[0]) + 1 //Sumo 1 por el salto de línea perdido
				lineas = lineas[1:]

			}

			if encontrado {
				var block = 0 //Bloque que se va a modificar
				punteroOr := punteroL
				if punteroL >= 64 {
					block = int(math.Round(float64(punteroL / 64)))
					punteroL = punteroL - ((block) * 64)

				}
				//Leo el bloque correcto
				file.Seek(int64(inodo.I_block[block]), 0)
				binary.Read(file, binary.BigEndian, &barchivo)

				if contenido[punteroOr+1] == ',' {
					barchivo.B_content[punteroL] = '0'
				} else {
					barchivo.B_content[punteroL] = '0'
					barchivo.B_content[punteroL+1] = '0'
				}
				//Actualizo el bloque de archivos
				file.Seek(int64(inodo.I_block[block]), 0)
				binary.Write(file, binary.BigEndian, &barchivo)

				consola += "¡Grupo eliminado con éxito!\n"

			} else {
				consola += "Error: No se encontró el grupo <" + name + ">\n"
			}

		} else {
			consola += "Error: Acción no disponible para el usuario actual\n"
		}
	} else {
		consola += "Error: No hay una sesión iniciada\n"
	}
}
func rmgrp(parametros []string) {
	fname := false
	var name string
	for len(parametros) > 0 {
		tmp := parametros[0]
		tipo, valor := getTipoValor(tmp)
		if tipo == ">name" {
			valor = regresarEspacio(valor)
			name = valor
			fname = true
		} else if tipo[0] == '#' {
			break
		} else {
			consola += "Error: Parámetro <" + valor + "> no válido\n"
		}
		parametros = parametros[1:]
	}

	if fname {
		eliminarGrupo(name)
	} else {
		consola += "Error: Faltan parámetros obligatorios\n"
	}
}

func existsGrp(contenido []byte, name string) bool {
	encontrado := false
	var gidaux int
	lineas := strings.Split(string(contenido[:]), "\n")
	lineas = lineas[:len(lineas)-1] //Por el salto de línea al final, elimino el último elemento
	for len(lineas) > 0 {
		linea := strings.Split(lineas[0], ",")
		for len(linea) > 0 {
			gidaux, _ = strconv.Atoi(string(linea[0]))
			linea = linea[1:]
			if linea[0] == "G" {

				if gidaux != 0 { //Si el grupo no ha sido eliminado, sigo analizando
					linea = linea[1:]
					//Leo el grupo
					if linea[0] == name {
						encontrado = true
					}
					linea = linea[1:]
				} else {
					linea = nil
				}

			} else {
				//Si estoy leyendo un usuario, salto la linea
				linea = nil

			}
			linea = nil

		}
		if encontrado {
			break
		}
		lineas = lineas[1:]
	}

	return encontrado
}

func crearUsuario(user string, password string, group string) {
	if logged {
		if userLog == "root" {
			file, err := os.OpenFile(montadas[idLog].ruta, os.O_RDWR, 0777)
			if err != nil {
				consola += "Error: No se puede abrir el disco duro\n"
			}
			defer file.Close()

			//Busco el inodo de users.txt, es decir, el segundo inodo de la tabla de inodos
			var superbloque SuperBloque
			file.Seek(int64(montadas[idLog].inicio), 0)
			binary.Read(file, binary.BigEndian, &superbloque)

			var inodo Inodo
			file.Seek(int64(superbloque.S_inode_start)+int64(unsafe.Sizeof(Inodo{})), 0)
			binary.Read(file, binary.BigEndian, &inodo)

			var contenido []byte
			var block, pos int //block -> Ultimo bloque de archivo ocupado, pos->posición en el bloque del último caracter usado
			var barchivo BloqueArchivos

			for i := 0; i <= 15; i++ {
				if inodo.I_block[i] != 0 {
					//Si el bloque está ocupado
					block = i
					file.Seek(int64(inodo.I_block[i]), 0)
					binary.Read(file, binary.BigEndian, &barchivo)

					for j := 0; j < 64; j++ {
						if barchivo.B_content[j] != 0 {
							contenido = append(contenido, barchivo.B_content[j])
							pos = j
						}
					}

				}
			}

			uid := 0 //Ultimo gid encontrado
			uidaux := 0
			encontrado := false
			lineas := strings.Split(string(contenido[:]), "\n")
			if existsGrp(contenido, group) {
				lineas = lineas[:len(lineas)-1] //Por el salto de línea al final, elimino el último elemento
				for len(lineas) > 0 {
					linea := strings.Split(lineas[0], ",")
					for len(linea) > 0 {
						uidaux, _ = strconv.Atoi(string(linea[0]))
						linea = linea[1:]
						if linea[0] == "G" {
							linea = nil //Si es un grupo, salto la línea
						} else {
							//Estoy leyendo un usuario

							if uidaux != 0 { //Si el usuario no ha sido eliminado
								uid = uidaux

								linea = linea[1:] //Gripo
								linea = linea[1:] //usuario
								if linea[0] == user {
									encontrado = true
									linea = nil

								}

								//linea = linea[1:]
								//Leo la contraseña

							} else {
								linea = nil
							}
							linea = nil

						}
						linea = nil

					}
					if encontrado {
						break
					}
					lineas = lineas[1:]
				}
			} else {
				consola += "Error: No se encontró el grupo <" + group + ">\n"
				return
			}

			if !encontrado {

				uid++
				pos++
				lineaNueva := ""
				var nuevoBloque bool = false
				var barchivo2 BloqueArchivos
				var bm byte
				var cont int
				file.Seek(int64(superbloque.S_bm_block_start), 0)
				for i := 0; i < int(superbloque.S_blocks_count); i++ {
					binary.Read(file, binary.BigEndian, &bm)
					if bm == '1' {
						cont++
					} else {
						break
						//Si es un cero, me detengo pues ya encontré el espacio
					}
				}
				if len(user) <= 10 {
					if len(password) <= 10 {
						lineaNueva = strconv.Itoa(uid) + ",U," + group + "," + user + "," + password + "\n"
						inodo.I_size += int32(len(lineaNueva))
						lineaNuevaByte := []byte(lineaNueva)
						for _, char := range lineaNuevaByte {
							if pos == 64 {
								//Escribo los cambios en el bloque actual
								file.Seek(int64(inodo.I_block[block]), 0)
								binary.Write(file, binary.BigEndian, &barchivo)

								barchivo = barchivo2
								nuevoBloque = true
								pos = 0
								block++
								inodo.I_block[block] = superbloque.S_block_start + int32((int32((cont + 1)) * int32(unsafe.Sizeof(BloqueArchivos{}))))

								//Actualizando el bitmap de bloques
								var uno byte = '1'
								file.Seek(int64(superbloque.S_bm_block_start), 0)
								for i := 0; i < int(superbloque.S_blocks_count); i++ {
									binary.Read(file, binary.BigEndian, &bm)
									if bm == '1' {
									} else {
										file.Seek(int64(superbloque.S_bm_block_start)+int64(i), 0)
										break
										//Si es un cero, me detengo pues ya encontré el espacio
									}
								}
								binary.Write(file, binary.BigEndian, uno)
							}
							barchivo.B_content[pos] = char
							pos++
						}

						if nuevoBloque {
							superbloque.S_free_blocks_count -= 1
							superbloque.S_first_blo += int32(unsafe.Sizeof(BloqueArchivos{}))

							//Escribo el nuevo bloque
							file.Seek(int64(inodo.I_block[block]), 0)
							binary.Write(file, binary.BigEndian, &barchivo)

							//Actualizo el superbloque
							file.Seek(int64(montadas[idLog].inicio), 0)
							binary.Write(file, binary.BigEndian, &superbloque)

						} else {
							//Actualizo el bloque de archivos
							file.Seek(int64(inodo.I_block[block]), 0)
							binary.Write(file, binary.BigEndian, &barchivo)
						}
						//Actualizo el inodo
						file.Seek(int64(superbloque.S_inode_start)+int64(unsafe.Sizeof(Inodo{})), 0)
						binary.Write(file, binary.BigEndian, &inodo)
						consola += "¡Usuario <" + user + "> creado con éxito!\n"
					} else {
						consola += "Error: La contraseña debe tener un máximo de 10 caracteres"
					}
				} else {
					consola += "Error: El nombre del usuario debe tener un máximo de 10 caracteres\n"
				}
			} else {
				consola += "Error: Ya existe un usuario llamado <" + user + ">\n"
			}

		} else {
			consola += "Error: Acción no disponible para el usuario actual\n"
		}
	} else {
		consola += "Error: No hay una sesión iniciada\n"
	}
}

func mkusr(parametros []string) {
	fuser, fpwd, fgrp := false, false, false
	var user, pwd, grp string
	for len(parametros) > 0 {
		tmp := parametros[0]
		tipo, valor := getTipoValor(tmp)
		if tipo == ">user" {
			valor = regresarEspacio(valor)
			user = valor
			fuser = true
		} else if tipo == ">pwd" {
			valor = regresarEspacio(valor)
			pwd = valor
			fpwd = true
		} else if tipo == ">grp" {
			valor = regresarEspacio(valor)
			grp = valor
			fgrp = true
		} else if tipo[0] == '#' {
			break
		} else {
			consola += "Error: Parámetro <" + valor + "> no válido\n"
		}
		parametros = parametros[1:]
	}

	if fuser && fpwd && fgrp {
		crearUsuario(user, pwd, grp)
	} else {
		consola += "Error: Faltan parámetros obligatorios\n"
	}
}

func eliminarUsuario(name string) {
	if logged {
		if userLog == "root" {
			file, err := os.OpenFile(montadas[idLog].ruta, os.O_RDWR, 0777)
			if err != nil {
				consola += "Error: No se puede abrir el disco duro\n"
			}
			defer file.Close()

			//Busco el inodo de users.txt, es decir, el segundo inodo de la tabla de inodos
			var superbloque SuperBloque
			file.Seek(int64(montadas[idLog].inicio), 0)
			binary.Read(file, binary.BigEndian, &superbloque)

			var inodo Inodo
			file.Seek(int64(superbloque.S_inode_start)+int64(unsafe.Sizeof(Inodo{})), 0)
			binary.Read(file, binary.BigEndian, &inodo)

			var contenido []byte

			var barchivo BloqueArchivos

			for i := 0; i <= 15; i++ {
				if inodo.I_block[i] != 0 {
					//Si el bloque está ocupado

					file.Seek(int64(inodo.I_block[i]), 0)
					binary.Read(file, binary.BigEndian, &barchivo)

					for j := 0; j < 64; j++ {
						if barchivo.B_content[j] != 0 {
							contenido = append(contenido, barchivo.B_content[j])

						}
					}

				}
			}

			uidaux := 0
			encontrado := false
			punteroL := 0 //Puntero de inicio de linea
			lineas := strings.Split(string(contenido[:]), "\n")

			lineas = lineas[:len(lineas)-1] //Por el salto de línea al final, elimino el último elemento
			for len(lineas) > 0 {
				linea := strings.Split(lineas[0], ",")
				for len(linea) > 0 {
					uidaux, _ = strconv.Atoi(string(linea[0]))
					linea = linea[1:]
					if linea[0] == "G" {
						linea = nil //Si es un grupo, salto la línea
					} else {
						//Estoy leyendo un usuario

						if uidaux != 0 { //Si el usuario no ha sido eliminado

							linea = linea[1:] //Gripo
							linea = linea[1:] //usuario
							if linea[0] == name {
								encontrado = true
								linea = nil

							}

							//linea = linea[1:]
							//Leo la contraseña

						} else {
							linea = nil
						}
						linea = nil

					}
					linea = nil

				}
				if encontrado {
					break
				}
				punteroL += len(lineas[0]) + 1 //Sumo 1 por el salto de línea perdido
				lineas = lineas[1:]
			}

			if encontrado {
				var block = 0 //Bloque que se va a modificar
				punteroOr := punteroL
				if punteroL >= 64 {
					block = int(math.Round(float64(punteroL / 64)))
					punteroL = punteroL - ((block) * 64)

				}
				//Leo el bloque correcto
				file.Seek(int64(inodo.I_block[block]), 0)
				binary.Read(file, binary.BigEndian, &barchivo)

				if contenido[punteroOr+1] == ',' {
					barchivo.B_content[punteroL] = '0'
				} else {
					barchivo.B_content[punteroL] = '0'
					barchivo.B_content[punteroL+1] = '0'
				}
				//Actualizo el bloque de archivos
				file.Seek(int64(inodo.I_block[block]), 0)
				binary.Write(file, binary.BigEndian, &barchivo)

				consola += "¡Grupo eliminado con éxito!\n"

			} else {
				consola += "Error: No se encontró el grupo <" + name + ">\n"
			}

		} else {
			consola += "Error: Acción no disponible para el usuario actual\n"
		}

	} else {
		consola += "Error: No existe una sesión iniciada\n"
	}

}

func rmusr(parametros []string) {
	fuser := false
	var user string
	for len(parametros) > 0 {
		tmp := parametros[0]
		tipo, valor := getTipoValor(tmp)
		if tipo == ">name" {
			valor = regresarEspacio(valor)
			user = valor
			fuser = true
		} else if tipo[0] == '#' {
			break
		} else {
			consola += "Error: Parámetro <" + valor + "> no válido\n"
		}

		parametros = parametros[1:]

	}

	if fuser {
		eliminarUsuario(user)
	} else {
		consola += "Error: Faltan parámetros obligatorios\n"
	}
}

func crearArchivo(ruta string, r bool, size int, cont string) {
	if logged {
		file, err := os.OpenFile(montadas[idLog].ruta, os.O_RDWR, 0777)
		if err != nil {
			consola += "Error: No se puede abrir el disco duro\n"
		}
		defer file.Close()

		var superbloque SuperBloque

		var nombreArchivo string
		var padre string
		file.Seek(int64(montadas[idLog].inicio), 0)
		binary.Read(file, binary.BigEndian, &superbloque)

		//Muevo el puntero al inicio de los inodos, para buscar el inodo raiz
		posinodo := 0 //Posición en el disco del inodo de carpeta donde será escrito el nuevo archivo

		file.Seek(int64(superbloque.S_inode_start), 0)
		posinodo = int(superbloque.S_inode_start)
		var inodo Inodo
		binary.Read(file, binary.BigEndian, &inodo)

		//Hago una lista con los dferentes directorios
		lista_ruta := strings.Split(ruta, "/")
		lista_ruta = lista_ruta[1:] //Elimino la primera posición, pues es un espacio en blanco

		var carpeta BloqueCarpetas

		pos := 0 //Indice del bloque de carpeta donde será escrito el nuevo archivo

		/*Primero voy a encontrar la carpeta que contendtá al archivo*/
		if ruta[0] == '/' { //Si no empieza con la raiz, la ruta está mal
			nombreArchivo = lista_ruta[len(lista_ruta)-1] //La ultima entrada de la lita corresponde al nombre de la carpeta
			lista_ruta = lista_ruta[:len(lista_ruta)-1]   //Elimino el nombre del archivo de la lista
			if len(lista_ruta) == 0 {
				padre = "/"

			} else {
				padre = lista_ruta[len(lista_ruta)-1]
			}
			if len(lista_ruta) > 0 {
				for i := 0; i <= 15; i++ {
					if inodo.I_type == '0' { //Si es carpeta
						if inodo.I_block[i] != 0 {
							//Si el inodo está ocupado
							file.Seek(int64(inodo.I_block[i]), 0)
							binary.Read(file, binary.BigEndian, &carpeta)
							if i == 0 {
								//Porque el primer bloque de cada inodo de carpeta apunta primero a si mismo y a su padre
								for j := 2; j < 4; j++ {
									if carpeta.B_content[j].B_inodo != 0 {
										var nombreArr []byte
										for _, char := range carpeta.B_content[j].B_name {
											if char != 0 {
												nombreArr = append(nombreArr, char)
											}
										}

										if string(nombreArr[:]) == lista_ruta[0] {
											lista_ruta = lista_ruta[1:]
											file.Seek(int64(carpeta.B_content[j].B_inodo), 0)
											binary.Read(file, binary.BigEndian, &inodo)
											posinodo = int(carpeta.B_content[i].B_inodo)
											i = 0
											break
										}
									}
								}
							} else {
								for j := 0; j < 4; j++ {
									if carpeta.B_content[j].B_inodo != 0 {
										var nombreArr []byte
										for _, char := range carpeta.B_content[j].B_name {
											if char != 0 {
												nombreArr = append(nombreArr, char)
											}
										}

										if string(nombreArr[:]) == lista_ruta[0] {
											lista_ruta = lista_ruta[1:]
											file.Seek(int64(carpeta.B_content[j].B_inodo), 0)
											binary.Read(file, binary.BigEndian, &inodo)
											posinodo = int(carpeta.B_content[i].B_inodo)
											i = 0
											break
										}
									}
								}
							}

						}
					}

				}
			}
			fmt.Println("listo", padre, nombreArchivo)
			/*Ahora obtengo el contenido del archivo*/

			contenido := ""
			if len(cont) > 0 {
				//Dandole prioridad al parámetro cont
				filecont, err := os.Open(cont)
				if err != nil {
					consola += "Error: No se encontró el archivo de contenido \n"
				}

				defer filecont.Close()
				reader := bufio.NewReader(filecont)
				content, _ := ioutil.ReadAll(reader)

				contenido = string(content[:])
			} else {
				index := 0
				for i := 0; i < size; i++ {
					contenido += strconv.Itoa(index)
					index++
					if index == 10 {
						index = 0
					}
				}
			}

			/*Ahora veo si debo crear carpetas padre o si no es poisible crear la carpeta*/
			if r {
				//Si se van a crear carpetas padre
			} else {
				if len(lista_ruta) == 0 {
					//Ahora que encontré la carpeta, creo el inodo y bloque correspondientes
					poscarpeta := 0

					//Busco en el inodo de carpeta un espacio para el nuevo inodo
					for i := 0; i <= 15; i++ {

						if inodo.I_block[i] != 0 {
							file.Seek(int64(inodo.I_block[i]), 0)
							binary.Read(file, binary.BigEndian, &carpeta)

							poscarpeta = int(inodo.I_block[i])

							//Busco espacio
							var nombreVacio [12]byte
							for j := 0; j < 4; j++ {
								if (carpeta.B_content[j].B_name == nombreVacio) && (carpeta.B_content[j].B_inodo == 0) {
									//Si está vacío, lo ocupo
									copy(carpeta.B_content[j].B_name[:], []byte(nombreArchivo))
									pos = j
									i = 15
									break

								}
							}
						} else {
							//Creo una nueva carpeta

							var carpetanueva BloqueCarpetas
							carpeta = carpetanueva

							/*Acá debo asociar el inodo con el nuevo bloque y escribir el nuevo bloque,
							además de actualizar el superbloque y el bitmap de bloques*/

							inodo.I_block[i] = superbloque.S_first_blo
							file.Seek(int64(superbloque.S_first_blo), 0)
							binary.Write(file, binary.BigEndian, &carpeta)

							poscarpeta = int(superbloque.S_first_blo)

							file.Seek(int64(posinodo), 0)
							binary.Write(file, binary.BigEndian, &inodo)

							superbloque.S_free_blocks_count--
							superbloque.S_first_blo += superbloque.S_block_size

							var registro byte
							var uno byte = '1'
							file.Seek(int64(superbloque.S_bm_block_start), 0)

							for i := 0; i < int(superbloque.S_blocks_count); i++ {
								binary.Read(file, binary.BigEndian, &registro)
								if registro == '0' {
									file.Seek(int64(superbloque.S_bm_block_start)+int64(i), 0)
									break //Si ya encontré el espacio libre, me detengo
								}
							}
							binary.Write(file, binary.BigEndian, &uno)
							copy(carpeta.B_content[0].B_name[:], []byte(nombreArchivo))

							pos = 0
							i = 15
						}

					}

					//Busco el primer inodo libre
					file.Seek(int64(superbloque.S_first_ino), 0)

					var nuevoi Inodo
					uid, _ := strconv.Atoi(uidLog)
					nuevoi.I_uid = int32(uid)
					nuevoi.I_gid = 1
					t := time.Now()
					fecha := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d",
						t.Year(), t.Month(), t.Day(),
						t.Hour(), t.Minute(), t.Second())
					copy(nuevoi.I_ctime[:], []byte(fecha))
					copy(nuevoi.I_mtime[:], []byte(fecha))

					nuevoi.I_type = '1' //Tipo archivo
					nuevoi.I_perm = 664
					//Cambio el bitmap de inodos
					file.Seek(int64(superbloque.S_bm_inode_start), 0)

					var registro byte
					var uno byte = '1'
					for i := 0; i < int(superbloque.S_inodes_count); i++ {
						binary.Read(file, binary.BigEndian, &registro)

						if registro == '0' {
							file.Seek(int64(superbloque.S_bm_inode_start)+int64(i), 0)

							break //Si ya encontré el espacio libre, me detengo

						}
					}
					binary.Write(file, binary.BigEndian, &uno)
					var archivonuevo BloqueArchivos

					//Escribo el primer bloque que de archivos y lo asocio con el inodo
					contblo := 0
					file.Seek(int64(superbloque.S_bm_block_start), 0)
					for i := 0; i < int(superbloque.S_blocks_count); i++ {
						binary.Read(file, binary.BigEndian, &registro)

						if registro == '0' {
							file.Seek(int64(superbloque.S_bm_block_start)+int64(i), 0)

							break //Si ya encontré el espacio libre, me detengo

						} else {

						}
					}
					binary.Write(file, binary.BigEndian, uno)
					nuevoi.I_block[0] = superbloque.S_first_blo
					file.Seek(int64(nuevoi.I_block[0]), 0)
					binary.Write(file, binary.BigEndian, &archivonuevo)
					superbloque.S_free_blocks_count--
					superbloque.S_first_blo += superbloque.S_block_size
					nuevoi.I_size = int32(len(contenido))

					var barchivo2 BloqueArchivos
					nuevoBloque := false
					puntero := 0
					var bm byte
					fmt.Println(contenido)
					for _, char := range []byte(contenido) {
						if puntero == 64 {
							//Escribo los cambios en el bloque actual
							file.Seek(int64(nuevoi.I_block[contblo]), 0)
							binary.Write(file, binary.BigEndian, &archivonuevo)

							archivonuevo = barchivo2
							nuevoBloque = true
							puntero = 0
							contblo++
							nuevoi.I_block[contblo] = superbloque.S_first_blo

							//Actualizando el bitmap de bloques
							var uno byte = '1'
							file.Seek(int64(superbloque.S_bm_block_start), 0)
							for i := 0; i < int(superbloque.S_blocks_count); i++ {
								binary.Read(file, binary.BigEndian, &bm)
								if bm == '1' {
								} else {
									file.Seek(int64(superbloque.S_bm_block_start)+int64(i), 0)
									break
									//Si es un cero, me detengo pues ya encontré el espacio
								}
							}
							binary.Write(file, binary.BigEndian, uno)
							superbloque.S_free_blocks_count -= 1
							superbloque.S_first_blo += int32(unsafe.Sizeof(BloqueArchivos{}))
						}
						archivonuevo.B_content[puntero] = char
						puntero++
					}

					if nuevoBloque {

						//Escribo el nuevo bloque
						file.Seek(int64(nuevoi.I_block[contblo]), 0)
						binary.Write(file, binary.BigEndian, &archivonuevo)

						//Actualizo el superbloque
						file.Seek(int64(montadas[idLog].inicio), 0)
						binary.Write(file, binary.BigEndian, &superbloque)

					} else {
						//Actualizo el bloque de archivos
						file.Seek(int64(nuevoi.I_block[0]), 0)
						binary.Write(file, binary.BigEndian, &archivonuevo)
					}

					//Asocio la carpeta padre con el nuevo inodo

					carpeta.B_content[pos].B_inodo = superbloque.S_first_ino
					//Escribo la carpeta modificada
					file.Seek(int64(poscarpeta), 0)
					binary.Write(file, binary.BigEndian, &carpeta)

					//Asocio el inodo con el archivo nuevo

					file.Seek(int64(superbloque.S_first_ino), 0) //Escribo el nuevo inodo
					binary.Write(file, binary.BigEndian, &nuevoi)

					superbloque.S_free_inodes_count--                   //Disminuyo los inodos libres
					superbloque.S_first_ino += superbloque.S_inode_size //Actualizo la posición del primer inodo libre

					file.Seek(int64(montadas[idLog].inicio), 0)
					binary.Write(file, binary.BigEndian, &superbloque)

					consola += "¡Archivo creado con exito!\n"

				} else {
					consola += "Error: No se encontró el directorio del archivo\n"
				}
			}

		} else {
			consola += "Error: Ruta inválida \n"
		}

	} else {
		consola += "Error: No existe una sesión iniciada\n"
	}
}

func mkfile(parametros []string) {
	fpath, fr := false, false
	var path, cont string
	var size int
	for len(parametros) > 0 {
		tmp := parametros[0]
		tipo, valor := getTipoValor(tmp)
		if tipo == ">path" {
			valor = regresarEspacio(valor)
			path = valor
			fpath = true
		} else if tipo == ">r" {
			if len(valor) > 0 {
				consola += "Error: El comando >r no necesita valor\n"
			}
			fr = true
		} else if tipo == ">size" {
			size, _ = strconv.Atoi(valor)
			if size < 0 {
				consola += "Error: Valor de >size negativo \n"
			}

		} else if tipo == ">cont" {
			valor = regresarEspacio(valor)
			cont = valor

		} else if tipo[0] == '#' {
			break
		} else {
			consola += "Error: Parámetro <" + valor + "> no válido\n"
		}

		parametros = parametros[1:]

	}

	if fpath {
		crearArchivo(path, fr, size, cont)
	} else {
		consola += "Error: Faltan parámetros obligatorios\n"
	}
}

func Analizar(lineas []string) string {
	consola = ""            //Reestableciendo la consola cada vez que se llama a analizar
	Reportes.Reportes = nil //Reestablesco la lista de reportes
	//fmt.Println(unsafe.Sizeof(MBR{}))
	for _, linea := range lineas {
		if len(linea) < 5 {
			continue //Si la línea solo incluye un salto de línea
		}
		//fmt.Println(linea)
		if linea[0] != '#' { //Si no es comentario, lo agrego a la consola
			consola += "\n-" + linea + "\n"
		}
		linea = espacioCadena(linea) //Cambio temporalmente los espacios dentro de cadenas por $

		params := strings.Split(linea, " ") //Separo por espacio

		if strings.EqualFold(params[0], "mkdisk") { //Comparación case insensitve
			//Elimino el primer elemento (el nombre del comando)
			params = params[1:]
			mkdisk(params)
		} else if strings.EqualFold(params[0], "rmdisk") {
			params = params[1:]
			rmdisk(params)
		} else if strings.EqualFold(params[0], "fdisk") {
			params = params[1:]
			fdisk(params)
		} else if strings.EqualFold(params[0], "mount") {
			params = params[1:]
			mount(params)
		} else if strings.EqualFold(params[0], "rep") {
			params = params[1:]
			rep(params)
		} else if strings.EqualFold(params[0], "mkfs") {
			params = params[1:]
			mkfs(params)
		} else if strings.EqualFold(params[0], "login") {
			params = params[1:]
			login(params)
		} else if strings.EqualFold(params[0], "logout") {
			Logout()
		} else if strings.EqualFold(params[0], "mkgrp") {
			params = params[1:]
			mkgrp(params)
		} else if strings.EqualFold(params[0], "rmgrp") {
			params = params[1:]
			rmgrp(params)
		} else if strings.EqualFold(params[0], "mkusr") {
			params = params[1:]
			mkusr(params)
		} else if strings.EqualFold(params[0], "rmusr") {
			params = params[1:]
			rmusr(params)
		} else if strings.EqualFold(params[0], "mkfile") {
			params = params[1:]
			mkfile(params)
		} else if params[0][0] == '#' {

			//Si es un comentario, no pasa nada
		} else {
			consola += "Error: Comando <" + params[0] + "> no reconocido\n"
		}
	}
	return consola
}
