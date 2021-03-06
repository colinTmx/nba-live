package main

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	figure "github.com/common-nighthawk/go-figure"
	tui "github.com/marcusolsson/tui-go"
)

var (
	updateChan = make(chan *GameInfoAndLive)
)

func newUI(game *Game) {
	grid := tui.NewGrid(0, 0)

	homeScore, visitScore := drawScore(game.HomeScore.(string), game.VisitScore.(string))
	homeLabel := tui.NewLabel(homeScore)
	visitLabel := tui.NewLabel(visitScore)
	homeLabel.SetStyleName("score")
	visitLabel.SetStyleName("score")

	grid.AppendRow(tui.NewHBox(homeLabel, visitLabel))

	scoreBox := tui.NewVBox(grid)
	scoreBox.SetTitle(fmt.Sprintf("%s VS %s", game.HomeTeam, game.VisitTeam))
	scoreBox.SetBorder(true)

	liveLabel := tui.NewLabel("")
	liveBox := tui.NewVBox(tui.NewScrollArea(liveLabel))
	liveBox.SetTitle(liveTitle(game.PeriodCn))
	liveBox.SetBorder(true)
	liveBox.SetSizePolicy(tui.Preferred, tui.Expanding)

	theme := tui.NewTheme()
	theme.SetStyle("label.score", tui.Style{Fg: tui.ColorCyan, Bold: true})

	root := tui.NewVBox(scoreBox, liveBox)
	ui := tui.New(root)
	ui.SetTheme(theme)
	ui.SetKeybinding("Esc", func() { ui.Quit() })
	ui.SetKeybinding("q", func() { ui.Quit() })
	ui.SetKeybinding("Ctrl+C", func() { ui.Quit() })
	ui.SetKeybinding("Ctrl+D", func() { ui.Quit() })

	go update(ui, homeLabel, visitLabel, liveLabel, liveBox)
	go fetch(game.ID)
	if err := ui.Run(); err != nil {
		panic(err)
	}
}

func update(ui tui.UI, homeLabel, visitLabel, liveLabel *tui.Label, liveBox *tui.Box) {
	defer close(updateChan)
	for {
		select {
		case data := <-updateChan:
			info, records := data.Info, data.LiveRecords
			ui.Update(func() {
				if info != nil {
					homeScore, visitScore := drawScore(info.HomeScore, info.VisitScore)
					homeLabel.SetText(homeScore)
					visitLabel.SetText(visitScore)
					liveBox.SetTitle(liveTitle(info.PeriodCn))
				}
				if records != nil {
					for _, record := range records {
						var liveTime string
						if lt, err := time.Parse("2006-01-02 15:04:05", record.LiveTime); err == nil {
							liveTime = lt.Format("15:04")
						}
						text := fmt.Sprintf("%s(%s): %s", record.UserChn, liveTime, strings.TrimSpace(record.LiveText))
						liveLabel.SetText(text + "\n" + liveLabel.Text())
					}
				}
			})
		}
	}
}

func fetch(gameId string) {
	lastMaxSid := 0
	for {
		maxSid, err := getMaxsid(gameId)
		if err != nil {
			//log.Println("get max sid error:", err.Error())
			time.Sleep(2 * time.Second)
			continue
		}
		if maxSid > lastMaxSid {
			gameInfo, _ := getGameInfo(gameId)
			//if err != nil {
			//log.Println("get game info error:", err.Error())
			//}
			liveRecords, _ := getLiveRecord(gameId, maxSid)
			//if err != nil {
			//log.Println("get game live records error:", err.Error())
			//}
			lastMaxSid = maxSid
			updateChan <- &GameInfoAndLive{
				Info:        gameInfo,
				LiveRecords: liveRecords,
			}
		}
		time.Sleep(1 * time.Second)
	}
}

func drawScore(homeScore, visitScore string) (string, string) {
	homeBuf := new(bytes.Buffer)
	figure.Write(homeBuf, figure.NewFigure(homeScore, "slant", true))
	visitBuf := new(bytes.Buffer)
	figure.Write(visitBuf, figure.NewFigure(visitScore, "slant", true))

	return homeBuf.String(), visitBuf.String()
}

func liveTitle(periodCn string) string {
	return fmt.Sprintf("直播 (%s)", periodCn)
}
