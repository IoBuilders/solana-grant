package contractcallbuilder

type Port interface {
	BuildCall(request *BuildCallRequest) (*CallData, error)
	DecodeResult(methodName string, result string) (map[string]any, error)
}

type BuildCallRequest struct {
	SmartContractId string
	MethodName      string
	MethodArgs      map[string]any
}

type CallData struct {
	To   string
	Data string
}
