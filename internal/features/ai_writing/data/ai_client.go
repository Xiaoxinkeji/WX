package data

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Xiaoxinkeji/WX/internal/features/ai_writing/domain"
)

type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

type OpenAIClient struct {
	BaseURL      string
	APIKey       string
	HTTPClient   HTTPDoer
	DefaultModel string
	UserAgent    string
}

func (c OpenAIClient) ProviderName() string { return "openai" }

func (c OpenAIClient) Chat(ctx context.Context, req domain.ChatRequest) (domain.ChatResponse, error) {
	return c.doChat(ctx, req, nil)
}

func (c OpenAIClient) StreamChat(ctx context.Context, req domain.ChatRequest, onDelta func(delta string) error) (domain.ChatResponse, error) {
	return c.doChat(ctx, req, onDelta)
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIChatRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	Temperature float64         `json:"temperature,omitempty"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
}

type openAIChatResponse struct {
	Model   string `json:"model"`
	Choices []struct {
		Message      openAIMessage `json:"message"`
		FinishReason string        `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type openAIStreamChunk struct {
	Model   string `json:"model"`
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}

func (c OpenAIClient) doChat(ctx context.Context, req domain.ChatRequest, onDelta func(delta string) error) (domain.ChatResponse, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	client := doerOrDefault(c.HTTPClient)

	model := strings.TrimSpace(req.Model)
	if model == "" {
		model = strings.TrimSpace(c.DefaultModel)
	}
	if model == "" {
		model = "gpt-4o-mini"
	}

	payload := openAIChatRequest{
		Model:       model,
		Messages:    toOpenAIMessages(req.Messages),
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		Stream:      onDelta != nil,
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return domain.ChatResponse{}, err
	}

	urlStr, err := buildURL(c.BaseURL, "https://api.openai.com", "/v1/chat/completions")
	if err != nil {
		return domain.ChatResponse{}, err
	}

	hreq, err := http.NewRequestWithContext(ctx, http.MethodPost, urlStr, bytes.NewReader(b))
	if err != nil {
		return domain.ChatResponse{}, err
	}
	hreq.Header.Set("Content-Type", "application/json")
	if ua := strings.TrimSpace(c.UserAgent); ua != "" {
		hreq.Header.Set("User-Agent", ua)
	}
	if key := strings.TrimSpace(c.APIKey); key != "" {
		hreq.Header.Set("Authorization", "Bearer "+key)
	}

	resp, err := client.Do(hreq)
	if err != nil {
		return domain.ChatResponse{}, errors.Join(domain.ErrProvider, err)
	}
	defer func() {
		// Ensure response body is fully read before closing to allow connection reuse
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := readSmallBody(resp.Body, 1024) // Reduced from 64KB to 1KB
		if msg == "" {
			msg = resp.Status
		}
		// Log detailed error but return generic message to user
		return domain.ChatResponse{}, errors.Join(domain.ErrProvider, fmt.Errorf("openai: status %d", resp.StatusCode))
	}

	if onDelta == nil {
		var out openAIChatResponse
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			return domain.ChatResponse{}, errors.Join(domain.ErrProvider, err)
		}
		content, finish := "", ""
		if len(out.Choices) > 0 {
			content = out.Choices[0].Message.Content
			finish = out.Choices[0].FinishReason
		}
		return domain.ChatResponse{
			Provider:     c.ProviderName(),
			Model:        out.Model,
			Content:      content,
			FinishReason: finish,
			Usage: domain.TokenUsage{
				PromptTokens:     out.Usage.PromptTokens,
				CompletionTokens: out.Usage.CompletionTokens,
				TotalTokens:      out.Usage.TotalTokens,
			},
		}, nil
	}

	return streamOpenAI(resp.Body, c.ProviderName(), model, onDelta)
}

func toOpenAIMessages(in []domain.Message) []openAIMessage {
	out := make([]openAIMessage, 0, len(in))
	for _, m := range in {
		out = append(out, openAIMessage{Role: string(m.Role), Content: m.Content})
	}
	return out
}

