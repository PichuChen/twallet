package twallet

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

type ExpireUnitType string

const (
	UnitTypeExpireDay   ExpireUnitType = "DAY"
	UnitTypeExpireMonth ExpireUnitType = "MONTH"
	UnitTypeExpireYear  ExpireUnitType = "YEAR"
)

type Field struct {
	Type                string `json:"type"`
	Cname               string `json:"cname"`
	Ename               string `json:"ename"`
	RegularExpressionId int    `json:"regularExpressionId"`
	CardCoverData       int    `json:"cardCoverData"`
}

// CreateVCItem will create a VC template item
// Parameters:
//
//	accessToken: 申請會員後會在信中得到的 access token
//	serialNo: VC 模板代碼，必須是唯一的，建議使用時間戳記
//	name: VC 模板名稱
//	expireNum: 有效期間數字部分，不能超過四位數字
//	expireUnit: 有效期間單位, 可選擇 DAY, MONTH, YEAR
//	expose: 對外顯示 true 的話可以開放其他人使用或是驗證
//	fields: 欄位資料，第一個欄位會是官方APP中卡片左下角的文字
//	cover: 封面圖片，傳 nil 代表不使用封面圖片
//		圖片規範下載: https://issuer-sandbox.wallet.gov.tw/assets/%E6%95%B8%E4%BD%8D%E6%86%91%E8%AD%89%E7%9A%AE%E5%A4%BE%EF%BC%BF%E5%8D%A1%E9%9D%A2%E4%B8%8A%E5%82%B3%E8%A6%8F%E7%AF%84.a093a088.pdf
//		圖片尺寸建議 320x200, 比例為 1.6:1 長度不大於 2048px, 大小介於 40kB ~ 500kB，檔案類行為 JPG 或 PNG
func CreateVCItem(accessToken, serialNo, name, expireNum string, expireUnit ExpireUnitType, expose bool, fields []Field, cover []byte) error {
	requestURL := "https://issuer-sandbox.wallet.gov.tw/api/vc-items"
	requestPayload := map[string]interface{}{
		"serialNo":           serialNo,
		"name":               name,
		"category":           4,
		"expose":             expose,
		"lengthExpire":       expireNum,
		"unitTypeExpire":     expireUnit,
		"vcItemFieldDTOList": fields,
	}

	if cover != nil {
		mimeType := http.DetectContentType(cover)
		base64EncodedCover := base64.StdEncoding.EncodeToString(cover)
		if mimeType == "image/jpeg" {
			requestPayload["cover"] = fmt.Sprintf("data:image/jpeg;base64,%s", base64EncodedCover)
		} else if mimeType == "image/png" {
			requestPayload["cover"] = fmt.Sprintf("data:image/png;base64,%s", base64EncodedCover)
		} else {
			slog.Warn("CreateVCItem", "detectedMimeType", mimeType)
		}
	}

	requestBody, err := json.Marshal(requestPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal request payload: %v", err)
	}

	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("access-token", accessToken)
	req.Header.Set("content-type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	var responseBody map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return fmt.Errorf("failed to decode response body: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		slog.Error("CreateVCItem", "status", resp.StatusCode, "responseBody", responseBody)
		return fmt.Errorf("unexpected response code: %v", responseBody["detail"])
	}
	slog.Debug("CreateVCItem", "Response", responseBody)
	return nil
}
