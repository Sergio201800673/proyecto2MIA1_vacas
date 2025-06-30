package diskmanager

type Montaje struct {
	ID              string
	DriveLetter     string
	PartitionName   string
	PartCorrelative int
}

var Montajes []Montaje

const carnet = "201800673" // <-- usa tus últimos dos dígitos aquí
