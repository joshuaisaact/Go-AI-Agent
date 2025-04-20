package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"agent/pkg/tools"

	"github.com/anthropics/anthropic-sdk-go"
)

// MessageHandler defines the signature for a function that gets user input
type MessageHandler func() (string, bool)

// Agent handles the conversation flow and tool execution
type Agent struct {
	client         *anthropic.Client
	getUserMessage MessageHandler
	tools          []tools.ToolDefinition
}

// NewAgent creates a new Agent instance
func NewAgent(
	client *anthropic.Client,
	getUserMessage MessageHandler,
	tools []tools.ToolDefinition,
) *Agent {
	return &Agent{
		client:         client,
		getUserMessage: getUserMessage,
		tools:          tools,
	}
}

// Run starts the agent's conversation loop
func (a *Agent) Run(ctx context.Context) error {
	conversation := []anthropic.MessageParam{}

	log.Println("Chat with Claude (use 'ctrl-c' to quit)")

	readUserInput := true
	for {
		if readUserInput {
			fmt.Print("\u001b[94mYou\u001b[0m: ")
			userInput, ok := a.getUserMessage()
			if !ok {
				break
			}

			userMessage := anthropic.NewUserMessage(anthropic.NewTextBlock(userInput))
			conversation = append(conversation, userMessage)
		}

		message, err := a.runInference(ctx, conversation)
		if err != nil {
			return fmt.Errorf("error running inference: %w", err)
		}
		conversation = append(conversation, message.ToParam())

		toolResults := []anthropic.ContentBlockParamUnion{}
		for _, content := range message.Content {
			switch content.Type {
			case "text":
				log.Printf("\u001b[93mClaude\u001b[0m: %s\n", content.Text)
			case "tool_use":
				log.Printf("\u001b[92mtool\u001b[0m: requesting %s(%s)\n", content.Name, content.Input)
				result := a.executeTool(content.ID, content.Name, content.Input)
				toolResults = append(toolResults, result)
			}
		}
		if len(toolResults) == 0 {
			readUserInput = true
			continue
		}
		readUserInput = false
		conversation = append(conversation, anthropic.NewUserMessage(toolResults...))
	}

	return nil
}

// executeTool handles execution of tools based on model requests
func (a *Agent) executeTool(id, name string, input json.RawMessage) anthropic.ContentBlockParamUnion {
	var toolDef tools.ToolDefinition
	var found bool
	for _, tool := range a.tools {
		if tool.Name == name {
			toolDef = tool
			found = true
			break
		}
	}
	if !found {
		log.Printf("Error: tool '%s' not found", name)
		return anthropic.NewToolResultBlock(id, "tool not found", true)
	}

	response, err := toolDef.Function(input)
	if err != nil {
		log.Printf("Error executing tool '%s': %v", name, err)
		return anthropic.NewToolResultBlock(id, err.Error(), true)
	}
	log.Printf("\u001b[92mtool\u001b[0m: result %s -> %s\n", name, response)
	return anthropic.NewToolResultBlock(id, response, false)
}