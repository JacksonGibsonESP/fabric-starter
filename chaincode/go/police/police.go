package main

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"strings"
)

var logger = shim.NewLogger("Dealer")

type PoliceChaincode struct {
}

// Define the car structure, with 6 properties.  Structure tags are used by encoding/json library
type Car struct {
	Make       string `json:"make"`
	Model      string `json:"model"`
	Color      string `json:"color"`
	Owner      string `json:"owner"`
	Restricted bool   `json:"restricted"`
	Reason     string `json:"reason"`
}

func (t *PoliceChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("Init")
	return shim.Success(nil)
}

func (t *PoliceChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("Invoke")

	creatorBytes, err := stub.GetCreator()
	if err != nil {
		return shim.Error(err.Error())
	}

	name, org := getCreator(creatorBytes)

	logger.Debug("transaction creator " + name + "@" + org)

	function, args := stub.GetFunctionAndParameters()

	if function == "queryCar" {
		return t.queryCar(stub, args)
	} else if function == "queryAllCars" {
		return t.queryAllCars(stub)
	} else if function == "addRestriction" {
		return t.addRestriction(stub, args)
	} else if function == "removeRestriction" {
		return t.removeRestriction(stub, args)
	} else if function == "helloWorld" {
		return t.helloWorld()
	}

	return pb.Response{Status: 403, Message: "Invalid invoke function name."}
}

func (s *PoliceChaincode) queryCar(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	carAsBytes, _ := stub.GetState(args[0])
	return shim.Success(carAsBytes)
}

func (s *PoliceChaincode) queryAllCars(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("Query all cars called")

	startKey := "CAR0"
	endKey := "CAR999"

	resultsIterator, err := stub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	logger.Debug("All cars in ledger:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func (s *PoliceChaincode) addRestriction(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	logger.Debug("Adding restriction")

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	carAsBytes, _ := stub.GetState(args[0])
	car := Car{}

	json.Unmarshal(carAsBytes, &car)
	car.Restricted = true
	car.Reason = args[1]
	carAsBytes, _ = json.Marshal(car)
	stub.PutState(args[0], carAsBytes)
	logger.Debug("Adding restriction successful. Restriction reason: ", car.Reason)
	return shim.Success([]byte("Adding restriction successful"))
}

func (s *PoliceChaincode) removeRestriction(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	logger.Debug("Removing restriction")

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	carAsBytes, _ := stub.GetState(args[0])
	car := Car{}

	json.Unmarshal(carAsBytes, &car)
	car.Restricted = false
	car.Reason = ""
	carAsBytes, _ = json.Marshal(car)
	stub.PutState(args[0], carAsBytes)
	logger.Debug("Removing restriction successful")
	return shim.Success([]byte("Removing restriction successful"))
}

func (t *PoliceChaincode) helloWorld() pb.Response {
	logger.Debug("Police hello world!")
	return shim.Success([]byte("Police hello world!"))
}

var getCreator = func(certificate []byte) (string, string) {
	data := certificate[strings.Index(string(certificate), "-----") : strings.LastIndex(string(certificate), "-----")+5]
	block, _ := pem.Decode([]byte(data))
	cert, _ := x509.ParseCertificate(block.Bytes)
	organization := cert.Issuer.Organization[0]
	commonName := cert.Subject.CommonName
	logger.Debug("commonName: " + commonName + ", organization: " + organization)

	organizationShort := strings.Split(organization, ".")[0]

	return commonName, organizationShort
}

func main() {
	err := shim.Start(new(PoliceChaincode))
	if err != nil {
		logger.Error(err.Error())
	}
}
