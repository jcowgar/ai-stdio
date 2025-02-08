package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jcowgar/acme-utils/internal/config"
	"github.com/jcowgar/acme-utils/internal/conversation"
	"github.com/jcowgar/acme-utils/internal/llm"
)

func actionSend(_ []string) {
	cfg, err := config.Load()
	if err != nil {
		log.Printf("failed to load configuration: %v\n", err)
		return
	}

	conv, err := readConversation(&cfg)
	if err != nil {
		log.Printf("failed to read conversation: %v\n", err)
		return
	}
	if conv == nil {
		// There is no new conversation data, ignore this request
		return
	}

	model := conv.Model
	if model == "" {
		model = cfg.LLM.DefaultProvider
	}

	providerConfig := cfg.LLM.Providers[model]
	provider, err := llm.NewProvider(providerConfig.Type, providerConfig)
	if err != nil {
		log.Printf("failed to create provider: %v\n", err)
		return
	}

	if err := sendLLMRequest(provider, conv); err != nil {
		log.Printf("error processing LLM request: %v\n", err)
	}
}

func readConversation(cfg *config.Config) (*conversation.Conversation, error) {
	projectDir, err := findProjectDirectory()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding project directory: %v\n", err)
		os.Exit(1)
	}

	chatFname := filepath.Join(projectDir, ".ai-stdio.md")
	rawContent, err := os.ReadFile(chatFname)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading chat filename: %v", err)
		os.Exit(1)
	}

	content := string(rawContent)
	conv, err := conversation.ParseContent(content)
	if err != nil {
		return nil, fmt.Errorf("could not parse conversation content: %w", err)
	}

	lastMessage, err := conv.GetLastUserMessage()
	if err != nil {
		return nil, fmt.Errorf("could not get last user message: %w", err)
	}

	// If the last message is empty, return early, we have nothing really to do or
	// send to the LLM
	if strings.TrimSpace(lastMessage) == "" {
		return nil, nil
	}

	for _, req := range conv.ResourceRequests {
		resources, err := req.Fetch(cfg, conv.ProjectDirectory)
		if err != nil {
			return nil, fmt.Errorf("could not fetch resource: %w", err)
		}

		for _, resource := range resources {
			conv.AddReferenceMaterial(resource.ResourceType, resource.Name, resource.Content)
		}
	}

	return conv, nil
}

func sendLLMRequest(provider llm.Provider, conv *conversation.Conversation) error {
	// Write immediately to give the user some feedback
	fmt.Fprintf(os.Stdout, "\n### Response\n\n")

	// Convert messages to provider format
	messages := make([]llm.Message, 0, len(conv.Messages))
	filesInserted := false

	for _, msg := range conv.Messages {
		role := "user"
		if msg.Role == "Response" {
			role = "assistant"
		}

		content := msg.Content
		if !filesInserted && role == "user" && len(conv.ReferenceMaterial) > 0 {
			var filesSection strings.Builder

			filesSection.WriteString("\n\n# Relevant Material\n\n")

			for _, file := range conv.ReferenceMaterial {
				filesSection.WriteString(fmt.Sprintf("Reference Material Type: %s\nName: %s\n```\n%s\n```\n\n",
					file.Typ,
					file.Name,
					file.Content))
			}

			content += filesSection.String()

			filesInserted = true
		}

		messages = append(messages, llm.Message{
			Role:    role,
			Content: content,
		})
	}

	response, err := provider.Chat(context.Background(), messages)
	if err != nil {
		return fmt.Errorf("failed to get response from provider: %w", err)
	}

	newContent := response + "\n\n## You\n\n"

	fmt.Fprintf(os.Stdout, "%s", newContent)

	return nil
}

func inIgnoreFilenames(s string) bool {
	ignoreAnywhere := []string{"+dirtree", "+watch", "+win", "+Errors"}
	ignoreJustFilename := []string{"", "guide"}

	for _, item := range ignoreAnywhere {
		if strings.Contains(s, item) {
			return true
		}
	}

	filename := filepath.Base(s)
	for _, item := range ignoreJustFilename {
		if item == filename {
			return true
		}
	}

	return false
}
