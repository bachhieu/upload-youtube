package helper

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

const (
	// spreadsheetID = "1ahCV5SJNVXsWR4MIPH3xlIaCb_iLZygw0ODpN08iHNA" // Thay thế bằng Spreadsheet ID của bạn
	spreadsheetID = "1DrzwIA04XJ4alIUjZs1T-T8Nxww2vhoImo_yZVxBEqk" // Thay thế bằng Spreadsheet ID của bạn
)

var (
	baseField = []interface{}{"Link Folder", "Channel", "Danh sách phát", "Link Youtube", "Date"}
	sheetName = time.Now().Format("01/2006")
	sheetID   = int64(0)
)

func UploadSheet(ups []*Upload) error {
	color.Yellow("Tiến hành quá trình cập nhật link youtube lên sheet")

	if len(ups) == 0 {
		color.Red("Không tìm thấy danh sách cần tải lên sheet")
		return fmt.Errorf("Không tìm thấy danh sách cần tải lên sheet")
	}

	// Đường dẫn tới file JSON chứa Service Account Key
	serviceAccountFile := "key.json"

	// Đọc file JSON của Service Account
	data, err := os.ReadFile(serviceAccountFile)
	if err != nil {
		color.Red("Đã có lỗi khi đọc file %s - Error: %v", serviceAccountFile, err)
	}

	// Xác thực với Service Account
	conf, err := google.JWTConfigFromJSON(data, sheets.SpreadsheetsScope)
	if err != nil {
		color.Red("Đã có lỗi khi tạo cấu hình từ file: %s - Error: %v", serviceAccountFile, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Tạo service cho Google Sheets API
	srv, err := sheets.NewService(ctx, option.WithHTTPClient(conf.Client(ctx)))
	if err != nil {
		color.Red("Không thể tạo được service cho google sheet: %v", err)
	}

	res, err := srv.Spreadsheets.Get(spreadsheetID).Do()
	if err != nil {
		color.Red("Đã có lỗi xảy ra trong quá trình lấy thông tin của spreadsheet: %s - Error: %v", spreadsheetID, err)
	}

	// Kiểm tra xem sheet đã tồn tại chưa
	sheetExists := false
	for _, sheet := range res.Sheets {
		if sheet.Properties.Title == sheetName {
			sheetExists = true
			sheetID = sheet.Properties.SheetId
			break
		}
	}
	if !sheetExists {
		color.Yellow("Không tồn tại sheet %s trong spreadsheet: %s", sheetName, spreadsheetID)
		createNewSheet(srv)
	}

	if err := addValues(srv, ups); err != nil {
		return err
	}
	color.Yellow("https://docs.google.com/spreadsheets/d/%s/?gid=%v#gid=%v", spreadsheetID, sheetID, sheetID)
	color.Yellow("Hoàn thành cập nhật thông tin mới lên sheet")
	return nil
}

func createNewSheet(srv *sheets.Service) error {
	rbr := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{{
			AddSheet: &sheets.AddSheetRequest{
				Properties: &sheets.SheetProperties{
					Title: sheetName,
				},
			},
		}},
	}

	sheet, err := srv.Spreadsheets.BatchUpdate(spreadsheetID, rbr).Context(context.Background()).Do()
	if err != nil {
		color.Red("Không thể tạo sheet mới: %s %v", sheetName, err)
		return err
	}

	sheetID = sheet.Replies[0].AddSheet.Properties.SheetId

	color.Green("Đã tạo sheet mới '%s'.\n", sheetName)
	return addBaseFields(srv)
}

func addBaseFields(srv *sheets.Service) error {
	valueRange := &sheets.ValueRange{
		Values: [][]interface{}{
			baseField,
		},
	}

	_, err := srv.Spreadsheets.Values.Update(spreadsheetID, fmt.Sprintf("'%s'!A1", sheetName), valueRange).
		ValueInputOption("RAW").Do()
	if err != nil {
		color.Red("Không thể tải các trường cơ bản lên sheet: %s - Error %v", sheetName, err)
		return err
	}

	color.Green("Đã thêm các trường dữ liệu thành công.")
	return nil
}

func addValues(srv *sheets.Service, ups []*Upload) error {
	// Đọc dữ liệu từ Google Sheets
	readRange := fmt.Sprintf("%s!A:A", sheetName)
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		color.Red("Đã có lỗi trong qua trình lấy dữ liệu trong sheet: %s - Error %v", sheetName, err)
		return err
	}

	lastRow := len(resp.Values)
	if lastRow == 0 {
		color.Red("Không tìm được dữ liệu trong sheet %s", sheetName)
		if err := addBaseFields(srv); err != nil {
			return err
		}
		lastRow += 1

	}

	// Ghi dữ liệu vào Google Sheets
	writeRange := fmt.Sprintf("%s!A%v", sheetName, lastRow+1)
	values := [][]interface{}{}

	for _, up := range ups {
		if up.Type == "video" {
			value := []interface{}{up.Path, up.Channel, up.Playlist, up.Link, up.Date}
			values = append(values, value)
		}
	}

	valueRange := &sheets.ValueRange{
		Values: values,
	}

	_, err = srv.Spreadsheets.Values.Update(spreadsheetID, writeRange, valueRange).ValueInputOption("RAW").Do()
	if err != nil {
		color.Red("Không thể tải thông tin lên sheet: %s - Error %v", sheetName, err)
		return err
	}
	return nil
}
