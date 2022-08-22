package api

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"signing-service/domain"
	"signing-service/logger"
	"signing-service/persistence"
	"signing-service/service"
	"signing-service/signingCrypto"
	"strconv"

	"github.com/go-playground/validator"
)

var DevicePayload domain.DevicePayload
var DeviceInfo *persistence.DeviceInfo
var SignTransactionPayload domain.SignTransactionPayload

type ALGORITHM int

const (
	ECC ALGORITHM = 1
	RSA ALGORITHM = 2
)

type CreateSignatureDeviceResponse service.CreateSignatureDeviceResponse
type SignatureResponse service.SignatureResponse
type RSAKeyPair signingCrypto.RSAKeyPair
type ECCKeyPair signingCrypto.ECCKeyPair

var ECCKeyGeneratorService signingCrypto.ECCGenerator

// swagger:route POST device/CreateSignatureDevice
// Create new signature device
//
// responses:
//
//	 405: Method not allowed
//		400: Bad Request
//		200: Success
//		201: Created
func (s *Server) CreateSignatureDevice(response http.ResponseWriter, request *http.Request) {

	if request.Method != http.MethodPost {
		WriteErrorResponse(response, http.StatusMethodNotAllowed, []string{
			http.StatusText(http.StatusMethodNotAllowed),
		})
		return
	}

	//Read request payload
	reqBody, err := ioutil.ReadAll(request.Body)
	if err != nil {
		WriteErrorResponse(response, http.StatusBadRequest, []string{
			err.Error(),
		})
	} else {
		json.Unmarshal(reqBody, &DevicePayload)

		deviceService := service.NewService(logger.Logger)

		validate := validator.New()
		//validate request payload
		if validationErr := validate.Struct(DevicePayload); validationErr != nil {
			WriteErrorResponse(response, http.StatusBadRequest, []string{
				validationErr.Error(),
			})
		} else {

			id := DevicePayload.Id
			//Check Requested device already exist or not
			storedDevice, storedErr := deviceService.GetSignatureDeviceInfo(id)
			if storedErr != nil && storedDevice == nil {
				deviceInfo := DeviceInfo
				deviceInfo = new(persistence.DeviceInfo)

				deviceInfo.Id = DevicePayload.Id
				deviceInfo.Algorithm = DevicePayload.Algorithm
				deviceInfo.Label = DevicePayload.Label

				//Will check key found or not
				err := deviceService.CreateSignatureDevice(deviceInfo)
				if err == nil {
					createSignatureDevice := CreateSignatureDeviceResponse{
						DeviceId: id,
						Message:  "Device Created Successfully",
					}
					WriteAPIResponse(response, http.StatusCreated, createSignatureDevice)
				} else {
					WriteErrorResponse(response, http.StatusBadRequest, []string{
						"Device Creation Failed, " + err.Error(),
					})
				}
			} else {
				if storedDevice == nil {
					WriteErrorResponse(response, http.StatusBadRequest, []string{
						"Device Info not parse, might be database issues",
					})
				} else {

					//device already stored, just return device id
					createSignatureDevice := CreateSignatureDeviceResponse{
						DeviceId: storedDevice.Id,
						Message:  "Already created",
					}
					WriteAPIResponse(response, http.StatusOK, createSignatureDevice)
				}

			}

		}
	}

}

