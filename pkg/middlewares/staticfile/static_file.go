package staticfile

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/log"
	"github.com/containous/traefik/v2/pkg/middlewares"
	"github.com/containous/traefik/v2/pkg/tracing"
	"github.com/opentracing/opentracing-go/ext"
)

const (
	typeName = "StaticFile"
)

// StaticFile is a middleware used to serve static files.
type staticFile struct {
	next   http.Handler
	root   string
	name   string
}

// New creates a new handler.
func New(ctx context.Context, next http.Handler, config dynamic.StaticFile, name string) (http.Handler, error) {
	log.FromContext(middlewares.GetLoggerCtx(ctx, name, typeName)).Debug("Creating middleware")
	var result *staticFile

	if len(config.Root) > 0 {
		result = &staticFile{
			root: config.Root,
			next: next,
			name: name,
		}
	} else {
		return nil, fmt.Errorf("root cannot be empty")
	}

	return result, nil
}

func (sf *staticFile) GetTracingInformation() (string, ext.SpanKindEnum) {
	return sf.name, tracing.SpanKindNoneEnum
}

func (sf *staticFile) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// TODO: use logger
	// logger := log.FromContext(middlewares.GetLoggerCtx(req.Context(), sf.name, typeName))

	fullPath := path.Join(sf.root, req.URL.Path);
	stat, err := os.Stat(fullPath);
	if err != nil {
		sf.next.ServeHTTP(rw, req)
	} else if stat.IsDir() {
		fullPath := path.Join(fullPath, "index.html") // TODO: config
		stat, err = os.Stat(fullPath)
		if err != nil || stat.IsDir() {
			sf.next.ServeHTTP(rw, req)
		} else {
			fmt.Println("serve index", fullPath);
			http.ServeFile(rw, req, fullPath);
		}
	} else {
		fmt.Println("serve file", fullPath);
		http.ServeFile(rw, req, fullPath);
	}
}
