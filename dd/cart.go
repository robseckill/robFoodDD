package dd

import (
	"errors"
	"fmt"
	"github.com/spf13/cast"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
	"net/url"
)

type Product struct {
	Id               string                   `json:"id"`
	ProductName      string                   `json:"-"`
	Price            string                   `json:"price"`
	Count            int                      `json:"count"`
	Sizes            []map[string]interface{} `json:"sizes"`
	TotalPrice       string                   `json:"total_money"`
	OriginPrice      string                   `json:"origin_price"`
	TotalOriginPrice string                   `json:"total_origin_money"`
}

func parseProduct(productMap gjson.Result) (error, Product) {
	var sizes []map[string]interface{}
	for _, size := range productMap.Get("sizes").Array() {
		sizes = append(sizes, size.Value().(map[string]interface{}))
	}
	product := Product{
		Id:          productMap.Get("id").Str,
		ProductName: productMap.Get("product_name").Str,
		Price:       productMap.Get("price").Str,
		Count:       int(productMap.Get("count").Num),
		TotalPrice:  productMap.Get("total_price").Str,
		OriginPrice: productMap.Get("origin_price").Str,
		Sizes:       sizes,
	}
	return nil, product
}

type Cart struct {
	ProdList        []Product `json:"effective_products"`
	ParentOrderSign string    `json:"parent_order_sign"`
}

func (s *DingdongSession) GetEffProd(result gjson.Result) error {
	var effProducts []Product
	effective := result.Get("data.product.effective").Array()
	for _, effProductMap := range effective {
		for _, productMap := range effProductMap.Get("products").Array() {
			_, product := parseProduct(productMap)
			effProducts = append(effProducts, product)
		}
	}
	s.Cart = Cart{
		ProdList:        effProducts,
		ParentOrderSign: result.Get("data.parent_order_info.parent_order_sign").Str,
	}
	return nil
}

func (s *DingdongSession) GetCheckProd(result gjson.Result) error {
	var products []Product
	orderProductList := result.Get("data.new_order_product_list").Array()
	for _, productList := range orderProductList {
		for _, productMap := range productList.Get("products").Array() {
			_, product := parseProduct(productMap)
			products = append(products, product)
		}
	}
	s.Cart = Cart{
		ProdList:        products,
		ParentOrderSign: result.Get("data.parent_order_info.parent_order_sign").Str,
	}
	return nil
}

func (s *DingdongSession) CheckCart() error {
	Url, _ := url.Parse("https://maicai.api.ddxq.mobi/cart/index")
	params := url.Values{}
	params.Set("api_version", ApiVersion)
	params.Set("app_version", AppVersion)
	params.Set("applet_source", "")
	params.Set("channel", "applet")
	params.Set("app_client_id", "3")
	params.Set("h5_source", "")
	params.Set("longitude", cast.ToString(s.Address.Longitude))
	params.Set("latitude", cast.ToString(s.Address.Latitude))
	params.Set("station_id", s.Address.StationId)
	params.Set("city_number", s.Address.CityNumber)
	params.Set("is_load", "1")

	Url.RawQuery = params.Encode()
	urlPath := Url.String()
	req, _ := http.NewRequest("POST", urlPath, nil)
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
	req.Header.Add("Referer", "https://servicewechat.com/wx111111111111111/422/page-frame.html")

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
			switch s.CartMode {
			case 1:
				return s.GetEffProd(result)
			case 2:
				return s.GetCheckProd(result)
			default:
				return errors.New("incorrect cart mode")
			}
		case -3000:
			return BusyErr
		default:
			return errors.New(string(body))
		}
	} else {
		return errors.New(fmt.Sprintf("[%v] %s", resp.StatusCode, body))
	}
}
