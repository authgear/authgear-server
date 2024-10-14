package authenticator

type FaceRecognitionSpec struct {
	B64ImageString string `json:"-"` // base64 encoded image string for facial recognition verification
}
