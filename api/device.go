package api

import (
	"encoding/base64"
	"encoding/json"
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
		createSignatureDeviceResponse := CreateSignatureDeviceResponse{
			Message: err.Error(),
		}
		WriteAPIResponse(response, http.StatusBadRequest, createSignatureDeviceResponse)
	} else {
		json.Unmarshal(reqBody, &DeviceInfo)

		deviceService := service.NewService(logger.Logger)

		validate := validator.New()
		//validate request payload
		if validationErr := validate.Struct(DeviceInfo); validationErr != nil {
			createSignatureDevice := CreateSignatureDeviceResponse{
				Message: validationErr.Error(),
			}
			WriteAPIResponse(response, http.StatusBadRequest, createSignatureDevice)
		} else {

			id := DeviceInfo.Id
			//Check Requested device already exist or not
			storedDevice, storedErr := deviceService.GetSignatureDeviceInfo(id)
			if storedErr != nil {
				//Will check key found or not
				err := deviceService.CreateSignatureDevice(DeviceInfo)
				if err == nil {
					createSignatureDevice := CreateSignatureDeviceResponse{
						Message: "Device Created Successfully",
					}
					WriteAPIResponse(response, http.StatusCreated, createSignatureDevice)
				} else {
					createSignatureDevice := CreateSignatureDeviceResponse{
						Message: "Device Creation Failed, " + err.Error(),
					}
					WriteAPIResponse(response, http.StatusBadRequest, createSignatureDevice)
				}
			} else {
				if storedDevice == nil {

					createSignatureDevice := CreateSignatureDeviceResponse{
						Message: "Device Info not parse, might be database issues",
					}
					WriteAPIResponse(response, http.StatusBadRequest, createSignatureDevice)
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
			signatureResponse := SignatureResponse{
				Message: validationErr.Error(),
			}
			WriteAPIResponse(response, http.StatusBadRequest, signatureResponse)
		} else {
			//Valid request found
			id := SignTransactionPayload.DeviceId
			//Check Requested device already exist or not
			storedDevice, storedErr := deviceService.GetSignatureDeviceInfo(id)
			if storedDevice != nil {

				/*For the signature creation, the client will have to provide data_to_be_signed through the API. In order to increase the security of the system, we will extend this raw data with the current signature_counter and the last_signature.

				The resulting string should follow this format: <signature_counter>_<data_to_be_signed>_<last_signature_base64_encoded>

				In the base case there is no last_signature (= signature_counter == 0). Use the base64 encoded device ID (last_signature = base64(device.id)) instead of the last_signature.

				This special string will be signed (Signer.sign(secured_data_to_be_signed)) and the resulting signature (base64 encoded) will be returned to the client. The signature response could look like this:

				{
				    "signature": <signature_base64_encoded>,
				    "signed_data": "<signature_counter>_<data_to_be_signed>_<last_signature_base64_encoded>"
				}*/
				//Get device signature counter
				signatureCounter := deviceService.GetSignatureCounter()
				signatureCounter++

				last_signature_base64_encoded := deviceService.GetLastSignatureBase64()

				if last_signature_base64_encoded == "" {
					//Use device id base64 as last signature device
					last_signature_base64_encoded = base64.StdEncoding.EncodeToString([]byte(id))
				}
				//Will do the sign
				signature_base64 := ""
				signatureResponse := SignatureResponse{
					Signature:   signature_base64,
					Signed_Data: strconv.Itoa(signatureCounter) + "_" + SignTransactionPayload.Data + "_" + last_signature_base64_encoded,
				}
				//To-Do
				//Update signature counter, last_signature_base64
				deviceService.SaveSignatureCounter(signatureCounter)
				deviceService.SaveLastSignatureBase64(last_signature_base64_encoded)
				WriteAPIResponse(response, http.StatusOK, signatureResponse)

			} else {
				//Device not registered
				signatureResponse := SignatureResponse{
					Message: storedErr.Error(),
				}
				WriteAPIResponse(response, http.StatusBadRequest, signatureResponse)
			}
		}

	} else {
		signatureResponse := SignatureResponse{
			Message: err.Error(),
		}
		WriteAPIResponse(response, http.StatusBadRequest, signatureResponse)
	}
}
