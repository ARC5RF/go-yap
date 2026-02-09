package yap

import (
	"os"

	iyap "github.com/ARC5RF/go-yap"
)

var legacy = iyap.New(nil, os.Stdout, 128, 100, iyap.INSERT_SPACES_IN_FILE|iyap.INSERT_SPACES_IN_TERMINAL)
var Log = legacy.Log
var ServeClientJS = iyap.ServeClientJS
var LogLN = legacy.LogLN
var Debug = legacy.Debug

var To = legacy.Websocket.Emit
var Add = legacy.Websocket.Add
var Rem = legacy.Websocket.Rem
var Keep = legacy.Websocket.Keep
var On = legacy.Websocket.On
var Purge = legacy.Websocket.Purge

func Huge(v any) any { return iyap.Huge(1, v) }

func Red(input any) any     { return iyap.Red(input) }
func Green(input any) any   { return iyap.Green(input) }
func Yellow(input any) any  { return iyap.Yellow(input) }
func Blue(input any) any    { return iyap.Blue(input) }
func Magenta(input any) any { return iyap.Magenta(input) }
func Cyan(input any) any    { return iyap.Cyan(input) }
func Gray(input any) any    { return iyap.Gray(input) }
func White(input any) any   { return iyap.White(input) }

var NL = iyap.NL
