package controllers

import (
	"encoding/json"
	"os"

	"purecore/core"
)

// ProjectInfo represents the structure of purecore.json
type ProjectInfo struct {
	Name         string                 `json:"name"`
	Description  map[string]string      `json:"description"`
	Version      string                 `json:"version"`
	ReleaseType  string                 `json:"release_type"`
	Author       map[string]string      `json:"author"`
	Repository   map[string]string      `json:"repository"`
	License      string                 `json:"license"`
	Keywords     []string               `json:"keywords"`
	GoVersion    string                 `json:"go_version"`
	Dependencies map[string]interface{} `json:"dependencies"`
}

var cachedInfo *ProjectInfo

func loadProjectInfo() (*ProjectInfo, error) {
	if cachedInfo != nil {
		return cachedInfo, nil
	}

	data, err := os.ReadFile("web/package.json")
	if err != nil {
		return nil, err
	}

	// web/package.json contains project metadata under the "purecore" key
	var pkg struct {
		PureCore ProjectInfo `json:"purecore"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, err
	}
	info := pkg.PureCore

	cachedInfo = &info
	return cachedInfo, nil
}

// SystemController handles system-level endpoints
type SystemController struct{}

// Info returns project information from purecore.json
func (sc *SystemController) Info(req *core.Request, res *core.Response) error {
	info, err := loadProjectInfo()
	if err != nil {
		return res.Error(core.GetLang().Trans("system.load_project_error")+": "+err.Error(), 500)
	}
	return res.Success(info)
}
