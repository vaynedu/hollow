package hlark

import (
	"context"
	"fmt"
	"testing"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larksheets "github.com/larksuite/oapi-sdk-go/v3/service/sheets/v3"
)

func TestLarkSheet(t *testing.T) {

	// 本来想写一下获取飞书表格数据，结果时间浪费在版本问题上，v2和v3版本兼容，很多对不上，对开发者就是灾难
	// 没有现成的api的话，就自己封装个读取数据的api
	myAppID := "1"
	myAppSecret := "1"

	// 创建 Client
	client := lark.NewClient(myAppID, myAppSecret)
	// 创建请求对象
	req := larksheets.NewGetSpreadsheetSheetReqBuilder().
		SpreadsheetToken("11").
		SheetId(`22`).
		Build()

	// 发起请求
	resp, err := client.Sheets.V3.SpreadsheetSheet.Get(context.Background(), req)

	// 处理错误
	if err != nil {
		fmt.Println(err)
		return
	}

	// 服务端错误处理
	if !resp.Success() {
		fmt.Printf("logId: %s, error response: \n%s", resp.RequestId(), larkcore.Prettify(resp.CodeError))
		return
	}

	// 业务处理
	fmt.Println(larkcore.Prettify(resp))

}
