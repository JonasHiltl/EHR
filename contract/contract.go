package contract

import (
	"fmt"
	"encoding/json"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

type Permission struct {
	PermissionId 	string `json:"permissionId"`
	DataCategory	string `json:"dataCategory"`
	PatientId		string `json:"patientId"`
	DoctorId		string `json:"doctorId"`
	Right			string `json:"right"`
	From			string `json:"from"`
	To				string `json:"to"`
}

// Patient creates new Permission
func (s *SmartContract) CreatePermission(ctx contractapi.TransactionContextInterface, doctorId string, dataCategory string, patientId string, right string, from string, to string) error {
	permissionId := doctorId + dataCategory + patientId

	// does a permission with this id already exist
	exists, err := s.PermissionExist(ctx, permissionId)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("A permission for this doctor already exists, use update instead.")
	}

	permission := Permission{
		PermissionId: doctorId + dataCategory + patientId,
		DataCategory: dataCategory,
		PatientId: patientId,
		DoctorId: doctorId,
		Right: right,
		From: from,
		To: to,
	}

	//create readable object for the database
	permissionJSON, err := json.Marshal(permission)
	if err != nil {
		return err
	}
	ctx.GetStub().PutState(permissionId, permissionJSON)

	//create composite Key for permission
	colorNameIndexKey, err := ctx.GetStub().CreateCompositeKey("permissionId", []string{doctorId, dataCategory, patientId})
	if err != nil {
		return err
	}

	value := []byte{0x00}
	return ctx.GetStub().PutState(colorNameIndexKey, value)
}

// Patient updates Permission
func (s *SmartContract) UpdatePermission(ctx contractapi.TransactionContextInterface, doctorId string, dataCategory string, patientId string, right string, from string, to string) error {
	permissionId := doctorId + dataCategory + patientId

	//does this permission exist?
	exists, err := s.PermissionExist(ctx, permissionId)
	if err != nil {
		return err
	}
	if exists == false {
		return fmt.Errorf("A permission for this doctor does not exist.")
	}

	permission := Permission{
		PermissionId: doctorId + dataCategory + patientId,
		DataCategory: dataCategory,
		DoctorId: doctorId,
		Right: right,
		From: from,
		To: to,
	}
	
	//create readable object for the database
	permissionJSON, err := json.Marshal(permission)
	if err != nil {
		return err
	} else {
		return ctx.GetStub().PutState(permissionId, permissionJSON)
	}
}

// Patient deletes Permission
func (s *SmartContract) DeletePermission(ctx contractapi.TransactionContextInterface, doctorId string, dataCategory string, patientId string) error {
	permissionId := doctorId + dataCategory + patientId

	// does the permission exist?
	exists, err := s.PermissionExist(ctx, permissionId)
	if err != nil {
		return err
	}
	if exists == false {
		return fmt.Errorf("A permission for this doctor does not exist.")
	}

	return ctx.GetStub().DelState(permissionId)
}

// Query one permission based on the patientId, doctorId and dataCategory to get back the read/write right  and the given time for the permission
func (s *SmartContract) ReadSpecificPermission(ctx contractapi.TransactionContextInterface, doctorId string, dataCategory string, patientId string) (*Permission, error) {
	permissionId := doctorId + dataCategory + patientId
	permissionJSON, err := ctx.GetStub().GetState(permissionId)
	if err != nil {
		return nil, err
	}

	var permission Permission
	json.Unmarshal(permissionJSON, &permission)

	return &permission, nil
}

// Change Permission period
func (s *SmartContract) ChangePermissionPeriod(ctx contractapi.TransactionContextInterface, doctorId string, dataCategory string, patientId string, from string, to string) error {
	permissionId := doctorId + dataCategory + patientId
	permission, err := s.ReadSpecificPermission(ctx, doctorId, dataCategory, patientId)
	if err != nil {
		return err
	}

	permission.From = from
	permission.To = to

	permissionJSON, err := json.Marshal(permission)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(permissionId, permissionJSON)
}

