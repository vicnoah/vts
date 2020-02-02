package transcode

const (
	vp9Vaapi = "docker run -i --rm -v=%s:/trs/%s --device /dev/dri/renderD128 jrottenberg/ffmpeg:4.1-vaapi -hwaccel vaapi -hwaccel_output_format vaapi -hwaccel_device /dev/dri/renderD128 -i /trs%s -c:v vp9_vaapi -c:a libvorbis /trs%s"
)
