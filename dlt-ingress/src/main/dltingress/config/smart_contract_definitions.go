package dltingressconfig

var SmartContractDefinitions = map[string]map[string]string{}

func SetupSmartContractsDefinitions(smartContractDefinitions map[string]map[string]string) {
	SmartContractDefinitions = smartContractDefinitions
}