// "signature": <signature_base64_encoded>,
// "signed_data": "<signature_counter>_<data_to_be_signed>_<last_signature_base64_encoded>"
/*For the signature creation, the client will have to provide data_to_be_signed through the API. In order to increase the security of the system, we will extend this raw data with the current signature_counter and the last_signature.

The resulting string should follow this format: <signature_counter>_<data_to_be_signed>_<last_signature_base64_encoded>

In the base case there is no last_signature (= signature_counter == 0). Use the base64 encoded device ID (last_signature = base64(device.id)) instead of the last_signature.

This special string will be signed (Signer.sign(secured_data_to_be_signed)) and the resulting signature (base64 encoded) will be returned to the client. The signature response could look like this:

{
    "signature": <signature_base64_encoded>,
    "signed_data": "<signature_counter>_<data_to_be_signed>_<last_signature_base64_encoded>"
}*/
func (s *Server) SignTransaction(response http.ResponseWriter, request *http.Request) {

	if request.Method != http.MethodPost {
		WriteErrorResponse(response, http.StatusMethodNotAllowed, []string{
			http.StatusText(http.StatusMethodNotAllowed),
		})
		return
	}

	reqBody, err := ioutil.ReadAll(request.Body)
	if err == nil {
		json.Unmarshal(reqBody, &SignTransactionPayload)

		deviceService := service.NewService(logger.Logger)

		validate := validator.New()
		//validate request payload
		if validationErr := validate.Struct(SignTransactionPayload); validationErr != nil {
			WriteErrorResponse(response, http.StatusBadRequest, []string{
				validationErr.Error(),
			})
		} else {
			//Valid request found
			id := SignTransactionPayload.DeviceId
			//Check Requested device already exist or not
			storedDevice, storedErr := deviceService.GetSignatureDeviceInfo(id)
			if storedDevice != nil {

				//Get device signature counter
				signatureCounter := deviceService.GetSignatureCounter()
				signatureCounter++

				last_signature_base64_encoded := deviceService.GetLastSignatureBase64()

				if last_signature_base64_encoded == "" {
					//Use device id base64 as last signature device
					last_signature_base64_encoded = base64.StdEncoding.EncodeToString([]byte(id))
				}

				signature_byte, err := GetSignatureByte(storedDevice)

				if err != nil {
					WriteErrorResponse(response, http.StatusBadRequest, []string{
						err.Error(),
					})
				}
				signature_base64 := base64.StdEncoding.EncodeToString([]byte(signature_byte))
				signatureResponse := SignatureResponse{
					Signature:   signature_base64,
					Signed_Data: strconv.Itoa(signatureCounter) + "_" + SignTransactionPayload.Data + "_" + last_signature_base64_encoded,
					Message:     "Data Signature Successfully",
				}
				//Update signature counter, last_signature_base6
				UpdateLastSignatureInformation(deviceService, signatureCounter, signature_base64)
				WriteAPIResponse(response, http.StatusOK, signatureResponse)

			} else {
				//Device not registered
				WriteErrorResponse(response, http.StatusBadRequest, []string{
					storedErr.Error(),
				})
			}
		}

	} else {
		WriteErrorResponse(response, http.StatusBadRequest, []string{
			err.Error(),
		})
	}
}

func UpdateLastSignatureInformation(service *service.Service, signatureCounter int, last_signature_base64_encoded string) {
	service.SaveSignatureCounter(signatureCounter)
	service.SaveLastSignatureBase64(last_signature_base64_encoded)
}

// Check signature algorithm and based on the algorithm will do the signature and return signature byte
func GetSignatureByte(deviceInfo *persistence.DeviceInfo) ([]byte, error) {

	var signature_byte []byte
	//Will do the sign
	if deviceInfo.Algorithm == "RSA" {

		signature_byte, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, &deviceInfo.RSAKeyPair.Private.PublicKey, []byte(SignTransactionPayload.Data), nil)
		if err != nil {
			panic(err)
		}
		decryptedBytes, err := deviceInfo.RSAKeyPair.Private.Decrypt(nil, signature_byte, &rsa.OAEPOptions{
			Hash:  crypto.SHA256,
			Label: []byte{},
		})
		if err != nil {
			return nil, err
		}

		fmt.Println("RSA Decrypted Message: ", string(decryptedBytes))
		return signature_byte, nil
	} else {
		eccKeyPair, err := ECCKeyGeneratorService.Generate()
		if err != nil {
			return nil, err
		}

		signature_byte = []byte(SignTransactionPayload.Data)
		sig1, sig2, err := ecdsa.Sign(rand.Reader, eccKeyPair.Private, signature_byte)

		if err != nil {
			return nil, err
		}

		ret := ecdsa.Verify(eccKeyPair.Public, signature_byte, sig1, sig2)
		if err != nil {
			return nil, err
		}

		fmt.Println("Is ECC Decryption Verified: ", ret)
		fmt.Printf("message: %#v\n\nsig1: %#v\nsig2: %#v", string(signature_byte[:]), sig1, sig2)

		signature_byte = append(sig1.Bytes(), sig2.Bytes()...)
		return signature_byte, nil
	}
}
