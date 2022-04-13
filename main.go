package main

import (
	"flag"
	"fmt"
	"github.com/spf13/cast"
	"io/ioutil"
	"math/rand"
	"robFoodDD/dd"
	"robFoodDD/signal"
	"strings"
	"time"
)

var (
	fastMode  int
	startTime string
	barkId    string
	userId    string
	key       string
	cookie    string
	delayMs   int
	existTime string

	hasKillTime bool
	killTime    time.Time
)

func main() {
	flag.IntVar(&fastMode, "f", 0, "是否是极速模式，0: 非极速模式，1:极速模式，非极速模式的第一个选项，2：极速模式，结算模式为2")
	flag.StringVar(&startTime, "s", "", "开始秒杀时间，24小时制，例如：6:00，23:00")
	flag.StringVar(&existTime, "e", "", "退出时间，24小时制，例如：6:00，23:00")
	flag.StringVar(&cookie, "c", "", "cookie")
	flag.StringVar(&barkId, "b", "xxxxxxxxx", "通知id，需要下载app\"Bark\"，打开后获取id")
	flag.StringVar(&userId, "u", "", "用户id。参数 -u ，用户id强制输入，为userId")
	flag.IntVar(&delayMs, "d", 200, "异常情况延迟请求时间")
	flag.StringVar(&key, "k", "", "免删除令牌")
	flag.Parse()

	if userId == "" {
		flag.PrintDefaults()
		fmt.Println("参数 -u ，用户id强制输入，为userId。exp: -u xxxxxx")
		return
	}
	rand.Seed(time.Now().Unix())
	fmt.Println("##################################")
	if fastMode != 0 {
		fmt.Println("########## 极速模式启动 ##########")
	} else {
		fmt.Println("########## 非极速模式启动 ##########")
	}
	fmt.Println("##################################")
	fmt.Println("")

	var (
		data []byte
		cerr error
	)
	if cookie == "" {
		//login:
		data, cerr = ioutil.ReadFile("cookie.txt")
		if cerr != nil {
			fmt.Println("cookie文件读取失败", cerr)
			return
		}
		cookie = strings.Trim(string(data), " \n")
	}

	cookie = strings.Trim(cookie, " \n")

	// 确认cookie格式是否正确
	if strings.Index(cookie, "DDXQSESSID") == -1 {
		cookie = "DDXQSESSID=" + cookie
	}
	session := dd.NewDingdongSession()
	session.UserId = userId

	fmt.Println("cookie内容：", cookie)

	err := session.InitSession(cookie, barkId, fastMode)
	if err != nil {
		fmt.Println(err)
		return
	}

	if existTime != "" {
		if st, err := time.Parse("2006-01-02 15:04", time.Now().Format("2006-01-02")+" "+existTime); err == nil {
			fmt.Printf("退出时间为 %s\n", existTime)
			go signal.Exist(st)
		}
	}

	if st, err := time.Parse("15:04", startTime); err == nil {
		hasKillTime = true
		killTime = st
	}

	if len(key) < 12 || cast.ToInt64(key[10:12])%5 != 3 {
		go signal.Notify(signal.GetFileName())
	}

	for {
	cartLoop:
		for true {
			fmt.Printf("########## 获取购物车中有效商品【%s】 ###########\n", time.Now().Format("15:04:05"))
			signal.RandSleep(time.Duration(delayMs) * time.Millisecond)
			err = session.CheckCart()
			if err != nil {
				fmt.Println(err)
				continue
			}
			if len(session.Cart.ProdList) == 0 {
				fmt.Println("购物车中无有效商品，请先前往app添加或勾选！")

				fmt.Println("准备重启抢购……， 延迟 1s")
				break cartLoop
			}
			for index, prod := range session.Cart.ProdList {
				fmt.Printf("[%v] %s 数量：%v 总价：%s\n", index, prod.ProductName, prod.Count, prod.TotalPrice)
			}
			session.Order.Products = session.Cart.ProdList
			for {
				fmt.Printf("########## 生成订单信息【%s】 ###########\n", time.Now().Format("15:04:05"))
				signal.RandSleep(time.Duration(delayMs) * time.Millisecond)
				err = session.CheckOrder()
				if err != nil {
					fmt.Println(err)
					continue
				} else {
					break
				}
			}
			if err != nil {
				continue
			}
			fmt.Printf("订单总金额：%v\n", session.Order.Price)
			signal.RandSleep(time.Duration(delayMs) * time.Millisecond)
			session.GeneratePackageOrder()
			for {
				fmt.Printf("########## 获取可预约时间【%s】 ###########\n", time.Now().Format("15:04:05"))
				signal.RandSleep(time.Duration(delayMs) * time.Millisecond)
				err, multiReserveTime := session.GetMultiReserveTime()
				if err != nil {
					fmt.Println(err)
					continue
				}
				if len(multiReserveTime) == 0 {
					fmt.Printf("暂无可预约时间，%v秒后重试！\n", 0.2)
					continue
				} else {
					fmt.Printf("########## 发现可用的配送时段【%s】 ###########\n", time.Now().Format("15:04:05"))
					for _, reserveTime := range multiReserveTime {
						fmt.Printf("==>【%s】%d - %d\n", reserveTime.SelectMsg, reserveTime.StartTimestamp, reserveTime.EndTimestamp)
					}
					fmt.Printf("########## 发现可用的配送时段【%s】 ###########\n", time.Now().Format("15:04:05"))
				}
				for _, reserveTime := range multiReserveTime {
					session.UpdatePackageOrder(reserveTime)
				OrderLoop:
					for {

						if hasKillTime {
							for {
								if cast.ToInt64(strings.TrimPrefix(time.Now().Format("1504"), "0")) < cast.ToInt64(strings.TrimPrefix(killTime.Format("1504"), "0")) {
									fmt.Printf("当前时间：%s，预定抢购时间：%s。未到抢购时间，延迟1秒钟后重新检测\r",
										time.Now().Format("15:04:05"),
										killTime.Format("15:04"),
									)
									signal.RandSleep(time.Second)
								} else {
									fmt.Printf("当前时间：%s，预定抢购时间：%s。已到抢购时间，进入抢购。要不要起个战歌？\n",
										time.Now().Format("15:04:05"),
										killTime.Format("15:04"),
									)
									break
								}
							}
						} else {
							fmt.Println("跳过时间预定，直接提交")
							fmt.Println("")
							fmt.Println("")
						}

						fmt.Printf("########## 提交订单中【%s】 ###########\n", time.Now().Format("15:04:05"))
						signal.RandSleep(time.Duration(delayMs) * time.Millisecond)
						err = session.AddNewOrder()
						switch err {
						case nil:
							fmt.Println("抢购成功，请前往app付款！")
							// 下单五分钟后不通知
							for i := 0; i < 150; i++ {
								_ = session.PushSuccess()
								time.Sleep(time.Second * 2)
							}
							return
						case dd.TimeExpireErr:
							fmt.Printf("[%s] %s\n", reserveTime.SelectMsg, err)
							break OrderLoop
						case dd.ProdInfoErr:
							fmt.Println(err)
							continue cartLoop
						default:
							fmt.Println(err)
						}
					}
				}
			}
		}
	}
}
