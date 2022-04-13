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

type Order struct {
	Products []Product `json:"products"`
	Price    string    `json:"price"`
	TicketId string    `json:"ticket"`
}

type Package struct {
	FirstSelectedBigTime string                   `json:"first_selected_big_time"`
	Products             []map[string]interface{} `json:"products"`
	EtaTraceId           string                   `json:"eta_trace_id"`
	PackageId            int                      `json:"package_id"`
	ReservedTimeStart    int                      `json:"reserved_time_start"`
	ReservedTimeEnd      int                      `json:"reserved_time_end"`
	SoonArrival          int                      `json:"soon_arrival"`
	PackageType          int                      `json:"package_type"`
}

type PaymentOrder struct {
	ReservedTimeStart    int    `json:"reserved_time_start"`
	ReservedTimeEnd      int    `json:"reserved_time_end"`
	Price                string `json:"price"`
	FreightDiscountMoney string `json:"freight_discount_money"`
	FreightMoney         string `json:"freight_money"`
	OrderFreight         string `json:"order_freight"`
	ParentOrderSign      string `json:"parent_order_sign"`
	ProductType          int    `json:"product_type"`
	AddressId            string `json:"address_id"`
	FormId               string `json:"form_id"`
	ReceiptWithoutSku    string `json:"receipt_without_sku"`
	PayType              int    `json:"pay_type"`
	UserTicketId         string `json:"user_ticket_id"`
	VipMoney             string `json:"vip_money"`
	VipBuyUserTicketId   string `json:"vip_buy_user_ticket_id"`
	CouponsMoney         string `json:"coupons_money"`
	CouponsId            string `json:"coupons_id"`
	UsedPointNum         int    `json:"used_point_num"`
	OrderType            int    `json:"order_type"`
	IsUseBalance         int    `json:"is_use_balance"`
}

type PackageOrder struct {
	Packages     []*Package   `json:"packages"`
	PaymentOrder PaymentOrder `json:"payment_order"`
}

type AddNewOrderReturnData struct {
	Success bool   `json:"success"`
	Code    int    `json:"code"`
	Msg     string `json:"msg"`
	Data    struct {
		PackageOrder     PackageOrder `json:"package_order"`
		StockoutProducts []Product    `json:"stockout_products"`
	} `json:"data"`
}

func (s *DingdongSession) CheckOrder() error {
	urlPath := "https://maicai.api.ddxq.mobi/order/checkOrder"

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
			"sizes":                product.Sizes,
		}
		products = append(products, prod)
	}
	packagesInfo := []map[string]interface{}{
		{
			"package_type": 1,
			"package_id":   1,
			"products":     products,
		},
	}
	packagesJson, _ := json.Marshal(packagesInfo)
	packagesStr := string(packagesJson)

	data := url.Values{}
	for k, v := range map[string]string{
		"uid":                      s.UserId,
		"longitude":                cast.ToString(s.Address.Longitude),
		"latitude":                 cast.ToString(s.Address.Latitude),
		"station_id":               s.Address.StationId,
		"city_number":              s.Address.CityNumber,
		"api_version":              ApiVersion,
		"app_version":              AppVersion,
		"applet_source":            "",
		"app_client_id":            "3",
		"h5_source":                "",
		"wx":                       "1",
		"address_id":               s.Address.Id,
		"user_ticket_id":           "default",
		"freight_ticket_id":        "default",
		"is_use_point":             "0",
		"is_use_balance":           "0",
		"is_buy_vip":               "0",
		"coupons_id":               "",
		"is_buy_coupons":           "0",
		"packages":                 packagesStr,
		"check_order_type":         "0",
		"is_support_merge_payment": "1",
		"showData":                 "true",
		"showMsg":                  "false",
	} {
		data.Set(k, v)
	}

	req, _ := http.NewRequest("POST", urlPath, strings.NewReader(data.Encode()))
	req.Header.Add("Host", "maicai.api.ddxq.mobi")
	req.Header.Add("Referer", "https://wx.m.ddxq.mobi/")
	req.Header.Add("Cookie", s.Cookie)
	req.Header.Add("User-Agent", UA)
	req.Header.Add("ddmc-city-number", s.Address.CityNumber)
	req.Header.Add("ddmc-api-version", ApiVersion)
	req.Header.Add("Origin", "https://wx.m.ddxq.mobi")
	req.Header.Add("ddmc-build-version", AppVersion)
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
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode == 200 {
		result := gjson.Parse(string(body))
		switch result.Get("code").Num {
		case 0:
			s.Order.Price = result.Get("data.order.total_money").Str
			if result.Get("data.order.total_money").Str !=
				result.Get("data.order.goods_real_money").Str {
				s.Order.TicketId = result.Get("data.order.default_coupon._id").Str
			}
			return nil
		case -3000:
			return BusyErr
		case -3001:
			return RateLimit
		case -3100:
			return DataLoadErr
		case 5010:
			s.Order.Price = result.Get("data.order.total_money").Str
			if result.Get("data.order.total_money").Str !=
				result.Get("data.order.goods_real_money").Str {
				s.Order.TicketId = result.Get("data.order.default_coupon._id").Str
			}
			fmt.Println("【5010】========xxxxxxxx")
			return nil
		case 5014:
			return errors.New("【5014】" + result.Get("msg").Str)
		default:
			return errors.New(string(body))
		}
	} else {
		return errors.New(fmt.Sprintf("[%v] %s", resp.StatusCode, body))
	}
}

