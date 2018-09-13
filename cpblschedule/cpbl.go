package cpblschedule

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Match struct {
	Date       string `json:"date"`
	Time       string `json:"time"`
	Ended      bool   `json:"ended"`
	Home       string `json:"home"`
	Away       string `json:"away"`
	Home_score string `json:"home_score"`
	Away_score string `json:"away_score"`
	Reason     string `json:"reason"`
	Game_type  string `json:"game_type"`
	No         string `json:"game_no"`
	Location   string `json:"loc"`
}

func parseTeam(str string) string {
	team_map := map[string]string{
		"A": "Lamigo桃猿",
		"B": "義大犀牛",
		"E": "中信兄弟",
		"L": "統一獅",
	}

	regex := regexp.MustCompile(".*\\/(.+)\\d\\d_logo_01\\.png")

	result := regex.FindStringSubmatch(str)[1]
	return team_map[result]
}

func ParseCPBLSchedule(year int, month int) ([]Match, error) {
	var days []string = make([]string, 7)
	url_patten := "http://www.cpbl.com.tw/schedule/index/%4d-%02d-01.html?&date=%4d-%02d-01&gameno=01&sfieldsub=&sgameno=01"
	url := fmt.Sprintf(url_patten, year, month, year, month)
	doc, err := goquery.NewDocument(url)

	if err != nil {
		return nil, err
	}

	var matches []Match = make([]Match, 0)
	doc.Find("table.schedule").Each(func(i int, schedule_table *goquery.Selection) {
		schedule_table.Children().ChildrenFiltered("tr").Each(func(i int, tr *goquery.Selection) {
			attr_val, existed_attr := tr.Attr("class")

			if existed_attr && attr_val == "day" {
				//Skip day header
			} else {
				td := tr.ChildrenFiltered("td")
				td_size := td.Size()

				if td_size > 0 {
					td.Each(func(i int, tds *goquery.Selection) {
						one_blocks := tds.Find(".one_block")
						if one_blocks.Size() == 0 {
							//no match today
						} else {
							day := days[i]
							one_blocks.Each(func(i int, block *goquery.Selection) {
								var match Match
								sch_tds := block.Find(".schedule_team").Find("td")
								match.Away = parseTeam(sch_tds.Nodes[0].FirstChild.Attr[0].Val)
								sch_tds = sch_tds.Next()
								match.Location = sch_tds.Text()
								sch_tds = sch_tds.Next()
								match.Home = parseTeam(sch_tds.Nodes[0].FirstChild.Attr[0].Val)

								sch_info := block.Find(".schedule_info").First()
								//Parse game type & no
								match.Game_type = strings.TrimSpace(sch_info.Find("th").Eq(0).Text())

								if match.Game_type != "補賽" {
									match.Game_type = "normal" //Normal
								} else {
									match.Game_type = "reschduled" //Rescheduled game
								}

								match.No = strings.TrimSpace(sch_info.Find("th").Eq(1).Text())

								sch_info = sch_info.Next()

								//Parse scores or time
								schedule_scores := sch_info.Find(".schedule_score")

								match.Ended = false
								if schedule_scores.Size() > 0 {
									//Has score
									match.Ended = true
									match.Away_score = schedule_scores.Eq(0).Text()
									match.Home_score = schedule_scores.Eq(1).Text()
								} else {
									sp_text := block.Find(".schedule_sp_txt")

									if sp_text.Size() == 0 {
										sch_info = sch_info.Next() //wtf
										match.Time = sch_info.Find("td").Eq(1).Text()
									} else {
										match.Game_type = "postponed" //Postponded or rain out
										match.Reason = sp_text.Text()
									}
								}

								match.Date = fmt.Sprintf("%02d/%s", month, day)
								matches = append(matches, match)
							})
						}
					})
				} else {
					th := tr.ChildrenFiltered("th")
					ds := th.Map(func(i int, ths *goquery.Selection) string {
						return ths.Text()
					})

					copy(days, ds)
				}
			}

		})
	})

	return matches, nil
}
