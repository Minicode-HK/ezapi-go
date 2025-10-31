package handlers

import (
    "github.com/gin-gonic/gin"
)

type ContentHandler struct{}

func NewContentHandler() *ContentHandler {
    return &ContentHandler{}
}

func (h *ContentHandler) ServeLayout(c *gin.Context) {
    c.File("./cms/static/private/main.html")
}

func (h *ContentHandler) ServeDashboard(c *gin.Context) {
    c.File("./cms/static/private/content/dashboard.html")
}

func (h *ContentHandler) ServeSnapshot(c *gin.Context) {
    c.File("./cms/static/private/content/snapshot.html")
}

func (h *ContentHandler) ServeModule(c *gin.Context) {
    moduleName := c.Param("name")

    c.HTML(200, "module.html", gin.H{
        "moduleName": moduleName,
    })
}