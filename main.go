package main

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"

	ap "github.com/Sorrow446/go-atomicparsley"
	"github.com/alexflint/go-arg"
	"github.com/grafov/m3u8"
)

const (
	regexString = `^https://www.beatport.com/release/[a-z0-9-]+/(\d+)$`
	userAgent   = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit" +
		"/537.36 (KHTML, like Gecko) Chrome/99.0.4844.82 Safari/537.36"
	baseUrl       = "https://www.beatport.com/"
	apiBase       = baseUrl + "api/v4/"
	trackTemplate = "{{.trackPad}}. {{.title}}"
	albumTemplate = "{{.albumArtist}} - {{.album}}"
)

var (
	jar, _ = cookiejar.New(nil)
	client = &http.Client{Transport: &Transport{}, Jar: jar}
)

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add(
		"User-Agent", userAgent,
	)
	return http.DefaultTransport.RoundTrip(req)
}

func getTempPath() (string, error) {
	return os.MkdirTemp(os.TempDir(), "")
}

func handleErr(errText string, err error, _panic bool) {
	errString := errText + "\n" + err.Error()
	if _panic {
		panic(errString)
	}
	fmt.Println(errString)
}

func wasRunFromSrc() bool {
	buildPath := filepath.Join(os.TempDir(), "go-build")
	return strings.HasPrefix(os.Args[0], buildPath)
}

func getScriptDir() (string, error) {
	var (
		ok    bool
		err   error
		fname string
	)
	runFromSrc := wasRunFromSrc()
	if runFromSrc {
		_, fname, _, ok = runtime.Caller(0)
		if !ok {
			return "", errors.New("Failed to get script filename.")
		}
	} else {
		fname, err = os.Executable()
		if err != nil {
			return "", err
		}
	}
	return filepath.Dir(fname), nil
}

func readTxtFile(path string) ([]string, error) {
	var lines []string
	f, err := os.OpenFile(path, os.O_RDONLY, 0755)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}
	if scanner.Err() != nil {
		return nil, scanner.Err()
	}
	return lines, nil
}

func contains(lines []string, value string) bool {
	for _, line := range lines {
		if strings.EqualFold(line, value) {
			return true
		}
	}
	return false
}

func processUrls(urls []string) ([]string, error) {
	var (
		processed []string
		txtPaths  []string
	)
	for _, _url := range urls {
		if strings.HasSuffix(_url, ".txt") && !contains(txtPaths, _url) {
			txtLines, err := readTxtFile(_url)
			if err != nil {
				return nil, err
			}
			for _, txtLine := range txtLines {
				if !contains(processed, txtLine) {
					processed = append(processed, txtLine)
				}
			}
			txtPaths = append(txtPaths, _url)
		} else {
			if !contains(processed, _url) {
				processed = append(processed, _url)
			}
		}
	}
	return processed, nil
}

func readConfig() (*Config, error) {
	data, err := ioutil.ReadFile("config.json")
	if err != nil {
		return nil, err
	}
	var obj Config
	err = json.Unmarshal(data, &obj)
	if err != nil {
		return nil, err
	}
	return &obj, nil
}

func parseArgs() *Args {
	var args Args
	arg.MustParse(&args)
	return &args
}

func parseCfg() (*Config, error) {
	cfg, err := readConfig()
	if err != nil {
		return nil, err
	}
	args := parseArgs()
	if args.OutPath != "" {
		cfg.OutPath = args.OutPath
	}
	if args.MaxCover {
		cfg.MaxCover = args.MaxCover
	}
	if args.AlbumTemplate != "" {
		cfg.AlbumTemplate = args.AlbumTemplate
	}
	if args.TrackTemplate != "" {
		cfg.TrackTemplate = args.TrackTemplate
	}
	if cfg.AlbumTemplate == "" {
		cfg.AlbumTemplate = albumTemplate
	}
	if cfg.TrackTemplate == "" {
		cfg.TrackTemplate = trackTemplate
	}
	if cfg.OutPath == "" {
		cfg.OutPath = "Beatport downloads"
	}
	cfg.Urls, err = processUrls(args.Urls)
	if err != nil {
		errString := fmt.Sprintf("Failed to process URLs.\n%s", err)
		return nil, errors.New(errString)
	}
	return cfg, nil
}

func makeDirs(path string) error {
	return os.MkdirAll(path, 0755)
}

