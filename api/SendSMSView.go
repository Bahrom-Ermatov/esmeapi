package api

import (
	"encoding/json"

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
		SuccessFalse(c, err.Error(), "Возникла ошибка при обработке")
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
		SuccessFalse(c, err.Error(), "Возникла ошибка при обработке")
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
		SuccessFalse(c, err.Error(), "Возникла ошибка при обработке")
		return
	}

	//Проверяем, достаточно ли денег на балансе
	if balanceSum < price+sumReserve {
		SuccessFalse(c, err.Error(), "Недостаточно средств на счету")
		return
	}

	r, err := newRabbit()
	if err != nil {
		SuccessFalse(c, err.Error(), "Возникла ошибка при обработке")
		return
	}
	s.r = r

	message := new(Message)
	if err := c.BindJSON(&message); err != nil {
		SuccessFalse(c, err.Error(), "Возникла ошибка при обработке")
		return
	}

	Now := time.Now()

	message.Price = price
	message.ClntId = 1
	message.CreatedAt = &Now
	message.LastUpdatedDate = &Now

	_, err = s.db.Model(message).Insert()

	if err != nil {
		SuccessFalse(c, err.Error(), "Возникла ошибка при обработке")
		return
	}

	_, err = s.db.Model(message).
		Set("state = 0").
		Where("id = ?", message.ID).
		Update()

	if err != nil {
		SuccessFalse(c, err.Error(), "Возникла ошибка при обработке")
		return
	}

	data, err := json.Marshal(message)
	if err != nil {
		SuccessFalse(c, err.Error(), "Возникла ошибка при обработке")
		return
	}

	if err := s.Publish(data); err != nil {
		SuccessFalse(c, err.Error(), "Возникла ошибка при обработке")
		return
	}

	SuccessTrue(c, "Сообщение принято на доставку")
}
