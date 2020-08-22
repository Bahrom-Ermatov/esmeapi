package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ClientBalance struct {
	BalanceSum float32
}

func (s *Server) GetClientBalance(c *gin.Context) {
	account := c.Query("account")

	clientBalance := new(ClientBalance)
	err := s.db.Model(clientBalance).
		ColumnExpr("client_balance.balance_sum").
		Join("JOIN clients AS c").
		JoinOn("client_balance.clnt_id = c.clnt_id").
		JoinOn("c.account = ?", account).
		Select()

	if err != nil {
		SuccessFalse(c, err.Error(), "Возникла ошибка при обработке")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"balance": clientBalance.BalanceSum,
	})
}
