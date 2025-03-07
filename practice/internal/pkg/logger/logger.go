package logger

import "go.uber.org/zap"

var Log *zap.SugaredLogger

func Init(env string) {
	if env == "development" {
		Log = zap.Must(zap.NewDevelopment()).Sugar()
	} else {
		Log = zap.Must(zap.NewProduction()).Sugar()
	}
}
