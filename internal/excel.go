package internal

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

func BuildMachineExcel(records []map[string]any) ([]byte, error) {
	excel := excelize.NewFile()
	sheet := "机具列表"
	_ = excel.SetSheetName(excel.GetSheetName(0), sheet)
	headers := []any{
		"机具号", "机具入库时间", "激活时间", "服务商编号", "服务商电话", "服务商姓名",
		"商户名称", "商户编号", "商户手机号", "机具类型", "机具标识", "达标奖励标识",
		"机具来源", "来源服务商编号", "来源服务商姓名", "扣费金额", "刷卡费率", "扫码费率",
		"手机pay费率", "单笔费率", "VIP扫码费率", "VIP扫码单笔", "京东白条", "费率政策编号",
		"首笔交易时间", "是否需要扣费", "是否完成扣费", "激活状态", "绑定状态", "是否二次绑定",
		"最近一次绑定时间", "冻结状态", "店铺编号", "渠道类型", "二级服务商编号", "二级服务商姓名",
	}
	_ = excel.SetSheetRow(sheet, "A1", &headers)
	headerStyle, _ := excel.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"D9E2F3"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "left"},
	})
	bodyStyle, _ := excel.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "left"},
	})
	_ = excel.SetCellStyle(sheet, "A1", "AJ1", headerStyle)
	_ = excel.SetRowHeight(sheet, 1, 22)
	_ = excel.SetColWidth(sheet, "A", "AJ", 24)

	keys := []string{
		"sn", "createTime", "activeTime", "psn", "pmobile", "pname",
		"businessName", "merchantNo", "businessMobile", "machineType", "termModeType", "standAwardFlagStr",
		"sourceDesc", "goldPartnerReferKey", "goldPartnerRelaName", "kfAmount", "shareConsumRate", "qrcodeConsumRate",
		"qpConsumRate", "businessConsumRate", "vipQrCodeConsumRate", "vipQrCodeConsumFee", "jdConsumRate", "showRateRemark",
		"firstTime", "kfFlag", "doneKfFlag", "activeStatus", "bindStatus", "twiceBindFlag",
		"lastBindTime", "freezeFlag", "merchantChannelCodeNo", "model", "secondReferkey", "secondPartnerName",
	}
	for row, record := range records {
		values := make([]any, len(keys))
		for column, key := range keys {
			value := record[key]
			if key == "termModeType" {
				switch fmt.Sprint(value) {
				case "1":
					value = "购机"
				case "2":
					value = "赠机"
				case "3":
					value = "兑换机具"
				case "4":
					value = "存量机具"
				}
			}
			values[column] = value
		}
		cell, err := excelize.CoordinatesToCellName(1, row+2)
		if err != nil {
			return nil, err
		}
		_ = excel.SetSheetRow(sheet, cell, &values)
	}
	if len(records) > 0 {
		_ = excel.SetCellStyle(sheet, "A2", fmt.Sprintf("AJ%d", len(records)+1), bodyStyle)
	}
	buffer, err := excel.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
