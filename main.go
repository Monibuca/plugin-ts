package ts

import (
	"io"

	. "github.com/Monibuca/engine/v4"
	"github.com/Monibuca/engine/v4/track"
	"github.com/Monibuca/utils/v3"
	astits "github.com/asticode/go-astits"
)

type TSConfig struct {
}

func (config *TSConfig) Update(override Config) {
	override.Unmarshal(config)

}

func init() {
	var config TSConfig
	InstallPlugin(&config)
	// http.HandleFunc("/api/ts/list", listTsDir)
	// http.HandleFunc("/api/ts/publish", publishTsDir)
}

type TSDir struct {
	StreamPath string
	TsCount    int
	TotalSize  int64
}
type TS struct {
	Publisher
	io.ReadCloser
	*astits.Demuxer
	TotalPesCount int
	IsSplitFrame  bool
	PTS           uint64
	DTS           uint64
	PesCount      int
	BufferLength  int //TsChan     chan io.Reader
	lastDts       uint64
}

func (ts *TS) Close() {
	ts.ReadCloser.Close()
}

func (ts *TS) Feed(source io.ReadCloser) {
	ts.ReadCloser = source
	ts.Demuxer = astits.NewDemuxer(ts, source)
	var at *track.Audio
	var vt *track.Video
	for {
		d, err := ts.NextData()
		if err != nil {
			return
		}
		if d.PMT != nil {
			// Loop through elementary streams
			for _, es := range d.PMT.ElementaryStreams {
				switch es.StreamType {
				case astits.StreamTypeH264Video:
					vt = (*track.Video)(ts.NewH264Track())
				case astits.StreamTypeH265Video:
					vt = (*track.Video)(ts.NewH265Track())
				case astits.StreamTypeAACAudio:
					at = (*track.Audio)(ts.NewAACTrack())
				}
			}
		}
		if d.PES != nil {
			if d.PES.Header.IsVideoStream() {
				vt.WriteAnnexB(uint32(d.PES.Header.OptionalHeader.PTS.Base), uint32(d.PES.Header.OptionalHeader.DTS.Base), d.PES.Data)
			} else {
				data := d.PES.Data
				at.Value.PTS = uint32(d.PES.Header.OptionalHeader.PTS.Base)
				at.Value.DTS = uint32(d.PES.Header.OptionalHeader.DTS.Base)
				for remainLen := len(data); remainLen > 0; {
					// AACFrameLength(13)
					// xx xxxxxxxx xxx
					frameLen := (int(data[3]&3) << 11) | (int(data[4]) << 3) | (int(data[5]) >> 5)
					if frameLen > remainLen {
						break
					}
					payload := data[:frameLen]
					if at.DecoderConfiguration.AVCC == nil {
						if payload[0] == 0xFF && (payload[1]&0xF0) == 0xF0 {
							//将ADTS转换成ASC
							at.WriteADTS(payload[:7])
							at.WriteSlice(payload[7:])
							at.Flush()
						} else {
							utils.Println("audio codec not support yet,want aac")
							return
							// ts.AudioTracks[0].SoundFormat = 2
							// ts.AudioTracks[0].Push(uint32(tsPesPkt.PesPkt.Header.Pts/90), payload)
						}
					} else if len(payload) > 7 {
						at.WriteSlice(payload[7:])
						at.Flush()
					}
					data = data[frameLen:remainLen]
					remainLen -= frameLen
				}
			}
		}
	}
}

// func publishTsDir(w http.ResponseWriter, r *http.Request) {
// 	streamPath := r.URL.Query().Get("streamPath")
// 	go new(TS).PublishDir(streamPath)
// }
// func readTsDir(currentDir string) []*TSDir {
// 	var list []*TSDir
// 	abDir := filepath.Join(config.Path, currentDir)
// 	if items, err := ioutil.ReadDir(abDir); err == nil {
// 		tscount := 0
// 		var totalSize int64
// 		for _, file := range items {
// 			if file.IsDir() {
// 				list = append(list, readTsDir(filepath.Join(currentDir, file.Name()))...)
// 			} else if filepath.Ext(filepath.Join(abDir, file.Name())) == ".ts" {
// 				tscount++
// 				totalSize = totalSize + file.Size()
// 			}
// 		}
// 		if tscount > 0 {
// 			info := TSDir{
// 				currentDir, tscount, totalSize,
// 			}
// 			list = append(list, &info)
// 		}
// 	}
// 	return list
// }
// func listTsDir(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Access-Control-Allow-Origin", "*")
// 	var list []*TSDir = readTsDir(".")
// 	bytes, err := json.Marshal(list)
// 	if err == nil {
// 		w.Write(bytes)
// 	} else {
// 		w.Write([]byte("{\"code\":1}"))
// 	}
// }
