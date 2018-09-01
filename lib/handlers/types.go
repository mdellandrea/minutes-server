package handlers

import (
	"github.com/rs/zerolog"
)

type Backend interface {
	SetTimeId(id, val string) error
	GetTimeId(id string) (string, error)
	DeleteTimeId(id string) error
	NotFoundErrCheck(err error) bool
}

type TimeHandler struct {
	Db  Backend
	Log zerolog.Logger
}

type NewTimeRequest struct {
	InitialTime string `json:"initialTime"`
}

type NewTime struct {
	TimeId      string `json:"timeId"`
	CurrentTime string `json:"currentTime"`
}

type CurrentTime struct {
	CurrentTime string `json:"currentTime"`
}

type ChangeTimeRequest struct {
	AddMinutes int `json:"addMinutes"`
}
