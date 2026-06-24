package ali

import (
	"bytes"
	"encoding/base64"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relay/constant"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newMultipartAudioContext builds a gin context whose request carries a
// multipart/form-data body with a "file" part (and optional extra fields),
// mirroring how /v1/audio/transcriptions clients upload audio.
func newMultipartAudioContext(t *testing.T, filename string, fileContent []byte, fields map[string]string) *gin.Context {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filename)
	require.NoError(t, err)
	_, err = part.Write(fileContent)
	require.NoError(t, err)

	for k, v := range fields {
		require.NoError(t, writer.WriteField(k, v))
	}
	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, "/v1/audio/transcriptions", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req
	return c
}

func TestConvertAudioToAliChatRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	audio := []byte("fake-wav-bytes\x00\x01\x02")
	c := newMultipartAudioContext(t, "speech.wav", audio, map[string]string{"language": "zh"})

	adaptor := &Adaptor{}
	info := &relaycommon.RelayInfo{RelayMode: constant.RelayModeAudioTranscription}
	request := dto.AudioRequest{Model: "qwen3-asr-flash", ResponseFormat: "json"}

	reader, err := adaptor.ConvertAudioRequest(c, info, request)
	require.NoError(t, err)
	require.Equal(t, "json", adaptor.ResponseFormat, "ConvertAudioRequest 应保存 ResponseFormat 供 handler 使用")

	raw, err := io.ReadAll(reader)
	require.NoError(t, err)

	var got aliASRChatRequest
	require.NoError(t, common.Unmarshal(raw, &got))

	assert.Equal(t, "qwen3-asr-flash", got.Model)
	assert.False(t, got.Stream, "ASR 走非流式 chat/completions")
	require.Len(t, got.Messages, 1)
	require.Len(t, got.Messages[0].Content, 1)

	content := got.Messages[0].Content[0]
	assert.Equal(t, "input_audio", content.Type)
	require.NotNil(t, content.InputAudio)

	// data URI 必须是 data:<mime>;base64,<payload>，且 payload 解码后等于原始音频字节。
	const prefix = "data:"
	require.True(t, strings.HasPrefix(content.InputAudio.Data, prefix))
	b64Idx := strings.Index(content.InputAudio.Data, ";base64,")
	require.Greater(t, b64Idx, len(prefix), "data URI 必须包含 ;base64, 段")
	payload := content.InputAudio.Data[b64Idx+len(";base64,"):]
	decoded, err := base64.StdEncoding.DecodeString(payload)
	require.NoError(t, err)
	assert.Equal(t, audio, decoded, "base64 解码后必须还原原始音频字节")

	// .wav 扩展名应映射到 audio/* MIME。
	mimePart := content.InputAudio.Data[len(prefix):b64Idx]
	assert.True(t, strings.HasPrefix(mimePart, "audio/"), "wav 应推断为 audio/* MIME，实际为 %q", mimePart)

	// language 表单字段应进入 extra_body.asr_options.language。
	require.NotNil(t, got.ExtraBody)
	require.NotNil(t, got.ExtraBody.ASROptions)
	assert.Equal(t, "zh", got.ExtraBody.ASROptions.Language)
}

func TestConvertAudioToAliChatRequest_NoLanguageOmitsExtraBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	c := newMultipartAudioContext(t, "speech.mp3", []byte("data"), nil)

	adaptor := &Adaptor{}
	info := &relaycommon.RelayInfo{RelayMode: constant.RelayModeAudioTranscription}
	reader, err := adaptor.ConvertAudioRequest(c, info, dto.AudioRequest{Model: "qwen3-asr-flash"})
	require.NoError(t, err)

	raw, err := io.ReadAll(reader)
	require.NoError(t, err)

	// 未提供 language 时 extra_body 必须被 omitempty 省略，避免向上游发送空对象。
	assert.NotContains(t, string(raw), "extra_body")
	assert.NotContains(t, string(raw), "asr_options")
}

