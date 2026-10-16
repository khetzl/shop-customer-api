package handler

import (
	"net/http"

	"github.com/askthechain/atcutils/pkg/logger"
	"github.com/gin-gonic/gin"
)

// ListItems returns every item in the shop.
func (h *Handler) ListItems(c *gin.Context) {
	h.respondWithAllItems(c)
}

// ListForYou returns the items recommended for the caller. There is no
// personalisation yet, so this is every item in the shop.
func (h *Handler) ListForYou(c *gin.Context) {
	h.respondWithAllItems(c)
}

func (h *Handler) respondWithAllItems(c *gin.Context) {
	items, err := h.itemStore.GetAll()
	if err != nil {
		h.logger.Error("failed to load items", atclogger.Params{"err": err, "path": c.FullPath()})
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": items, "error": nil})
}
