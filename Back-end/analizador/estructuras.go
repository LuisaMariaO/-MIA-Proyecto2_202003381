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
