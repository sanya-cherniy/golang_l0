package main

import (
	"l0/internal/order"
	"net/http"
	"strings"
	"testing"

	"context"
	"net/http/httptest"

	"github.com/gin-gonic/gin"

	"l0/internal/config"

	"github.com/stretchr/testify/assert"

	"l0/pkg/client/broker"
	"l0/pkg/client/postgres"
	"l0/pkg/logging"
	"os"
)

const logFile = "logs/test.log"

func TestPost(t *testing.T) {
	router := gin.Default()
	logging.InitLogger(logFile)
	cfg := config.GetConfig(logFile)

	logger, err := logging.GetLogger(logFile)
	if err != nil {
		panic(err)
	}
	memoryStorage := order.NewMemoryStorage()
	pool, err := postgres.NewClient(context.Background(), cfg.Storage)

	if err != nil {
		logger.Fatal(err)
	}

	brokerProducer, err := broker.NewProducer(cfg.Brokers)
	if err != nil {
		logger.Fatal(err)
	}
	go memoryStorage.CashChecker(cfg.LifeTime)

	handler := order.NewHandler(memoryStorage, logger, pool, brokerProducer)
	handler.Register(router)
	w := httptest.NewRecorder()

	testJson1, err := os.ReadFile("test_order0.json")
	if err != nil {
		panic(err)
	}
	testJson2, err := os.ReadFile("test_order1.json")
	if err != nil {
		panic(err)
	}
	testJson3, err := os.ReadFile("test_order2.json")
	if err != nil {
		panic(err)
	}

	testJson4, err := os.ReadFile("test_order3.json")
	if err != nil {
		panic(err)
	}
	testJson5, err := os.ReadFile("test_order4.json")
	if err != nil {
		panic(err)
	}
	testJson6, err := os.ReadFile("test_order5.json")
	if err != nil {
		panic(err)
	}

	req, _ := http.NewRequest("POST", "/order", strings.NewReader(string(testJson1)))
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, strings.TrimSuffix(string(testJson1), "\n"), w.Body.String())
	w = httptest.NewRecorder()

	req, _ = http.NewRequest("POST", "/order", strings.NewReader(string(testJson2)))
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, strings.TrimSuffix(string(testJson2), "\n"), w.Body.String())
	w = httptest.NewRecorder()

	req, _ = http.NewRequest("POST", "/order", strings.NewReader(string(testJson3)))
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, strings.TrimSuffix(string(testJson3), "\n"), w.Body.String())
	w = httptest.NewRecorder()

	req, _ = http.NewRequest("POST", "/order", strings.NewReader(string(testJson4)))
	router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)
	w = httptest.NewRecorder()

	req, _ = http.NewRequest("POST", "/order", strings.NewReader(string(testJson5)))
	router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)
	w = httptest.NewRecorder()

	req, _ = http.NewRequest("POST", "/order", strings.NewReader(string(testJson6)))
	router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)
	w = httptest.NewRecorder()

	req, _ = http.NewRequest("GET", "/order", strings.NewReader(string(testJson1)))
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	w = httptest.NewRecorder()

	req, _ = http.NewRequest("GET", "/order/RANDOM", strings.NewReader(string(testJson1)))
	router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

}
