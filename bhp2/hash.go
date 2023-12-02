package bhp2

import (
	"fmt"
	"runtime/debug"
	"time"
)

var Hash string = fmt.Sprintf("%d", time.Now().Unix())

func init() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		panic("failed to read build info")
	}
	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" {
			Hash = setting.Value
		}
	}

	time.Local = time.UTC
}
