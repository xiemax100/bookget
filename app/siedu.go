package app

import (
	"bookget/config"
	"bookget/model/iiif"
	"bookget/pkg/downloader"
	"bookget/pkg/gohttp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

type SiEdu struct {
	dt  *DownloadTask
	ctx context.Context
}

func NewSiEdu() *SiEdu {
	return &SiEdu{
		// 初始化字段
		dt:  new(DownloadTask),
		ctx: context.Background(),
	}
}

func (r *SiEdu) GetRouterInit(sUrl string) (map[string]interface{}, error) {
	msg, err := r.Run(sUrl)
	return map[string]interface{}{
		"type": "iiif",
		"url":  sUrl,
		"msg":  msg,
	}, err
}

func (r *SiEdu) Run(sUrl string) (msg string, err error) {

	r.dt.UrlParsed, err = url.Parse(sUrl)
	r.dt.Url = sUrl

	r.dt.BookId = r.getBookId(r.dt.Url)
	if r.dt.BookId == "" {
		return "requested URL was not found.", err
	}
	r.dt.Jar, _ = cookiejar.New(nil)
	return r.download()
}

func (r *SiEdu) getBookId(sUrl string) (bookId string) {
	//https://ids.si.edu/ids/manifest/FS-F1904.61_006
	m := regexp.MustCompile(`manifest/([A-z0-9._-]+)`).FindStringSubmatch(sUrl)
	if m != nil {
		return m[1]
	}
	if strings.Contains(sUrl, "/object/") {
		bs, err := r.getBody(sUrl, nil)
		if err != nil {
			return ""
		}
		//<div  data-idsid="FS-F1904.61_006" class="media-metadata">
		m = regexp.MustCompile(`data-idsid="([A-z0-9._-]+)"`).FindStringSubmatch(string(bs))
		if m != nil {
			bookId = m[1]
		}
		return bookId
	}
	return bookId
}

func (r *SiEdu) download() (msg string, err error) {
	log.Printf("Get %s\n", r.dt.Url)
	r.dt.SavePath = config.Conf.Directory
	apiUrl := "https://" + r.dt.UrlParsed.Host + "/ids/manifest/" + r.dt.BookId
	canvases, err := r.getCanvases(apiUrl, r.dt.Jar)
	if err != nil || canvases == nil {
		return "requested URL was not found.", err
	}
	log.Printf(" %d images \n", len(canvases))
	return r.do(canvases)
}

func (r *SiEdu) do(iiifUrls []string) (msg string, err error) {
	if iiifUrls == nil {
		return
	}
	referer := url.QueryEscape(r.dt.Url)

	args := []string{
		"-H", "Origin:" + referer,
		"-H", "Referer:" + referer,
	}
	size := len(iiifUrls)
	iiifDownloader := downloader.NewIIIFDownloader(&config.Conf)
	for i, uri := range iiifUrls {
		if uri == "" || !config.PageRange(i, size) {
			continue
		}
		sortId := fmt.Sprintf("%04d", i+1)
		filename := sortId + config.Conf.FileExt
		inputUri := filepath.Join(r.dt.SavePath, sortId+"_info.json")
		bs, err := r.getBody(uri, r.dt.Jar)
		if err != nil {
			continue
		}
		body := strings.Replace(string(bs), `"http://iiif.io/api/image/2/level2.json",`, "", -1)
		body = strings.Replace(body, `"sizeByH",`, "", -1)
		os.WriteFile(inputUri, []byte(body), os.ModePerm)
		dest := path.Join(r.dt.SavePath, filename)
		if FileExist(dest) {
			continue
		}
		log.Printf("Get %s  %s\n", sortId, uri)
		if err := iiifDownloader.Dezoomify(r.ctx, inputUri, dest, args); err == nil {
			os.Remove(inputUri)
		}
	}
	return "", nil
}

func (r *SiEdu) getCanvases(sUrl string, jar *cookiejar.Jar) (canvases []string, err error) {
	bs, err := r.getBody(sUrl, jar)
	var manifest = new(iiif.ManifestResponse)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(bs, manifest); err != nil {
		log.Printf("json.Unmarshal failed: %s\n", err)
		return
	}
	if len(manifest.Sequences) == 0 {
		return
	}
	size := len(manifest.Sequences[0].Canvases)
	canvases = make([]string, 0, size)
	for _, canvase := range manifest.Sequences[0].Canvases {
		for _, image := range canvase.Images {
			iiiInfo := fmt.Sprintf("%s/info.json", image.Resource.Service.Id)
			canvases = append(canvases, iiiInfo)
		}
	}
	return canvases, nil
}

func (r *SiEdu) getBody(apiUrl string, jar *cookiejar.Jar) ([]byte, error) {
	ctx := context.Background()
	cli := gohttp.NewClient(ctx, gohttp.Options{
		CookieFile: config.Conf.CookieFile,
		CookieJar:  jar,
		Headers: map[string]interface{}{
			"User-Agent": config.Conf.UserAgent,
			"Referer":    r.dt.Url,
			"authority":  "www.si.edu",
			"origin":     "https://www.si.edu/",
		},
	})
	resp, err := cli.Get(apiUrl)
	if err != nil {
		return nil, err
	}
	bs, _ := resp.GetBody()
	if resp.GetStatusCode() != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("ErrCode:%d, %s", resp.GetStatusCode(), resp.GetReasonPhrase()))
	}
	return bs, nil
}
