package api

import (
	"encoding/json"
	"net/http"

	"time"

	"github.com/gin-gonic/gin"
)

type Price struct {
	ServName string
	Price    float32
	Stime    *time.Time
	Etime    *time.Time
	Comments string
}

func (s *Server) sendSMSView(c *gin.Context) {

	//Получаем баланс клиента
	var balanceSum float32
	err := s.db.Model((*ClientBalance)(nil)).
		ColumnExpr("client_balance.balance_sum").
		Join("JOIN clients AS c").
		JoinOn("client_balance.clnt_id = c.clnt_id").
		JoinOn("c.account = ?", 1).
		Select(&balanceSum)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка при получении баланса клиента"})
		ErrorLog(false, err.Error())
		return
	}

	//Получаем стоимость SMS
	var price float32
	err = s.db.Model((*Price)(nil)).
		ColumnExpr("price").
		Where("serv_name='SMS'").
		Where("? between stime and etime", time.Now()).
		Select(&price)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка при получении стоимости SMS"})
		ErrorLog(false, err.Error())
		return
	}

	//Получаем сумму резервов SMS
	var sumReserve float32
	err = s.db.Model((*Message)(nil)).
		ColumnExpr("sum(price)").
		Where("clnt_id=?", 1).
		Where("state in (0, 1)").
		Select(&sumReserve)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка при получении резервов по SMS"})
		ErrorLog(false, err.Error())
		return
	}

	//Проверяем, достаточно ли денег на балансе
	if balanceSum < price+sumReserve {
		c.JSON(http.StatusOK, gin.H{
			"errorCode":    -1,
			"errorMessage": "Недостаточно средств на счету",
		})
		ErrorLog(false, err.Error())
	}

	r, err := newRabbit()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка при запуске брокера сообщений"})
		ErrorLog(false, err.Error())
		return
	}
	s.r = r

	message := new(Message)
	if err := c.BindJSON(&message); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка при форматировании сообщений в JSON"})
		ErrorLog(false, err.Error())
		return
	}

	Now := time.Now()

	message.Price = price
	message.ClntId = 1
	message.CreatedAt = &Now
	message.LastUpdatedDate = &Now

	_, err = s.db.Model(message).Insert()

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка при добавлении сообщений в БД"})
		ErrorLog(false, err.Error())
		return
	}

	_, err = s.db.Model(message).
		Set("state = 0").
		Where("id = ?", message.ID).
		Update()

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка при изменении статуса сообщения"})
		ErrorLog(false, err.Error())
		return
	}

	data, err := json.Marshal(message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot marshal data to json"})
		ErrorLog(false, err.Error())
		return
	}

	if err := s.Publish(data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot publish message"})
		ErrorLog(false, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errorCode":    0,
		"errorMessage": "Сообщение принято на доставку",
	})

}
