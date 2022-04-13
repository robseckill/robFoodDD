package dd

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/cast"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type ReserveTime struct {
	StartTimestamp int    `json:"start_timestamp"`
	EndTimestamp   int    `json:"end_timestamp"`
	SelectMsg      string `json:"select_msg"`
}

func (s *DingdongSession) GetMultiReserveTime() (error, []ReserveTime) {
	urlPath := "https://maicai.api.ddxq.mobi/order/getMultiReserveTime"
	var products []map[string]interface{}
	for _, product := range s.Order.Products {
		prod := map[string]interface{}{
			"id":                   product.Id,
			"total_money":          product.TotalPrice,
			"total_origin_money":   product.OriginPrice,
			"count":                product.Count,
			"price":                product.Price,
			"instant_rebate_money": "0.00",
			"origin_price":         product.OriginPrice,
		}
		products = append(products, prod)
	}
	productsList := [][]map[string]interface{}{
		products,
	}
	productsJson, _ := json.Marshal(productsList)
	productsStr := string(productsJson)

	data := fmt.Sprintf("uid=%s&longitude=%f&latitude=%f&station_id=%s&city_number=%s&api_version=%s&app_version=%s"+
		"&applet_source=&channel=applet&app_client_id=%d&sharer_uid=&h5_source="+
		"&address_id=%s&group_config_id=&products=%s"+
		"&isBridge=false",
		s.UserId, s.Address.Longitude,
		s.Address.Longitude, s.Address.StationId,
		s.Address.CityNumber, ApiVersion, AppVersion,
		AppClientId, s.Address.Id, url.QueryEscape(productsStr),
	)
	req, _ := http.NewRequest("POST", urlPath, strings.NewReader(data))
	req.Header.Add("Host", "maicai.api.ddxq.mobi")
	req.Header.Add("Referer", "https://wx.m.ddxq.mobi/")
	req.Header.Add("Cookie", s.Cookie)
	req.Header.Add("User-Agent", UA)
	req.Header.Add("ddmc-city-number", s.Address.CityNumber)
	req.Header.Add("ddmc-api-version", ApiVersion)
	req.Header.Add("Origin", "https://wx.m.ddxq.mobi")
	req.Header.Add("ddmc-build-version", BuildVersion)
	req.Header.Add("ddmc-longitude", cast.ToString(s.Address.Longitude))
	req.Header.Add("ddmc-latitude", cast.ToString(s.Address.Latitude))
	req.Header.Add("ddmc-app-client-id", "3")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("ddmc-uid", s.UserId)
	req.Header.Add("Accept-Language", "zh-CN,zh-Hans;q=0.9")
	req.Header.Add("ddmc-channel", "undefined")
	req.Header.Add("ddmc-device-id", "")
	req.Header.Add("Accept", "application/json, text/plain, */*")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("ddmc-station-id", s.Address.StationId)
	req.Header.Add("ddmc-ip", "")
	req.Header.Add("ddmc-os-version", "undefined")

	resp, err := s.Client.Do(req)
	if err != nil {
		return err, nil
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err, nil
	}
	defer resp.Body.Close()
	fmt.Println(string(body))
	if resp.StatusCode == 200 {
		var reserveTimeList []ReserveTime
		result := gjson.Parse(string(body))

		isSecKill := strings.Index(result.Get("data.0.time.0.times.0.select_msg").Str, "自动尝试") != -1

		// 秒杀
		if isSecKill && len(result.Get("data.0.time.0.times").Array()) == 1 {
			fmt.Println("生成的秒杀时间段。。。。")
			reserveTimeList = s.generateSecKillTimeList(s.missTimePoint...)
		} else {
			fmt.Println("获取成功的时间段。。。。")
			for _, reserveTimeInfo := range result.Get("data.0.time.0.times").Array() {
				if reserveTimeInfo.Get("disableType").Num == 0 {
					reserveTime := ReserveTime{
						StartTimestamp: int(reserveTimeInfo.Get("start_timestamp").Num),
						EndTimestamp:   int(reserveTimeInfo.Get("end_timestamp").Num),
						SelectMsg:      reserveTimeInfo.Get("select_msg").Str,
					}
					reserveTimeList = append(reserveTimeList, reserveTime)
				}
			}
		}
		return nil, reserveTimeList
	} else {
		return errors.New(fmt.Sprintf("[%v] %s", resp.StatusCode, body)), nil
	}
}

func (s *DingdongSession) generateSecKillTimeList(missPoint ...int64) []ReserveTime {
	timePoints := []string{
		"06:30",
		"08:30",
		"10:30",
		"12:30",
		"14:30",
		"16:30",
		"18:30",
		"20:30",
		"22:30",
	}

	// 时间云配
	// 抢购时间段强制指定时间段，不走叮咚列表。
	resp, err := http.Get("https://xxxxx.com/time.json")
	if err == nil {
		if bytes, berr := ioutil.ReadAll(resp.Body); berr == nil {
			var t []string
			if err = json.Unmarshal(bytes, &t); err == nil {
				timePoints = t
			}
		}
	}
	var cstSh, _ = time.LoadLocation("Asia/Shanghai")
	var multiReserveTime []ReserveTime
	today := time.Now().Format("2006-01-02")
	for i := 0; i < len(timePoints)-1; i++ {
		t1s := today + " " + timePoints[i] + ":00"
		t1, _ := time.ParseInLocation("2006-01-02 15:04:05", t1s, cstSh)

		t2s := today + " " + timePoints[i+1] + ":00"
		t2, _ := time.ParseInLocation("2006-01-02 15:04:05", t2s, cstSh)

		if t1.Before(time.Now()) {
			continue
		}

		// 检查生成的时间段是否包含失效时间
		if isMiss(t1, t2, missPoint...) {
			continue
		}

		multiReserveTime = append(multiReserveTime, ReserveTime{
			StartTimestamp: int(t1.Unix()),
			EndTimestamp:   int(t2.Unix()),
			SelectMsg:      "今天 " + timePoints[i] + "-" + timePoints[i+1],
		})
	}
	return s.reTime(multiReserveTime)
}

func (s *DingdongSession) reTime(multiReserveTime []ReserveTime) []ReserveTime {
	var ret []ReserveTime
	for i := len(multiReserveTime) - 1; i >= 0; i-- {
		ret = append(ret, multiReserveTime[i])
	}
	return ret
}

func isMiss(t1, t2 time.Time, missPoint ...int64) bool {
	for _, i2 := range missPoint {
		if t1.Unix() < i2 && t2.Unix() > i2 {
			return true
		}
	}
	return false
}
