package ts

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Monibuca/engine/v3"
	. "github.com/Monibuca/engine/v3"
	"github.com/Monibuca/utils/v3"
	"github.com/Monibuca/utils/v3/codec"
	"github.com/Monibuca/utils/v3/codec/mpegts"
)

var config = struct {
	BufferLength int
	Path         string
	// AutoPublish  bool
}{2048, "ts"}

func init() {
	InstallPlugin(&PluginConfig{
		Name:   "TS",
		Config: &config,
		// HotConfig: map[string]func(interface{}){
		// 	"AutoPublish": func(value interface{}) {
		// 		config.AutoPublish = value.(bool)
		// 	},
		// },
		Run: func() {
			http.HandleFunc("/api/ts/list", listTsDir)
			http.HandleFunc("/api/ts/publish", publishTsDir)
			// AddHook(HOOK_SUBSCRIBE, func(x interface{}) {
			// 	s := x.(*Subscriber)
			// 	if config.AutoPublish && s.Publisher == nil {
			// 		new(TS).PublishDir(s.StreamPath)
			// 	}
			// })
		},
	})
}

type TSDir struct {
	StreamPath string
	TsCount    int
	TotalSize  int64
}
type TS struct {
	*Stream
	*mpegts.MpegTsStream `json:"-"`
	TotalPesCount        int
	IsSplitFrame         bool
	PTS                  uint64
	DTS                  uint64
	PesCount             int
	BufferLength         int //TsChan     chan io.Reader
	lastDts              uint64
}

func (ts *TS) run() {
	//defer close(ts.TsChan)
	totalBuffer := cap(ts.TsPesPktChan)
	var at *AudioTrack
	vt := ts.NewVideoTrack(7)
	defer ts.Close()
	for {
		select {
		case <-ts.Done():
			return
		case tsPesPkt, ok := <-ts.TsPesPktChan:
			ts.BufferLength = len(ts.TsPesPktChan)
			if ok {
				ts.TotalPesCount++
				switch tsPesPkt.PesPkt.Header.StreamID & 0xF0 {
				case mpegts.STREAM_ID_AUDIO:
					data := tsPesPkt.PesPkt.Payload
					for remainLen := len(data); remainLen > 0; {
						// AACFrameLength(13)
						// xx xxxxxxxx xxx
						frameLen := (int(data[3]&3) << 11) | (int(data[4]) << 3) | (int(data[5]) >> 5)
						if frameLen > remainLen {
							break
						}
						payload := data[:frameLen]
						if at == nil {
							if payload[0] == 0xFF && (payload[1]&0xF0) == 0xF0 {
								at = ts.NewAudioTrack(10)
								//将ADTS转换成ASC
								at.SoundRate = codec.SamplingFrequencies[(payload[2]&0x3c)>>2]
								at.Channels = ((payload[2] & 0x1) << 2) | ((payload[3] & 0xc0) >> 6)
								at.ExtraData = codec.ADTSToAudioSpecificConfig(payload)
								at.PushRaw(engine.AudioPack{Timestamp: uint32(tsPesPkt.PesPkt.Header.Dts / 90), Raw: payload[7:]})
							} else {
								utils.Println("audio codec not support yet,want aac")
								return
								// ts.AudioTracks[0].SoundFormat = 2
								// ts.AudioTracks[0].Push(uint32(tsPesPkt.PesPkt.Header.Pts/90), payload)
							}
						} else {
							at.PushRaw(engine.AudioPack{Timestamp: uint32(tsPesPkt.PesPkt.Header.Dts / 90), Raw: payload[7:]})
						}
						data = data[frameLen:remainLen]
						remainLen = remainLen - frameLen
					}

				case mpegts.STREAM_ID_VIDEO:
					var err error
					ts.PTS = tsPesPkt.PesPkt.Header.Pts
					ts.DTS = tsPesPkt.PesPkt.Header.Dts
					lastDts := ts.lastDts
					dts := ts.DTS
					pts := ts.PTS
					if dts == 0 {
						dts = pts
					}
					if ts.lastDts == 0 {
						ts.lastDts = dts
					}
					compostionTime := uint32((pts - dts) / 90)
					t1 := time.Now()
					duration := time.Millisecond * time.Duration((dts-ts.lastDts)/90)
					ts.lastDts = dts
					vt.PushAnnexB(VideoPack{Timestamp: uint32(dts / 90), CompositionTime: compostionTime, Payload: tsPesPkt.PesPkt.Payload})
					if utils.MayBeError(err) {
						return
					}
					t2 := time.Since(t1)
					if duration != 0 && t2 < duration {
						if duration < time.Second {
							//if ts.BufferLength > 50 {
							duration = duration - t2
							//}
							if ts.BufferLength > 300 {
								duration = duration - duration*time.Duration(ts.BufferLength)/time.Duration(totalBuffer)
							}
							time.Sleep(duration)
						} else {
							time.Sleep(time.Millisecond * 20)
							log.Printf("stream:%s,duration:%d,dts:%d,lastDts:%d\n", ts.StreamPath, duration/time.Millisecond, tsPesPkt.PesPkt.Header.Dts, lastDts)
						}
					}
				}
			}
		}
	}
}
func (ts *TS) Publish(streamPath string) (result bool) {
	ts.Stream = &Stream{
		StreamPath: streamPath,
		Type:       "TS",
	}
	if result = ts.Stream.Publish(); result {
		ts.MpegTsStream = mpegts.NewMpegTsStream(config.BufferLength)
		go ts.run()
	}
	return
}
func (ts *TS) PublishDir(streamPath string) {
	dirPath := filepath.Join(config.Path, streamPath)
	files, err := ioutil.ReadDir(dirPath)
	if err != nil || len(files) == 0 {
		return
	}
	ts.Stream = &Stream{
		StreamPath: strings.ReplaceAll(streamPath, "\\", "/"),
		Type:       "TSFiles",
	}
	if ts.Stream.Publish() {
		ts.Type = "TSFiles"
		ts.MpegTsStream = mpegts.NewMpegTsStream(0)
		go ts.run()
		for _, file := range files {
			fullPath := filepath.Join(dirPath, file.Name())
			if filepath.Ext(fullPath) == ".ts" {
				if data, err := os.Open(fullPath); err == nil {
					ts.Feed(data)
					data.Close()
				}
			}
		}
		ts.Close()
	}
}
func publishTsDir(w http.ResponseWriter, r *http.Request) {
	streamPath := r.URL.Query().Get("streamPath")
	go new(TS).PublishDir(streamPath)
}
func readTsDir(currentDir string) []*TSDir {
	var list []*TSDir
	abDir := filepath.Join(config.Path, currentDir)
	if items, err := ioutil.ReadDir(abDir); err == nil {
		tscount := 0
		var totalSize int64
		for _, file := range items {
			if file.IsDir() {
				list = append(list, readTsDir(filepath.Join(currentDir, file.Name()))...)
			} else if filepath.Ext(filepath.Join(abDir, file.Name())) == ".ts" {
				tscount++
				totalSize = totalSize + file.Size()
			}
		}
		if tscount > 0 {
			info := TSDir{
				currentDir, tscount, totalSize,
			}
			list = append(list, &info)
		}
	}
	return list
}
func listTsDir(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	var list []*TSDir = readTsDir(".")
	bytes, err := json.Marshal(list)
	if err == nil {
		w.Write(bytes)
	} else {
		w.Write([]byte("{\"code\":1}"))
	}
}