func streamOpenAI(r io.Reader, providerName, model string, onDelta func(delta string) error) (domain.ChatResponse, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var content strings.Builder
	finishReason := ""
	modelOut := model

	for scanner.Scan() {
		line := strings.TrimSpace(strings.TrimRight(scanner.Text(), "\r"))
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "data:") {
			data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if data == "[DONE]" {
				break
			}

			var chunk openAIStreamChunk
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				return domain.ChatResponse{}, errors.Join(domain.ErrStream, err)
			}
			if chunk.Model != "" {
				modelOut = chunk.Model
			}
			if len(chunk.Choices) == 0 {
				continue
			}
			delta := chunk.Choices[0].Delta.Content
			if delta != "" {
				content.WriteString(delta)
				if err := onDelta(delta); err != nil {
					return domain.ChatResponse{}, err
				}
			}
			if chunk.Choices[0].FinishReason != nil {
				finishReason = *chunk.Choices[0].FinishReason
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return domain.ChatResponse{}, errors.Join(domain.ErrStream, err)
	}

	return domain.ChatResponse{
		Provider:     providerName,
		Model:        modelOut,
		Content:      content.String(),
		FinishReason: finishReason,
	}, nil
}

type ClaudeClient struct {
	BaseURL           string
	APIKey            string
	HTTPClient        HTTPDoer
	DefaultModel      string
	AnthropicVersion  string
	UserAgent         string
	DefaultMaxTokens  int
	DefaultTemperature float64
}

func (c ClaudeClient) ProviderName() string { return "claude" }

func (c ClaudeClient) Chat(ctx context.Context, req domain.ChatRequest) (domain.ChatResponse, error) {
	return c.doChat(ctx, req, nil)
}

