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
	"unsafe"
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
	if len(parametro) < 2 {
		return "#vacio", ""
	}
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
	binary.Read(file, binary.LittleEndian, &mbr)

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
		binary.Write(file, binary.LittleEndian, &mbr)
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
	binary.Read(file, binary.LittleEndian, &mbr)

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
		binary.Write(file, binary.LittleEndian, &mbr) //Escribiendo el mbr actualizado

		var ebr EBR
		ebr.Part_status = '0'
		ebr.Part_fit = fit
		ebr.Part_start = 0
		ebr.Part_size = 0
		ebr.Part_next = -1

		file.Seek(int64(particion.Part_start), 0)
		binary.Write(file, binary.LittleEndian, &ebr)

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
	binary.Read(file, binary.LittleEndian, &mbr)

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
		binary.Read(file, binary.LittleEndian, &tmp)

		file.Seek(int64(ext.Part_start), 0)
		binary.Read(file, binary.LittleEndian, &tmp)

		ocupado += int(tmp.Part_size)

		if tmp.Part_status == '0' {
			//Si el primer EBR no está siendo usado, se inserta la partición
			if (ocupado + int(logica.Part_size)) <= int(ext.Part_size) {
				//Si cabe en la partición extendida
				logica.Part_start = ext.Part_start
				file.Seek(int64(ext.Part_start), 0) //Para actualizar el ebr
				binary.Write(file, binary.LittleEndian, &logica)
				consola += "¡Partición lógica <" + string(logica.Part_name[:]) + "> creada!\n"
			} else {
				consola += "Error: Espacio insuficiente en la partición extendida\n"
			}
		} else {
			for tmp.Part_next != -1 {
				file.Seek(int64(tmp.Part_next), 0)
				binary.Read(file, binary.LittleEndian, &tmp)
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
				binary.Write(file, binary.LittleEndian, &ultimo)
				//Escribo el nuevo
				file.Seek(int64(logica.Part_start), 0)
				binary.Write(file, binary.LittleEndian, &logica)
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
	binary.Read(file, binary.LittleEndian, &mbr)

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
			binary.Read(file, binary.LittleEndian, &tmp)

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
				binary.Read(file, binary.LittleEndian, &tmp)

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

func Analizar(lineas []string) string {
	consola = "" //Reestableciendo la consola cada vez que se llama a analizar
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
		} else if params[0][0] == '#' {

			//Si es un comentario, no pasa nada
		} else {
			consola += "Error: Comando <" + params[0] + "> no reconocido\n"
		}
	}
	return consola
}
