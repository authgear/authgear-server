package authenticator

type FaceRecognitionSpec struct {
	B64ImageString   string `json:"-"`                 // base64 encoded image string for facial recognition verification
	OpenCVFRPersonID string `json:"opencv_fr_user_id"` // OpenCV face recognition Person ID
}
