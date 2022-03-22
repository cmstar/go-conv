//go:build !debug
// +build !debug

package conv

import "sync"

type syncMap = sync.Map
