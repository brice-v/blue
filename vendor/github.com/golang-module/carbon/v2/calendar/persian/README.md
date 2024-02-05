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

#### Persian(Jalaali) Calendar

##### Convert `Gregorian` calendar to `Persian` calendar

```go
// Get persian year
carbon.Parse("2020-08-05 13:14:15").Persian().Year() // 1399
// Get persian month
carbon.Parse("2020-08-05 13:14:15").Persian().Month() // 5
// Get persian day
carbon.Parse("2020-08-05 13:14:15").Persian().Day() // 15
// Get persian hour
carbon.Parse("2020-08-05 13:14:15").Persian().Hour() // 13
// Get persian minute
carbon.Parse("2020-08-05 13:14:15").Lunar().Minute() // 14
// Get persian second
carbon.Parse("2020-08-05 13:14:15").Lunar().Second() // 15

// Get persian date and time string
carbon.Parse("2020-08-05 13:14:15").Lunar().String() // 1399-05-15 13:14:15
fmt.Printf("%s", carbon.Parse("2020-08-05 13:14:15").Lunar()) // 1399-05-15 13:14:15
// // Get persian month as string
carbon.Parse("2020-08-05 13:14:15").Persian().ToMonthString() // مرداد
// // Get persian week as string
carbon.Parse("2020-08-05 13:14:15").Persian().ToWeekString() // چهارشنبه

// Whether is a persian zero time
carbon.Parse("0000-00-00 00:00:00").Persian().IsZero() // true
carbon.Parse("2020-08-05 13:14:15").Persian().IsZero() // false

// Whether is a persian leap year
carbon.Parse("2016-03-20 00:00:00").Persian().IsLeapYear() // true
carbon.Parse("2020-08-05 13:14:15").Persian().IsLeapYear() // false
```

##### Convert `Persian` calendar to `Gregorian` calendar

```go
carbon.CreateFromPersian(1395, 1, 1, 0, 0, 0).ToDateTimeString() // 2016-03-20 00:00:00
carbon.CreateFromPersian(1399, 5, 15, 13, 14, 15).ToDateTimeString() // 2020-08-05 13:14:15
```