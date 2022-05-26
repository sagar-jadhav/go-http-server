package connector

type Connector interface {
	Get(url string, retryCount int) ([]byte, error)
}
