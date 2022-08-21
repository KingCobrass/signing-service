package persistence

import (
	"encoding/json"
	"fmt"
	"log"
	"signing-service/signingCrypto"
	"strings"

	badger "github.com/dgraph-io/badger/v3"
)

type RSAKeyPair *signingCrypto.RSAKeyPair
type ECCKeyPair *signingCrypto.ECCKeyPair

type DeviceInfo struct {
	Id         string     `json:"id"`
	Algorithm  string     `json:"algorithm"`
	Label      string     `json:"label"`
	RSAKeyPair RSAKeyPair `json:"rsaKeyPair"`
	ECCKeyPair ECCKeyPair `json:"eccKeyPair"`
}

type IDevice interface {
	CreateSignatureDevice(deviceInfo *DeviceInfo) error
	GetSignatureDeviceInfo(id string) (*DeviceInfo, error)
	SaveLastSignatureBase64(last_signature_base64_encoded string) error
	GetLastSignatureBase64() string
	SaveSignatureCounter(signature_counter int) error
	GetSignatureCounter() int
}

type DeviceDAO struct {
	log    *log.Logger
	dbConn *badger.DB
}

func New(l *log.Logger) *DeviceDAO {
	return &DeviceDAO{
		log:    l,
		dbConn: DBConn,
	}
}

// Save last_signature_base64_encoded writes the key value to the db
func (s *DeviceDAO) CreateSignatureDevice(deviceInfo *DeviceInfo) error {

	objJson, err := json.Marshal(deviceInfo)
	if err == nil {

		err := s.dbConn.Update(func(txn *badger.Txn) error {

			err := txn.Set(
				[]byte(deviceInfo.Id),
				[]byte(objJson),
			)
			return err
		})
		return err
	}

	return err
}
func (s *DeviceDAO) GetSignatureDeviceInfo(id string) (*DeviceInfo, error) {
	var deviceInfo_session *badger.Item
	var deviceInfo *DeviceInfo
	//var tmpDeviceInfo DeviceInfo

	var deviceByte []byte
	err := s.dbConn.View(func(txn *badger.Txn) error {
		var err error
		deviceInfo_session, err = txn.Get([]byte(id))
		if err != nil {
			return err
		}

		err = deviceInfo_session.Value(func(v []byte) error {

			deviceByte = append([]byte{}, v...)

			if err := json.Unmarshal(deviceByte, &deviceInfo); err != nil {
				resolveUnmarshalErr(deviceByte, err)
				return err
			}
			return err

		})
		return err
	})

	return deviceInfo, err
}

// Save last_signature_base64_encoded writes the key value to the db
func (s *DeviceDAO) SaveLastSignatureBase64(last_signature_base64_encoded string) error {

	objJson, err := json.Marshal(last_signature_base64_encoded)
	if err == nil {
		err := s.dbConn.Update(func(txn *badger.Txn) error {

			err := txn.Set(
				[]byte("last_signature_base64_encoded"),
				[]byte(objJson),
			)
			return err
		})
		return err
	}
	return err
}

func (s *DeviceDAO) GetLastSignatureBase64() string {
	var last_signature_base64_encoded_session *badger.Item
	var last_signature_base64_encoded string

	var last_signature_base64_encodedByte []byte

	err := s.dbConn.View(func(txn *badger.Txn) error {
		var err error
		last_signature_base64_encoded_session, err = txn.Get([]byte("last_signature_base64_encoded"))
		if err != nil {
			return err
		}

		err = last_signature_base64_encoded_session.Value(func(v []byte) error {
			last_signature_base64_encodedByte = append([]byte{}, v...)
			if err := json.Unmarshal(last_signature_base64_encodedByte, &last_signature_base64_encoded); err != nil {
				return err
			}
			return nil
		})

		return err
	})

	if err != nil {
		last_signature_base64_encoded = ""
	}
	return last_signature_base64_encoded
}

// Save Signature Counter writes the key value to the db
func (s *DeviceDAO) SaveSignatureCounter(signature_counter int) error {

	objJson, err := json.Marshal(signature_counter)
	if err == nil {
		err := s.dbConn.Update(func(txn *badger.Txn) error {

			err := txn.Set(
				[]byte("signature_counter"),
				[]byte(objJson),
			)
			return err
		})
		return err
	}
	return err
}

func (s *DeviceDAO) GetSignatureCounter() int {
	var counterSession *badger.Item
	var counterByte []byte
	var counter int

	err := s.dbConn.View(func(txn *badger.Txn) error {
		var err error
		counterSession, err = txn.Get([]byte("signature_counter"))
		if err != nil {
			return err
		}

		err = counterSession.Value(func(v []byte) error {

			//log.Printf("key=%s, value=%s\n", counterSession.Key(), v)

			counterByte = append([]byte{}, v...)
			if err := json.Unmarshal(counterByte, &counter); err != nil {
				return err
			}
			return err
		})

		return err
	})

	if err != nil {
		counter = 0
	}

	return counter
}

func resolveUnmarshalErr(data []byte, err error) string {
	if e, ok := err.(*json.UnmarshalTypeError); ok {
		// grab stuff ahead of the error
		var i int
		for i = int(e.Offset) - 1; i != -1 && data[i] != '\n' && data[i] != ','; i-- {
		}
		info := strings.TrimSpace(string(data[i+1 : int(e.Offset)]))
		s := fmt.Sprintf("%s - at: %s", e.Error(), info)
		return s
	}
	if e, ok := err.(*json.UnmarshalFieldError); ok {
		return e.Error()
	}
	if e, ok := err.(*json.InvalidUnmarshalError); ok {
		return e.Error()
	}
	return err.Error()
}
