//go:build darwin

package protocol

func ensure() Result {
	return Result{
		Supported: false,
		Message:   "macOS üzerinde bu sürümde manuel bağlantı kodu kullanılmalıdır.",
	}
}
