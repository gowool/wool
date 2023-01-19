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

func HandleUI(pattern string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		path := req.URL.Path
		if pattern != "" {
			req.URL.Path = strings.TrimPrefix(req.URL.Path, pattern)
		}
		req.URL.Path = strings.TrimSuffix(req.URL.Path, "/")
		h.ServeHTTP(w, req)
		req.URL.Path = path
		return
	})
}

func UIFileServer(fs http.FileSystem) http.Handler {
	return http.FileServer(&UIAssetWrapper{FileSystem: fs})
}

func (wool *Wool) UI(pattern string, fs http.FileSystem, methods ...string) {
	handler := ToHandler(HandleUI(pattern, UIFileServer(fs)))

	wool.Add(pattern, handler, methods...)
	wool.Add(pattern+"/...", handler, methods...)
}
