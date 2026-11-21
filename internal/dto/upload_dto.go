package dto

// UploadResponse is returned after a successful file upload.
type UploadResponse struct {
	URL         string `json:"url"`
	Path        string `json:"path"`
	Filename    string `json:"filename"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type"`
	Folder      string `json:"folder"`
}
