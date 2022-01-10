package golog

import (
	"github.com/sirupsen/logrus"
)

type Option struct {
	AmqpURL string

	Level logrus.Level
}