func TestConvertAudioRequest_RejectsUnsupportedMode(t *testing.T) {
	gin.SetMode(gin.TestMode)

	c := newMultipartAudioContext(t, "speech.wav", []byte("x"), nil)
	adaptor := &Adaptor{}
	info := &relaycommon.RelayInfo{RelayMode: constant.RelayModeChatCompletions}

	_, err := adaptor.ConvertAudioRequest(c, info, dto.AudioRequest{Model: "qwen3-asr-flash"})
	require.Error(t, err, "非音频 relay mode 必须被拒绝")
}

func TestConvertAudioToAliChatRequest_MissingFile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	require.NoError(t, writer.WriteField("language", "en"))
	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, "/v1/audio/transcriptions", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req

	adaptor := &Adaptor{}
	info := &relaycommon.RelayInfo{RelayMode: constant.RelayModeAudioTranscription}
	_, err := adaptor.ConvertAudioRequest(c, info, dto.AudioRequest{Model: "qwen3-asr-flash"})
	require.Error(t, err, "缺少 file 字段必须报错")
}

func newResponseContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/audio/transcriptions", nil)
	return c, rec
}

func makeResp(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestAliSTTHandler_JSONResponse(t *testing.T) {
	c, rec := newResponseContext()
	resp := makeResp(`{"choices":[{"message":{"role":"assistant","content":"你好世界"}}],"usage":{"prompt_tokens":12,"completion_tokens":4,"total_tokens":16}}`)
	info := &relaycommon.RelayInfo{RelayMode: constant.RelayModeAudioTranscription}

	apiErr, usage := aliSTTHandler(c, resp, info, "json")
	require.Nil(t, apiErr)
	require.NotNil(t, usage)

	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var out dto.AudioResponse
	require.NoError(t, common.Unmarshal(rec.Body.Bytes(), &out))
	assert.Equal(t, "你好世界", out.Text)

	assert.Equal(t, 12, usage.PromptTokens)
	assert.Equal(t, 4, usage.CompletionTokens)
	assert.Equal(t, 16, usage.TotalTokens)
}

func TestAliSTTHandler_TextResponse(t *testing.T) {
	c, rec := newResponseContext()
	resp := makeResp(`{"choices":[{"message":{"content":"plain text"}}],"usage":{"prompt_tokens":3,"completion_tokens":2,"total_tokens":5}}`)
	info := &relaycommon.RelayInfo{RelayMode: constant.RelayModeAudioTranscription}

	apiErr, usage := aliSTTHandler(c, resp, info, "text")
	require.Nil(t, apiErr)
	require.NotNil(t, usage)

	assert.True(t, strings.HasPrefix(rec.Header().Get("Content-Type"), "text/plain"))
	// text 格式必须返回裸文本，而不是 JSON 包裹。
	assert.Equal(t, "plain text", rec.Body.String())
}

func TestAliSTTHandler_UsageFallbackToInputOutputTokens(t *testing.T) {
	c, _ := newResponseContext()
	// DashScope 有时只返回 input_tokens/output_tokens，需回退映射到 prompt/completion。
	resp := makeResp(`{"choices":[{"message":{"content":"hi"}}],"usage":{"input_tokens":20,"output_tokens":6,"total_tokens":26}}`)
	info := &relaycommon.RelayInfo{RelayMode: constant.RelayModeAudioTranscription}

	apiErr, usage := aliSTTHandler(c, resp, info, "json")
	require.Nil(t, apiErr)
	require.NotNil(t, usage)
	assert.Equal(t, 20, usage.PromptTokens)
	assert.Equal(t, 6, usage.CompletionTokens)
	assert.Equal(t, 26, usage.TotalTokens)
}

func TestAliSTTHandler_UpstreamError(t *testing.T) {
	c, _ := newResponseContext()
	resp := makeResp(`{"code":"InvalidApiKey","message":"Invalid API-key provided.","request_id":"abc"}`)
	info := &relaycommon.RelayInfo{RelayMode: constant.RelayModeAudioTranscription}

	apiErr, usage := aliSTTHandler(c, resp, info, "json")
	require.NotNil(t, apiErr, "上游返回 code 时必须转成错误")
	assert.Nil(t, usage)
	assert.Contains(t, apiErr.Error(), "Invalid API-key")
}
