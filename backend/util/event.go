package util

import (
	"context"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func Message(ctx context.Context, content string ){
	runtime.EventsEmit(ctx, "message", content)

}
