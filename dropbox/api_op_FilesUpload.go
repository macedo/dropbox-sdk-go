package dropbox

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"
)

func (c *Client) FilesUpload(params *FilesUploadInput) (*FilesUploadOutput, error) {
	var out FilesUploadOutput

	if params == nil {
		params = &FilesUploadInput{}
	}

	uploadURL, _ := url.Parse(ContentUploadEndpoint)
	uploadURL.Path = ServiceAPIVersion + "/files/upload"

	request, err := http.NewRequest(
		http.MethodPost,
		uploadURL.String(),
		params.Body,
	)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Authorization", c.bearerAuth())
	request.Header.Set("Content-Type", "application/octet-stream")
	request.Header.Set("Dropbox-API-Arg", jsonStringify(params.UploadArg))

	err = c.sendRequest(request, &out)

	return &out, err
}

type UploadArg struct {
	Autorename     bool   `json:"auto_rename,omitempty"`
	ClientModified string `json:"client_modified,omitempty"`
	ContentHash    string `json:"content_hash,omitempty"`
	Mode           string `json:"mode"`
	Mute           bool   `json:"mute"`
	Path           string `json:"path"`
	StrictConflict bool   `json:"strict_conflict,omitempty"`
}

type FilesUploadInput struct {
	Body      io.Reader
	UploadArg *UploadArg
}

type FilesUploadOutput struct {
	ClientModified time.Time `json:"client_modified"`
	ContentHash    string    `json:"content_hash"`
	FileLockInfo   struct {
		Created        time.Time `json:"created"`
		IsLockholder   bool      `json:"is_lockholder"`
		LockholderName string    `json:"lockholder_name"`
	} `json:"file_lock_info"`
	HasExplicitSharedMembers bool   `json:"has_explicit_shared_members"`
	ID                       string `json:"id"`
	IsDownloadable           bool   `json:"is_downloadable"`
	Name                     string `json:"name"`
	PathDisplay              string `json:"path_display"`
	PathLower                string `json:"path_lower"`
	PropertyGroups           []struct {
		Fields []struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		} `json:"fields"`
		TemplateID string `json:"template_id"`
	} `json:"property_groups"`
	Rev            string    `json:"rev"`
	ServerModified time.Time `json:"server_modified"`
	SharingInfo    struct {
		ModifiedBy           string `json:"modified_by"`
		ParentSharedFolderID string `json:"parent_shared_folder_id"`
		ReadOnly             bool   `json:"read_only"`
	} `json:"sharing_info"`
	Size int `json:"size"`
}

func jsonStringify(arg *UploadArg) string {
	b, _ := json.Marshal(arg)
	return string(b)
}
