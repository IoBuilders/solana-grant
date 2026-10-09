package portcommon

type Invocation struct {
	SmartContractName string
	MethodName        string
	MethodArgs        map[string]any
}