func (s *DingdongSession) GeneratePackageOrder() {
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
			"sizes":                product.Sizes,
		}
		products = append(products, prod)
	}

	p := Package{
		FirstSelectedBigTime: "0",
		Products:             products,
		EtaTraceId:           "",
		PackageId:            1,
		PackageType:          1,
	}
	paymentOrder := PaymentOrder{
		FreightDiscountMoney: "5.00",
		FreightMoney:         "5.00",
		OrderFreight:         "0.00",
		AddressId:            s.Address.Id,
		UsedPointNum:         0,
		ParentOrderSign:      s.Cart.ParentOrderSign,
		ProductType:          1,
		PayType:              s.PayType,
		UserTicketId:         s.Order.TicketId, // 不使用优惠券
		VipMoney:             "",
		VipBuyUserTicketId:   "",
		CouponsMoney:         "",
		CouponsId:            "",
		OrderType:            1,
		IsUseBalance:         0,
		ReceiptWithoutSku:    "1",
		Price:                s.Order.Price,
	}
	packageOrder := PackageOrder{
		Packages: []*Package{
			&p,
		},
		PaymentOrder: paymentOrder,
	}
	s.PackageOrder = packageOrder
}

func (s *DingdongSession) UpdatePackageOrder(reserveTime ReserveTime) {
	s.PackageOrder.PaymentOrder.ReservedTimeStart = reserveTime.StartTimestamp
	s.PackageOrder.PaymentOrder.ReservedTimeEnd = reserveTime.EndTimestamp
	for _, p := range s.PackageOrder.Packages {
		p.ReservedTimeStart = reserveTime.StartTimestamp
		p.ReservedTimeEnd = reserveTime.EndTimestamp
	}
	fmt.Println("可用时间为：" + reserveTime.SelectMsg)
}

func (s *DingdongSession) AddNewOrder() error {
	urlPath := "https://maicai.api.ddxq.mobi/order/addNewOrder"

	packageOrderJson, _ := json.Marshal(s.PackageOrder)
	packageOrderStr := string(packageOrderJson)

	data := url.Values{}
	for k, v := range map[string]string{
		"uid":           s.UserId,
		"longitude":     cast.ToString(s.Address.Longitude),
		"latitude":      cast.ToString(s.Address.Latitude),
		"station_id":    s.Address.StationId,
		"city_number":   s.Address.CityNumber,
		"api_version":   ApiVersion,
		"app_version":   AppVersion,
		"applet_source": "",
		"app_client_id": AppClientId,
		"h5_source":     "",
		"package_order": packageOrderStr,
		"showMsg":       "false",
		"showData":      "true",
	} {
		data.Set(k, v)
	}

	req, _ := http.NewRequest("POST", urlPath, strings.NewReader(data.Encode()))

	req.Header.Add("Host", "maicai.api.ddxq.mobi")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("content-type", "application/x-www-form-urlencoded")
	req.Header.Add("ddmc-city-number", s.Address.CityNumber)
	req.Header.Add("ddmc-build-version", AppVersion)
	req.Header.Add("ddmc-station-id", s.Address.StationId)
	req.Header.Add("ddmc-channel", "applet")
	req.Header.Add("ddmc-os-version", "[object Undefined]")
	req.Header.Add("ddmc-app-client-id", "3")
	req.Header.Add("Cookie", s.Cookie)
	req.Header.Add("ddmc-ip", "")
	req.Header.Add("ddmc-longitude", cast.ToString(s.Address.Longitude))
	req.Header.Add("ddmc-latitude", cast.ToString(s.Address.Latitude))
	req.Header.Add("ddmc-api-version", ApiVersion)
	req.Header.Add("ddmc-uid", s.UserId)
	req.Header.Add("User-Agent", UA)
	req.Header.Add("Referer", "https://servicewechat.com/wx1111111111111/422/page-frame.html")

	resp, err := s.Client.Do(req)
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode == 200 {
		result := AddNewOrderReturnData{}
		err := json.Unmarshal(body, &result)
		if err != nil {
			return err
		}
		switch result.Code {
		case 0:
			return nil
		case 5001:
			s.PackageOrder = result.Data.PackageOrder
			return OOSErr
		case 5003:
			fmt.Println(string(body))
			return ProdInfoErr
		case 5004:
			fmt.Println(result.Msg)
			s.missTimePoint = append(s.missTimePoint, cast.ToInt64(s.PackageOrder.Packages[0].ReservedTimeStart))
			fmt.Printf("时间：%s 已过期\n", time.Unix(cast.ToInt64(s.PackageOrder.Packages[0].ReservedTimeStart), 0).Format("2006-01-02 15:04:05"))
			return TimeExpireErr
		case 5014:
			fmt.Println(result.Msg)
			return NotStart
		case -3001:
			fmt.Println(result.Msg)
			return RateLimit
		case -3000:
			fmt.Println(result.Msg)
			return BusyErr
		default:
			return errors.New(string(body))
		}
	} else {
		return errors.New(fmt.Sprintf("[%v] %s", resp.StatusCode, body))
	}
}
