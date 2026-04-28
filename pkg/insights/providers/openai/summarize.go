package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/mynaparrot/plugnmeet-protocol/plugnmeet"
)

func summarize(ctx context.Context, p *OpenAIProvider, model string, history []*plugnmeet.InsightsAITextChatContent) (string, uint32, uint32, error) {
	var messages []Message
	for _, h := range history {
		role := "user"
		if string(h.Role) == "model" {
			role = "assistant"
		}
		messages = append(messages, Message{
			Role:    role,
			Content: h.Text,
		})
	}

	messages = append(messages, Message{
		Role:    "user",
		Content: "Summarize the following conversation in a concise paragraph.",
	})

	reqBody := ChatCompletionRequest{
		Model:    model,
		Messages: messages,
		Stream:   false,
	}

	if val, ok := p.options["max_tokens"].(float64); ok {
		reqBody.MaxTokens = uint32(val)
	}
	if val, ok := p.options["temperature"].(float64); ok {
		reqBody.Temperature = float32(val)
	}
	if val, ok := p.options["top_p"].(float64); ok {
		reqBody.TopP = float32(val)
	}
	if val, ok := p.options["reasoning_budget"].(float64); ok {
		reqBody.ReasoningBudget = uint32(val)
	}
	if val, ok := p.options["enable_thinking"].(bool); ok && val {
		reqBody.ChatTemplateKwargs = &TemplateKwargs{EnableThinking: true}
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", 0, 0, fmt.Errorf("failed to marshal openai request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", 0, 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return "", 0, 0, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", 0, 0, fmt.Errorf("openai API error: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var chatResp ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", 0, 0, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", 0, 0, fmt.Errorf("no choices in response")
	}

	var promptTokens, completionTokens uint32
	if chatResp.Usage != nil {
		promptTokens = chatResp.Usage.PromptTokens
		completionTokens = chatResp.Usage.CompletionTokens
	}

	return chatResp.Choices[0].Message.Content, promptTokens, completionTokens, nil
}
