module TiendaSupported

go 1.21

require (
	github.com/google/uuid v1.6.0
	golang.org/x/crypto v0.21.0
)

// No se necesita 'replace' si los módulos están en la estructura correcta
// Si models.go está en modules/, se importa como TiendaSupported/modules
