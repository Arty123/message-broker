package command

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/message-broker/internal/controller"
	"github.com/message-broker/internal/service"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

var CmdHttpServer = &cobra.Command{
	Use:   "http_server",
	Short: "Start http server",
	Args:  cobra.MinimumNArgs(1),
	Run:   cmdHttpServer,
}

func cmdHttpServer(cmd *cobra.Command, args []string) {
	engine := gin.Default()
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
		return
	}

	port := args[1]
	if _, err := strconv.Atoi(port); err != nil {
		logger.Error("port should be numeric", zap.Error(err))
		return
	}

	validate := validator.New()
	queueResolver := service.NewQueueResolver()

	nameQueue := queueResolver.ResolveQueue(service.NameQueue)
	nameController := controller.NameController(logger, nameQueue, validate)
	nameController.SetRoutes(engine)

	colorQueue := queueResolver.ResolveQueue(service.ColorQueue)
	colorController := controller.ColorController(logger, colorQueue, validate)
	colorController.SetRoutes(engine)

	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	errChan := make(chan error)

	go func(port string) {
		if err := engine.Run(fmt.Sprintf(":%s", port)); err != nil {
			logger.Error("could not start http server", zap.Error(err))
			errChan <- err
		}
	}(port)

	// wait for stop application
	select {
	case <-signalChan:
		logger.Info("stop signal received")
		return
	case <-errChan:
		logger.Error("an error occurred while serve http request", zap.Error(err))
		return
	}
}
