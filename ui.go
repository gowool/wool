package wool

import (
	"net/http"
	"os"
	"strings"
)

type UIAssetWrapper struct {
	FileSystem http.FileSystem
}

func (fsw *UIAssetWrapper) Open(name string) (http.File, error) {
	file, err := fsw.FileSystem.Open(name)
	if err == nil {
		return file, nil
	}
	if os.IsNotExist(err) {
		return fsw.FileSystem.Open("index.html")
	}
	return nil, err
}

func HandleUI(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// do not allow paths like ../../resource
		// only specified folder and resources in it
		// https://lgtm.com/rules/1510366186013/
		if strings.Contains(req.URL.Path, "..") {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		path := req.URL.Path
		fp := strings.TrimSuffix(path, "/")
		fp = strings.ReplaceAll(fp, "\n", "")
		fp = strings.ReplaceAll(fp, "\r", "")
		req.URL.Path = fp
		h.ServeHTTP(w, req)
		req.URL.Path = path
		return
	})
}

func UIFileServer(fs http.FileSystem) http.Handler {
	return http.FileServer(&UIAssetWrapper{FileSystem: fs})
}

func (wool *Wool) UI(pattern string, fs http.FileSystem, methods ...string) {
	handler := ToHandler(HandleUI(UIFileServer(fs)))

	wool.Add(pattern, handler, methods...)
	wool.Add(pattern+"/...", handler, methods...)
}
