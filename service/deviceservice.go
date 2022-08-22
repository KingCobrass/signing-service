package service

import (
	"log"
	"signing-service/persistence"
	"signing-service/signingCrypto"
)

var ECCKeyGeneratorService signingCrypto.ECCGenerator
var RSAKeyGeneratorService signingCrypto.RSAGenerator

type CreateSignatureDeviceResponse struct {
	DeviceId string `json:"deviceId"`
	Message  string `json:"message"`
}

type SignatureResponse struct {
	Signature   string `json:"signature"`
	Signed_Data string `json:"signed_data"`
	Message     string `json:"message"`
}

// Service struct is returned by the NewService function
type Service struct {
	log     *log.Logger
	queryer persistence.IDevice
}

// NewService creates a new connection as well as logger entry
func NewService(logger *log.Logger) *Service {
	return &Service{
		log:     logger,
		queryer: persistence.New(logger),
	}
}

// Generate key-pair for the current request, and create device information and save into db
func (s Service) CreateSignatureDevice(deviceInfo *persistence.DeviceInfo) error {

	s.log.Println(">> [deviceService][CreateSignatureDeviceResponse][Received]")
	//We have to generate Key-pair
	if deviceInfo.Algorithm == "RSA" {
		rsaKeyPair, err := RSAKeyGeneratorService.Generate()
		if err != nil {
			return err
		} else {

			deviceInfo.RSAKeyPair = rsaKeyPair
			deviceInfo.ECCKeyPair = nil

		}
	} else if deviceInfo.Algorithm == "ECC" {
		eccKeyPair, err := ECCKeyGeneratorService.Generate()
		if err != nil {
			return err
		} else {
			deviceInfo.ECCKeyPair = eccKeyPair
			deviceInfo.RSAKeyPair = nil

		}
	}
	err := s.queryer.CreateSignatureDevice(deviceInfo)
	return err
}

// Get stored device info
func (s Service) GetSignatureDeviceInfo(id string) (*persistence.DeviceInfo, error) {
	s.log.Println(">> [deviceService][GetSignatureDeviceInfo][Received]")
	deviceInfo, err := s.queryer.GetSignatureDeviceInfo(id)
	return deviceInfo, err
}

// Save last_signature_base64_encoded writes the key value to the db
func (s Service) SaveLastSignatureBase64(last_signature_base64_encoded string) error {
	s.log.Println(">> [deviceService][SaveLastSignatureBase64][Received]")
	err := s.queryer.SaveLastSignatureBase64(last_signature_base64_encoded)
	return err
}

func (s Service) GetLastSignatureBase64() string {
	s.log.Println(">> [deviceService][GetSaveLastSignatureBase64][Received]")
	last_signature_base64_encoded := s.queryer.GetLastSignatureBase64()

	return last_signature_base64_encoded
}

// Save Signature Counter writes the key value to the db
func (s Service) SaveSignatureCounter(signature_counter int) error {
	s.log.Println(">> [deviceService][SaveSignatureCounter][Received]")
	err := s.queryer.SaveSignatureCounter(signature_counter)
	return err
}

// Get signature counter
func (s Service) GetSignatureCounter() int {
	s.log.Println(">> [deviceService][GetSignatureCounter][Received]")
	counter := s.queryer.GetSignatureCounter()
	return counter
}
