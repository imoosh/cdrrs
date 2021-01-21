package logrus

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"os"
	"path"
	"time"
)

func LoggerToFile() gin.HandlerFunc {
	var logFilePath, logFileName string

	fileName := path.Join(logFilePath, logFileName)
	src, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		fmt.Println(err)
	}

	logger := logrus.New()
	logger.Out = src
	logger.SetLevel(logrus.DebugLevel)
	logger.SetFormatter(&logrus.TextFormatter{})

	return func(ctx *gin.Context) {
		startTime := time.Now()
		ctx.Next()

		endTime := time.Now()
		latencyTime := endTime.Sub(startTime)
		reqMethod := ctx.Request.Method
		reqUrl := ctx.Request.RequestURI
		statusCode := ctx.Writer.Status()
		clientIP := ctx.ClientIP()
		logger.Infof("| %03d | %13v | %15s | %s | %s |",
			statusCode, latencyTime, clientIP, reqMethod, reqUrl)
	}
}

func LoggerToMongo() gin.HandlerFunc {
	return func(context *gin.Context) {

	}
}

func LoggerToES() gin.HandlerFunc {
	return func(context *gin.Context) {

	}
}

func LoggerToMQ() gin.HandlerFunc {
	return func(context *gin.Context) {

	}
}
