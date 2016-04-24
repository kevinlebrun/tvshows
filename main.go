package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Show struct {
	Name     string
	Episodes []Episode
}

type Episode struct {
	Name   string
	Season int64
	Num    int64
	Aired  bool
}

type Catalog struct {
	Client *http.Client
}

func NewCatalog() *Catalog {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}
	return &Catalog{client}
}

func (c *Catalog) Auth(username, password string) error {
	form := make(url.Values)
	form.Add("username", "lebrun.k@gmail.com")
	form.Add("password", "vinke")
	form.Add("sub_login", "Account Login")

	data := strings.NewReader(form.Encode())

	req, err := http.NewRequest("POST", "http://www.pogdesign.co.uk/cat/", data)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")

	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}

	resp.Body.Close()

	return nil
}

func (c *Catalog) Followed() ([]Show, error) {
	req, err := http.NewRequest("GET", "http://www.pogdesign.co.uk/cat/profile/all-shows", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	shows := make([]Show, 0)

	doc.Find("a.prfimg.prfmed").Each(func(i int, s *goquery.Selection) {
		s.Find("span > strong").Remove()
		show := Show{
			Name: strings.Trim(s.Find("span").Text(), " \n\t"),
		}
		shows = append(shows, show)
	})

	return shows, nil
}

func (c *Catalog) Unwatched() ([]Show, error) {
	req, err := http.NewRequest("GET", "http://www.pogdesign.co.uk/cat/profile/unwatched-episodes", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	shows := make([]Show, 0)

	doc.Find("a.prfimg.prfmed").Each(func(i int, s *goquery.Selection) {
		if url, exists := s.Attr("href"); exists {
			episodes, err := c.UnwatchedEpisodesByURL(url)
			if err != nil {
				panic(err)
			}

			show := Show{
				Name:     strings.Trim(s.Find("span").Text(), " \n\t"),
				Episodes: episodes,
			}
			shows = append(shows, show)
		}
	})

	return shows, nil
}

func (c *Catalog) UnwatchedEpisodesByURL(url string) ([]Episode, error) {
	req, err := http.NewRequest("GET", "http://www.pogdesign.co.uk"+url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	episodes := make([]Episode, 0)

	doc.Find(".ep.info").Each(func(i int, s *goquery.Selection) {
		num, _ := strconv.ParseInt(s.Find(".pnumber").Text(), 10, 64)
		season, _ := strconv.ParseInt(s.PrevAllFiltered("h2.xxla").Eq(0).AttrOr("id", ""), 10, 64)

		name := s.Clone()
		name.Find("span").Remove()
		name.Find("label").Remove()

		episode := Episode{
			Name:   strings.Trim(name.Text(), " \n\t"),
			Num:    num,
			Season: season,
			Aired:  s.Children().Eq(1).Text() == "AIRED",
		}
		episodes = append(episodes, episode)
	})

	return episodes, nil
}

func main() {
	var err error
	var shows []Show

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s <command>\n", os.Args[0])
		flag.PrintDefaults()
	}

	var (
		username = flag.String("username", "", "www.pogdesign.co.uk/cat username")
		password = flag.String("password", "", "www.pogdesign.co.uk/cat password")
	)

	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	command := flag.Arg(0)

	catalog := NewCatalog()

	err = catalog.Auth(*username, *password)
	if err != nil {
		panic(err)
	}

	switch command {
	case "followed":
		shows, err = catalog.Followed()
		if err != nil {
			panic(err)
		}

		for _, show := range shows {
			fmt.Println(show.Name)
		}

	case "unwatched":
		shows, err = catalog.Unwatched()
		if err != nil {
			panic(err)
		}

		for _, show := range shows {
			for _, episode := range show.Episodes {
				if episode.Aired {
					fmt.Printf("%s s%02d e%02d [%s]\n", show.Name, episode.Season, episode.Num, episode.Name)
				}
			}
		}
	default:
		fmt.Printf("Unknown command %q\n", command)
		os.Exit(1)
	}
}
