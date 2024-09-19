package helper

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/fatih/color"
)

type Upload struct {
	Path     string `json:"path"`
	Playlist string `json:"playlist"`
	Channel  string `json:"channel"`
	Type     string `json:"type"`
	Link     string `json:"link"`
}

var userDataDir = ""

func SetupContextChrome() (context.Context, context.CancelFunc) {
	opts := []chromedp.ExecAllocatorOption{}
	opts = append(opts,
		chromedp.Flag("headless", false),
		chromedp.UserDataDir(userDataDir),
		chromedp.Flag("remote-debugging-port", "9222"),
		chromedp.Flag("allow-running-insecure-content", true),
	)
	ctx, _ := context.WithTimeout(context.Background(), 300*time.Second)
	allocCtx, _ := chromedp.NewExecAllocator(ctx, opts...)
	return chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
}

func ListFilesInDirectory(directory string) ([]*Upload, error) {
	uploads := make([]*Upload, 0)
	// Walk through all files and directories in the specified directory
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Kiểm tra xem đó có phải là file không (không phải thư mục)
		if !info.IsDir() {
			up := &Upload{
				Path: path,
				Type: getFileType(path),
			}
			uploads = append(uploads, up) // Thêm đường dẫn file vào slice
		}
		return nil
	})

	return uploads, err
}

func getFileType(file string) string {
	ext := strings.ToLower(filepath.Ext(file))

	// So sánh với các phần mở rộng của video và text file
	switch ext {
	case ".mp4", ".avi", ".mkv", ".mov":
		return "video"
	case ".txt", ".md", ".log", ".json", ".xml":
		return "text"
	case ".jpg", ".jpeg", ".png", ".gif":
		return "image"
	default:
		return "unknown"
	}
}

func IsChromeBrowserVisible() bool {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("powershell", "-Command", `WMIC PROCESS WHERE "name='chrome.exe'" GET CommandLine | findstr "User Data\\youtube"`)
	case "darwin":
		cmd = exec.Command("osascript", "-e", `tell application "System Events" to count (every process whose name is "Google Chrome" and visible is true)`)
	default: // Linux
		cmd = exec.Command("bash", "-c", "xdotool search --onlyvisible --class chrome")
	}

	output, err := cmd.Output()
	if err != nil && err.Error() != "exit status 1" {
		color.Red("Error checking for visible Chrome windows: %v", err)
		return false
	}
	return len(strings.TrimSpace(string(output))) > 0
}

func init() {
	usr, err := user.Current()
	if err != nil {
		userDataDir = ""
		return
	}
	userDataDir = fmt.Sprintf("%s\\AppData\\Local\\Google\\Chrome\\User Data\\youtube", usr.HomeDir)
	fmt.Printf("Thiết lập chrome profile cho quá trình tải lên video: %s \n", color.MagentaString(userDataDir))
}
