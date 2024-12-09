package model

type EthereumNetwork string

const (
	EthereumNetworkEthereumMainnet EthereumNetwork = "1"
	EthereumNetworkEthereumGoerli  EthereumNetwork = "5"
	EthereumNetworkPolygonMainnet  EthereumNetwork = "137"
	EthereumNetworkPolygonMumbai   EthereumNetwork = "80001"
)

func ParseEthereumNetwork(s string) (EthereumNetwork, bool) {
	switch s {
	case "1":
		return EthereumNetworkEthereumMainnet, true
	case "5":
		return EthereumNetworkEthereumGoerli, true
	case "137":
		return EthereumNetworkPolygonMainnet, true
	case "80001":
		return EthereumNetworkPolygonMumbai, true
	default:
		return "", false
	}
}
