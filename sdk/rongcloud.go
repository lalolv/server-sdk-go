// The MIT License (MIT)

// Copyright (c) 2014 融云 Rong Cloud

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

/*
 * 融云 Server API go 客户端
 * create by RongCloud
 * create datetime : 2018-11-28
 * v3
 */

package sdk

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/url"
	"reflect"
	"strconv"
	"time"

	"github.com/astaxie/beego/httplib"
)

const (
	// RONGCLOUDSMSURI 容云默认 SMS API 地址
	RONGCLOUDSMSURI = "http://api.sms.ronghub.com"
	// RONGCLOUDURI 容云默认 API 地址
	RONGCLOUDURI = "http://api-cn.ronghub.com"
	// RONGCLOUDURI_2 容云备用 API 地址
	RONGCLOUDURI_2 = "http://api2-cn.ronghub.com"
	// ReqType body类型
	ReqType = "json"
	// USERAGENT sdk 名称
	USERAGENT = "rc-go-sdk/3.0"
	// DEFAULTTIMEOUT 默认超时时间
	DEFAULTTIMEOUT = 10
	// NUMTIMEOUT 默认超时次数切换 api 地址
	NUMTIMEOUT = 3
)

// RongCloud appKey appSecret extra
type RongCloud struct {
	appKey    string
	appSecret string
	*RongCloudExtra
}

// RongCloudExtra RongCloud扩展增加自定义容云服务器地址,请求超时时间
type RongCloudExtra struct {
	RongCloudURI    string
	RongCloudSMSURI string
	TimeOut         time.Duration
	numTimeOut      int
}

// CodeResult 容云返回状态码和错误码
type CodeResult struct {
	Code         int    `json:"code"`
	ErrorMessage string `json:"errorMessage"`
}

// getSignature 本地生成签名
// Signature (数据签名)计算方法：将系统分配的 App Secret、Nonce (随机数)、
// Timestamp (时间戳)三个字符串按先后顺序拼接成一个字符串并进行 SHA1 哈希计算。如果调用的数据签名验证失败，接口调用会返回 HTTP 状态码 401。
func (rc *RongCloud) getSignature() (nonce, timestamp, signature string) {
	nonceInt := rand.Int()
	nonce = strconv.Itoa(nonceInt)
	timeInt64 := time.Now().Unix()
	timestamp = strconv.FormatInt(timeInt64, 10)
	h := sha1.New()
	io.WriteString(h, rc.appSecret+nonce+timestamp)
	signature = fmt.Sprintf("%x", h.Sum(nil))
	return
}

// FillHeader 在http header 增加API签名
func (rc *RongCloud) FillHeader(req *httplib.BeegoHTTPRequest) {
	nonce, timestamp, signature := rc.getSignature()
	req.Header("App-Key", rc.appKey)
	req.Header("Nonce", nonce)
	req.Header("Timestamp", timestamp)
	req.Header("Signature", signature)
	req.Header("Content-Type", "application/x-www-form-urlencoded")
	req.Header("User-Agent", USERAGENT)
}

// FillJSONHeader 在http header Content-Type 设置为josn格式
func FillJSONHeader(req *httplib.BeegoHTTPRequest) {
	req.Header("Content-Type", "application/json")
}

// NewRongCloud 创建RongCloud对象
func NewRongCloud(appKey, appSecret string, extra *RongCloudExtra) *RongCloud {
	// 默认扩展配置
	defaultExtra := RongCloudExtra{
		RongCloudURI:    RONGCLOUDURI,
		RongCloudSMSURI: RONGCLOUDSMSURI,
		TimeOut:         DEFAULTTIMEOUT,
		numTimeOut:      NUMTIMEOUT,
	}
	// 使用默认服务器地址
	if extra == nil {
		rc := RongCloud{
			appKey:         appKey,    //app key
			appSecret:      appSecret, //app secret
			RongCloudExtra: &defaultExtra,
		}
		return &rc
	}
	if extra.TimeOut == 0 {
		extra.TimeOut = DEFAULTTIMEOUT
	}
	// RongCloudSMSURI RongCloudURI 必须同时修改
	if extra.RongCloudSMSURI == "" || extra.RongCloudURI == "" {
		extra.RongCloudURI = RONGCLOUDURI
		extra.RongCloudSMSURI = RONGCLOUDSMSURI
	}
	// 使用扩展配置地址
	rc := RongCloud{
		appKey:         appKey,    //app key
		appSecret:      appSecret, //app secret
		RongCloudExtra: extra,
	}
	return &rc
}

// SetRongCloudURI 设置 api 服务器地址
func (rc *RongCloud) setRongCloudURI() {
	if rc.RongCloudURI == RONGCLOUDURI {
		rc.RongCloudURI = RONGCLOUDURI_2
		return
	}
	if rc.RongCloudURI != RONGCLOUDURI && rc.RongCloudURI == RONGCLOUDURI_2 {
		return
	}
	rc.RongCloudURI = RONGCLOUDURI
}

// SetNumTimeOut 设置切换 api 超时次数
func (rc *RongCloud) SetNumTimeOut(num int) error {
	if num < 0 {
		return errors.New("invalid num, only allow greater than and equal to the 0 ")
	}
	rc.numTimeOut = num
	return nil
}

// URLError 判断是否为 url.Error
func (rc *RongCloud) URLError(err error) {
	if reflect.TypeOf(err) == reflect.TypeOf(&url.Error{}) {
		if rc.numTimeOut == 0 {
			return
		}
		if rc.numTimeOut >= 3 {
			rc.setRongCloudURI()
			rc.numTimeOut = 1
		} else {
			rc.numTimeOut += 1
		}
	}
}