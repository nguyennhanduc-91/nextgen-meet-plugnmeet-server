package openai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mynaparrot/plugnmeet-protocol/plugnmeet"
)

func newChatStream(ctx context.Context, p *OpenAIProvider, model string, history []*plugnmeet.InsightsAITextChatContent) (<-chan *plugnmeet.InsightsAITextChatStreamResult, error) {
	resultChan := make(chan *plugnmeet.InsightsAITextChatStreamResult)
	streamId := uuid.NewString()

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

	reqBody := ChatCompletionRequest{
		Model:    model,
		Messages: messages,
		Stream:   true,
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
		return nil, fmt.Errorf("failed to marshal openai request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("openai API error: status %d", resp.StatusCode)
	}

	go func() {
		defer resp.Body.Close()
		defer close(resultChan)

		reader := bufio.NewReader(resp.Body)
		var totalUsage *Usage

		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				break
			}

			lineStr := string(bytes.TrimSpace(line))
			if !strings.HasPrefix(lineStr, "data: ") {
				continue
			}

			dataStr := strings.TrimPrefix(lineStr, "data: ")
			if dataStr == "[DONE]" {
				break
			}

			var chunk ChatCompletionStreamResponse
			if err := json.Unmarshal([]byte(dataStr), &chunk); err != nil {
				continue
			}

			if chunk.Usage != nil {
				totalUsage = chunk.Usage
			}

			if len(chunk.Choices) > 0 {
				deltaContent := chunk.Choices[0].Delta.Content
				reasoningContent := chunk.Choices[0].Delta.ReasoningContent
				
				var outputText string
				if reasoningContent != "" {
					outputText += reasoningContent
				}
				if deltaContent != "" {
					outputText += deltaContent
				}

				if outputText != "" {
					resultChan <- &plugnmeet.InsightsAITextChatStreamResult{
						Id:        streamId,
						Text:      outputText,
						CreatedAt: fmt.Sprintf("%d", time.Now().UnixMilli()),
					}
				}
			}
		}

		var promptTokens, completionTokens, totalTokens uint32
		if totalUsage != nil {
			promptTokens = totalUsage.PromptTokens
			completionTokens = totalUsage.CompletionTokens
			totalTokens = totalUsage.TotalTokens
		}

		resultChan <- &plugnmeet.InsightsAITextChatStreamResult{
			Id:               streamId,
			IsLastChunk:      true,
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      totalTokens,
			CreatedAt:        fmt.Sprintf("%d", time.Now().UnixMilli()),
		}
	}()

	return resultChan, nil
}