func checkUrl(url string) string {
	regex := regexp.MustCompile(regexString)
	match := regex.FindStringSubmatch(url)
	if match == nil {
		return ""
	}
	return match[1]
}

func getCsrfToken() (string, error) {
	_url := baseUrl + "account/login"
	req, err := http.NewRequest(http.MethodGet, _url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Referer", _url)
	do, err := client.Do(req)
	if err != nil {
		return "", err
	}
	do.Body.Close()
	if do.StatusCode != http.StatusOK {
		return "", errors.New(do.Status)
	}
	var csrf string
	for _, c := range do.Cookies() {
		if c.Name == "_csrf_token" {
			csrf = c.Value
			break
		}
	}
	if csrf == "" {
		return "", errors.New("Csrf token wasn't returned by server.")
	}
	return csrf, nil
}

func auth(email, pwd string) error {
	csrfToken, err := getCsrfToken()
	if err != nil {
		return errors.New("Failed to get CSRF token.\n" + err.Error())
	}
	_url := baseUrl + "account/login"
	data := url.Values{}
	data.Set("_csrf_token", csrfToken)
	data.Set("next", "")
	data.Set("username", email)
	data.Set("password", pwd)
	req, err := http.NewRequest(http.MethodPost, _url, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Referer", _url)
	do, err := client.Do(req)
	if err != nil {
		return err
	}
	do.Body.Close()
	if do.StatusCode != http.StatusOK {
		return errors.New(do.Status)
	}
	if do.Request.URL.String() == _url {
		return errors.New("Redirected to login page. Bad credentials?")
	}
	return nil
}

func getPlan() (string, error) {
	req, err := http.NewRequest(http.MethodGet, apiBase+"my/subscriptions", nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Referer", baseUrl+"subscriptions")
	do, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer do.Body.Close()
	if do.StatusCode != http.StatusOK {
		return "", errors.New(do.Status)
	}
	var obj UserSub
	err = json.NewDecoder(do.Body).Decode(&obj)
	if err != nil {
		return "", err
	}
	return obj.Subscription.Bundle.Name, nil
}

func getAlbumMeta(albumId, ref string) (*AlbumMeta, error) {
	req, err := http.NewRequest(http.MethodGet, apiBase+"catalog/releases/"+albumId, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Referer", ref)
	do, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer do.Body.Close()
	if do.StatusCode != http.StatusOK {
		return nil, errors.New(do.Status)
	}
	var obj AlbumMeta
	err = json.NewDecoder(do.Body).Decode(&obj)
	if err != nil {
		return nil, err
	}
	sort.Strings(obj.Tracks)
	return &obj, nil
}

func sanitize(filename string) string {
	regex := regexp.MustCompile(`[\/:*?"><|]`)
	sanitized := regex.ReplaceAllString(filename, "_")
	return sanitized
}

func parseArtists(artists []Artist) string {
	var parsedArtists string
	for _, artist := range artists {
		parsedArtists += artist.Name + ", "
	}
	return parsedArtists[:len(parsedArtists)-2]
}

func parseAlbumMeta(meta *AlbumMeta) map[string]string {
	parsedMeta := map[string]string{
		"album":       meta.Name,
		"albumArtist": parseArtists(meta.Artists),
		"year":        meta.PublishDate[:4],
	}
	upc := meta.Upc
	if upc != nil {
		parsedMeta["upc"] = upc.(string)
	}
	return parsedMeta
}

func parseTrackMeta(meta *TrackMeta, albMeta map[string]string, trackNum, trackTotal int, omit bool) (map[string]string, string) {
	albMeta["artist"] = parseArtists(meta.Artists)
	albMeta["bpm"] = strconv.Itoa(meta.Bpm)
	albMeta["genre"] = meta.Genre.Name
	albMeta["track"] = strconv.Itoa(trackNum)
	albMeta["trackPad"] = fmt.Sprintf("%02d", trackNum)
	albMeta["trackTotal"] = strconv.Itoa(trackTotal)
	isrc := meta.Isrc
	if isrc != nil {
		albMeta["isrc"] = isrc.(string)
	}
	mixName := meta.MixName
	titleWithMixName := meta.Name + " (" + mixName + ")"
	if omit {
		if mixName != "Original Mix" {
			albMeta["title"] = titleWithMixName
		} else {
			albMeta["title"] = meta.Name
		}
	} else {
		albMeta["title"] = titleWithMixName
	}
	return albMeta, titleWithMixName
}

func getTrackId(trackUrl string) (string, error) {
	u, err := url.Parse(trackUrl)
	if err != nil {
		return "", err
	}
	// Path only in case queries are added in the future.
	return path.Base(u.Path), nil
}

func getTrackMeta(trackId, ref string) (*TrackMeta, error) {
	req, err := http.NewRequest(http.MethodGet, apiBase+"catalog/tracks/"+trackId, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Referer", ref)
	do, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer do.Body.Close()
	if do.StatusCode != http.StatusOK {
		return nil, errors.New(do.Status)
	}
	var obj TrackMeta
	err = json.NewDecoder(do.Body).Decode(&obj)
	if err != nil {
		return nil, err
	}
	return &obj, nil
}

func fileExists(path string) (bool, error) {
	f, err := os.Stat(path)
	if err == nil {
		return !f.IsDir(), nil
	} else if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func getTrackStreamUrl(trackId, ref string, sampleEnd int) (string, error) {
	req, err := http.NewRequest(http.MethodGet, apiBase+"catalog/tracks/"+trackId+"/stream", nil)
	if err != nil {
		return "", err
	}
	query := url.Values{}
	query.Set("start", "0")
	query.Set("end", strconv.Itoa(sampleEnd))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Referer", ref)
	req.URL.RawQuery = query.Encode()
	do, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer do.Body.Close()
	if do.StatusCode != http.StatusOK {
		return "", errors.New(do.Status)
	}
	var obj TrackStream
	err = json.NewDecoder(do.Body).Decode(&obj)
	if err != nil {
		return "", err
	}
	streamUrl := strings.Replace(obj.StreamURL, ".128k.", ".256k.", 1)
	return streamUrl, nil
}

func getBaseUrl(manifestUrl string) (string, error) {
	u, err := url.Parse(manifestUrl)
	if err != nil {
		return "", err
	}
	return u.Scheme + "://" + u.Host + path.Dir(u.Path) + "/", nil
}

func getKey(keyUrl string) ([]byte, error) {
	req, err := client.Get(keyUrl)
	if err != nil {
		return nil, err
	}
	defer req.Body.Close()
	if req.StatusCode != http.StatusOK {
		return nil, errors.New(req.Status)
	}
	return ioutil.ReadAll(req.Body)
}

func parseKeyIv(segments *Segments, keyUrl, iv string) error {
	keyBytes, err := getKey(keyUrl)
	if err != nil {
		return err
	}
	ivBytes, err := hex.DecodeString(strings.TrimPrefix(iv, "0x"))
	if err != nil {
		return err
	}
	segments.Key = keyBytes
	segments.IV = ivBytes
	return nil
}

func parseSegments(manifestUrl string) (*Segments, error) {
	var segments Segments
	req, err := client.Get(manifestUrl)
	if err != nil {
		return nil, err
	}
	defer req.Body.Close()
	if req.StatusCode != http.StatusOK {
		return nil, errors.New(req.Status)
	}
	playlist, _, err := m3u8.DecodeFrom(req.Body, true)
	if err != nil {
		return nil, err
	}
	baseUrl, err := getBaseUrl(manifestUrl)
	if err != nil {
		return nil, err
	}
	media := playlist.(*m3u8.MediaPlaylist)
	for i, segment := range media.Segments {
		if segment == nil {
			break
		}
		if i == 0 {
			err := parseKeyIv(&segments, baseUrl+segment.Key.URI, segment.Key.IV)
			if err != nil {
				return nil, err
			}
		}
		segments.SegmentUrls = append(segments.SegmentUrls, baseUrl+segment.URI)
	}
	return &segments, nil
}

func pkcs5Trimming(data []byte) []byte {
	padding := data[len(data)-1]
	return data[:len(data)-int(padding)]
}

func decryptSegment(segmentBytes, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	cbc := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(segmentBytes))
	cbc.CryptBlocks(decrypted, segmentBytes)
	return pkcs5Trimming(decrypted), nil
}

func downloadSegments(tempPath string, segments *Segments) ([]string, error) {
	var segPaths []string
	segTotal := len(segments.SegmentUrls)
	for segNum, segmentUrl := range segments.SegmentUrls {
		segNum++
		segPath := filepath.Join(tempPath, fmt.Sprintf("%03d.aac", segNum))
		req, err := client.Get(segmentUrl)
		if err != nil {
			return nil, err
		}
		if req.StatusCode != http.StatusOK {
			req.Body.Close()
			return nil, errors.New(req.Status)
		}
		fmt.Printf("\rSegment %d of %d.", segNum, segTotal)
		segBytes, err := ioutil.ReadAll(req.Body)
		if err != nil {
			req.Body.Close()
			return nil, err
		}
		decSegBytes, err := decryptSegment(segBytes, segments.Key, segments.IV)
		if err != nil {
			req.Body.Close()
			return nil, err
		}
		f, err := os.OpenFile(segPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
		if err != nil {
			req.Body.Close()
			return nil, err
		}
		_, err = f.Write(decSegBytes)
		req.Body.Close()
		f.Close()
		if err != nil {
			return nil, err
		}
		segPaths = append(segPaths, segPath)
	}
	fmt.Println("")
	return segPaths, nil
}

func parseTemplate(templateText, defTemplate string, tags map[string]string) string {
	var buffer bytes.Buffer
	for {
		err := template.Must(template.New("").Parse(templateText)).Execute(&buffer, tags)
		if err == nil {
			break
		}
		fmt.Println("Failed to parse template. Default will be used instead.")
		templateText = defTemplate
		buffer.Reset()
	}
	return buffer.String()
}

func cleanup(tempPath string) {
	files, _ := ioutil.ReadDir(tempPath)
	for _, f := range files {
		os.Remove(filepath.Join(tempPath, f.Name()))
	}
}

func writeConcatFile(txtPath string, segPaths []string) error {
	f, err := os.OpenFile(txtPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, segPath := range segPaths {
		line := fmt.Sprintf("file '%s'\n", segPath)
		_, err := f.WriteString(line)
		if err != nil {
			return err
		}
	}
	return nil
}

func concatSegments(trackPath, tempPath string, segPaths []string) error {
	txtPath := filepath.Join(tempPath, "tmp.txt")
	defer cleanup(tempPath)
	err := writeConcatFile(txtPath, segPaths)
	if err != nil {
		return err
	}
	var (
		errBuffer bytes.Buffer
		args      = []string{"-f", "concat", "-safe", "0", "-i", txtPath, "-c:a", "copy", trackPath}
	)
	cmd := exec.Command("ffmpeg", args...)
	cmd.Stderr = &errBuffer
	err = cmd.Run()
	if err != nil {
		errString := fmt.Sprintf("%s\n%s", err, errBuffer.String())
		return errors.New(errString)
	}
	return nil
}

// Neither FFmpeg nor AtomicParsley support writing ISRC or UPC :(. Gib Go mp4 tag writing lib.
func writeTags(trackPath, coverPath string, _tags map[string]string) error {
	tags := map[string]string{
		"album":       _tags["album"],
		"albumArtist": _tags["albumArtist"],
		"artist":      _tags["artist"],
		"bpm":         _tags["bpm"],
		"genre":       _tags["genre"],
		"title":       _tags["title"],
		"tracknum":    _tags["track"] + "/" + _tags["trackTotal"],
		"year":        _tags["year"],
	}
	if coverPath != "" {
		tags["artwork"] = coverPath
	}
	return ap.WriteTags(trackPath, tags)
}

func downloadCover(maxUrl, dynamicUrl, coverPath string, maxCover bool) error {
	var _url string
	if maxCover {
		_url = maxUrl
	} else {
		_url = strings.Replace(dynamicUrl, "{w}x{h}", "600x600", 1)
	}
	f, err := os.OpenFile(coverPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	defer f.Close()
	req, err := client.Get(_url)
	if err != nil {
		return err
	}
	defer req.Body.Close()
	if req.StatusCode != http.StatusOK {
		return errors.New(req.Status)
	}
	_, err = io.Copy(f, req.Body)
	return err
}

func init() {
	fmt.Println(`
 _____         _               _      ____                _           _         
| __  |___ ___| |_ ___ ___ ___| |_   |    \ ___ _ _ _ ___| |___ ___ _| |___ ___ 
| __ -| -_| .'|  _| . | . |  _|  _|  |  |  | . | | | |   | | . | .'| . | -_|  _|
|_____|___|__,|_| |  _|___|_| |_|    |____/|___|_____|_|_|_|___|__,|___|___|_|  
                  |_|
`)
}

func main() {
	scriptDir, err := getScriptDir()
	if err != nil {
		panic(err)
	}
	tempPath, err := getTempPath()
	if err != nil {
		panic(err)
	}
	err = os.Chdir(scriptDir)
	if err != nil {
		panic(err)
	}
	cfg, err := parseCfg()
	if err != nil {
		handleErr("Failed to parse config file.", err, true)
	}
	err = makeDirs(cfg.OutPath)
	if err != nil {
		handleErr("Failed to make output path.", err, true)
	}
	err = auth(cfg.Email, cfg.Password)
	if err != nil {
		handleErr("Failed to auth.", err, true)
	}
	plan, err := getPlan()
	if err != nil {
		handleErr("Failed to get subscription info.", err, true)
	}
	if !strings.Contains(plan, "LINK") {
		panic("LINK or LINK Pro subscription required.")
	}
	fmt.Println("Signed in successfully - " + plan + "\n")
	albumTotal := len(cfg.Urls)
	for albumNum, _url := range cfg.Urls {
		fmt.Printf("Album %d of %d:\n", albumNum+1, albumTotal)
		albumId := checkUrl(_url)
		if albumId == "" {
			fmt.Println("Invalid URL:", _url)
			continue
		}
		albumMeta, err := getAlbumMeta(albumId, _url)
		if err != nil {
			handleErr("Failed to get album metadata.", err, false)
			continue
		}
		parsedAlbMeta := parseAlbumMeta(albumMeta)
		albumFolder := parseTemplate(cfg.AlbumTemplate, albumTemplate, parsedAlbMeta)
		fmt.Println(parsedAlbMeta["albumArtist"] + " - " + parsedAlbMeta["album"])
		if len(albumFolder) > 120 {
			fmt.Println("Album folder was chopped as it exceeds 120 characters.")
			albumFolder = albumFolder[:120]
		}
		albumPath := filepath.Join(cfg.OutPath, sanitize(albumFolder))
		err = makeDirs(albumPath)
		if err != nil {
			handleErr("Failed to make album folder.", err, false)
			continue
		}
		coverPath := filepath.Join(albumPath, "cover.jpg")
		err = downloadCover(albumMeta.Image.URI, albumMeta.Image.DynamicURI, coverPath, cfg.MaxCover)
		if err != nil {
			handleErr("Failed to get cover.", err, false)
			coverPath = ""
		}
		trackTotal := len(albumMeta.Tracks)
		for trackNum, trackUrl := range albumMeta.Tracks {
			trackNum++
			trackId, err := getTrackId(trackUrl)
			if err != nil {
				handleErr("Failed to get track ID.", err, false)
				continue
			}
			trackMeta, err := getTrackMeta(trackId, _url)
			if err != nil {
				handleErr("Failed to get track metadata.", err, false)
				continue
			}
			parsedMeta, titleWithMixName := parseTrackMeta(trackMeta, parsedAlbMeta, trackNum, trackTotal, cfg.OmitOrigMix)
			trackFname := parseTemplate(cfg.TrackTemplate, trackTemplate, parsedMeta)
			sanTrackFname := sanitize(trackFname)
			trackPath := filepath.Join(albumPath, sanTrackFname+".m4a")
			exists, err := fileExists(trackPath)
			if err != nil {
				handleErr("Failed to check if track already exists locally.", err, false)
				continue
			}
			if err != nil {
				handleErr("Failed to check if track already exists locally.", err, false)
				continue
			}
			if exists {
				fmt.Println("Track already exists locally.")
				continue
			}
			fmt.Printf(
				"Downloading track %d of %d: %s - AAC 256\n", trackNum, trackTotal, titleWithMixName,
			)
			streamUrl, err := getTrackStreamUrl(trackId, _url, trackMeta.SampleEndMs)
			if err != nil {
				handleErr("Failed to get track stream URL.", err, false)
				continue
			}
			segments, err := parseSegments(streamUrl)
			if err != nil {
				handleErr("Failed to parse segments.", err, false)
				continue
			}
			segPaths, err := downloadSegments(tempPath, segments)
			if err != nil {
				handleErr("Failed to download segments.", err, false)
				continue
			}
			err = concatSegments(trackPath, tempPath, segPaths)
			if err != nil {
				handleErr("Failed to concat segments.", err, false)
				continue
			}
			err = writeTags(trackPath, coverPath, parsedMeta)
			if err != nil {
				handleErr("Failed to write tags.", err, false)
			}
		}
		if coverPath != "" && !cfg.KeepCover {
			err := os.Remove(coverPath)
			if err != nil {
				handleErr("Failed to delete cover.", err, false)
			}
		}
	}
}
