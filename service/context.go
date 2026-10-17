package service

import (
	"context"
	"time"
)

func databaseContext(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, 10*time.Second)
}
