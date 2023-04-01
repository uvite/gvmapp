package bot

import (
	"fmt"
	"github.com/dannywolfmx/go-tts/tts"
	"sync"
	"time"
)

const lang = "zh"
const sampleRate = 27000

type Qvm struct {
	Voice *tts.TTS
	Wg    sync.WaitGroup
}

func NewVoice() *Qvm {
	voice := tts.NewTTS(lang, sampleRate)
	q := &Qvm{
		Voice: voice,
	}
	return q

}
func (q *Qvm) Alert(msg string) {
	fmt.Println(msg)
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

