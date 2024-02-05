# Carbon

[![Carbon Release](https://img.shields.io/github/release/golang-module/carbon.svg)](https://github.com/golang-module/carbon/releases)
[![Go Test](https://github.com/golang-module/carbon/actions/workflows/test.yml/badge.svg)](https://github.com/golang-module/carbon/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/golang-module/carbon/v2)](https://goreportcard.com/report/github.com/golang-module/carbon/v2)
[![Go Coverage](https://codecov.io/gh/golang-module/carbon/branch/master/graph/badge.svg)](https://codecov.io/gh/golang-module/carbon)
[![Goproxy.cn](https://goproxy.cn/stats/github.com/golang-module/carbon/badges/download-count.svg)](https://goproxy.cn)
[![Carbon Doc](https://img.shields.io/badge/go.dev-reference-brightgreen?logo=go&logoColor=white&style=flat)](https://pkg.go.dev/github.com/golang-module/carbon/v2)
[![License](https://img.shields.io/github/license/golang-module/carbon)](https://github.com/golang-module/carbon/blob/master/LICENSE)

日本語 | [English](README.md) | [简体中文](README.cn.md)

一个轻量级、语义化、对开发者友好的 golang 时间处理库，支持链式调用，已被 [awesome-go-cn](https://github.com/yinggaozhen/awesome-go-cn#日期和时间 "awesome-go-cn") 收录

[github.com/golang-module/carbon](https://github.com/golang-module/carbon "github.com/golang-module/carbon")

[gitee.com/golang-module/carbon](https://gitee.com/golang-module/carbon "gitee.com/golang-module/carbon")

#### ペルシア暦（イラン暦）

##### `西暦` を `ペルシャ暦` に変換

```go
// ペルシャ暦の取得
carbon.Parse("2020-08-05 13:14:15").Persian().Year() // 1399
// ペルシャ暦月の取得
carbon.Parse("2020-08-05 13:14:15").Persian().Month() // 5
// ペルシャ暦の取得日
carbon.Parse("2020-08-05 13:14:15").Persian().Day() // 15
// ペルシャ暦時間の取得
carbon.Parse("2020-08-05 13:14:15").Persian().Hour() // 13
// ペルシャ暦分の取得
carbon.Parse("2020-08-05 13:14:15").Lunar().Minute() // 14
// ペルシャ暦秒の取得
carbon.Parse("2020-08-05 13:14:15").Lunar().Second() // 15

// ペルシャ暦日時文字列の取得
carbon.Parse("2020-08-05 13:14:15").Lunar().String() // 1399-05-15 13:14:15
fmt.Printf("%s", carbon.Parse("2020-08-05 13:14:15").Lunar()) // 1399-05-15 13:14:15
// ペルシア暦月文字列の取得
carbon.Parse("2020-08-05 13:14:15").Persian().ToMonthString() // مرداد
// ペルシャ暦週文字列の取得
carbon.Parse("2020-08-05 13:14:15").Persian().ToWeekString() // چهارشنبه

// ペルシャ暦ゼロ時間かどうか
carbon.Parse("0000-00-00 00:00:00").Persian().IsZero() // true
carbon.Parse("2020-08-05 13:14:15").Persian().IsZero() // false

// ペルシア暦閏年かどうか
carbon.Parse("2016-03-20 00:00:00").Persian().IsLeapYear() // true
carbon.Parse("2020-08-05 13:14:15").Persian().IsLeapYear() // false
```

##### ペルシャ暦を西暦に変換する

```go
carbon.CreateFromPersian(1395, 1, 1, 0, 0, 0).ToDateTimeString() // 2016-03-20 00:00:00
carbon.CreateFromPersian(1399, 5, 15, 13, 14, 15).ToDateTimeString() // 2020-08-05 13:14:15
```