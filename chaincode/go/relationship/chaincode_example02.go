package main

import (
	"encoding/json"
	"github.com/hyperledger/fabric/core/chaincode/lib/cid"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"strconv"
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

type Item struct {
	Key    string `json:"Key"`
	Record Car    `json:"Record"`
}

func (t *DealerChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("Init")
	return shim.Success(nil)
}

func (t *DealerChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("Invoke")

	mspid, err := cid.GetMSPID(stub)
	if err != nil {
		logger.Debug("MSP detection error. ", err.Error())
		return shim.Error(err.Error())
	}

	function, args := stub.GetFunctionAndParameters()

	logger.Debug("Function: ", function)

	if function == "queryCar" {
		return t.queryCar(stub, args)
	} else if function == "queryAllCars" {
		return t.queryAllCars(stub)
	} else if function == "helloWorld" {
		return t.helloWorld()
	} else if function == "checkMSP" {
		return t.checkMSP(stub)
	}

	logger.Debug("MSP: ", mspid)

	switch mspid {
	case "aMSP":
		if function == "initLedger" {
			return t.initLedger(stub)
		} else if function == "createCar" {
			return t.createCar(stub, args)
		} else if function == "changeCarOwner" {
			return t.changeCarOwner(stub, args)
		} else {
			logger.Debug("Invalid invoke function name or caller MSP.")
			return shim.Error("Invalid invoke function name or caller MSP.")
		}
	case "bMSP":
		if function == "addRestriction" {
			return t.addRestriction(stub, args)
		} else if function == "removeRestriction" {
			return t.removeRestriction(stub, args)
		} else {
			logger.Debug("Invalid invoke function name or caller MSP.")
			return shim.Error("Invalid invoke function name or caller MSP.")
		}
	default:
		logger.Debug("Wrong caller MSP.")
		return shim.Error("Wrong caller MSP.")
	}

	logger.Debug("Invalid invoke function name.")
	return shim.Error("Invalid invoke function name.")
}

func (s *DealerChaincode) queryCar(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	logger.Debug("Query car called")
	if len(args) != 1 {
		logger.Debug("Incorrect number of arguments. Expecting 1")
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	carAsBytes, _ := stub.GetState(args[0])
	logger.Debug("Car queried: ", carAsBytes)
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
		logger.Debug("Incorrect number of arguments. Expecting 5")
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
		logger.Debug("Error: ", err.Error())
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	l := []Item{}

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			logger.Debug("Error: ", err.Error())
			return shim.Error(err.Error())
		}

		car := Car{}

		json.Unmarshal(queryResponse.Value, &car)

		item := Item{
			queryResponse.Key,
			car,
		}

		l = append(l, item)
	}

	lAsBytes, err := json.Marshal(l)

	if err != nil {
		logger.Debug("Marshalling error. ", err.Error())
		return shim.Error(err.Error())
	}

	logger.Debug("All cars in ledger:", string(lAsBytes))

	return shim.Success(lAsBytes)
}

func (s *DealerChaincode) changeCarOwner(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	logger.Debug("Attempt to change car owner")

	if len(args) != 2 {
		logger.Debug("Incorrect number of arguments. Expecting 2")
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

func (s *DealerChaincode) addRestriction(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	logger.Debug("Adding restriction")

	if len(args) != 2 {
		logger.Debug("Incorrect number of arguments. Expecting 2")
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

func (s *DealerChaincode) removeRestriction(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	logger.Debug("Removing restriction")

	if len(args) != 1 {
		logger.Debug("Incorrect number of arguments. Expecting 1")
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

func (t *DealerChaincode) checkMSP(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("Check MSP called")
	mspid, err := cid.GetMSPID(stub)
	logger.Debug("MSP: ", mspid)
	if err != nil {
		logger.Debug("Error: ", err.Error())
		return shim.Error(err.Error())
	}
	return shim.Success([]byte(mspid))
}

func (t *DealerChaincode) helloWorld() pb.Response {
	logger.Debug("Hello world called")
	return shim.Success([]byte("Hello world!"))
}

func main() {
	err := shim.Start(new(DealerChaincode))
	if err != nil {
		logger.Error(err.Error())
	}
}
