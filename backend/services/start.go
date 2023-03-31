package services

import (
	"context"
	"fmt"
	"github.com/uvite/gvmapp/backend/lib"
)

type StartService struct {
	Ctx context.Context
}

func NewStartService() *StartService {
	return &StartService{}
}

func (f *StartService) ReadFileContents(path string) string {
	contents, err :=  lib.ReadFileContents(path)

	if err != nil {
		fmt.Println(err)
		return ""
	}

	return contents
}

func (f *StartService) OpenFileContents() string {
	contents, _ := lib.OpenFileContents(f.Ctx)

	return contents
}