// List all Permissions given to a doctor
func (s *SmartContract) ListDoctorPermissions(ctx contractapi.TransactionContextInterface, doctorId string) ([]byte, error) {
	doctorIterator, err := ctx.GetStub().GetStateByPartialCompositeKey("permissionId", []string{doctorId})
	if err != nil {
		return nil, err
	}
	fmt.Printf("the doctor Iterator is: %s", doctorIterator)

	defer doctorIterator.Close()

	var dataCategory string
	var patientId string
	var permissionId string

	var permissions []byte
	bArrayPermissionAlreadyWritten := false

	for doctorIterator.HasNext() {
		responseRange, err := doctorIterator.Next()
		if err != nil {
			return nil, err
		}

		objectType, compositeKeyParts, err := ctx.GetStub().SplitCompositeKey(responseRange.Key)
		if err != nil {
			return nil, err
		}
		fmt.Printf("the objectType is: %s", objectType)

		dataCategory = compositeKeyParts[1]
		patientId = compositeKeyParts[2]
		permissionId = doctorId + dataCategory + patientId
		fmt.Printf("the compositeKeyParts are: %s", compositeKeyParts[0], compositeKeyParts[1], compositeKeyParts[2])

		permissionAsBytes, err := ctx.GetStub().GetState(permissionId)
		if err != nil {
			return nil, err
		}

		if bArrayPermissionAlreadyWritten == true {
			newBytes := append([]byte(","), permissionAsBytes...)
			permissions = append(permissions, newBytes...)
		} else {
			permissions = append(permissions, permissionAsBytes...)
			fmt.Print(permissions)
		}

		fmt.Printf("Found a asset for index : %s asset id : ", objectType, compositeKeyParts[0], compositeKeyParts[1], compositeKeyParts[2])
		bArrayPermissionAlreadyWritten = true

	}

	permissions = append(permissions, []byte("]")...)
	fmt.Print(permissions)
	return permissions, nil
}

// List all Permissions given by a patient
func (s *SmartContract) ListPatientPermissions(ctx contractapi.TransactionContextInterface, patientId string) ( []byte, error ) {
	patientIterator, err := ctx.GetStub().GetStateByPartialCompositeKey("permissionId", []string{patientId})
	if err != nil {
		return nil, err
	}
	fmt.Printf("the patient Iterator is: %s", patientIterator)

	defer patientIterator.Close()

	var dataCategory string
	var doctorId string
	var permissionId string

	var permissions []byte
	bArrayPermissionAlreadyWritten := false

	for patientIterator.HasNext() {
		responseRange, err := patientIterator.Next()
		if err != nil {
			return nil, err
		}

		objectType, compositeKeyParts, err := ctx.GetStub().SplitCompositeKey(responseRange.Key)
		if err != nil {
			return nil, err
		}
		fmt.Printf("the objectType is: %s", objectType)

		doctorId = compositeKeyParts[0]
		dataCategory = compositeKeyParts[1]
		permissionId = doctorId + dataCategory + patientId
		fmt.Printf("the compositeKeyParts are: %s", compositeKeyParts[0], compositeKeyParts[1], compositeKeyParts[2])

		permissionAsBytes, err := ctx.GetStub().GetState(permissionId)
		if err != nil {
			return nil, err
		}

		if bArrayPermissionAlreadyWritten == true {
			newBytes := append([]byte(","), permissionAsBytes...)
			permissions = append(permissions, newBytes...)
		} else {
			permissions = append(permissions, permissionAsBytes...)
			fmt.Print(permissions)
		}

		fmt.Printf("Found a asset for index : %s asset id : ", objectType, compositeKeyParts[0], compositeKeyParts[1], compositeKeyParts[2])
		bArrayPermissionAlreadyWritten = true

	}

	permissions = append(permissions, []byte("]")...)
	fmt.Print(permissions)
	return permissions, nil
}

// does the permission exist function
func (s *SmartContract) PermissionExist(ctx contractapi.TransactionContextInterface, permissionId string) ( bool, error ) {
	permissionJSON, err := ctx.GetStub().GetState(permissionId)
	if err != nil {
		return false, err
	}

	if permissionJSON != nil {
		return true, nil
	}


	return false, nil
}