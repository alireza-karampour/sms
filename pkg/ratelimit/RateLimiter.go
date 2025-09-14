package ratelimit

import (
	"time"

	"github.com/sirupsen/logrus"
)

type HardCapFunc[T any] func(mch *ManagedChan[T], msg T, mainChan *chan T) error
type SoftCapFunc func() error

type ManagedChan[T any] struct {
	HardCap          uint
	SoftCap          uint
	OnHardCapReached HardCapFunc[T]
	OnSoftCapReached SoftCapFunc
	In               chan T
	Out              chan T
	core             *chan T
	counter          uint
}

func NewManagedChan[T any](hardCap uint, softCap uint, onHardCapReached HardCapFunc[T], onSoftCapReached SoftCapFunc) *ManagedChan[T] {
	if onHardCapReached == nil || onSoftCapReached == nil {
		panic("cap callbacks should not be nil")
	}
	mch := new(ManagedChan[T])

	*mch.core = make(chan T, hardCap)
	mch.HardCap = hardCap
	mch.SoftCap = softCap
	mch.OnHardCapReached = onHardCapReached
	mch.OnSoftCapReached = onSoftCapReached
	go func() {
		for msg := range mch.In {
			if mch.counter < hardCap {
				*mch.core <- msg
				mch.counter++
				if mch.counter == softCap {
					err := mch.OnSoftCapReached()
					logrus.Errorf("soft CAP reached and it's callback failed: %s\n", err.Error())
				}
			} else {

			}
		}
	}()
	return mch
}

type MsgPerUnitTime struct {
	*time.Ticker
}

func NewMPUT(minRate time.Duration, maxRate time.Duration) *MsgPerUnitTime {
	rl := new(MsgPerUnitTime)

	return rl
}

func (rl *MsgPerUnitTime) StartTimer() {

}

func (rl *MsgPerUnitTime) StopTimer() {

}
