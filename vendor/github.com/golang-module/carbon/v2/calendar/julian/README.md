# Carbon

[![Carbon Release](https://img.shields.io/github/release/golang-module/carbon.svg)](https://github.com/golang-module/carbon/releases)
[![Go Test](https://github.com/golang-module/carbon/actions/workflows/test.yml/badge.svg)](https://github.com/golang-module/carbon/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/golang-module/carbon/v2)](https://goreportcard.com/report/github.com/golang-module/carbon/v2)
[![Go Coverage](https://codecov.io/gh/golang-module/carbon/branch/master/graph/badge.svg)](https://codecov.io/gh/golang-module/carbon)
[![Goproxy.cn](https://goproxy.cn/stats/github.com/golang-module/carbon/badges/download-count.svg)](https://goproxy.cn)
[![Carbon Doc](https://img.shields.io/badge/go.dev-reference-brightgreen?logo=go&logoColor=white&style=flat)](https://pkg.go.dev/github.com/golang-module/carbon/v2)
[![License](https://img.shields.io/github/license/golang-module/carbon)](https://github.com/golang-module/carbon/blob/master/LICENSE)

English | [简体中文](README.cn.md) | [日本語](README.jp.md)

#### Introduction

A simple, semantic and developer-friendly golang package for time, has been included
by [awesome-go](https://github.com/avelino/awesome-go#date-and-time "awesome-go")

[github.com/golang-module/carbon](https://github.com/golang-module/carbon "github.com/golang-module/carbon")

[gitee.com/golang-module/carbon](https://gitee.com/golang-module/carbon "gitee.com/golang-module/carbon")

#### Julian Day/Modified Julian Day

##### Convert `Gregorian` calendar to `Julian Day`
```go
// By default, 6 decimal places are retained for precision
carbon.Parse("2024-01-24 12:00:00").Julian().JD() // 2460334
carbon.Parse("2024-01-24 13:14:15").Julian().JD() // 2460334.051563

// 4 decimal places are retained for precision
carbon.Parse("2024-01-24 12:00:00").Julian().JD(4) // 2460334
carbon.Parse("2024-01-24 13:14:15").Julian().JD(4) // 2460334.0516
```

##### Convert `Gregorian` calendar to `Modified Julian Day`
```go
// By default, 6 decimal places are retained for precision
carbon.Parse("2024-01-24 12:00:00").Julian().MJD() // 60333.5
carbon.Parse("2024-01-24 13:14:15").Julian().MJD() // 60333.551563

// 4 decimal places are retained for precision
carbon.Parse("2024-01-24 12:00:00").Julian().MJD(4) // 60333.5
carbon.Parse("2024-01-24 13:14:15").Julian().MJD(4) // 60333.5516
```

##### Convert `Julian Day` to `Modified Julian Day`
```go
// By default, 6 decimal places are retained for precision
carbon.CreateFromJulian(2460334).Julian().MJD() // 60333.5
carbon.CreateFromJulian(2460334.051563).Julian().MJD() // 60332.551563

// 4 decimal places are retained for precision
carbon.CreateFromJulian(2460334).Julian().MJD(4) // 60333.5
carbon.CreateFromJulian(2460334.051563).Julian().MJD(4) // 60332.5516
```

##### Convert `Modified Julian Day` to `Julian Day`
```go
// By default, 6 decimal places are retained for precision
carbon.CreateFromJulian(60333.5).Julian().JD()() // 2460334
carbon.CreateFromJulian(60333.551563).Julian().JD()() // 2460333.051563

// 4 decimal places are retained for precision
carbon.CreateFromJulian(60333.5).Julian().JD(4) // 2460334
carbon.CreateFromJulian(60333.551563).Julian().JD(4) // 2460333.0516
```

##### Convert `Julian Day`/`Modified Julian Day` to `Gregorian` calendar
```go
// Convert Julian Day to Gregorian calendar`
carbon.CreateFromJulian(2460334).ToDateTimeString() // 2024-01-24 12:00:00
carbon.CreateFromJulian(2460334.051563).ToDateTimeString() // 2024-01-24 13:14:15

// Convert Modified Julian Day to Gregorian calendar`
carbon.CreateFromJulian(60333.5).ToDateTimeString() // 2024-01-24 12:00:00
carbon.CreateFromJulian(60333.551563).ToDateTimeString() // 2024-01-24 13:14:15
```