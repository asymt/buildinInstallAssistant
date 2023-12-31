package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"unsafe"
)

var cookieJar, _ = cookiejar.New(nil)

func BasicGet(requestGetURL string, headers map[string]string, cookies map[string]string, params map[string]string) (ret *string, err error) {
	client := &http.Client{
		Jar: cookieJar,
	}
	getURL := requestGetURL
	if params != nil && len(params) > 0 {
		rawUrl, err := url.Parse(getURL)
		if err != nil {
			return nil, err
		}
		urlParams := url.Values{}
		for key, value := range params {
			urlParams.Add(key, value)
		}
		if len(urlParams) > 0 {
			rawUrl.RawQuery = urlParams.Encode()
		}
		getURL = rawUrl.String()
	}
	req, err := http.NewRequest(http.MethodGet, getURL, nil)
	if err != nil {
		log.Println("创建请求失败", err.Error())
		return nil, err
	}
	if headers != nil && len(headers) > 0 {
		// 添加请求头
		for key, value := range headers {
			req.Header.Add(key, value)
		}
	}
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0")
	if cookies != nil && len(cookies) > 0 {
		// 添加cookie
		for key, value := range cookies {
			cookie := &http.Cookie{
				Name:  key,
				Value: value,
			}
			req.AddCookie(cookie)
		}
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		log.Println("请求链接："+getURL+"失败", err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprint("请求失败！错误代码：", resp.StatusCode))
	}
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("读取响应内容失败", err.Error())
		return nil, err
	}

	//fmt.Println(string(content))       // 直接打印
	str := (*string)(unsafe.Pointer(&content)) //转化为string,优化内存
	return str, nil
}

func BasicPostJson(requestPostURL string, headers map[string]string, cookies map[string]string, data interface{}) (ret *string, err error) {
	client := &http.Client{
		Jar: cookieJar,
	}
	var dataReader io.Reader
	if data != nil {
		bytesData, _ := json.Marshal(data)
		dataReader = bytes.NewReader(bytesData)
	}
	req, err := http.NewRequest(http.MethodPost, requestPostURL, dataReader)
	if err != nil {
		log.Println("创建请求失败", err.Error())
		return nil, err
	}
	if headers != nil && len(headers) > 0 {
		// 添加请求头
		for key, value := range headers {
			req.Header.Add(key, value)
		}
	}

	//增加application/json的Content-Type
	req.Header.Set("content-type", "application/json; charset=UTF-8")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0")
	if cookies != nil && len(cookies) > 0 {
		// 添加cookie
		for key, value := range cookies {
			cookie := &http.Cookie{
				Name:  key,
				Value: value,
			}
			req.AddCookie(cookie)
		}
	}
	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		log.Println("请求链接："+requestPostURL+"失败", err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprint("请求失败！错误代码：", resp.StatusCode))
	}
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("读取响应内容失败", err.Error())
		return nil, err
	}

	//fmt.Println(string(content))       // 直接打印
	str := (*string)(unsafe.Pointer(&content)) //转化为string,优化内存
	return str, nil
}

func BasicPostForm(requestPostURL string, headers map[string]string, cookies map[string]string, data map[string]interface{}) (ret *string, err error) {
	client := &http.Client{
		Jar: cookieJar,
	}
	var dataReader io.Reader
	if data != nil && len(data) > 0 {
		formParams := url.Values{}
		for key, value := range data {
			if value != nil {
				bytesValue, err := json.Marshal(value)
				if err != nil {
					return nil, err
				}
				formParams.Add(key, string(bytesValue))
			}
		}
		if len(formParams) > 0 {
			dataReader = strings.NewReader(formParams.Encode())
		}
	}
	req, err := http.NewRequest(http.MethodPost, requestPostURL, dataReader)
	if err != nil {
		log.Println("创建请求失败", err.Error())
		return nil, err
	}
	if headers != nil && len(headers) > 0 {
		// 添加请求头
		for key, value := range headers {
			req.Header.Add(key, value)
		}
	}

	//增加application/json的Content-Type
	req.Header.Set("content-type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0")
	if cookies != nil && len(cookies) > 0 {
		// 添加cookie
		for key, value := range cookies {
			cookie := &http.Cookie{
				Name:  key,
				Value: value,
			}
			req.AddCookie(cookie)
		}
	}
	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		log.Println("请求链接："+requestPostURL+"失败", err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprint("请求失败！错误代码：", resp.StatusCode))
	}
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("读取响应内容失败", err.Error())
		return nil, err
	}

	//fmt.Println(string(content))       // 直接打印
	str := (*string)(unsafe.Pointer(&content)) //转化为string,优化内存
	return str, nil
}

func Get(requestGetUrl string, params map[string]string) (ret *string, err error) {
	return BasicGet(requestGetUrl, nil, nil, params)
}

func PostJson(requestPostUrl string, data map[string]interface{}) (ret *string, err error) {
	return BasicPostJson(requestPostUrl, nil, nil, data)
}

func PostForm(requestPostUrl string, data map[string]interface{}) (ret *string, err error) {
	return BasicPostForm(requestPostUrl, nil, nil, data)
}
