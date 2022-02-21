package ts

import (
	. "github.com/Monibuca/engine/v4"
	"github.com/Monibuca/engine/v4/track"
	"github.com/asticode/go-astits"
)

type TSConfig struct {
}

func (config *TSConfig) OnEvent(event any) {
}

var plugin = InstallPlugin(&TSConfig{})

type TSDir struct {
	StreamPath string
	TsCount    int
	TotalSize  int64
}
type TSPuller struct {
	Publisher
	Puller
	PesCount int
}

func (ts *TSPuller) Pull() {
	demuxer := astits.NewDemuxer(ts, ts)
	for d, err := demuxer.NextData(); err == nil; d, err = demuxer.NextData() {
		if d.PMT != nil && (ts.VideoTrack == nil || ts.AudioTrack == nil) {
			// Loop through elementary streams
			for _, es := range d.PMT.ElementaryStreams {
				switch es.StreamType {
				case astits.StreamTypeH264Video:
					if ts.VideoTrack == nil {
						ts.VideoTrack = track.NewH264(ts.Stream)
					}
				case astits.StreamTypeH265Video:
					if ts.VideoTrack == nil {
						ts.VideoTrack = track.NewH265(ts.Stream)
					}
				case astits.StreamTypeAACAudio:
					if ts.AudioTrack == nil {
						ts.AudioTrack = track.NewAAC(ts.Stream)
					}
				}
			}
		}
		if d.PES != nil {
			ts.PesCount++
			if d.PES.Header.IsVideoStream() {
				ts.VideoTrack.WriteAnnexB(uint32(d.PES.Header.OptionalHeader.PTS.Base), uint32(d.PES.Header.OptionalHeader.DTS.Base), d.PES.Data)
			} else {
				data := d.PES.Data
				ts.AudioTrack.CurrentFrame().PTS = uint32(d.PES.Header.OptionalHeader.PTS.Base)
				ts.AudioTrack.CurrentFrame().DTS = uint32(d.PES.Header.OptionalHeader.DTS.Base)
				for remainLen := len(data); remainLen > 0; {
					// AACFrameLength(13)
					// xx xxxxxxxx xxx
					frameLen := (int(data[3]&3) << 11) | (int(data[4]) << 3) | (int(data[5]) >> 5)
					if frameLen > remainLen {
						break
					}
					payload := data[:frameLen]
					if ts.AudioTrack.GetDecoderConfiguration().AVCC == nil {
						if payload[0] == 0xFF && (payload[1]&0xF0) == 0xF0 {
							//将ADTS转换成ASC
							ts.AudioTrack.WriteADTS(payload[:7])
							ts.AudioTrack.WriteSlice(payload[7:])
							ts.AudioTrack.Flush()
						} else {
							plugin.Warn("audio codec not support yet,want aac")
							continue
							// ts.AudioTracks[0].SoundFormat = 2
							// ts.AudioTracks[0].Push(uint32(tsPesPkt.PesPkt.Header.Pts/90), payload)
						}
					} else if len(payload) > 7 {
						ts.AudioTrack.WriteSlice(payload[7:])
						ts.AudioTrack.Flush()
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
