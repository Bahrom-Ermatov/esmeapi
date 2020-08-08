package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/lib/pq"
)

func (s *Server) GetSMSDetail(c *gin.Context) {
	msgId := c.Query("msg_id")

	db := s.ConnectDB()
	defer db.Close()

	message := new(Message)
	err := db.Model(message).
		ColumnExpr("dst, message, src, state, created_at, last_updated_date").
		Where("id = ?", msgId).
		Select()

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"dst":              message.Dst,
		"message":          message.Message,
		"src":              message.Src,
		"state":            message.State,
		"create_at":        message.CreatedAt,
		"last_update_date": message.LastUpdatedDate,
	})

}
