package yap

import (
	"os"

	iyap "github.com/ARC5RF/go-yap"
)

var legacy = iyap.New(nil, os.Stdout, 128, 100, true, false)
var Log = legacy.Log

var To = legacy.Websocket.Emit
var Add = legacy.Websocket.Add
var Rem = legacy.Websocket.Rem
var Keep = legacy.Websocket.Keep
var On = legacy.Websocket.On
var Purge = legacy.Websocket.Purge
