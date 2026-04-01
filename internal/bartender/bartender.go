package bartender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	apiURL          = "https://api.openai.com/v1/chat/completions"
	chatModel       = "gpt-4.1-nano"
	memoryModel     = "gpt-4.1-nano"
	maxTokens       = 150
	memoryMaxTokens = 100
	cooldownPerUser = 10 * time.Second
)

// ChatMsg is a minimal chat message for building context.
type ChatMsg struct {
	Nickname string
	Text     string
}

// MemoryStore is the interface the bartender needs from the store.
type MemoryStore interface {
	AddBartenderMemory(text string) error
	BartenderMemories(limit int) []string
	SetBartenderUserNote(fingerprint, note string) error
	BartenderUserNote(fingerprint string) string
}

// Bartender handles the tavern bartender AI persona.
type Bartender struct {
	apiKey    string
	soul      string
	store     MemoryStore
	mu        sync.Mutex
	cooldowns map[string]time.Time // fingerprint → last response time
}

// New creates a bartender. Returns nil if apiKey is empty.
func New(apiKey, soul string, store MemoryStore) *Bartender {
	if apiKey == "" {
		return nil
	}
	return &Bartender{
		apiKey:    apiKey,
		soul:      soul,
		store:     store,
		cooldowns: make(map[string]time.Time),
	}
}

// ShouldRespond checks if a message triggers the bartender.
func ShouldRespond(text, room string) bool {
	if room != "lounge" {
		return false
	}
	lower := strings.ToLower(text)
	return strings.Contains(lower, "@bartender")
}

// CanRespond checks the per-user cooldown. Returns true if allowed.
func (b *Bartender) CanRespond(fingerprint string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	last, ok := b.cooldowns[fingerprint]
	if ok && time.Since(last) < cooldownPerUser {
		return false
	}
	b.cooldowns[fingerprint] = time.Now()
	return true
}

// Respond generates a bartender response given recent chat context.
// It fetches long-term memories and user notes to enrich the prompt.
func (b *Bartender) Respond(recentMessages []ChatMsg, triggerFingerprint, triggerNick, triggerText string) (string, error) {
	// Build conversation context
	var contextParts []string
	for _, m := range recentMessages {
		contextParts = append(contextParts, fmt.Sprintf("%s: %s", m.Nickname, m.Text))
	}
	chatContext := strings.Join(contextParts, "\n")

	// Fetch long-term memories
	memories := b.store.BartenderMemories(20)
	var memoryBlock string
	if len(memories) > 0 {
		memoryBlock = "\n\nThings you remember from past shifts:\n- " + strings.Join(memories, "\n- ")
	}

	// Fetch user-specific notes
	userNote := b.store.BartenderUserNote(triggerFingerprint)
	var userBlock string
	if userNote != "" {
		userBlock = fmt.Sprintf("\n\nWhat you know about %s:\n%s", triggerNick, userNote)
	}

	systemPrompt := b.soul + memoryBlock + userBlock

	messages := []apiMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: fmt.Sprintf("Recent tavern chat:\n%s\n\n%s says to you: %s", chatContext, triggerNick, triggerText)},
	}

	reply, err := b.callAPI(chatModel, messages, maxTokens)
	if err != nil {
		return "", err
	}

	log.Printf("bartender: replied to %s (%d chars)", triggerNick, len(reply))

	// Async: extract memories from this exchange
	go b.extractMemory(triggerFingerprint, triggerNick, triggerText, reply)

	return reply, nil
}

// extractMemory asks the model if anything from this exchange is worth remembering.
func (b *Bartender) extractMemory(fingerprint, nick, userMsg, bartenderReply string) {
	prompt := fmt.Sprintf(`You are the memory system for a tavern bartender. Given this exchange, decide if anything is worth remembering long-term.

%s said: %s
bartender replied: %s

Rules:
- Only save genuinely interesting facts: where someone is from, what they like, recurring jokes, nicknames, memorable moments.
- Do NOT save greetings, drink orders, or generic small talk.
- If nothing is worth saving, respond with exactly: NOTHING
- If something is worth saving about the tavern/regulars in general, respond with: MEMORY: <one short sentence>
- If something is worth noting about this specific person, respond with: USER: <one short sentence>
- Only one line. Pick the most important thing if multiple.`, nick, userMsg, bartenderReply)

	messages := []apiMessage{
		{Role: "user", Content: prompt},
	}

	result, err := b.callAPI(memoryModel, messages, memoryMaxTokens)
	if err != nil {
		log.Printf("bartender memory error: %v", err)
		return
	}

	result = strings.TrimSpace(result)

	if strings.HasPrefix(result, "MEMORY:") {
		mem := strings.TrimSpace(strings.TrimPrefix(result, "MEMORY:"))
		if mem != "" {
			b.store.AddBartenderMemory(mem)
			log.Printf("bartender: saved memory: %s", mem)
		}
	} else if strings.HasPrefix(result, "USER:") {
		note := strings.TrimSpace(strings.TrimPrefix(result, "USER:"))
		if note != "" {
			// Append to existing note
			existing := b.store.BartenderUserNote(fingerprint)
			if existing != "" {
				note = existing + "\n" + note
				// Cap at 500 chars
				if len(note) > 500 {
					note = note[len(note)-500:]
				}
			}
			b.store.SetBartenderUserNote(fingerprint, note)
			log.Printf("bartender: saved user note for %s: %s", nick, note)
		}
	}
}

type apiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type apiRequest struct {
	Model     string       `json:"model"`
	Messages  []apiMessage `json:"messages"`
	MaxTokens int          `json:"max_tokens"`
}

type apiChoice struct {
	Message apiMessage `json:"message"`
}

type apiResponse struct {
	Choices []apiChoice `json:"choices"`
	Error   *apiError   `json:"error,omitempty"`
}

type apiError struct {
	Message string `json:"message"`
}

func (b *Bartender) callAPI(model string, messages []apiMessage, tokens int) (string, error) {
	reqBody := apiRequest{
		Model:     model,
		Messages:  messages,
		MaxTokens: tokens,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+b.apiKey)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("api call: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}

	var apiResp apiResponse
	if err := json.Unmarshal(respBytes, &apiResp); err != nil {
		return "", fmt.Errorf("unmarshal: %w", err)
	}

	if apiResp.Error != nil {
		return "", fmt.Errorf("api error: %s", apiResp.Error.Message)
	}

	if len(apiResp.Choices) == 0 {
		return "", fmt.Errorf("no choices returned")
	}

	reply := strings.TrimSpace(apiResp.Choices[0].Message.Content)
	if reply == "" {
		return "", fmt.Errorf("empty response")
	}

	return reply, nil
}
