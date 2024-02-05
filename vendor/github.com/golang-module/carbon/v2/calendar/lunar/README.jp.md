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

#### 中国の旧暦

> 現在は西暦` 1900 `年から` 2100 `年までの` 200 `年の時間スパンのみをサポートしている

##### `西暦`を`旧暦`に変換する

```go
// 旧暦の干支を手に入れる
carbon.Parse("2020-08-05 13:14:15").Lunar().Animal() // 鼠
// 旧暦の祝日を取得する
carbon.Parse("2021-02-12 13:14:15").Lunar().Festival() // 春节

// 旧正月の取得
carbon.Parse("2020-08-05 13:14:15").Lunar().Year() // 2020
// 旧暦月の取得
carbon.Parse("2020-08-05 13:14:15").Lunar().Month() // 6
// 旧暦うるう月の取得
carbon.Parse("2020-08-05 13:14:15").Lunar().LeapMonth() // 4
// 旧暦日の取得
carbon.Parse("2020-08-05 13:14:15").Lunar().Day() // 16
// 旧暦時刻の取得
carbon.Parse("2020-08-05 13:14:15").Lunar().Hour() // 13
// 旧暦分の取得
carbon.Parse("2020-08-05 13:14:15").Lunar().Minute() // 14
// 旧暦の取得秒数
carbon.Parse("2020-08-05 13:14:15").Lunar().Second() // 15

// 旧暦日時文字列の取得
carbon.Parse("2020-08-05 13:14:15").Lunar().String() // 2020-06-16 13:14:15
fmt.Printf("%s", carbon.Parse("2020-08-05 13:14:15").Lunar()) // 2020-06-16 13:14:15
// 旧正月文字列の取得
carbon.Parse("2020-08-05 13:14:15").Lunar().ToYearString() // 二零二零
// 旧暦月文字列の取得
carbon.Parse("2020-08-05 13:14:15").Lunar().ToMonthString() // 六月
// 旧暦うるう月文字列の取得
carbon.Parse("2020-04-23 13:14:15").Lunar().ToMonthString() // 闰四月
// 旧暦週文字列の取得
carbon.Parse("2020-04-23 13:14:15").Lunar().ToWeekString() // 周四
// 旧暦日文字列の取得
carbon.Parse("2020-08-05 13:14:15").Lunar().ToDayString() // 十六
// 旧暦日付文字列の取得
carbon.Parse("2020-08-05 13:14:15").Lunar().ToDateString() // 二零二零年六月十六

// ゼロ値の時間ですか
carbon.Parse("0000-00-00 00:00:00").Lunar().IsZero() // true
carbon.Parse("2020-08-05 13:14:15").Lunar().IsZero() // false

// 旧暦うるう年かどうか
carbon.Parse("2020-08-05 13:14:15").Lunar().IsLeapYear() // true
// 旧暦うるう月かどうか
carbon.Parse("2020-08-05 13:14:15").Lunar().IsLeapMonth() // false

// ねずみ年かどうか
carbon.Parse("2020-08-05 13:14:15").Lunar().IsRatYear() // true
// 牛年かどうか
carbon.Parse("2020-08-05 13:14:15").Lunar().IsOxYear() // false
// 寅年かどうか
carbon.Parse("2020-08-05 13:14:15").Lunar().IsTigerYear() // false
// うさぎ年かどうか
carbon.Parse("2020-08-05 13:14:15").Lunar().IsRabbitYear() // false
// 龍年かどうか
carbon.Parse("2020-08-05 13:14:15").Lunar().IsDragonYear() // false
// 蛇の年かどうか
carbon.Parse("2020-08-05 13:14:15").Lunar().IsSnakeYear() // false
// 馬年かどうか
carbon.Parse("2020-08-05 13:14:15").Lunar().IsHorseYear() // false
// 羊年かどうか
carbon.Parse("2020-08-05 13:14:15").Lunar().IsGoatYear() // false
// 申年かどうか
carbon.Parse("2020-08-05 13:14:15").Lunar().IsMonkeyYear() // false
// 鶏の年かどうか
carbon.Parse("2020-08-05 13:14:15").Lunar().IsRoosterYear() // false
// 犬年かどうか
carbon.Parse("2020-08-05 13:14:15").Lunar().IsDogYear() // false
// 豚年かどうか
carbon.Parse("2020-08-05 13:14:15").Lunar().IsPigYear() // false
```

##### `旧暦`を`西暦`に変換する

```go
// 2023 年の旧暦 12 月 11 日をグレゴリオ暦に変換します
carbon.CreateFromLunar(2023, 12, 11, 0, 0, 0, false).ToDateTimeString() // 2024-01-21 00:00:00
// 旧暦の 2023 年 2 月 11 日をグレゴリオ暦に変換します
carbon.CreateFromLunar(2023, 2, 11, 0, 0, 0, false).ToDateTimeString() // 2023-03-02 00:00:00
// 旧暦 2023 年、閏 2 月 11 日をグレゴリオ暦に変換します
carbon.CreateFromLunar(2023, 2, 11, 0, 0, 0, true).ToDateTimeString() // 2023-04-01 00:00:00
```