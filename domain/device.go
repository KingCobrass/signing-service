package domain

// TODO: signature device domain model ...

type DevicePayload struct {
	Id        string `json:"id" validate:"required"`
	Algorithm string `json:"algorithm" validate:"required"`
	Label     string `json:"label"`
}

type SignTransactionPayload struct {
	DeviceId string `json:"deviceId" validate:"required"`
	Data     string `json:"data" validate:"required"`
}

type ALGORITHM int

const (
	ECC ALGORITHM = 1
	RSA ALGORITHM = 2
)
