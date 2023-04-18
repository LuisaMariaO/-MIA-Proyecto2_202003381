package analizador

//Estructuras a utilizar a lo largo del proyecto

type MBR struct {
	Mbr_tamano         int32
	Mbr_fecha_creacion [30]byte
	Mbr_dsk_signature  int32
	Dsk_fit            byte
	mbr_partition      [4]Partition
}
type Partition struct {
	part_status byte
	part_type   byte
	part_fit    byte
	part_start  int32
	part_size   int32
	part_name   [16]byte
}
