package sams

import (
	"SAMS_buyer/conf"
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"time"
)

type CapCityResponse struct {
	StrDate        string `json:"strDate"`
	DeliveryDesc   string `json:"deliveryDesc"`
	DeliveryDescEn string `json:"deliveryDescEn"`
	DateISFull     bool   `json:"dateISFull"`
	List           []List `json:"list"`
}

type List struct {
	StartTime     string `json:"startTime"`
	EndTime       string `json:"endTime"`
	TimeISFull    bool   `json:"timeISFull"`
	Disabled      bool   `json:"disabled"`
	CloseDate     string `json:"closeDate"`
	CloseTime     string `json:"closeTime"`
	StartRealTime string `json:"startRealTime"` //1649984400000
	EndRealTime   string `json:"endRealTime"`   //1650016800000
}

type Capacity struct {
	Data                      string            `json:"data"`
	CapCityResponseList       []CapCityResponse `json:"capcityResponseList"`
	PortalPerformanceTemplate string            `json:"getPortalPerformanceTemplateResponse"`
}

func parseCapacity(result gjson.Result) (error, CapCityResponse) {
	var list []List
	for _, v := range result.Get("list").Array() {
		list = append(list, List{
			StartTime:     v.Get("startTime").Str,
			EndTime:       v.Get("endTime").Str,
			TimeISFull:    v.Get("timeISFull").Bool(),
			Disabled:      v.Get("disabled").Bool(),
			CloseDate:     v.Get("closeDate").Str,
			CloseTime:     v.Get("closeTime").Str,
			StartRealTime: v.Get("startRealTime").Str,
			EndRealTime:   v.Get("endRealTime").Str,
		})
	}
	capacity := CapCityResponse{
		StrDate:        result.Get("strDate").Str,
		DeliveryDesc:   result.Get("deliveryDesc").Str,
		DeliveryDescEn: result.Get("deliveryDescEn").Str,
		DateISFull:     result.Get("dateISFull").Bool(),
		List:           list,
	}
	return nil, capacity
}

func (session *Session) GetCapacity(result gjson.Result) error {
	var capCityResponseList []CapCityResponse
	for _, v := range result.Get("data.capcityResponseList").Array() {
		_, product := parseCapacity(v)
		capCityResponseList = append(capCityResponseList, product)
	}
	session.Capacity = Capacity{
		Data:                      result.String(),
		CapCityResponseList:       capCityResponseList,
		PortalPerformanceTemplate: result.Get("data.getPortalPerformanceTemplateResponse").Str,
	}
	return nil
}

func (session *Session) SetCapacity() error {
	session.SettleDeliveryInfo = SettleDeliveryInfo{}
	isSet := false
	for _, caps := range session.Capacity.CapCityResponseList {
		if isSet {
			break
		}
		for _, v := range caps.List {
			fmt.Printf("配送时间： %s %s - %s, 是否可用：%v\n", caps.StrDate, v.StartTime, v.EndTime, !v.TimeISFull && !v.Disabled)
			if v.TimeISFull == false && v.Disabled == false && session.SettleDeliveryInfo.ArrivalTimeStr == "" {
				session.SettleDeliveryInfo.ArrivalTimeStr = fmt.Sprintf("%s %s - %s", caps.StrDate, v.StartTime, v.EndTime)
				session.SettleDeliveryInfo.ExpectArrivalTime = v.StartRealTime
				session.SettleDeliveryInfo.ExpectArrivalEndTime = v.EndRealTime
				isSet = true
				break
			}
		}
	}
	if isSet {
		return nil
	}
	return conf.CapacityFullErr
}

func (session *Session) CheckCapacity() error {
	data := make(map[string]interface{})
	data["perDateList"] = []string{
		time.Now().Format("2006-01-02"),
		time.Now().AddDate(0, 0, 1).Format("2006-01-02"),
		time.Now().AddDate(0, 0, 2).Format("2006-01-02"),
		time.Now().AddDate(0, 0, 3).Format("2006-01-02"),
		time.Now().AddDate(0, 0, 4).Format("2006-01-02"),
		time.Now().AddDate(0, 0, 5).Format("2006-01-02"),
		time.Now().AddDate(0, 0, 6).Format("2006-01-02"),
	}
	data["storeDeliveryTemplateId"] = session.Cart.FloorInfoList[0].StoreInfo.StoreDeliveryTemplateId
	dataStr, _ := json.Marshal(data)
	err, result := session.Request.POST(CapacityDataAPI, dataStr)
	if err != nil {
		return nil
	}

	err = session.GetCapacity(result)
	if err != nil {
		return err
	}

	err = session.SetCapacity()
	if err != nil {
		return err
	}
	return nil
}