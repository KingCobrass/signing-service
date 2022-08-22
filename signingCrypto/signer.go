package signingCrypto

// Signer defines a contract for different types of signing implementations.
type Signer interface {
	RSASign(dataToBeSigned []byte) ([]byte, error)
	ECCSign(dataToBeSigned []byte) ([]byte, error)
}

// TODO: implement RSA and ECDSA signing ...
