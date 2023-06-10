package lib

import (
	"embed"
)

//go:embed core/core.b
var CoreFile string

//go:embed std/*
var stdFs embed.FS

//go:embed web/*
var WebEmbedFiles embed.FS

func ReadStdFileToString(fname string) string {
	bs, err := stdFs.ReadFile("std/" + fname)
	if err != nil {
		panic(err)
	}
	return string(bs)
}
