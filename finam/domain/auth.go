package domain

import "time"

type Token struct {
	Token string
	Exp   time.Time
}
