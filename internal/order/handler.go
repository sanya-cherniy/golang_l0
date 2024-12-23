package order

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"l0/pkg/client/broker"
	"l0/pkg/client/postgres"
	"l0/pkg/logging"
)

const (
	ordersURL = "/order"
	orderURL  = "/order/:id"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

type Handler struct {
	storage  Storage
	logger   *logging.Logger
	db       postgres.Pool
	producer *broker.Producer
}

func (h *Handler) Register(router *gin.Engine) {
	router.POST(ordersURL, h.CreateOrder)
	router.GET(orderURL, h.GetOrder)
	router.GET(ordersURL, h.GetOrders)
}

func NewHandler(storage Storage, logger *logging.Logger, db postgres.Pool, producer *broker.Producer) *Handler {
	return &Handler{storage: storage, logger: logger, db: db, producer: producer}
}

func (h *Handler) CreateOrder(c *gin.Context) {
	var order Order
	if err := c.ShouldBindJSON(&order); err != nil {
		h.logger.Errorf("failed to create order %s\n", err.Error())
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: err.Error(),
		})
		return
	}
	h.storage.Insert(&order)
	h.logger.Infof("order with ID %s was created", order.OrderUID)
	err := insertOrder(h.db, order)
	if err != nil {
		h.logger.Errorf("error Insert to db: %s", err)
	}

	jsonData, _ := json.Marshal(order)
	strBody := string(jsonData)
	message := fmt.Sprintf("received order: %s", strBody)
	if err := h.producer.ProducePartitionAny(message, "order"); err != nil {
		h.logger.Error(err)
	} else {
		h.logger.Infof("produce order %s in topic order", order.OrderUID)
	}
	c.JSON(http.StatusOK, order)
}

func (h *Handler) GetOrders(c *gin.Context) {
	orders, err := GetAllOrdersFromDB(h.db)
	if err != nil {
		h.logger.Errorf("failed to get orders %s\n", err.Error())
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: err.Error(),
		})
		return
	}
	h.logger.Info("get all orders")
	c.JSON(http.StatusOK, orders)
}

func (h *Handler) GetOrder(c *gin.Context) {
	id := c.Param("id")
	order, err := h.storage.Get(id)
	if err != nil {
		exist, errDb := CheckRecordExists(h.db, id)
		if errDb != nil {
			err = errDb
		}
		if exist {
			order, err = GetOrderFromDB(h.db, id)
			if err == nil {
				h.logger.Infof("get order from DB: %s\n", id)
				c.JSON(http.StatusOK, order)
				return
			}
		}
		h.logger.Errorf("failed to get order %s\n", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: err.Error(),
		})
		return
	}
	h.logger.Infof("get order from cache: %s\n", id)
	c.JSON(http.StatusOK, order)
}
