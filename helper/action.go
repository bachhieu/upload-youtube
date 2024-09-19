package helper

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/fatih/color"
)

const sleep500ms = 500 * time.Millisecond

func CommandUpload(u *Upload, ctx context.Context) chromedp.Tasks {
	return chromedp.Tasks{
		actionCreateUpload(),
		actionUploadFile(u.Path),
		actionChoiceNotChildren(),
		actionChoicePlaylist(u.Playlist),
		chromedp.Sleep(3 * sleep500ms),
		actionNext(),
		actionNext(),
		actionNext(),
		actionChoiceDisplayModeUnlisted(),
		actionCopyLink(&u.Link),
		actionSave(),
		actionClose(),
	}
}

func actionCreateUpload() chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.WaitVisible("#create-icon > ytcp-button-shape", chromedp.ByQuery),
		chromedp.Click("#create-icon > ytcp-button-shape", chromedp.ByQuery),
		chromedp.Sleep(400 * time.Millisecond),
		chromedp.Click("#text-item-0", chromedp.ByQuery),
	}
}

func actionUploadFile(filePath string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.ActionFunc(func(ctx context.Context) error {
			color.Green("Tiến hành tải lên file %s", filePath)
			return nil
		}),
		chromedp.SetUploadFiles(`input[type="file"]`, []string{filePath}),
		chromedp.EvaluateAsDevTools(`document.querySelector('input[type="file"]').style.removeProperty('display');`, nil),
	}
}

func actionNext() chromedp.Tasks {
	key := "#next-button button"
	return actionClick(key)
}

func actionChoiceNotChildren() chromedp.Tasks {
	key := "#audience tp-yt-paper-radio-button:nth-child(2)"
	return actionClick(key)
}

func actionChoicePlaylist(name string) chromedp.Tasks {
	if name == "" {
		return chromedp.Tasks{
			chromedp.Sleep(sleep500ms),
		}
	}
	key := "#items > ytcp-ve > li > label"
	return chromedp.Tasks{
		chromedp.Sleep(sleep500ms),
		chromedp.WaitVisible(key),
		chromedp.Evaluate(fmt.Sprintf(`
				const checkboxes = document.querySelectorAll('%s');
				checkboxes.forEach(checkbox => {
					const name = checkbox.closest('label') || document.querySelector('label[for="' + checkbox.id + '"]');
					if (label && label.textContent.trim() === "%s") {
						if (!checkbox.checked) {
							checkbox.click();
						}
					}
				});
			`, key, name), nil),
	}
}

func actionChoiceDisplayModeUnlisted() chromedp.Tasks {
	key := "#privacy-radios > tp-yt-paper-radio-button:nth-child(13)"
	return actionClick(key)
}

func actionSave() chromedp.Tasks {
	key := "#done-button > ytcp-button-shape > button"
	return actionClick(key)
}

func actionCopyLink(linkText *string) chromedp.Tasks {
	key := "#details > ytcp-video-metadata-editor-sidepanel > ytcp-video-info > div > div.row.style-scope.ytcp-video-info > div.left.style-scope.ytcp-video-info > div.value.style-scope.ytcp-video-info > span > a"
	key = "#scrollable-content > ytcp-uploads-review > div.right-col.style-scope.ytcp-uploads-review > ytcp-video-info > div > div.row.style-scope.ytcp-video-info > div.left.style-scope.ytcp-video-info > div.value.style-scope.ytcp-video-info > span > a"
	return chromedp.Tasks{
		chromedp.Sleep(4 * sleep500ms),
		chromedp.WaitVisible(key, chromedp.ByQuery),
		chromedp.Text(key, linkText, chromedp.ByID),
	}
}

func actionClose() chromedp.Tasks {
	key := "#close-button > ytcp-button-shape > button > yt-touch-feedback-shape > div"
	return actionClick(key)
}

func actionClick(key string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Sleep(sleep500ms),
		chromedp.WaitVisible(key, chromedp.ByQuery),
		chromedp.Click(key, chromedp.ByQuery),
	}
}
