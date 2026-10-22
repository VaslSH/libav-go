package libav

/*
#cgo pkg-config: libavcodec libavformat libavutil libswscale
#include <libavcodec/packet.h>
#include <libavcodec/codec.h>
#include <libavcodec/codec_id.h>
#include <libavcodec/codec_par.h>
#include <libavcodec/avcodec.h>
#include <libavformat/avformat.h>
#include <libavutil/avutil.h>
#include <libavutil/rational.h>
#include <libavutil/error.h>
#include <errno.h>
#include <string.h>
#include <libswscale/swscale.h>
#include <libavutil/imgutils.h>
#include <libavutil/frame.h>

static inline int av_error_eagain(void) { return AVERROR(EAGAIN); }
*/
import "C"
import "fmt"

type SWRescale struct {
	Inner *C.SwsContext
	Dst   Frame
}
type SWRescaleInfo struct {
	Height, Width int
	PixelFormat   PixelFormat
}

func NewSWRescale(src Frame, target SWRescaleInfo) SWRescale {
	ctx := C.sws_getContext(
		C.int(src.Width()), C.int(src.Height()), int32(src.Format()),
		C.int(target.Width), C.int(target.Height), int32(target.PixelFormat),
		C.SWS_BILINEAR, nil, nil, nil,
	)
	if ctx == nil {
		panic("failed to create scaling context")
	}

	Dst, err := NewFrameWithVideoBuffers(target.Height, target.Width, target.PixelFormat)
	if err != nil {
		panic("failed to allocate sws context! err = " + err.Error())
	}

	return SWRescale{ctx, Dst}
}

func (r *SWRescale) IsValid () bool { // invalid when no abr
	return r.Inner != nil
}
func (r *SWRescale) Rescale(f Frame) Frame {
	ret := C.sws_scale(
		r.Inner,
		f.Data(), f.LineSize(),
		0, C.int(f.Height()),
		r.Dst.Data(), r.Dst.LineSize(),
	)

	r.Dst.SetPts(f.Pts())
	if f.KeyFrame() {
		r.Dst.SetKeyFrame()
	}

	if ret < 0 {
		err := fmt.Errorf("failed %d: %s", -ret, AvErrorString(int(ret)))
		panic(err)
	}
	return r.Dst
}

func (r *SWRescale) Close() {
	// Cleanup when done
	C.sws_freeContext(r.Inner)
}
