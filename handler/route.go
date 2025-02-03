package handler

import (
	"github.com/sidra-api/sidra/dto"
	"github.com/valyala/fasthttp"
	"strings"
)

func (h *Handler) findRoute(ctx *fasthttp.RequestCtx) (dto.SerializeRoute, bool) {
	requestPath := string(ctx.Request.URI().Path())
	key := string(ctx.Host()) + requestPath
	route, exists := h.dataSet.SerializeRoute[key]

	if exists {
		return route, exists
	}

	if requestPath != "/" {
		segments := strings.Split(requestPath, "/")

		for i := len(segments); i > 0; i-- {
			path := strings.Join(segments[:i], "/")

			if path == "" {
				path = "/"
			}

			r := h.dataSet.SerializeRoute[string(ctx.Host())+path]

			if r.PathType == "prefix" {
				route = r
				exists = true
				break
			}
		}
	}

	return route, exists
}
