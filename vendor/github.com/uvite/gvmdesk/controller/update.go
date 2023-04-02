package gtools

import (
	"github.com/minio/selfupdate"
	"net/http"
)

func (a *App) doUpdate(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	err = selfupdate.Apply(resp.Body, selfupdate.Options{})
	if err != nil {
		// error handling
	}
	return err
}
