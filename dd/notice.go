package dd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

func (s *DingdongSession) PushSuccess() error {
	if s.BarkId == "" {
		fmt.Println("抢购成功，请前往app付款！\r")
		return nil
	}
	urlPath := fmt.Sprintf("https://api.day.app/%s/cookie:%s抢到菜了，请速去支付?sound=alarm", s.BarkId, s.Cookie)
	req, _ := http.NewRequest("GET", urlPath, nil)
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
		return nil
	} else {
		return errors.New(fmt.Sprintf("[%v] %s", resp.StatusCode, body))
	}
}
