package domain

// TODO: signature device domain model ...

type Device struct {
	Id        string `json:"id"`
	Algorithm string `json:"algorithm"`
	Abc       string `json:"abc"`
}

type Direction int

const (
	ECC Direction = iota
	RSA
)

type ISignature interface {
	CreateSignatureDevice(id string, algorithm Direction, label string) error
	SignTransaction(deviceId string, data string) ([]byte, error)
}