func (c ClaudeClient) StreamChat(ctx context.Context, req domain.ChatRequest, onDelta func(delta string) error) (domain.ChatResponse, error) {
	return c.doChat(ctx, req, onDelta)
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeChatRequest struct {
	Model       string         `json:"model"`
	MaxTokens   int            `json:"max_tokens"`
	Messages    []claudeMessage `json:"messages"`
	System      string         `json:"system,omitempty"`
	Temperature float64        `json:"temperature,omitempty"`
	Stream      bool           `json:"stream,omitempty"`
}

type claudeChatResponse struct {
	Model      string `json:"model"`
	StopReason string `json:"stop_reason"`
	Content    []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

func (c ClaudeClient) doChat(ctx context.Context, req domain.ChatRequest, onDelta func(delta string) error) (domain.ChatResponse, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	client := doerOrDefault(c.HTTPClient)

	model := strings.TrimSpace(req.Model)
	if model == "" {
		model = strings.TrimSpace(c.DefaultModel)
	}
	if model == "" {
		model = "claude-3-5-sonnet-20241022"
	}
	maxTokens := req.MaxTokens
	if maxTokens <= 0 {
		maxTokens = c.DefaultMaxTokens
	}
	if maxTokens <= 0 {
		maxTokens = 1024
	}
	temp := req.Temperature
	if temp == 0 {
		temp = c.DefaultTemperature
	}

	systemText, msgs := splitSystemMessages(req.Messages)
	payload := claudeChatRequest{
		Model:       model,
		MaxTokens:   maxTokens,
		Messages:    toClaudeMessages(msgs),
		System:      systemText,
		Temperature: temp,
		Stream:      onDelta != nil,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return domain.ChatResponse{}, err
	}

	urlStr, err := buildURL(c.BaseURL, "https://api.anthropic.com", "/v1/messages")
	if err != nil {
		return domain.ChatResponse{}, err
	}

	hreq, err := http.NewRequestWithContext(ctx, http.MethodPost, urlStr, bytes.NewReader(b))
	if err != nil {
		return domain.ChatResponse{}, err
	}
	hreq.Header.Set("Content-Type", "application/json")
	if ua := strings.TrimSpace(c.UserAgent); ua != "" {
		hreq.Header.Set("User-Agent", ua)
	}
	if key := strings.TrimSpace(c.APIKey); key != "" {
		hreq.Header.Set("x-api-key", key)
	}
	version := strings.TrimSpace(c.AnthropicVersion)
	if version == "" {
		version = "2023-06-01"
	}
	hreq.Header.Set("anthropic-version", version)

	resp, err := client.Do(hreq)
	if err != nil {
		return domain.ChatResponse{}, errors.Join(domain.ErrProvider, err)
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := readSmallBody(resp.Body, 1024)
		if msg == "" {
			msg = resp.Status
		}
		return domain.ChatResponse{}, errors.Join(domain.ErrProvider, fmt.Errorf("claude: status %d", resp.StatusCode))
	}

	if onDelta == nil {
		var out claudeChatResponse
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			return domain.ChatResponse{}, errors.Join(domain.ErrProvider, err)
		}
		content := joinTextBlocks(out.Content)
		return domain.ChatResponse{
			Provider:     c.ProviderName(),
			Model:        out.Model,
			Content:      content,
			FinishReason: out.StopReason,
			Usage: domain.TokenUsage{
				PromptTokens:     out.Usage.InputTokens,
				CompletionTokens: out.Usage.OutputTokens,
				TotalTokens:      out.Usage.InputTokens + out.Usage.OutputTokens,
			},
		}, nil
	}

	return streamClaude(resp.Body, c.ProviderName(), model, onDelta)
}

func toClaudeMessages(in []domain.Message) []claudeMessage {
	out := make([]claudeMessage, 0, len(in))
	for _, m := range in {
		if m.Role == domain.RoleSystem {
			continue
		}
		role := "user"
		if m.Role == domain.RoleAssistant {
			role = "assistant"
		}
		out = append(out, claudeMessage{Role: role, Content: m.Content})
	}
	return out
}

type GeminiClient struct {
	BaseURL      string
	APIKey       string
	HTTPClient   HTTPDoer
	DefaultModel string
	UserAgent    string
}

func (c GeminiClient) ProviderName() string { return "gemini" }

func (c GeminiClient) Chat(ctx context.Context, req domain.ChatRequest) (domain.ChatResponse, error) {
	return c.doChat(ctx, req, nil)
}

func (c GeminiClient) StreamChat(ctx context.Context, req domain.ChatRequest, onDelta func(delta string) error) (domain.ChatResponse, error) {
	return c.doChat(ctx, req, onDelta)
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiGenerationConfig struct {
	Temperature     float64 `json:"temperature,omitempty"`
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
}

type geminiChatRequest struct {
	Contents          []geminiContent          `json:"contents"`
	SystemInstruction *geminiContent          `json:"system_instruction,omitempty"`
	GenerationConfig  *geminiGenerationConfig `json:"generationConfig,omitempty"`
}

type geminiChatResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
		FinishReason string `json:"finishReason"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		TotalTokenCount      int `json:"totalTokenCount"`
	} `json:"usageMetadata"`
}

func (c GeminiClient) doChat(ctx context.Context, req domain.ChatRequest, onDelta func(delta string) error) (domain.ChatResponse, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	client := doerOrDefault(c.HTTPClient)

	model := strings.TrimSpace(req.Model)
	if model == "" {
		model = strings.TrimSpace(c.DefaultModel)
	}
	if model == "" {
		model = "gemini-1.5-flash"
	}

	systemText, msgs := splitSystemMessages(req.Messages)
	contents := toGeminiContents(msgs)
	var systemInstruction *geminiContent
	if strings.TrimSpace(systemText) != "" {
		systemInstruction = &geminiContent{Parts: []geminiPart{{Text: systemText}}}
	}
	var generationConfig *geminiGenerationConfig
	if req.Temperature != 0 || req.MaxTokens != 0 {
		generationConfig = &geminiGenerationConfig{Temperature: req.Temperature, MaxOutputTokens: req.MaxTokens}
	}

	payload := geminiChatRequest{Contents: contents, SystemInstruction: systemInstruction, GenerationConfig: generationConfig}
	b, err := json.Marshal(payload)
	if err != nil {
		return domain.ChatResponse{}, err
	}

	action := ":generateContent"
	if onDelta != nil {
		action = ":streamGenerateContent"
	}
	path := fmt.Sprintf("/v1beta/models/%s%s", url.PathEscape(model), action)
	urlStr, err := buildURL(c.BaseURL, "https://generativelanguage.googleapis.com", path)
	if err != nil {
		return domain.ChatResponse{}, err
	}
	urlStr, _ = addAPIKeyQueryParam(urlStr, c.APIKey)

	hreq, err := http.NewRequestWithContext(ctx, http.MethodPost, urlStr, bytes.NewReader(b))
	if err != nil {
		return domain.ChatResponse{}, err
	}
	hreq.Header.Set("Content-Type", "application/json")
	if ua := strings.TrimSpace(c.UserAgent); ua != "" {
		hreq.Header.Set("User-Agent", ua)
	}
	if key := strings.TrimSpace(c.APIKey); key != "" {
		hreq.Header.Set("x-goog-api-key", key)
	}

	resp, err := client.Do(hreq)
	if err != nil {
		return domain.ChatResponse{}, errors.Join(domain.ErrProvider, err)
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := readSmallBody(resp.Body, 1024)
		if msg == "" {
			msg = resp.Status
		}
		return domain.ChatResponse{}, errors.Join(domain.ErrProvider, fmt.Errorf("gemini: status %d", resp.StatusCode))
	}

	if onDelta == nil {
		var out geminiChatResponse
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			return domain.ChatResponse{}, errors.Join(domain.ErrProvider, err)
		}
		content, finish := extractGeminiContentAndFinish(out)
		return domain.ChatResponse{
			Provider:     c.ProviderName(),
			Model:        model,
			Content:      content,
			FinishReason: finish,
			Usage: domain.TokenUsage{
				PromptTokens:     out.UsageMetadata.PromptTokenCount,
				CompletionTokens: out.UsageMetadata.CandidatesTokenCount,
				TotalTokens:      out.UsageMetadata.TotalTokenCount,
			},
		}, nil
	}

	return streamGemini(resp.Body, c.ProviderName(), model, onDelta)
}

func toGeminiContents(in []domain.Message) []geminiContent {
	out := make([]geminiContent, 0, len(in))
	for _, m := range in {
		if m.Role == domain.RoleSystem {
			continue
		}
		role := "user"
		if m.Role == domain.RoleAssistant {
			role = "model"
		}
		out = append(out, geminiContent{Role: role, Parts: []geminiPart{{Text: m.Content}}})
	}
	return out
}

func extractGeminiContentAndFinish(out geminiChatResponse) (string, string) {
	if len(out.Candidates) == 0 {
		return "", ""
	}
	var sb strings.Builder
	for _, p := range out.Candidates[0].Content.Parts {
		sb.WriteString(p.Text)
	}
	return sb.String(), out.Candidates[0].FinishReason
}

func streamClaude(r io.Reader, providerName, model string, onDelta func(delta string) error) (domain.ChatResponse, error) {
	var content strings.Builder
	finishReason := ""
	modelOut := model

	err := scanSSE(r, func(event, data string) error {
		data = strings.TrimSpace(data)
		if data == "" {
			return nil
		}

		switch event {
		case "message_start":
			var payload struct {
				Message struct {
					Model string `json:"model"`
				} `json:"message"`
			}
			if err := json.Unmarshal([]byte(data), &payload); err != nil {
				return errors.Join(domain.ErrStream, err)
			}
			if payload.Message.Model != "" {
				modelOut = payload.Message.Model
			}
		case "content_block_delta":
			var payload struct {
				Delta struct {
					Text string `json:"text"`
				} `json:"delta"`
			}
			if err := json.Unmarshal([]byte(data), &payload); err != nil {
				return errors.Join(domain.ErrStream, err)
			}
			delta := payload.Delta.Text
			if delta == "" {
				return nil
			}
			content.WriteString(delta)
			if err := onDelta(delta); err != nil {
				return err
			}
		case "message_delta":
			var payload struct {
				Delta struct {
					StopReason string `json:"stop_reason"`
				} `json:"delta"`
			}
			if err := json.Unmarshal([]byte(data), &payload); err != nil {
				return errors.Join(domain.ErrStream, err)
			}
			if payload.Delta.StopReason != "" {
				finishReason = payload.Delta.StopReason
			}
		case "message_stop":
			return nil
		default:
			return nil
		}
		return nil
	})
	if err != nil {
		return domain.ChatResponse{}, err
	}
	return domain.ChatResponse{Provider: providerName, Model: modelOut, Content: content.String(), FinishReason: finishReason}, nil
}

func streamGemini(r io.Reader, providerName, model string, onDelta func(delta string) error) (domain.ChatResponse, error) {
	var content strings.Builder
	finishReason := ""

	err := scanSSE(r, func(event, data string) error {
		_ = event
		data = strings.TrimSpace(data)
		if data == "" {
			return nil
		}
		if data == "[DONE]" {
			return nil
		}

		var chunk geminiChatResponse
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			return errors.Join(domain.ErrStream, err)
		}
		delta, finish := extractGeminiContentAndFinish(chunk)
		if delta != "" {
			content.WriteString(delta)
			if err := onDelta(delta); err != nil {
				return err
			}
		}
		if finish != "" {
			finishReason = finish
		}
		return nil
	})
	if err != nil {
		return domain.ChatResponse{}, err
	}
	return domain.ChatResponse{Provider: providerName, Model: model, Content: content.String(), FinishReason: finishReason}, nil
}

func splitSystemMessages(in []domain.Message) (string, []domain.Message) {
	if len(in) == 0 {
		return "", nil
	}
	var systemParts []string
	out := make([]domain.Message, 0, len(in))
	for _, m := range in {
		if m.Role == domain.RoleSystem {
			if s := strings.TrimSpace(m.Content); s != "" {
				systemParts = append(systemParts, s)
			}
			continue
		}
		out = append(out, m)
	}
	return strings.Join(systemParts, "\n\n"), out
}

func joinTextBlocks(blocks []struct {
	Type string `json:"type"`
	Text string `json:"text"`
}) string {
	var sb strings.Builder
	for _, b := range blocks {
		if b.Text != "" {
			sb.WriteString(b.Text)
		}
	}
	return sb.String()
}

func scanSSE(r io.Reader, onEvent func(event, data string) error) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	event := ""
	var data strings.Builder

	flush := func() error {
		if event == "" && data.Len() == 0 {
			return nil
		}
		d := data.String()
		data.Reset()
		e := event
		event = ""
		return onEvent(e, d)
	}

	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r")
		if line == "" {
			if err := flush(); err != nil {
				return err
			}
			continue
		}
		if strings.HasPrefix(line, "event:") {
			event = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
			continue
		}
		if strings.HasPrefix(line, "data:") {
			if data.Len() > 0 {
				data.WriteByte('\n')
			}
			data.WriteString(strings.TrimSpace(strings.TrimPrefix(line, "data:")))
			continue
		}
	}
	if err := scanner.Err(); err != nil {
		return errors.Join(domain.ErrStream, err)
	}
	return flush()
}

func doerOrDefault(doer HTTPDoer) HTTPDoer {
	if doer != nil {
		return doer
	}
	return &http.Client{Timeout: 60 * time.Second}
}

func buildURL(baseURL, fallback, path string) (string, error) {
	base := strings.TrimSpace(baseURL)
	if base == "" {
		base = strings.TrimSpace(fallback)
	}
	base = strings.TrimRight(base, "/")
	u, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	u.Path = strings.TrimRight(u.Path, "/") + path
	return u.String(), nil
}

func addAPIKeyQueryParam(rawURL, apiKey string) (string, error) {
	key := strings.TrimSpace(apiKey)
	if rawURL == "" || key == "" {
		return rawURL, nil
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL, err
	}
	q := u.Query()
	if q.Get("key") == "" {
		q.Set("key", key)
		u.RawQuery = q.Encode()
	}
	return u.String(), nil
}

func readSmallBody(r io.Reader, limit int64) string {
	if r == nil || limit <= 0 {
		return ""
	}
	b, _ := io.ReadAll(io.LimitReader(r, limit))
	return strings.TrimSpace(string(b))
}
