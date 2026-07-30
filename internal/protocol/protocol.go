package protocol

type Result struct {
	Supported bool
	Ready     bool
	Message   string
}

func Ensure() Result {
	return ensure()
}
