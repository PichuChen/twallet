package twallet

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
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

type VCItemDataField struct {
	Ename   string `json:"ename"`
	Content string `json:"content"`
}

type VCItemDataResponse struct {
	ID                    int    `json:"id"`
	BusinessId            string `json:"businessId"`
	Content               string `json:"content"`
	CrDatetime            string `json:"crDatetime"`
	CrUser                int    `json:"crUser"`
	DeepLink              string `json:"deepLink"`
	Expired               string `json:"expired"`
	PureContent           string `json:"pureContent"`
	QrCode                string `json:"qrCode"`
	ScheduleRevokeMessage string `json:"scheduleRevokeMessage"`
	TransactionId         string `json:"transactionId"`
	Valid                 int    `json:"valid"`
	VcCid                 string `json:"vcCid"`
	VcItemName            string `json:"vcItemName"`
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

// CreateVCItemData will create a VC item data (取得卡片用的 QR Code)
// Parameters:
//
//	accessToken: 申請會員後會在信中得到的 access token
//	vcID: VC 序號 (不是模板代碼)
//	fields: 欄位資料
//		completion: 如果想知道使用者是否已經掃描過 QR Code，可以傳入這個參數，
//		如果傳入 nil，則不會有任何回傳
//		如果傳入的參數不為 nil，則會在掃描後回傳 vcCid
func CreateVCItemData(accessToken string, vcID int, fields []VCItemDataField, completion func(vcCid string)) (*VCItemDataResponse, error) {

	requestURL := "https://issuer-sandbox.wallet.gov.tw/api/vc-item-data"
	requestPayload := map[string]interface{}{
		"vcId":   vcID,
		"fields": fields,
	}
	requestBody, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request payload: %v", err)
	}
	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("access-token", accessToken)
	req.Header.Set("content-type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()
	var responseBody []byte
	responseBody, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		slog.Error("CreateVCItemData", "status", resp.StatusCode, "responseBody", response)
		return nil, fmt.Errorf("unexpected response code: %v", response["detail"])
	}
	slog.Debug("CreateVCItemData", "Response", responseBody)

	var vcItemDataResponse VCItemDataResponse
	if err := json.Unmarshal(responseBody, &vcItemDataResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %v", err)
	}
	if completion != nil {
		timeout := 300 * time.Second
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		done := make(chan bool)
		go func() {
			for {
				select {
				case <-done:
					return
				case <-ticker.C:
					// Check if the timeout has been reached
					if timeout <= 0 {
						slog.Error("CreateVCItemData", "timeout", timeout)
						return
					}
					timeout -= 5 * time.Second
					// Call GetVCItemData to check if the vcCid is available
					// If the vcCid is available, call the completion function
					// and break the loop
					vcCid, err := GetVCItemData(accessToken, vcItemDataResponse.ID)
					if err != nil {
						slog.Error("GetVCItemData", "error", err)
						done <- true
						return
					}
					if vcCid != "" {
						completion(vcCid)
						slog.Info("CreateVCItemData", "vcCid", vcCid)
						done <- true
						return
					}
				}
			}
		}()
	}

	return &vcItemDataResponse, nil

}

// GetVCItemData will get the VC item data (取得卡片的vcCid)
// Parameters:
//
//	accessToken: 申請會員後會在信中得到的 access token
//	id: VC 序號 (不是模板代碼)
func GetVCItemData(accessToken string, id int) (string, error) {
	requestURL := fmt.Sprintf("https://issuer-sandbox.wallet.gov.tw/api/vc-item-data/%d", id)
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("access-token", accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	var responseBody map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return "", fmt.Errorf("failed to decode response body: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		slog.Error("GetVCItemData", "status", resp.StatusCode, "responseBody", responseBody)
		return "", fmt.Errorf("unexpected response code: %v", responseBody["detail"])
	}
	slog.Debug("GetVCItemData", "Response", responseBody)
	if vccid := responseBody["vcCid"]; vccid == nil {
		return "", nil
	}

	return responseBody["vcCid"].(string), nil
}
