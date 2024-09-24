package main

import (
	"bytes"
	"flag"
	"fmt"
	"time"

	"example/helper"

	"github.com/chromedp/chromedp"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
)

// Selectors

var (
	path     = flag.String("path", "", "Là đường dẫn tới file muốn upload")
	channel  = flag.String("channel", "", "Là ID của kênh muốn upload")
	playlist = flag.String("playlist", "", "Danh sách phát muốn thêm vào")
)

var (
	baseUri             = "https://studio.youtube.com/channel/%s/videos/upload"
	ChannelName *string = new(string)
)

func main() {
	flag.Parse()

	if *path == "" {
		color.Red("Bạn cần phải chỉ định đường dãn thư mục hoặc dường dẫn file upload!")
		return
	}

	if *channel == "" {
		color.Red("Bạn cần phải chỉ định channel cần upload file!")
		return
	}

	fmt.Println("Tiến hành tải video trong", color.YellowString("path:"), color.GreenString(*path), "lên", color.YellowString("channel ID:"), color.GreenString(*channel))

	uploads, err := helper.ListFilesInDirectory(*path)
	if err != nil {
		color.Red("Đã có lỗi sảy ra khi lấy danh sách file theo đường dẫn %s \n", *path)
		return
	}

	if len(uploads) == 0 {
		color.Red("Vui lòng kiểm tra lại!. Không tìm thấy video nào trong đường dẫn %s \n", *path)
		return
	}

	// Check for open Chrome windows
	if helper.IsChromeBrowserVisible() {
		color.Red("Vui lòng tắt cửa sổ trình duyệt trước khi chạy ứng dụng")
		return
	}
	ctx, cancel := helper.SetupContextChrome(len(uploads))
	defer cancel()
	tasks := chromedp.Tasks{
		chromedp.Navigate(fmt.Sprintf(baseUri, *channel)),
		helper.ActionGetChannelName(ChannelName),
	}
	for _, up := range uploads {
		if up.Type != "video" {
			continue
		}
		up.Playlist = *playlist
		up.Channel = ChannelName

		task := chromedp.Tasks{
			chromedp.Sleep(1 * time.Second),
			helper.CommandUpload(up, ctx),
			chromedp.Sleep(1 * time.Second),
		}
		tasks = append(tasks, task)
	}

	if err := chromedp.Run(ctx, tasks); err != nil {
		color.Red("Lỗi khi chạy cửa sổ trình duyệt: %v", err)
	}

	// Keep the browser open
	color.Green("Đã hoàn thành tải video trong đường dẫn %s lên channelID: %s \n", *path, *channel)
	printf(uploads)
	helper.UploadSheet(uploads)
	color.Green("kết thúc!")
}

func printf(uploads []*helper.Upload) string {
	buf := new(bytes.Buffer)
	table := tablewriter.NewWriter(buf)
	table.SetRowLine(true)
	table.SetHeader([]string{"Đường dẫn", "kiểu dữ liệu", "Link youtube"})
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetRowSeparator("-")
	table.SetAlignment(tablewriter.ALIGN_CENTER)
	table.SetHeaderAlignment(tablewriter.ALIGN_CENTER)
	table.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_CENTER, tablewriter.ALIGN_LEFT})
	table.SetHeaderColor(tablewriter.Colors{tablewriter.Bold, tablewriter.BgGreenColor},
		tablewriter.Colors{tablewriter.FgHiRedColor, tablewriter.Bold, tablewriter.BgBlackColor},
		tablewriter.Colors{tablewriter.BgCyanColor, tablewriter.FgWhiteColor})

	for _, u := range uploads {
		table.Append([]string{
			u.Path,
			u.Type,
			u.Link,
		})
	}

	table.Render()
	fmt.Println(buf.String())
	return buf.String()
}
