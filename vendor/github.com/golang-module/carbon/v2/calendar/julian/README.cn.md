# Carbon

[![Carbon Release](https://img.shields.io/github/release/golang-module/carbon.svg)](https://github.com/golang-module/carbon/releases)
[![Go Test](https://github.com/golang-module/carbon/actions/workflows/test.yml/badge.svg)](https://github.com/golang-module/carbon/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/golang-module/carbon/v2)](https://goreportcard.com/report/github.com/golang-module/carbon/v2)
[![Go Coverage](https://codecov.io/gh/golang-module/carbon/branch/master/graph/badge.svg)](https://codecov.io/gh/golang-module/carbon)
[![Goproxy.cn](https://goproxy.cn/stats/github.com/golang-module/carbon/badges/download-count.svg)](https://goproxy.cn)
[![Carbon Doc](https://img.shields.io/badge/go.dev-reference-brightgreen?logo=go&logoColor=white&style=flat)](https://pkg.go.dev/github.com/golang-module/carbon/v2)
[![License](https://img.shields.io/github/license/golang-module/carbon)](https://github.com/golang-module/carbon/blob/master/LICENSE)

简体中文 | [English](README.md) | [日本語](README.jp.md)

一个轻量级、语义化、对开发者友好的 golang 时间处理库，支持链式调用，已被 [awesome-go-cn](https://github.com/yinggaozhen/awesome-go-cn#日期和时间 "awesome-go-cn") 收录

[github.com/golang-module/carbon](https://github.com/golang-module/carbon "github.com/golang-module/carbon")

[gitee.com/golang-module/carbon](https://gitee.com/golang-module/carbon "gitee.com/golang-module/carbon")

#### 儒略日/简化儒略日

##### 将 `公历` 转换成 `儒略日`
```go
// 默认保留 6 位小数精度
carbon.Parse("2024-01-24 12:00:00").Julian().JD() // 2460334
carbon.Parse("2024-01-24 13:14:15").Julian().JD() // 2460334.051563

// 保留 4 位小数精度
carbon.Parse("2024-01-24 12:00:00").Julian().JD(4) // 2460334
carbon.Parse("2024-01-24 13:14:15").Julian().JD(4) // 2460334.0516
```

##### 将 `公历` 转换成 `简化儒略日`
```go
// 默认保留 6 位小数精度
carbon.Parse("2024-01-24 12:00:00").Julian().MJD() // 60333.5
carbon.Parse("2024-01-24 13:14:15").Julian().MJD() // 60333.551563

// 保留 4 位小数精度
carbon.Parse("2024-01-24 12:00:00").Julian().MJD(4) // 60333.5
carbon.Parse("2024-01-24 13:14:15").Julian().MJD(4) // 60333.5516
```

##### 将 `儒略日` 转换成 `简化儒略日`
```go
// 默认保留 6 位小数精度
carbon.CreateFromJulian(2460334).Julian().MJD() // 60333.5
carbon.CreateFromJulian(2460334.051563).Julian().MJD() // 60332.551563

// 保留 4 位小数精度
carbon.CreateFromJulian(2460334).Julian().MJD(4) // 60333.5
carbon.CreateFromJulian(2460334.051563).Julian().MJD(4) // 60332.5516
```

##### 将 `简化儒略日` 转换成 `儒略日`
```go
// 默认保留 6 位小数精度
carbon.CreateFromJulian(60333.5).Julian().JD()() // 2460334
carbon.CreateFromJulian(60333.551563).Julian().JD()() // 2460333.051563

// 保留 4 位小数精度
carbon.CreateFromJulian(60333.5).Julian().JD(4) // 2460334
carbon.CreateFromJulian(60333.551563).Julian().JD(4) // 2460333.0516
```

##### 将 `儒略日`/`简化儒略日` 转换成 `公历`
```go
// 将 儒略日 转换成 公历
carbon.CreateFromJulian(2460334).ToDateTimeString() // 2024-01-24 12:00:00
carbon.CreateFromJulian(2460334.051563).ToDateTimeString() // 2024-01-24 13:14:15

// 将 简化儒略日 转换成 公历
carbon.CreateFromJulian(60333.5).ToDateTimeString() // 2024-01-24 12:00:00
carbon.CreateFromJulian(60333.551563).ToDateTimeString() // 2024-01-24 13:14:15
```


