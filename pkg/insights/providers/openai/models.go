package openai

type ChatCompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Stream           bool        `json:"stream,omitempty"`
	MaxTokens        uint32      `json:"max_tokens,omitempty"`
	Temperature      float32     `json:"temperature,omitempty"`
	TopP             float32     `json:"top_p,omitempty"`
	ReasoningBudget  uint32      `json:"reasoning_budget,omitempty"`
	ChatTemplateKwargs *TemplateKwargs `json:"chat_template_kwargs,omitempty"`
}

type TemplateKwargs struct {
	EnableThinking bool `json:"enable_thinking"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionResponse struct {
	ID      string `json:"id"`
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Usage *Usage `json:"usage"`
}

type ChatCompletionStreamResponse struct {
	ID      string `json:"id"`
	Choices []struct {
		Delta struct {
			Content          string `json:"content"`
			ReasoningContent string `json:"reasoning_content,omitempty"`
		} `json:"delta"`
	} `json:"choices"`
	Usage *Usage `json:"usage,omitempty"`
}

type Usage struct {
	PromptTokens     uint32 `json:"prompt_tokens"`
	CompletionTokens uint32 `json:"completion_tokens"`
	TotalTokens      uint32 `json:"total_tokens"`
}
