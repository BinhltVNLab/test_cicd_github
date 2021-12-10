package excel

import (
	"encoding/csv"
	"mime/multipart"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	cf "gitlab.vietnamlab.vn/micro_erp/frontend-api/configs"
)

func ReadExcelFile(file *multipart.FileHeader) ([][]string, string) {
	var records [][]string
	var errEx string
	extensionFile := filepath.Ext(strings.TrimSpace(file.Filename))
	if extensionFile != ".csv" && extensionFile != ".xlsx" && extensionFile != ".xls" {
		errEx = "Invalid file"
		return records, errEx
	}

	src, err := file.Open()
	if err != nil {
		errEx = "Invalid file"
		return records, errEx
	}
	defer src.Close()

	if extensionFile == ".csv" {
		records, err = csv.NewReader(src).ReadAll()
		if err != nil {
			errEx = "System Error"
			return records, errEx
		}
	} else {
		f, err := excelize.OpenReader(src)
		if err != nil {
			errEx = "System Error"
			return records, errEx
		}
		records, _ = f.GetRows("Sheet1")
	}

	return records, ""
}

func ConvertExcelDate(excelDate string) string {
	var excelEpoch = time.Date(1899, time.December, 30, 0, 0, 0, 0, time.UTC)
	var days, _ = strconv.Atoi(excelDate)
	return excelEpoch.Add(time.Second * time.Duration(days*86400)).Format(cf.FormatDateDatabase)
}
