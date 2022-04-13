package dd

import "errors"

var TimeExpireErr = errors.New("送达时间已失效")
var OOSErr = errors.New("部分商品已缺货")
var BusyErr = errors.New("【3000】当前人多拥挤")
var RateLimit = errors.New("【3001】【限流】当前人多拥挤")
var NotStart = errors.New("抢购未开始")
var DataLoadErr = errors.New("【-3100】限流了：部分数据加载失败")
var ProdInfoErr = errors.New("商品信息有变化")
