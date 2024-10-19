package vmedis

type tokenProvider interface {
	GetActiveToken() (string, error)
}

type staticTokenProvider string

func (s staticTokenProvider) GetActiveToken() (string, error) {
	return string(s), nil
}
