package main

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"strconv"
	"strings"
)

var logger = shim.NewLogger("Dealer")

type DealerChaincode struct {
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

func (t *DealerChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("Init")
	return shim.Success(nil)
}

func (t *DealerChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
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
	} else if function == "initLedger" {
		return t.initLedger(stub)
	} else if function == "createCar" {
		return t.createCar(stub, args)
	} else if function == "queryAllCars" {
		return t.queryAllCars(stub)
	} else if function == "changeCarOwner" {
		return t.changeCarOwner(stub, args)
	} else if function == "helloWorld" {
		return t.helloWorld()
	}

	return pb.Response{Status: 403, Message: "Invalid invoke function name."}
}

func (s *DealerChaincode) queryCar(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	carAsBytes, _ := stub.GetState(args[0])
	return shim.Success(carAsBytes)
}

func (s *DealerChaincode) initLedger(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("Init Ledger called")

	cars := []Car{
		Car{Make: "Toyota", Model: "Prius", Color: "blue", Owner: "Tomoko", Restricted: false, Reason: ""},
		Car{Make: "Ford", Model: "Mustang", Color: "red", Owner: "Brad", Restricted: false, Reason: ""},
		Car{Make: "Hyundai", Model: "Tucson", Color: "green", Owner: "Jin Soo", Restricted: false, Reason: ""},
		Car{Make: "Volkswagen", Model: "Passat", Color: "yellow", Owner: "Max", Restricted: false, Reason: ""},
		Car{Make: "Tesla", Model: "S", Color: "black", Owner: "Adriana", Restricted: false, Reason: ""},
		Car{Make: "Peugeot", Model: "205", Color: "purple", Owner: "Michel", Restricted: false, Reason: ""},
		Car{Make: "Chery", Model: "S22L", Color: "white", Owner: "Aarav", Restricted: false, Reason: ""},
		Car{Make: "Fiat", Model: "Punto", Color: "violet", Owner: "Pari", Restricted: false, Reason: ""},
		Car{Make: "Tata", Model: "Nano", Color: "indigo", Owner: "Valeria", Restricted: false, Reason: ""},
		Car{Make: "Holden", Model: "Barina", Color: "brown", Owner: "Shotaro", Restricted: false, Reason: ""},
	}

	i := 0
	for i < len(cars) {
		logger.Debug("i is ", i)
		carAsBytes, _ := json.Marshal(cars[i])
		stub.PutState("CAR"+strconv.Itoa(i), carAsBytes)
		logger.Debug("Added", cars[i])
		i = i + 1
	}

	return shim.Success([]byte("Ledger successfully initiated"))
}

func (s *DealerChaincode) createCar(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	logger.Debug("Create Car called")

	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}

	var car = Car{Make: args[1], Model: args[2], Color: args[3], Owner: args[4], Restricted: false, Reason: ""}

	carAsBytes, _ := json.Marshal(car)
	stub.PutState(args[0], carAsBytes)
	logger.Debug("Car created with key ", args[0], "and value ", car)

	return shim.Success([]byte("Car successfully created"))
}

func (s *DealerChaincode) queryAllCars(stub shim.ChaincodeStubInterface) pb.Response {
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

func (s *DealerChaincode) changeCarOwner(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	logger.Debug("Attempt to change car owner")

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	carAsBytes, _ := stub.GetState(args[0])
	car := Car{}

	json.Unmarshal(carAsBytes, &car)
	if !car.Restricted {
		car.Owner = args[1]
		carAsBytes, _ = json.Marshal(car)
		stub.PutState(args[0], carAsBytes)
		logger.Debug("Car owner successfully changed")
		return shim.Success([]byte("Car owner successfully changed"))
	} else {
		logger.Debug("Car has restrictions: ", car.Reason)
		return shim.Error("Car has restrictions: " + car.Reason)
	}
}

func (t *DealerChaincode) helloWorld() pb.Response {
	logger.Debug("Dealer hello world!")
	return shim.Success([]byte("Dealer hello world!"))
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
	err := shim.Start(new(DealerChaincode))
	if err != nil {
		logger.Error(err.Error())
	}
}
