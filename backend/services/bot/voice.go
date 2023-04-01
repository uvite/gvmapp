package bot

import (
	"context"
	"github.com/dannywolfmx/go-tts/tts"
	"github.com/labstack/gommon/log"
	"sync"
	"time"
)

const lang = "zh"
const sampleRate = 27000

type Qvm struct {
	Voice *tts.TTS
	Wg    sync.WaitGroup
	Ctx   context.Context
}

func NewVoice() *Qvm {
	voice := tts.NewTTS(lang, sampleRate)
	q := &Qvm{
		Voice: voice,
	}
	return q

}
func (q *Qvm) Alert(msg string) {
	_, cancel := context.WithCancel(context.Background())
	defer cancel()
	log.Info(msg)

	q.Voice.Play()
	q.Wg.Add(1)
	go func() {
		time.Sleep(10 * time.Second)
		q.Voice.Next()
		q.Wg.Done()
	}()
	q.Voice.Add(msg)
	q.Wg.Wait()

}
