package utils

type ConfigNotFoundError struct{}

func (e ConfigNotFoundError) Error() string {
	return "config not found"
}

func ConfigNotFound() ConfigNotFoundError {
	return ConfigNotFoundError{}
}
