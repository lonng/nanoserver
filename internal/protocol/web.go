package protocol

type Version struct {
	Version int    `json:"version"`
	Android string `json:"android"`
	IOS     string `json:"ios"`
}
