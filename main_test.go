package twallet_test

import (
	"fmt"
	"time"

	"github.com/PichuChen/twallet"
)

// 請在這邊填入你的 access token
var accessToken = "{{access_token}}"

func ExampleCreateVCItem() {
	// VC 模板代碼
	serialNo := fmt.Sprintf("t_%v", time.Now().Unix())
	// VC 模板名稱
	name := "我們家的會員卡"
	// 有效期間數字部分，不能超過四位數字
	expireNum := "1"
	// 有效期間單位, 可選擇 DAY, MONTH, YEAR
	expireUnit := twallet.UnitTypeExpireMonth
	// 對外顯示 true 的話可以開放其他人使用或是驗證
	expose := false
	// 欄位資料，第一個欄位會是官方APP中卡片左下角的文字
	fields := []twallet.Field{
		{
			Type:  "BASIC",
			Cname: "姓名",
			Ename: "name",
			// 4: 只允許輸入英文
			// 5: 只允許輸入英數
			// 6: 電子郵件格式檢查
			// 7: 台灣行動電話號碼規則
			// 8: 不允許輸入 ~!@#$%^&*()_-+*/
			// 9: 只能輸入中英文數字和_
			// 10: 只能輸入英文數字和_
			// 11: 台灣身分證字號
			// 12: 民國出生年月日(0991231)
			// 14: 居留證統一證號
			// 15: 外來人口統一證號
			// 16: 性別格式檢查
			// 17: 國籍格式檢查
			// 18: 西元出生日期
			// 19: 郵遞區號
			// 20: 護照號碼
			// 22: 只能輸入中文
			// 23: 只能輸入3碼數字
			// 101: 網址格式檢查
			RegularExpressionId: 9,
			CardCoverData:       1,
		},
	}
	var cover []byte = nil // no cover image

	// Create a VC item
	err := twallet.CreateVCItem(accessToken, serialNo, name, expireNum, expireUnit, expose, fields, cover)
	if err != nil {
		panic(err)
	}

	// Output:
}
