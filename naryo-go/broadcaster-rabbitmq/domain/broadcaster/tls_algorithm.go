package broadcaster

type TLSAlgorithm string

const (
	TLSAlgorithmTLS12 TLSAlgorithm = "TLSv1.2"
	TLSAlgorithmTLS13 TLSAlgorithm = "TLSv1.3"
)

func (a TLSAlgorithm) IsValid() bool {
	switch a {
	case TLSAlgorithmTLS12, TLSAlgorithmTLS13:
		return true
	}
	return false
}

func (a TLSAlgorithm) String() string {
	return string(a)
}
