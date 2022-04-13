package dd

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
)

const (
	UA           = "Mozilla/5.0 (Linux; Android 9; LIO-AN00 Build/LIO-AN00; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/92.0.4515.131 Mobile Safari/537.36 xzone/9.47.0 station_id/null"
	ApiVersion   = "9.49.2"
	AppVersion   = "2.82.0"
	BuildVersion = "2.82.1"
	AppClientId  = "4"
)

type DingdongSession struct {
	Address      Address      `json:"address"`
	BarkId       string       `json:"bark_id"`
	Client       *http.Client `json:"client"`
	Cookie       string       `json:"cookie"`
	Cart         Cart         `json:"cart"`
	Order        Order        `json:"order"`
	PackageOrder PackageOrder `json:"package_order"`
	PayType      int          `json:"pay_type"`
	CartMode     int          `json:"cart_mode"`
	UserId       string       `json:"user_id"`

	missTimePoint []int64 `json:"reserve_time_data"`
}

func NewDingdongSession() *DingdongSession {
	return &DingdongSession{
		Client: &http.Client{},
	}
}

func (s *DingdongSession) InitSession(cookie string, barkId string, fastMode int) error {
	fmt.Println("########## 初始化 ##########")
	s.Cookie = cookie
	s.BarkId = barkId
	err, addrList := s.GetAddress()
	if err != nil {
		return err
	}
	if len(addrList) == 0 {
		return errors.New("未查询到有效收货地址，请前往app添加或检查cookie是否正确！")
	}

	msgs := map[int]string{
		1: "结算所有有效商品（不包括换购）",
		2: "结算所有勾选商品（包括换购)",
	}

	if fastMode > 0 && fastMode < 3 {
		s.Address = addrList[0]
		s.PayType = 2
		s.CartMode = fastMode
		modeStr := msgs[fastMode]
		addr := addrList[0]
		fmt.Println()
		fmt.Println("##################################")
		fmt.Printf("收货地址: %s %s %s %s\n", addr.Name, addr.AddrDetail, addr.UserName, addr.Mobile)
		fmt.Printf("支付方式: %s\n", "支付宝")
		fmt.Printf("结算模式: %s\n", modeStr)
		fmt.Println("##################################")
		fmt.Println()
		return nil
	}
	fmt.Println("########## 选择收货地址 ##########")
	for i, addr := range addrList {
		fmt.Printf("[%v] %s %s %s %s \n", i, addr.Name, addr.AddrDetail, addr.UserName, addr.Mobile)
	}
	var index int
	for true {
		fmt.Println("请输入地址序号（0, 1, 2...)：")
		stdin := bufio.NewReader(os.Stdin)
		_, err := fmt.Fscanln(stdin, &index)
		if err != nil {
			fmt.Printf("输入有误：%s!\n", err)
		} else if index >= len(addrList) {
			fmt.Println("输入有误：超过最大序号！")
		} else {
			break
		}
	}
	s.Address = addrList[index]
	fmt.Println("########## 选择支付方式 ##########")
	for true {
		fmt.Println("请输入支付方式序号（1：支付宝 2：微信)：")
		stdin := bufio.NewReader(os.Stdin)
		_, err := fmt.Fscanln(stdin, &index)
		if err != nil {
			fmt.Printf("输入有误：%s!\n", err)
		} else if index == 1 {
			s.PayType = 2
			break
		} else if index == 2 {
			s.PayType = 4
			break
		} else {
			fmt.Println("输入有误：序号无效！")
		}
	}
	fmt.Println("########## 选择购物车商品结算模式 ##########")
	for true {
		fmt.Println("请输入结算模式序号（1：结算所有有效商品（不包括换购） 2：结算所有勾选商品（包括换购)：")
		stdin := bufio.NewReader(os.Stdin)
		_, err := fmt.Fscanln(stdin, &index)
		if err != nil {
			fmt.Printf("输入有误：%s!\n", err)
		} else if index == 1 {
			s.CartMode = 1
			break
		} else if index == 2 {
			s.CartMode = 2
			break
		} else {
			fmt.Println("输入有误：序号无效！")
		}
	}
	return nil
}

func ToJson(data interface{}) string {
	a, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	return string(a)
}
