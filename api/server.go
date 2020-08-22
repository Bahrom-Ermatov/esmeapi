package api

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-pg/pg"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	channel     string = "SMS_CHANNEL"
	logFilePath string = "logs/esmeapi.log"
)

// Server is main server instance
type Server struct {
	l  *log.Logger
	e  *gin.Engine
	r  *rabbit
	db *pg.DB
}

type Message struct {
	ID              int32
	ExternalId      int32
	Dst             string
	Message         string
	Src             string
	State           int
	CreatedAt       *time.Time
	LastUpdatedDate *time.Time
	SMSCMessageID   string
	Price           float32
	ClntId          int
}

// Run is the entry point to the program
func (s *Server) Run() error {
	lumberjackLogRotate := &lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    2,   // Max megabytes before log is rotated
		MaxBackups: 500, // Max number of old log files to keep
		MaxAge:     60,  // Max number of days to retain log files
		Compress:   true,
	}
	log.SetOutput(lumberjackLogRotate)

	s.l = log.StandardLogger()
	s.db = ConnectDB()
	s.e = gin.Default()
	s.getRoutes()

	if err := s.e.Run(os.Getenv("ESMEAPI_HOST")); err != nil {
		return errors.Wrapf(err, "cannot start server")
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	s.db.Close()

	return nil
}

func ConnectDB() *pg.DB {
	db := pg.Connect(&pg.Options{
		Addr:     ":5432",
		User:     "postgres",
		Password: "123",
		Database: "messages",
	})
	return db
}

func SuccessFalse(c *gin.Context, errorMessage string, outErrMessage string) {
	c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": outErrMessage})
	log.Errorf(errorMessage)
}

func SuccessTrue(c *gin.Context, errorMessage string) {
	c.JSON(http.StatusOK, gin.H{"ok": true, "message": errorMessage})
	log.Errorf(errorMessage)
}
