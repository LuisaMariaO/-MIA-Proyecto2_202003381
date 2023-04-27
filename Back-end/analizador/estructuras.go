package analizador

//Estructuras a utilizar a lo largo del proyecto

type MBR struct {
	Mbr_tamano         int32
	Mbr_fecha_creacion [30]byte
	Mbr_dsk_signature  int32
	Dsk_fit            byte
	Mbr_partition      [4]Partition
}
type Partition struct {
	Part_status byte     //Indica si la partición está activa o no 1->Activa, 0->Inactiva
	Part_type   byte     //P->Primaria E->Extendida
	Part_fit    byte     //B->BestFit F->FirstFit W->WorstFit
	Part_start  int32    //Indica en qué byte del disco inicia la partición
	Part_size   int32    //Contiene el tamaño de la partición en bytes
	Part_name   [16]byte //Nombre de la partición
}

type EBR struct {
	Part_status byte     //Indica si la partición está activa o no
	Part_fit    byte     //B->BestFit F->FirstFit W->WorstFit
	Part_start  int32    //Indica en qué byte del disco inicia la partición
	Part_size   int32    //Contiene el tamaño de la partición en bytes
	Part_next   int32    //Byte en el que está el próximo EBR. -1 si no hay siguiente
	Part_name   [16]byte //Nombre de la partición
}
type SuperBloque struct {
	S_filesystem_type   int32    //Guarda el número que identifica al sistema de archivos, 2 pues es EXT2
	S_inodes_count      int32    //Guarda el número total de inodos
	S_blocks_count      int32    //Guarda el número total de bloques
	S_free_blocks_count int32    //Contiene el número de bloques libres
	S_free_inodes_count int32    //Contiene el número de inodos libres
	S_mtime             [30]byte //Última fecha en el que el sistema fue montado
	S_mnt_count         int32    //Indica cuántas veces se ha montado el sistema
	S_magic             int32    //Valor que identifica al sistema de archivos, tendrá el valor 0xEF53
	S_inode_size        int32    //Tamaño del inodo
	S_block_size        int32    //Tamaño del bloque
	S_first_ino         int32    //Primer inodo libre
	S_first_blo         int32    //Primer bloque libre
	S_bm_inode_start    int32    //Guardará el inicio del bitmap de inodos
	S_bm_block_start    int32    //Guardará el inicio del bitmap de bloques
	S_inode_start       int32    //Guardará el inicio de la tabla de inodos
	S_block_start       int32    //Guardará el inicio de la tabla de bloques
}
type Inodo struct {
	I_uid   int32     //UID del usuario propietario del archivo o carpeta
	I_gid   int32     //GID del grupo al que pertenece el archivo o carpeta
	I_size  int32     //Tamaño del archivo em bytes
	I_atime [30]byte  //Última fecha en que se leyó el inodo sin modificarlo
	I_ctime [30]byte  //Fecha en que se creó el inodo
	I_mtime [30]byte  //Última fecha en que se modifica el inodo
	I_block [16]int32 //Array en los que los primeros 16 registros son bloques directos
	I_type  byte      //Indica si es archivo o carpeta 1->Archivo, 0->Carpeta
	I_perm  int32     /*Guardará los permisos del archivo o carpeta. Se trabajará a
	nivel de bits, estará dividido de la siguiente forma:
	Los primeros tres bits serán para el Usuario i_uid. Los siguientes
	tres bits serán para el Grupo al que pertenece el usuario. Y los
	últimos tres bits serán para los permisos de Otros usuarios.
	Cada grupo de tres bits significa lo siguiente: El primer bit indica
	el permiso de lectura R. El segundo bit indica el permiso de
	escritura
	W. El tercer bit indica el permiso de ejecución X.*/
}

type BloqueCarpetas struct {
	B_content [4]Content //Array con el contenido de la carpeta
}
type Content struct {
	B_name  [12]byte //Nombre de la carpeta o archivo
	B_inodo int32    //Apuntador hacia un inodo asociado al archivo o carpeta
}
type BloqueArchivos struct {
	B_content [64]byte
}
