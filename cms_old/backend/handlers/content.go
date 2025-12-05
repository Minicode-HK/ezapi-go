package handlers

import (
    "github.com/gin-gonic/gin"
)

type ContentHandler struct{}

func NewContentHandler() *ContentHandler {
    return &ContentHandler{}
}

func (h *ContentHandler) ServeLayout(c *gin.Context) {
    c.File("./cms/frontend/private/main.html")
}

func (h *ContentHandler) ServeDashboard(c *gin.Context) {
    c.File("./cms/frontend/private/content/dashboard.html")
}

func (h *ContentHandler) ServeSnapshot(c *gin.Context) {
    c.File("./cms/frontend/private/content/snapshot.html")
}

func (h *ContentHandler) ServeSystemRoutes(c *gin.Context) {
    c.Header("Content-Type", "text/html; charset=utf-8")
    c.File("cms/frontend/private/content/system_routes.html")
}

func (h *ContentHandler) ServeModule(c *gin.Context) {
    moduleName := c.Param("name")

    c.HTML(200, "module.html", gin.H{
        "moduleName": moduleName,
    })
}

func (h *ContentHandler) ServePlayground(c *gin.Context) {
    c.Header("Content-Type", "text/html; charset=utf-8")
    c.File("cms/frontend/private/content/playground.html")
}

func (h *ContentHandler) ServeAPILogger(c *gin.Context) {
    c.File("cms/frontend/private/content/api-logger.html")
}